// Package app exposes product operations shared by the command line and desktop UI.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/adenauersampaio/nuvem/internal/config"
	"github.com/adenauersampaio/nuvem/internal/googleauth"
	"github.com/adenauersampaio/nuvem/internal/nativesync"
	"github.com/adenauersampaio/nuvem/internal/runlock"
	"github.com/adenauersampaio/nuvem/internal/service"
	"github.com/adenauersampaio/nuvem/internal/syncer"
)

type Dashboard struct {
	Configured         bool
	LocalPath          string
	Remote             string
	Interval           time.Duration
	Service            string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleFolderLinked bool
}

func DashboardState(ctx context.Context) (Dashboard, error) {
	cfg, err := config.LoadDefault()
	if errors.Is(err, os.ErrNotExist) {
		return Dashboard{}, nil
	}
	if err != nil {
		return Dashboard{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Dashboard{}, err
	}
	unitPath, err := service.DefaultUnitPath()
	if err != nil {
		return Dashboard{}, err
	}
	state, err := service.New(unitPath, service.OSRunner{}).Status(ctx)
	if err != nil {
		state = "unknown"
	}
	return Dashboard{Configured: true, LocalPath: cfg.Sync.LocalPath, Remote: cfg.Sync.Remote, Interval: cfg.Sync.Interval, Service: state, GoogleClientID: cfg.GoogleDrive.ClientID, GoogleClientSecret: cfg.GoogleDrive.ClientSecret, GoogleFolderLinked: cfg.GoogleDrive.RootFolderID != ""}, nil
}

func Save(localPath, remote string, interval time.Duration, clientID, clientSecret string) error {
	cfg, err := config.LoadDefault()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if errors.Is(err, os.ErrNotExist) {
		cfg = config.Config{Version: config.CurrentVersion}
	}
	cfg.Sync = config.SyncConfig{LocalPath: localPath, Remote: remote, Interval: interval}
	if clientID != "" {
		cfg.GoogleDrive.ClientID = clientID
	}
	if clientSecret != "" {
		cfg.GoogleDrive.ClientSecret = clientSecret
	}
	return config.SaveDefault(cfg)
}

func ConnectGoogleDrive(ctx context.Context, clientID, clientSecret string) error {
	if err := SaveCurrentClient(clientID, clientSecret); err != nil {
		return err
	}
	result, err := googleauth.Connect(ctx, clientID, clientSecret)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(result.Token)
	if err != nil {
		return err
	}
	cfg, err := config.LoadDefault()
	if err != nil {
		return err
	}
	cfg.GoogleDrive.Token = string(encoded)
	cfg.GoogleDrive.RootFolderID = result.PickedFolderID
	// The selected folder is the root. No name/path must be created inside it.
	cfg.Sync.Remote = "GoogleDrive:"
	cfg.Sync.Engine = "native"
	return config.SaveDefault(cfg)
}

func RestartService(ctx context.Context) error {
	returnCode, err := service.OSRunner{}.Run(ctx, "systemctl", "--user", "restart", service.UnitName)
	if err != nil {
		return fmt.Errorf("reiniciar serviço: %s%w", returnCode, err)
	}
	return nil
}

func SaveCurrentClient(clientID, clientSecret string) error {
	cfg, err := config.LoadDefault()
	if err != nil {
		return err
	}
	if clientID == "" {
		return fmt.Errorf("o ID do cliente OAuth do Google é obrigatório")
	}
	cfg.GoogleDrive.ClientID = clientID
	cfg.GoogleDrive.ClientSecret = clientSecret
	return config.SaveDefault(cfg)
}

func RunOnce(ctx context.Context) error {
	cfg, err := config.LoadDefault()
	if err != nil {
		return err
	}
	return Run(ctx, cfg.Sync)
}

// Run executes one synchronization using the selected engine. Empty retains
// the established embedded engine; native is opt-in during its rollout.
func Run(ctx context.Context, syncConfig config.SyncConfig) error {
	if err := (config.Config{Version: config.CurrentVersion, Sync: syncConfig}).Validate(); err != nil {
		return err
	}
	lock, err := runlock.AcquireDefault()
	if errors.Is(err, runlock.ErrHeld) {
		return fmt.Errorf("a Nuvem synchronization is already running")
	}
	if err != nil {
		return err
	}
	defer lock.Release()
	if syncConfig.Engine == "native" {
		_, err := nativesync.Run(ctx, syncConfig)
		return err
	}
	return runEmbedded(ctx, syncConfig)
}

func RunWithEngine(ctx context.Context, selected string) error {
	cfg, err := config.LoadDefault()
	if err != nil {
		return err
	}
	cfg.Sync.Engine = selected
	return Run(ctx, cfg.Sync)
}

func runEmbedded(ctx context.Context, syncConfig config.SyncConfig) error {
	path, err := config.DefaultRcloneConfigPath()
	if err != nil {
		return err
	}
	engine := syncer.NewEmbeddedEngine(path)
	return engine.Sync(ctx, syncConfig)
}
