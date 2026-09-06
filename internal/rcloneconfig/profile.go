package rcloneconfig

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
)

// DriveProfile is the small, provider-neutral subset read during a one-time
// migration from an existing rclone configuration. Nuvem never invokes rclone
// after importing these credentials.
type DriveProfile struct {
	Token        *oauth2.Token
	ClientID     string
	ClientSecret string
	RootFolderID string
}

func ParseDriveProfile(contents, remote string) (DriveProfile, error) {
	section, ok := findSection(contents, remote)
	if !ok {
		return DriveProfile{}, fmt.Errorf("remote %q não encontrado", remote)
	}
	values := make(map[string]string)
	for _, line := range strings.Split(section, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if values["type"] != "drive" {
		return DriveProfile{}, fmt.Errorf("remote %q não é do Google Drive", remote)
	}
	if values["token"] == "" {
		return DriveProfile{}, fmt.Errorf("remote %q não possui autorização", remote)
	}
	var token oauth2.Token
	if err := json.Unmarshal([]byte(values["token"]), &token); err != nil {
		return DriveProfile{}, fmt.Errorf("ler autorização do Drive: %w", err)
	}
	if token.AccessToken == "" {
		return DriveProfile{}, fmt.Errorf("remote %q possui uma autorização inválida", remote)
	}
	rootFolderID := values["root_folder_id"]
	if rootFolderID == "" {
		rootFolderID = values["team_drive"]
	}
	return DriveProfile{
		Token:        &token,
		ClientID:     values["client_id"],
		ClientSecret: values["client_secret"],
		RootFolderID: rootFolderID,
	}, nil
}
