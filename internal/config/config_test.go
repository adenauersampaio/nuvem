package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoad(t *testing.T) {
	local := t.TempDir()
	path := filepath.Join(t.TempDir(), "nuvem", "config.json")
	want := Config{
		Version: CurrentVersion,
		Sync:    SyncConfig{LocalPath: local, Remote: "GoogleDrive:Modelos", Interval: time.Minute},
	}
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("configuração = %#v; quer %#v", got, want)
	}
}
