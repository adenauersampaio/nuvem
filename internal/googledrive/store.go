// Package googledrive adapts the Google Drive API to Nuvem's storage contract.
package googledrive

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/adenauersampaio/nuvem/internal/engine"
	"github.com/adenauersampaio/nuvem/internal/rcloneconfig"
	"github.com/rclone/rclone/fs/config/obscure"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const folderMimeType = "application/vnd.google-apps.folder"

// These are only used when migrating an existing profile that was authorized
// with rclone's documented default OAuth application. New Nuvem connections
// should carry their own client ID and secret in the migrated profile.
const (
	legacyClientID              = "202264815644.apps.googleusercontent.com"
	legacyEncryptedClientSecret = "eX8GpZTVx3vxMWVkuuBdDWmAUE6rGhTwVrvG9GhllYccSdj2-mvHVg"
)

type Store struct {
	service     *drive.Service
	tokenSource oauth2.TokenSource
	rootID      string
	mu          sync.RWMutex
	files       map[string]*drive.File
}

func New(ctx context.Context, profile rcloneconfig.DriveProfile, remotePath string) (*Store, error) {
	if profile.Token == nil {
		return nil, fmt.Errorf("a autorização do Google Drive é obrigatória")
	}
	var source oauth2.TokenSource
	clientID, clientSecret := profile.ClientID, profile.ClientSecret
	if clientID == "" {
		var err error
		clientSecret, err = obscure.Reveal(legacyEncryptedClientSecret)
		if err != nil {
			return nil, fmt.Errorf("ler a autorização migrada: %w", err)
		}
		clientID = legacyClientID
	}
	config := oauth2.Config{ClientID: clientID, ClientSecret: clientSecret, Endpoint: oauth2.Endpoint{AuthURL: "https://accounts.google.com/o/oauth2/auth", TokenURL: "https://oauth2.googleapis.com/token"}}
	source = config.TokenSource(ctx, profile.Token)
	service, err := drive.NewService(ctx, option.WithTokenSource(source))
	if err != nil {
		return nil, fmt.Errorf("criar cliente do Google Drive: %w", err)
	}
	store := &Store{service: service, tokenSource: source, rootID: profile.RootFolderID, files: make(map[string]*drive.File)}
	if store.rootID == "" {
		store.rootID = "root"
	}
	for _, segment := range strings.Split(strings.Trim(remotePath, "/"), "/") {
		if segment == "" {
			continue
		}
		folderID, err := store.ensureFolder(ctx, store.rootID, segment)
		if err != nil {
			return nil, err
		}
		store.rootID = folderID
	}
	return store, nil
}

// CurrentToken returns the latest OAuth token after any automatic refresh.
func (s *Store) CurrentToken() (*oauth2.Token, error) { return s.tokenSource.Token() }

func (s *Store) Snapshot(ctx context.Context) ([]engine.Entry, error) {
	files := make(map[string]*drive.File)
	entries, err := s.snapshotFolder(ctx, s.rootID, "", files)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.files = files
	s.mu.Unlock()
	return entries, nil
}

func (s *Store) snapshotFolder(ctx context.Context, parentID, prefix string, index map[string]*drive.File) ([]engine.Entry, error) {
	children, err := s.children(ctx, parentID)
	if err != nil {
		return nil, err
	}
	entries := make([]engine.Entry, 0)
	for _, file := range children {
		relative := file.Name
		if prefix != "" {
			relative = prefix + "/" + file.Name
		}
		if file.MimeType == folderMimeType {
			nested, err := s.snapshotFolder(ctx, file.Id, relative, index)
			if err != nil {
				return nil, err
			}
			entries = append(entries, nested...)
			continue
		}
		if strings.HasPrefix(file.MimeType, "application/vnd.google-apps.") {
			continue
		}
		modified, err := time.Parse(time.RFC3339, file.ModifiedTime)
		if err != nil {
			return nil, fmt.Errorf("ler data de %s: %w", relative, err)
		}
		index[relative] = file
		entries = append(entries, engine.Entry{Path: relative, ModTime: modified.UTC(), Size: file.Size, Hash: file.Md5Checksum})
	}
	return entries, nil
}

