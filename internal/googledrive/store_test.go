package googledrive

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adenauersampaio/nuvem/internal/engine"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func TestStorePutUpdateDoesNotSendParents(t *testing.T) {
	var updateReceived bool
	var updateContainedParents bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Handle findOne list request
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/files") {
			resp := drive.FileList{
				Files: []*drive.File{
					{
						Id:       "existing-file-id",
						Name:     "document.docx",
						MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Handle update request
		if strings.Contains(r.URL.Path, "existing-file-id") {
			updateReceived = true
			mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err == nil && strings.HasPrefix(mediaType, "multipart/") {
				mr := multipart.NewReader(r.Body, params["boundary"])
				for {
					part, err := mr.NextPart()
					if err == io.EOF {
						break
					}
					if err != nil {
						break
					}
					partType, _, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
					if strings.Contains(partType, "json") {
						body, _ := io.ReadAll(part)
						var m map[string]any
						if err := json.Unmarshal(body, &m); err == nil {
							if _, exists := m["parents"]; exists {
								updateContainedParents = true
								w.WriteHeader(http.StatusForbidden)
								_, _ = w.Write([]byte(`{"error": {"code": 403, "message": "The parents field is not directly writable in update requests."}}`))
								return
							}
						}
					}
				}
			}

			resp := drive.File{
				Id:       "existing-file-id",
				Name:     "document.docx",
				MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	store := &Store{
		service: srv,
		rootID:  "root",
		files:   make(map[string]*drive.File),
	}

	err = store.Put(ctx, "document.docx", engine.Entry{Path: "document.docx"}, strings.NewReader("hello world"))
	if err != nil {
		t.Fatalf("store.Put failed: %v", err)
	}

	if !updateReceived {
		t.Fatalf("expected update request, but none was received")
	}
	if updateContainedParents {
		t.Fatalf("update request contained 'parents' field which causes Google Drive API 403 error")
	}
}

func TestStorePutCreateSendsParents(t *testing.T) {
	var createReceived bool
	var createContainedParents bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Handle findOne list request returning empty (file does not exist)
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/files") {
			resp := drive.FileList{
				Files: []*drive.File{},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Handle create request
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/files") {
			createReceived = true
			mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err == nil && strings.HasPrefix(mediaType, "multipart/") {
				mr := multipart.NewReader(r.Body, params["boundary"])
				for {
					part, err := mr.NextPart()
					if err == io.EOF {
						break
					}
					if err != nil {
						break
					}
					partType, _, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
					if strings.Contains(partType, "json") {
						body, _ := io.ReadAll(part)
						var m map[string]any
						if err := json.Unmarshal(body, &m); err == nil {
							if parents, exists := m["parents"]; exists && parents != nil {
								createContainedParents = true
							}
						}
					}
				}
			}

			resp := drive.File{
				Id:       "new-file-id",
				Name:     "new-document.docx",
				MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	store := &Store{
		service: srv,
		rootID:  "root",
		files:   make(map[string]*drive.File),
	}

	err = store.Put(ctx, "new-document.docx", engine.Entry{Path: "new-document.docx"}, strings.NewReader("hello world"))
	if err != nil {
		t.Fatalf("store.Put failed: %v", err)
	}

	if !createReceived {
		t.Fatalf("expected create request, but none was received")
	}
	if !createContainedParents {
		t.Fatalf("create request should contain 'parents' field")
	}
}

func TestStoreSnapshotIgnoresLockAndHiddenFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/files") {
			resp := drive.FileList{
				Files: []*drive.File{
					{
						Id:           "valid-id",
						Name:         "document.docx",
						MimeType:     "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
						ModifiedTime: "2026-09-15T10:00:00Z",
						Size:         100,
						Md5Checksum:  "abc123",
					},
					{
						Id:           "lock-id",
						Name:         ".~lock.document.docx#",
						MimeType:     "application/octet-stream",
						ModifiedTime: "2026-09-15T10:00:00Z",
						Size:         50,
						Md5Checksum:  "def456",
					},
					{
						Id:           "hidden-dir-id",
						Name:         ".git",
						MimeType:     folderMimeType,
						ModifiedTime: "2026-09-15T10:00:00Z",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	store := &Store{
		service: srv,
		rootID:  "root",
		files:   make(map[string]*drive.File),
	}

	entries, err := store.Snapshot(ctx)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}
	if len(entries) != 1 || entries[0].Path != "document.docx" {
		t.Fatalf("unexpected entries: %#v", entries)
	}
}

