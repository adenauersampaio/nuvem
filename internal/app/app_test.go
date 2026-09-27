package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adenauersampaio/nuvem/internal/config"
)

func TestDashboardStateWithNonExistentFolder(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempHome)

	// Save a config with a folder that doesn't exist
	cfg := config.Config{
		Version: config.CurrentVersion,
		Sync: config.SyncConfig{
			LocalPath: filepath.Join(tempHome, "deleted-dir"),
			Remote:    "GoogleDrive:Test",
			Interval:  time.Minute,
		},
	}
	cfgPath := filepath.Join(tempHome, "nuvem", "config.json")
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	state, err := DashboardState(context.Background())
	if err != nil {
		t.Fatalf("DashboardState should not fail on non-existent folder, got error: %v", err)
	}
	if state.Configured {
		t.Fatal("expected Configured to be false for non-existent directory")
	}
	if state.LocalPath != filepath.Join(tempHome, "deleted-dir") {
		t.Fatalf("expected LocalPath to be preserved, got %s", state.LocalPath)
	}
	if state.Remote != "GoogleDrive:Test" {
		t.Fatalf("expected Remote to be preserved, got %s", state.Remote)
	}
}

func TestSaveCurrentClientWithoutPreexistingConfig(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempHome)

	err := SaveCurrentClient("test-client-id", "test-secret")
	if err != nil {
		t.Fatalf("SaveCurrentClient should succeed without preexisting config, got: %v", err)
	}

	loaded, err := config.LoadDefault()
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}
	if loaded.GoogleDrive.ClientID != "test-client-id" {
		t.Fatalf("expected client ID test-client-id, got %s", loaded.GoogleDrive.ClientID)
	}
}

func TestSaveWithModeCreatesLocalFolder(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempHome)

	newFolder := filepath.Join(tempHome, "new-sync-folder")
	err := SaveWithMode(newFolder, "GoogleDrive:Sync", 15*time.Minute, "id", "sec", config.ModeMonodirectional, config.DirectionLocalToRemote)
	if err != nil {
		t.Fatalf("SaveWithMode failed: %v", err)
	}

	info, err := os.Stat(newFolder)
	if err != nil || !info.IsDir() {
		t.Fatalf("expected newFolder to be created as a directory, err: %v", err)
	}
}
