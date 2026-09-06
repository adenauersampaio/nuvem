package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Decision struct {
	Action      Action    `json:"action"`
	CompletedAt time.Time `json:"completed_at"`
}

// Journal is an append-only local audit trail for Nuvem-native decisions.
type Journal struct{ Path string }

func DefaultJournalPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "nuvem", "decisions.jsonl"), nil
}

func (j Journal) Record(ctx context.Context, action Action) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if j.Path == "" {
		return fmt.Errorf("journal path is required")
	}
	if err := os.MkdirAll(filepath.Dir(j.Path), 0o700); err != nil {
		return err
	}
	encoded, err := json.Marshal(Decision{Action: action, CompletedAt: time.Now().UTC()})
	if err != nil {
		return err
	}
	file, err := os.OpenFile(j.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(encoded, '\n'))
	return err
}

func (j Journal) Recent(limit int) ([]Decision, error) {
	if limit <= 0 {
		return nil, nil
	}
	file, err := os.Open(j.Path)
	if os.IsNotExist(err) {
		return []Decision{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decisions := make([]Decision, 0, limit)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var decision Decision
		if err := json.Unmarshal(scanner.Bytes(), &decision); err != nil {
			return nil, fmt.Errorf("read journal: %w", err)
		}
		decisions = append(decisions, decision)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(decisions) > limit {
		decisions = decisions[len(decisions)-limit:]
	}
	return decisions, nil
}
