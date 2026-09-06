package engine

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStoresSynchronizeAndJournal(t *testing.T) {
	sourceRoot, targetRoot := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(sourceRoot, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "nested", "note.txt"), []byte("nuvem"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(sourceRoot, "nested", "note.txt"), 0o740); err != nil {
		t.Fatal(err)
	}
	journal := Journal{Path: filepath.Join(t.TempDir(), "decisions.jsonl")}
	actions, err := (Coordinator{
		Local:  LocalStore{Root: sourceRoot},
		Remote: LocalStore{Root: targetRoot},
		Audit:  journal,
	}).Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Reason != "local-only" {
		t.Fatalf("unexpected actions: %#v", actions)
	}
	content, err := os.ReadFile(filepath.Join(targetRoot, "nested", "note.txt"))
	if err != nil || string(content) != "nuvem" {
		t.Fatalf("copied content = %q, error = %v", content, err)
	}
	info, err := os.Stat(filepath.Join(targetRoot, "nested", "note.txt"))
	if err != nil || info.Mode().Perm() != 0o740 {
		t.Fatalf("copied mode = %v, error = %v", info.Mode().Perm(), err)
	}
	decisions, err := journal.Recent(5)
	if err != nil || len(decisions) != 1 || decisions[0].Action.Path != "nested/note.txt" {
		t.Fatalf("journal = %#v, error = %v", decisions, err)
	}
}

func TestLocalSnapshotUsesDriveCompatibleMD5(t *testing.T) {
	root := t.TempDir()
	contents := []byte("same content")
	if err := os.WriteFile(filepath.Join(root, "note.txt"), contents, 0o600); err != nil {
		t.Fatal(err)
	}
	entries, err := (LocalStore{Root: root}).Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("%x", md5.Sum(contents))
	if len(entries) != 1 || entries[0].Hash != want {
		t.Fatalf("entries = %#v, want MD5 %s", entries, want)
	}
}
