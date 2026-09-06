// Package engine contains provider-independent synchronization decisions.
package engine

import (
	"cmp"
	"io/fs"
	"sort"
	"time"
)

type Side string

const (
	Local  Side = "local"
	Remote Side = "remote"
)

type Entry struct {
	Path    string
	ModTime time.Time
	Size    int64
	Hash    string
	Mode    fs.FileMode
}

type Action struct {
	Path      string
	Source    Side
	Target    Side
	Reason    string
	Operation Operation
}

type Operation string

const (
	Copy   Operation = "copy"
	Delete Operation = "delete"
)

// ResolveConflict chooses the most recently modified version. Identical
// timestamps fall back to local deterministically and are marked for audit.
func ResolveConflict(local, remote Entry) Action {
	if local.ModTime.After(remote.ModTime) {
		return Action{Path: local.Path, Source: Local, Target: Remote, Reason: "local-newer", Operation: Copy}
	}
	if remote.ModTime.After(local.ModTime) {
		return Action{Path: remote.Path, Source: Remote, Target: Local, Reason: "remote-newer", Operation: Copy}
	}
	return Action{Path: local.Path, Source: Local, Target: Remote, Reason: "same-timestamp-local-wins", Operation: Copy}
}

// Plan compares two provider snapshots. Entries present on only one side are
// copied; divergent entries are resolved by the configured newest-wins policy.
func Plan(local, remote []Entry) []Action {
	byPath := func(entries []Entry) map[string]Entry {
		result := make(map[string]Entry, len(entries))
		for _, entry := range entries {
			result[entry.Path] = entry
		}
		return result
	}
	localByPath, remoteByPath := byPath(local), byPath(remote)
	paths := make(map[string]struct{}, len(localByPath)+len(remoteByPath))
	for path := range localByPath {
		paths[path] = struct{}{}
	}
	for path := range remoteByPath {
		paths[path] = struct{}{}
	}

	actions := make([]Action, 0)
	for path := range paths {
		left, localExists := localByPath[path]
		right, remoteExists := remoteByPath[path]
		switch {
		case localExists && !remoteExists:
			actions = append(actions, Action{Path: path, Source: Local, Target: Remote, Reason: "local-only", Operation: Copy})
		case !localExists && remoteExists:
			actions = append(actions, Action{Path: path, Source: Remote, Target: Local, Reason: "remote-only", Operation: Copy})
		case left.Hash != "" && left.Hash == right.Hash:
			continue
		case left.Size == right.Size && left.ModTime.Equal(right.ModTime):
			continue
		default:
			actions = append(actions, ResolveConflict(left, right))
		}
	}
	sort.Slice(actions, func(i, j int) bool { return cmp.Compare(actions[i].Path, actions[j].Path) < 0 })
	return actions
}

// PlanWithBaseline adds safe deletion propagation to the ordinary plan. A
// deletion is only propagated when the surviving counterpart is unchanged
// since the last successful convergence. Otherwise the modified file wins.
func PlanWithBaseline(local, remote []Entry, baseline Snapshot) []Action {
	localByPath, remoteByPath := entriesByPath(local), entriesByPath(remote)
	paths := make(map[string]struct{}, len(localByPath)+len(remoteByPath)+len(baseline.Entries))
	for path := range localByPath {
		paths[path] = struct{}{}
	}
	for path := range remoteByPath {
		paths[path] = struct{}{}
	}
	for path := range baseline.Entries {
		paths[path] = struct{}{}
	}
	actions := make([]Action, 0)
	for path := range paths {
		left, localExists := localByPath[path]
		right, remoteExists := remoteByPath[path]
		previous, existedBefore := baseline.Entries[path]
		switch {
		case localExists && remoteExists:
			if Same(left, right) {
				continue
			}
			actions = append(actions, ResolveConflict(left, right))
		case localExists && !remoteExists:
			if !existedBefore {
				actions = append(actions, Action{Path: path, Source: Local, Target: Remote, Reason: "local-only", Operation: Copy})
				continue
			}
			if Same(left, previous) {
				actions = append(actions, Action{Path: path, Source: Remote, Target: Local, Reason: "remote-deleted", Operation: Delete})
				continue
			}
			actions = append(actions, Action{Path: path, Source: Local, Target: Remote, Reason: "local-changed-after-remote-delete", Operation: Copy})
		case !localExists && remoteExists:
			if !existedBefore {
				actions = append(actions, Action{Path: path, Source: Remote, Target: Local, Reason: "remote-only", Operation: Copy})
				continue
			}
			if Same(right, previous) {
				actions = append(actions, Action{Path: path, Source: Local, Target: Remote, Reason: "local-deleted", Operation: Delete})
				continue
			}
			actions = append(actions, Action{Path: path, Source: Remote, Target: Local, Reason: "remote-changed-after-local-delete", Operation: Copy})
		}
	}
	sort.Slice(actions, func(i, j int) bool { return cmp.Compare(actions[i].Path, actions[j].Path) < 0 })
	return actions
}

func Same(left, right Entry) bool {
	if left.Hash != "" && right.Hash != "" {
		return left.Hash == right.Hash
	}
	return left.Size == right.Size && left.ModTime.Equal(right.ModTime)
}

func entriesByPath(entries []Entry) map[string]Entry {
	result := make(map[string]Entry, len(entries))
	for _, entry := range entries {
		result[entry.Path] = entry
	}
	return result
}
