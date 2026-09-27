package engine

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"
)

type memoryStore struct {
	entries map[string]Entry
	content map[string][]byte
}

func (s *memoryStore) Snapshot(context.Context) ([]Entry, error) {
	entries := make([]Entry, 0, len(s.entries))
	for _, entry := range s.entries {
		entries = append(entries, entry)
	}
	return entries, nil
}
func (s *memoryStore) Open(_ context.Context, path string) (io.ReadCloser, Entry, error) {
	return io.NopCloser(bytes.NewReader(s.content[path])), s.entries[path], nil
}
func (s *memoryStore) Put(_ context.Context, path string, entry Entry, content io.Reader) error {
	data, _ := io.ReadAll(content)
	s.entries[path], s.content[path] = entry, data
	return nil
}
func (s *memoryStore) Remove(_ context.Context, path string) error {
	delete(s.entries, path)
	delete(s.content, path)
	return nil
}

type records []Action

func (r *records) Record(_ context.Context, action Action) error { *r = append(*r, action); return nil }

type memoryBaseline struct{ snapshot Snapshot }

func (b *memoryBaseline) Load() (Snapshot, error)      { return b.snapshot, nil }
func (b *memoryBaseline) Save(snapshot Snapshot) error { b.snapshot = snapshot; return nil }

func TestCoordinatorCopiesNewerRemoteAndRecordsDecision(t *testing.T) {
	then := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	local := &memoryStore{entries: map[string]Entry{"a.txt": {Path: "a.txt", ModTime: then, Size: 3}}, content: map[string][]byte{"a.txt": []byte("old")}}
	remote := &memoryStore{entries: map[string]Entry{"a.txt": {Path: "a.txt", ModTime: then.Add(time.Minute), Size: 3}}, content: map[string][]byte{"a.txt": []byte("new")}}
	var audit records
	actions, err := (Coordinator{Local: local, Remote: remote, Audit: &audit, Mode: "bidirectional"}).Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Source != Remote {
		t.Fatalf("unexpected actions: %#v", actions)
	}
	if string(local.content["a.txt"]) != "new" {
		t.Fatalf("local content = %q", local.content["a.txt"])
	}
	if len(audit) != 1 || audit[0].Reason != "remote-newer" {
		t.Fatalf("unexpected audit: %#v", audit)
	}
}

func TestCoordinatorPropagatesDeletionOnlyAfterBaseline(t *testing.T) {
	then := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	entry := Entry{Path: "a.txt", ModTime: then, Size: 3, Hash: "same"}
	local := &memoryStore{entries: map[string]Entry{"a.txt": entry}, content: map[string][]byte{"a.txt": []byte("old")}}
	remote := &memoryStore{entries: map[string]Entry{}, content: map[string][]byte{}}
	baseline := &memoryBaseline{snapshot: Snapshot{Entries: map[string]Entry{"a.txt": entry}}}
	actions, err := (Coordinator{Local: local, Remote: remote, Baseline: baseline, Mode: "bidirectional"}).Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Operation != Delete || actions[0].Target != Local {
		t.Fatalf("unexpected actions: %#v", actions)
	}
	if len(local.entries) != 0 {
		t.Fatalf("local file was not removed: %#v", local.entries)
	}
	if len(baseline.snapshot.Entries) != 0 {
		t.Fatalf("baseline was not updated: %#v", baseline.snapshot)
	}
}

func TestCoordinatorRetainsFileChangedAfterOtherSideDeletion(t *testing.T) {
	then := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	previous := Entry{Path: "a.txt", ModTime: then, Size: 3, Hash: "old"}
	changed := Entry{Path: "a.txt", ModTime: then.Add(time.Minute), Size: 3, Hash: "new"}
	local := &memoryStore{entries: map[string]Entry{"a.txt": changed}, content: map[string][]byte{"a.txt": []byte("new")}}
	remote := &memoryStore{entries: map[string]Entry{}, content: map[string][]byte{}}
	baseline := &memoryBaseline{snapshot: Snapshot{Entries: map[string]Entry{"a.txt": previous}}}
	actions, err := (Coordinator{Local: local, Remote: remote, Baseline: baseline, Mode: "bidirectional"}).Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Operation != Copy || actions[0].Reason != "local-changed-after-remote-delete" {
		t.Fatalf("unexpected actions: %#v", actions)
	}
	if string(remote.content["a.txt"]) != "new" {
		t.Fatalf("remote content = %q", remote.content["a.txt"])
	}
}

func TestCoordinatorMonodirectionalDefaultNeverDeletesLocal(t *testing.T) {
	then := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	entry := Entry{Path: "data.txt", ModTime: then, Size: 5, Hash: "data"}
	local := &memoryStore{entries: map[string]Entry{"data.txt": entry}, content: map[string][]byte{"data.txt": []byte("hello")}}
	// Remote was deleted (e.g. user deleted folder on Google Drive)
	remote := &memoryStore{entries: map[string]Entry{}, content: map[string][]byte{}}
	baseline := &memoryBaseline{snapshot: Snapshot{Entries: map[string]Entry{"data.txt": entry}}}

	// Monodirectional Local -> Remote by default
	c := Coordinator{Local: local, Remote: remote, Baseline: baseline}
	actions, err := c.Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Local must NOT be deleted. Instead, local is master so it copies to remote.
	if len(local.entries) != 1 {
		t.Fatalf("local file was deleted in monodirectional mode! entries = %#v", local.entries)
	}
	if len(actions) != 1 || actions[0].Operation != Copy || actions[0].Target != Remote {
		t.Fatalf("unexpected actions: %#v", actions)
	}
	if string(remote.content["data.txt"]) != "hello" {
		t.Fatalf("remote content was not synced: %q", remote.content["data.txt"])
	}
}

func TestCoordinatorMonodirectionalRemoteToLocal(t *testing.T) {
	then := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	entry := Entry{Path: "remote_doc.txt", ModTime: then, Size: 7, Hash: "h"}
	remote := &memoryStore{entries: map[string]Entry{"remote_doc.txt": entry}, content: map[string][]byte{"remote_doc.txt": []byte("content")}}
	local := &memoryStore{entries: map[string]Entry{}, content: map[string][]byte{}}

	c := Coordinator{Local: local, Remote: remote, Mode: "monodirectional", Direction: "remote-to-local"}
	actions, err := c.Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 1 || actions[0].Operation != Copy || actions[0].Target != Local {
		t.Fatalf("unexpected actions: %#v", actions)
	}
	if string(local.content["remote_doc.txt"]) != "content" {
		t.Fatalf("local content = %q; want %q", local.content["remote_doc.txt"], "content")
	}
}
