package engine

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStore is the portable reference implementation of a storage provider.
// Cloud adapters use the same Store interface.
type LocalStore struct{ Root string }

func (s LocalStore) Snapshot(ctx context.Context) ([]Entry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries := make([]Entry, 0)
	err := filepath.WalkDir(s.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		trashRoot := filepath.Join(s.Root, ".nuvem-trash")
		if d.IsDir() && path == trashRoot {
			return filepath.SkipDir
		}
		if d.IsDir() || d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(s.Root, path)
		if err != nil {
			return err
		}
		hash, err := fileMD5(path)
		if err != nil {
			return err
		}
		entries = append(entries, Entry{Path: filepath.ToSlash(relative), ModTime: info.ModTime().UTC(), Size: info.Size(), Hash: hash, Mode: info.Mode().Perm()})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// Remove moves a file to a private, ignored trash folder under the local root.
func (s LocalStore) Remove(ctx context.Context, relative string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fullPath, err := s.path(relative)
	if err != nil {
		return err
	}
	trash := filepath.Join(s.Root, ".nuvem-trash", time.Now().UTC().Format("20060102T150405.000000000Z"), filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(trash), 0o700); err != nil {
		return err
	}
	return os.Rename(fullPath, trash)
}

func (s LocalStore) Open(ctx context.Context, path string) (io.ReadCloser, Entry, error) {
	if err := ctx.Err(); err != nil {
		return nil, Entry{}, err
	}
	fullPath, err := s.path(path)
	if err != nil {
		return nil, Entry{}, err
	}
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, Entry{}, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, Entry{}, err
	}
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		_ = file.Close()
		return nil, Entry{}, err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		_ = file.Close()
		return nil, Entry{}, err
	}
	return file, Entry{Path: path, Size: info.Size(), ModTime: info.ModTime().UTC(), Hash: fmt.Sprintf("%x", hash.Sum(nil)), Mode: info.Mode().Perm()}, nil
}

func fileMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (s LocalStore) Put(ctx context.Context, path string, entry Entry, source io.Reader) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fullPath, err := s.path(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(fullPath), ".nuvem-*")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if _, err := io.Copy(temporary, source); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if entry.Mode != 0 {
		if err := os.Chmod(temporaryName, entry.Mode.Perm()); err != nil {
			return err
		}
	}
	if !entry.ModTime.IsZero() {
		if err := os.Chtimes(temporaryName, time.Now(), entry.ModTime); err != nil {
			return err
		}
	}
	return os.Rename(temporaryName, fullPath)
}

func (s LocalStore) path(path string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(path))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe relative path %q", path)
	}
	return filepath.Join(s.Root, clean), nil
}
