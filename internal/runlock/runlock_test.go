package runlock

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestAcquireExcludesConcurrentOwners(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sync.lock")
	first, err := Acquire(path)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Release()

	second, err := Acquire(path)
	if !errors.Is(err, ErrHeld) || second != nil {
		t.Fatalf("lock concorrente = %#v, %v", second, err)
	}
	if err := first.Release(); err != nil {
		t.Fatal(err)
	}
	third, err := Acquire(path)
	if err != nil {
		t.Fatalf("bloqueio após liberar: %v", err)
	}
	if err := third.Release(); err != nil {
		t.Fatal(err)
	}
}
