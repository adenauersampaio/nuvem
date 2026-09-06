// Package nativesync connects Nuvem's provider-neutral engine to Google Drive.
package nativesync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/adenauersampaio/nuvem/internal/config"
	"github.com/adenauersampaio/nuvem/internal/engine"
	"github.com/adenauersampaio/nuvem/internal/googledrive"
	"github.com/adenauersampaio/nuvem/internal/rcloneconfig"
	"golang.org/x/oauth2"
)

func Run(ctx context.Context, syncConfig config.SyncConfig) ([]engine.Action, error) {
	remoteName, remotePath, ok := strings.Cut(syncConfig.Remote, ":")
	if !ok || remoteName == "" {
		return nil, fmt.Errorf("o Drive remoto deve usar o formato Perfil:Pasta")
	}
	applicationConfig, err := config.LoadDefault()
	if err != nil {
		return nil, err
	}
	var profile rcloneconfig.DriveProfile
	if applicationConfig.GoogleDrive.Token != "" {
		if applicationConfig.GoogleDrive.RootFolderID == "" {
			return nil, fmt.Errorf("escolha uma pasta do Google Drive antes de sincronizar")
		}
		var token oauth2.Token
		if err := json.Unmarshal([]byte(applicationConfig.GoogleDrive.Token), &token); err != nil {
			return nil, fmt.Errorf("ler autorização do Nuvem: %w", err)
		}
		profile = rcloneconfig.DriveProfile{Token: &token, ClientID: applicationConfig.GoogleDrive.ClientID, ClientSecret: applicationConfig.GoogleDrive.ClientSecret, RootFolderID: applicationConfig.GoogleDrive.RootFolderID}
		remotePath = ""
	} else {
		path, err := config.DefaultRcloneConfigPath()
		if err != nil {
			return nil, err
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("ler autorização importada: %w", err)
		}
		profile, err = rcloneconfig.ParseDriveProfile(string(contents), remoteName)
		if err != nil {
			return nil, err
		}
	}
	remote, err := googledrive.New(ctx, profile, remotePath)
	if err != nil {
		return nil, err
	}
	journalPath, err := engine.DefaultJournalPath()
	if err != nil {
		return nil, err
	}
	baselinePath, err := engine.DefaultBaselinePath()
	if err != nil {
		return nil, err
	}
	actions, err := (engine.Coordinator{Local: engine.LocalStore{Root: syncConfig.LocalPath}, Remote: remote, Audit: engine.Journal{Path: journalPath}, Baseline: engine.FileBaseline{Path: baselinePath}}).Sync(ctx)
	if err != nil {
		return nil, err
	}
	if applicationConfig.GoogleDrive.Token != "" {
		token, err := remote.CurrentToken()
		if err != nil {
			return nil, fmt.Errorf("atualizar autorização do Nuvem: %w", err)
		}
		encoded, err := json.Marshal(token)
		if err != nil {
			return nil, err
		}
		applicationConfig.GoogleDrive.Token = string(encoded)
		if err := config.SaveDefault(applicationConfig); err != nil {
			return nil, fmt.Errorf("salvar autorização renovada: %w", err)
		}
	}
	return actions, nil
}