func (s *Store) Open(ctx context.Context, relative string) (io.ReadCloser, engine.Entry, error) {
	s.mu.RLock()
	file := s.files[relative]
	s.mu.RUnlock()
	if file == nil {
		return nil, engine.Entry{}, fmt.Errorf("arquivo remoto %q não foi encontrado", relative)
	}
	response, err := s.service.Files.Get(file.Id).SupportsAllDrives(true).Context(ctx).Download()
	if err != nil {
		return nil, engine.Entry{}, fmt.Errorf("baixar %s: %w", relative, err)
	}
	modified, err := time.Parse(time.RFC3339, file.ModifiedTime)
	if err != nil {
		_ = response.Body.Close()
		return nil, engine.Entry{}, err
	}
	return response.Body, engine.Entry{Path: relative, ModTime: modified.UTC(), Size: file.Size, Hash: file.Md5Checksum}, nil
}

func (s *Store) Put(ctx context.Context, relative string, entry engine.Entry, content io.Reader) error {
	parent, name := path.Dir(relative), path.Base(relative)
	parentID := s.rootID
	if parent != "." {
		for _, segment := range strings.Split(parent, "/") {
			folderID, err := s.ensureFolder(ctx, parentID, segment)
			if err != nil {
				return err
			}
			parentID = folderID
		}
	}
	existing, err := s.findOne(ctx, parentID, name)
	if err != nil {
		return err
	}
	contentType := mime.TypeByExtension(path.Ext(name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	metadata := &drive.File{Name: name, Parents: []string{parentID}, MimeType: contentType}
	if existing == nil {
		_, err = s.service.Files.Create(metadata).SupportsAllDrives(true).Media(content, googleapi.ContentType(contentType)).Context(ctx).Do()
	} else {
		if existing.MimeType == folderMimeType {
			return fmt.Errorf("%q is a folder in Google Drive", relative)
		}
		_, err = s.service.Files.Update(existing.Id, metadata).SupportsAllDrives(true).Media(content, googleapi.ContentType(contentType)).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("enviar %s: %w", relative, err)
	}
	return nil
}

// Remove moves a remote file to Google Drive's trash, where it remains
// recoverable under the account's regular Drive retention policy.
func (s *Store) Remove(ctx context.Context, relative string) error {
	s.mu.RLock()
	file := s.files[relative]
	s.mu.RUnlock()
	if file == nil {
		return fmt.Errorf("arquivo remoto %q não foi encontrado", relative)
	}
	_, err := s.service.Files.Update(file.Id, &drive.File{Trashed: true}).SupportsAllDrives(true).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enviar %s para a lixeira: %w", relative, err)
	}
	return nil
}

func (s *Store) ensureFolder(ctx context.Context, parentID, name string) (string, error) {
	folder, err := s.findOne(ctx, parentID, name)
	if err != nil {
		return "", err
	}
	if folder != nil {
		if folder.MimeType != folderMimeType {
			return "", fmt.Errorf("%q não é uma pasta no Google Drive", name)
		}
		return folder.Id, nil
	}
	created, err := s.service.Files.Create(&drive.File{Name: name, MimeType: folderMimeType, Parents: []string{parentID}}).SupportsAllDrives(true).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("criar pasta remota %q: %w", name, err)
	}
	return created.Id, nil
}

func (s *Store) findOne(ctx context.Context, parentID, name string) (*drive.File, error) {
	query := fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false", escapeQuery(name), escapeQuery(parentID))
	list, err := s.service.Files.List().Q(query).Spaces("drive").IncludeItemsFromAllDrives(true).SupportsAllDrives(true).Fields("files(id,name,mimeType,modifiedTime,size,md5Checksum)").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("procurar %q: %w", name, err)
	}
	if len(list.Files) > 1 {
		return nil, fmt.Errorf("há mais de um item chamado %q na pasta remota", name)
	}
	if len(list.Files) == 0 {
		return nil, nil
	}
	return list.Files[0], nil
}

func (s *Store) children(ctx context.Context, parentID string) ([]*drive.File, error) {
	query := fmt.Sprintf("'%s' in parents and trashed = false", escapeQuery(parentID))
	files := make([]*drive.File, 0)
	page := ""
	for {
		list, err := s.service.Files.List().Q(query).Spaces("drive").IncludeItemsFromAllDrives(true).SupportsAllDrives(true).Fields("nextPageToken, files(id,name,mimeType,modifiedTime,size,md5Checksum)").PageToken(page).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("listar pasta remota: %w", err)
		}
		files = append(files, list.Files...)
		if list.NextPageToken == "" {
			return files, nil
		}
		page = list.NextPageToken
	}
}

func escapeQuery(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\\", "\\\\"), "'", "\\'")
}
