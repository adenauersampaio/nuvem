package engine

import (
	"context"
	"fmt"
	"io"
)

// Store is a provider endpoint. Local folders and cloud providers implement
// the same small contract, so the synchronisation policy never depends on a
// particular service.
type Store interface {
	Snapshot(context.Context) ([]Entry, error)
	Open(context.Context, string) (io.ReadCloser, Entry, error)
	Put(context.Context, string, Entry, io.Reader) error
	Remove(context.Context, string) error
}

// Recorder retains decisions made by the synchronisation engine.
type Recorder interface {
	Record(context.Context, Action) error
}

// Coordinator applies the newest-wins plan between two stores. It purposely
// does not delete files: deletion tracking needs a stable historical baseline
// and will be added only once it can be performed safely.
type Coordinator struct {
	Local    Store
	Remote   Store
	Audit    Recorder
	Baseline BaselineStore
}

func (c Coordinator) Sync(ctx context.Context) ([]Action, error) {
	if c.Local == nil || c.Remote == nil {
		return nil, fmt.Errorf("local and remote stores are required")
	}
	local, err := c.Local.Snapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("read local snapshot: %w", err)
	}
	remote, err := c.Remote.Snapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("read remote snapshot: %w", err)
	}
	baseline := Snapshot{Entries: map[string]Entry{}}
	if c.Baseline != nil {
		baseline, err = c.Baseline.Load()
		if err != nil {
			return nil, fmt.Errorf("load baseline: %w", err)
		}
	}
	actions := PlanWithBaseline(local, remote, baseline)
	for _, action := range actions {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		source, target := c.Local, c.Remote
		if action.Source == Remote {
			source, target = c.Remote, c.Local
		}
		if action.Operation == Delete {
			if err := target.Remove(ctx, action.Path); err != nil {
				return nil, fmt.Errorf("remove %s: %w", action.Path, err)
			}
			if c.Audit != nil {
				if err := c.Audit.Record(ctx, action); err != nil {
					return nil, fmt.Errorf("record %s: %w", action.Path, err)
				}
			}
			continue
		}
		content, entry, err := source.Open(ctx, action.Path)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", action.Path, err)
		}
		err = target.Put(ctx, action.Path, entry, content)
		closeErr := content.Close()
		if err != nil {
			return nil, fmt.Errorf("copy %s: %w", action.Path, err)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close %s: %w", action.Path, closeErr)
		}
		if c.Audit != nil {
			if err := c.Audit.Record(ctx, action); err != nil {
				return nil, fmt.Errorf("record %s: %w", action.Path, err)
			}
		}
	}
	if c.Baseline != nil {
		local, err = c.Local.Snapshot(ctx)
		if err != nil {
			return nil, fmt.Errorf("refresh local snapshot: %w", err)
		}
		remote, err = c.Remote.Snapshot(ctx)
		if err != nil {
			return nil, fmt.Errorf("refresh remote snapshot: %w", err)
		}
		converged := convergedSnapshot(local, remote)
		if err := c.Baseline.Save(converged); err != nil {
			return nil, fmt.Errorf("save baseline: %w", err)
		}
	}
	return actions, nil
}

func convergedSnapshot(local, remote []Entry) Snapshot {
	localByPath, remoteByPath := entriesByPath(local), entriesByPath(remote)
	entries := make(map[string]Entry)
	for path, left := range localByPath {
		if right, ok := remoteByPath[path]; ok && Same(left, right) {
			entries[path] = left
		}
	}
	return Snapshot{Entries: entries}
}
