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

func TestValidateRejectsUnknownEngine(t *testing.T) {
	local := t.TempDir()
	cfg := Config{Version: CurrentVersion, Sync: SyncConfig{LocalPath: local, Remote: "GoogleDrive:Modelos", Interval: time.Minute, Engine: "other"}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid engine to be rejected")
	}
}

func TestModeAndDirectionDefaults(t *testing.T) {
	s := SyncConfig{}
	if s.EffectiveMode() != ModeMonodirectional {
		t.Fatalf("EffectiveMode() = %s; want %s", s.EffectiveMode(), ModeMonodirectional)
	}
	if s.EffectiveDirection() != DirectionLocalToRemote {
		t.Fatalf("EffectiveDirection() = %s; want %s", s.EffectiveDirection(), DirectionLocalToRemote)
	}

	s.Mode = "bidirectional"
	if s.EffectiveMode() != ModeBidirectional {
		t.Fatalf("EffectiveMode() = %s; want %s", s.EffectiveMode(), ModeBidirectional)
	}

	s.Direction = "remote-to-local"
	if s.EffectiveDirection() != DirectionRemoteToLocal {
		t.Fatalf("EffectiveDirection() = %s; want %s", s.EffectiveDirection(), DirectionRemoteToLocal)
	}
}

func TestValidateModeAndDirection(t *testing.T) {
	local := t.TempDir()
	cfg := Config{
		Version: CurrentVersion,
		Sync: SyncConfig{
			LocalPath: local,
			Remote:    "GoogleDrive:Modelos",
			Interval:  time.Minute,
			Mode:      "invalid-mode",
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid mode to be rejected")
	}

	cfg.Sync.Mode = "monodirectional"
	cfg.Sync.Direction = "invalid-dir"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid direction to be rejected")
	}

	cfg.Sync.Direction = "local-to-remote"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got: %v", err)
	}
}
