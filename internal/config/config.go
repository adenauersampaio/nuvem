package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const CurrentVersion = 1

type Config struct {
	Version     int               `json:"version"`
	Sync        SyncConfig        `json:"sync"`
	GoogleDrive GoogleDriveConfig `json:"google_drive,omitempty"`
}

// GoogleDriveConfig contains credentials for the application's own OAuth
// client. Config files are created with owner-only permissions.
type GoogleDriveConfig struct {
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	Token        string `json:"token,omitempty"`
	RootFolderID string `json:"root_folder_id,omitempty"`
}

type SyncConfig struct {
	LocalPath string        `json:"local_path"`
	Remote    string        `json:"remote"`
	Interval  time.Duration `json:"interval"`
	Engine    string        `json:"engine,omitempty"`
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("localizar diretório de configuração: %w", err)
	}
	return filepath.Join(dir, "nuvem", "config.json"), nil
}

func DefaultRcloneConfigPath() (string, error) {
	path, err := DefaultPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(path), "rclone.conf"), nil
}

func LoadDefault() (Config, error) {
	path, err := DefaultPath()
	if err != nil {
		return Config{}, err
	}
	return Load(path)
}

func SaveDefault(c Config) error {
	path, err := DefaultPath()
	if err != nil {
		return err
	}
	return Save(path, c)
}

func Save(path string, c Config) error {
	if err := c.Validate(); err != nil {
		return err
	}
	contents, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar configuração: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("criar diretório de configuração: %w", err)
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, contents, 0o600); err != nil {
		return fmt.Errorf("salvar configuração: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("finalizar configuração: %w", err)
	}
	return nil
}

func Load(path string) (Config, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(contents, &cfg); err != nil {
		return Config{}, fmt.Errorf("ler configuração: %w", err)
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.Version != CurrentVersion {
		return fmt.Errorf("versão de configuração incompatível: %d", c.Version)
	}
	if !filepath.IsAbs(c.Sync.LocalPath) {
		return errors.New("a pasta local deve ter caminho absoluto")
	}
	info, err := os.Stat(c.Sync.LocalPath)
	if err != nil {
		return fmt.Errorf("acessar pasta local: %w", err)
	}
	if !info.IsDir() {
		return errors.New("o caminho local não é uma pasta")
	}
	if c.Sync.Remote == "" {
		return errors.New("a pasta remota é obrigatória")
	}
	if c.Sync.Interval < 15*time.Second {
		return errors.New("o intervalo mínimo é de 15 segundos")
	}
	if c.Sync.Engine != "" && c.Sync.Engine != "embedded" && c.Sync.Engine != "native" {
		return fmt.Errorf("motor de sincronização desconhecido: %s", c.Sync.Engine)
	}
	return nil
}
