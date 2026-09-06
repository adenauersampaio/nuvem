package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Snapshot is the shared file state captured after a fully successful sync.
type Snapshot struct {
	Entries map[string]Entry `json:"entries"`
}

type BaselineStore interface {
	Load() (Snapshot, error)
	Save(Snapshot) error
}

type FileBaseline struct{ Path string }

func DefaultBaselinePath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "nuvem", "baseline.json"), nil
}

func (b FileBaseline) Load() (Snapshot, error) {
	contents, err := os.ReadFile(b.Path)
	if os.IsNotExist(err) {
		return Snapshot{Entries: map[string]Entry{}}, nil
	}
	if err != nil {
		return Snapshot{}, err
	}
	var snapshot Snapshot
	if err := json.Unmarshal(contents, &snapshot); err != nil {
		return Snapshot{}, err
	}
	if snapshot.Entries == nil {
		snapshot.Entries = map[string]Entry{}
	}
	return snapshot, nil
}

func (b FileBaseline) Save(snapshot Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(b.Path), 0o700); err != nil {
		return err
	}
	contents, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	temporary := b.Path + ".tmp"
	if err := os.WriteFile(temporary, contents, 0o600); err != nil {
		return err
	}
	return os.Rename(temporary, b.Path)
}
