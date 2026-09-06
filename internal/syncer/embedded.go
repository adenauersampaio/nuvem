package syncer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	_ "github.com/rclone/rclone/backend/drive"
	_ "github.com/rclone/rclone/backend/local"
	_ "github.com/rclone/rclone/cmd/bisync"
	"github.com/rclone/rclone/librclone/librclone"

	"github.com/adenauersampaio/nuvem/internal/config"
)

var initializeOnce sync.Once

// EmbeddedEngine executes rclone's synchronization engine inside the Nuvem
// process. The rclone executable is not required on the user's PATH.
type EmbeddedEngine struct{ configPath string }

func NewEmbeddedEngine(configPath string) EmbeddedEngine {
	return EmbeddedEngine{configPath: configPath}
}

func (e EmbeddedEngine) Check() error {
	if e.configPath != "" {
		if _, err := os.Stat(e.configPath); err != nil {
			return fmt.Errorf("configuração privada do Drive: %w", err)
		}
		if err := os.Setenv("RCLONE_CONFIG", e.configPath); err != nil {
			return fmt.Errorf("definir configuração privada: %w", err)
		}
	}
	initializeOnce.Do(librclone.Initialize)
	return nil
}

func (e EmbeddedEngine) Sync(_ context.Context, syncConfig config.SyncConfig) error {
	if err := e.Check(); err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]string{
		"path1":     syncConfig.LocalPath,
		"path2":     syncConfig.Remote,
		"checkSync": "true",
	})
	if err != nil {
		return fmt.Errorf("criar pedido de sincronização: %w", err)
	}
	output, status := librclone.RPC("sync/bisync", string(payload))
	if status != http.StatusOK {
		return fmt.Errorf("motor de sincronização retornou HTTP %d: %s", status, output)
	}
	return nil
}
