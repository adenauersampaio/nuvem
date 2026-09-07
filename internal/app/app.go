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
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
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
	if cfg.GoogleDrive.Token != "" && cfg.GoogleDrive.RootFolderID != "" {
		if cfg.GoogleDrive.RootFolderName == "" || cfg.Sync.Remote == "GoogleDrive:" {
			var token oauth2.Token
			if err := json.Unmarshal([]byte(cfg.GoogleDrive.Token), &token); err == nil {
				name := fetchGoogleDriveFolderName(ctx, cfg.GoogleDrive.ClientID, cfg.GoogleDrive.ClientSecret, &token, cfg.GoogleDrive.RootFolderID)
				if name != "" {
					cfg.GoogleDrive.RootFolderName = name
					cfg.Sync.Remote = "GoogleDrive:" + name
					_ = config.SaveDefault(cfg)
				}
			}
		}
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

func fetchGoogleDriveFolderName(ctx context.Context, clientID, clientSecret string, token *oauth2.Token, folderID string) string {
	if token == nil || folderID == "" {
		return ""
	}
	if clientID == "" {
		clientID = googleauth.DefaultClientID
	}
	if clientSecret == "" {
		clientSecret = googleauth.DefaultClientSecret()
	}
	oauthConfig := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}
	ts := oauthConfig.TokenSource(ctx, token)
	srv, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return ""
	}
	file, err := srv.Files.Get(folderID).Fields("name").Context(ctx).Do()
	if err != nil {
		return ""
	}
	return file.Name
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
	} else if cfg.GoogleDrive.ClientID == "" {
		cfg.GoogleDrive.ClientID = googleauth.DefaultClientID
	}
	if clientSecret != "" {
		cfg.GoogleDrive.ClientSecret = clientSecret
	} else if cfg.GoogleDrive.ClientSecret == "" {
		cfg.GoogleDrive.ClientSecret = googleauth.DefaultClientSecret()
	}
	return config.SaveDefault(cfg)
}

func ConnectGoogleDrive(ctx context.Context, clientID, clientSecret string) error {
	if clientID == "" {
		clientID = googleauth.DefaultClientID
	}
	if clientSecret == "" {
		clientSecret = googleauth.DefaultClientSecret()
	}
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
	folderName := fetchGoogleDriveFolderName(ctx, clientID, clientSecret, result.Token, result.PickedFolderID)
	cfg.GoogleDrive.Token = string(encoded)
	cfg.GoogleDrive.RootFolderID = result.PickedFolderID
	cfg.GoogleDrive.RootFolderName = folderName
	if folderName != "" {
		cfg.Sync.Remote = "GoogleDrive:" + folderName
	} else {
		cfg.Sync.Remote = "GoogleDrive:"
	}
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

func StopService(ctx context.Context) error {
	returnCode, err := service.OSRunner{}.Run(ctx, "systemctl", "--user", "stop", service.UnitName)
	if err != nil {
		return fmt.Errorf("parar serviço: %s%w", returnCode, err)
	}
	return nil
}

func StartService(ctx context.Context) error {
	returnCode, err := service.OSRunner{}.Run(ctx, "systemctl", "--user", "start", service.UnitName)
	if err != nil {
		return fmt.Errorf("iniciar serviço: %s%w", returnCode, err)
	}
	return nil
}

func SaveCurrentClient(clientID, clientSecret string) error {
	cfg, err := config.LoadDefault()
	if err != nil {
		return err
	}
	if clientID == "" {
		clientID = googleauth.DefaultClientID
	}
	if clientSecret == "" {
		clientSecret = googleauth.DefaultClientSecret()
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
