// Package runlock keeps concurrent Nuvem synchronizations from sharing state.
package runlock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

var ErrHeld = errors.New("another Nuvem synchronization is already running")

type Lock struct {
	file *os.File
}

func AcquireDefault() (*Lock, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("localizar diretório de estado: %w", err)
	}
	return Acquire(filepath.Join(dir, "nuvem", "sync.lock"))
}

func Acquire(path string) (*Lock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("criar diretório de estado: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("abrir bloqueio de sincronização: %w", err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, ErrHeld
		}
		return nil, fmt.Errorf("bloquear sincronização: %w", err)
	}
	return &Lock{file: file}, nil
}

func (l *Lock) Release() error {
	if l == nil || l.file == nil {
		return nil
	}
	err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	closeErr := l.file.Close()
	l.file = nil
	if err != nil {
		return err
	}
	return closeErr
}
