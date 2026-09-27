package engine

import (
	"testing"
	"time"
)

func TestPlanUsesNewestVersionForConflict(t *testing.T) {
	now := time.Now().UTC()
	actions := Plan(
		[]Entry{{Path: "draft.docx", ModTime: now, Size: 10, Hash: "local"}},
		[]Entry{{Path: "draft.docx", ModTime: now.Add(time.Minute), Size: 12, Hash: "remote"}},
	)
	if len(actions) != 1 || actions[0].Source != Remote || actions[0].Reason != "remote-newer" {
		t.Fatalf("actions = %#v", actions)
	}
}

func TestPlanCopiesOneSidedEntries(t *testing.T) {
	actions := Plan([]Entry{{Path: "local.txt"}}, []Entry{{Path: "remote.txt"}})
	if len(actions) != 2 || actions[0].Source != Local || actions[1].Source != Remote {
		t.Fatalf("actions = %#v", actions)
	}
}

func TestPlanMonodirectionalLocalToRemoteNeverTouchesLocal(t *testing.T) {
	now := time.Now().UTC()
	// Scenario: local has file a.txt and b.txt; remote has a.txt and c.txt (c.txt exists only on remote).
	// a.txt on local is newer.
	local := []Entry{
		{Path: "a.txt", ModTime: now.Add(time.Minute), Size: 10},
		{Path: "b.txt", ModTime: now, Size: 20},
	}
	remote := []Entry{
		{Path: "a.txt", ModTime: now, Size: 5},
		{Path: "c.txt", ModTime: now, Size: 30},
	}
	baseline := Snapshot{Entries: map[string]Entry{
		"a.txt": {Path: "a.txt", ModTime: now, Size: 5},
	}}

	actions := PlanMonodirectional(local, remote, baseline, Local, Remote)
	for _, action := range actions {
		if action.Target == Local {
			t.Fatalf("PlanMonodirectional in Local->Remote must NEVER target Local: %#v", action)
		}
	}

	// Should copy a.txt (source-modified) and b.txt (source-only) to Remote
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got: %#v", actions)
	}
	if actions[0].Path != "a.txt" || actions[0].Operation != Copy || actions[0].Target != Remote {
		t.Errorf("action 0 error: %#v", actions[0])
	}
	if actions[1].Path != "b.txt" || actions[1].Operation != Copy || actions[1].Target != Remote {
		t.Errorf("action 1 error: %#v", actions[1])
	}
}

func TestPlanMonodirectionalRemoteDeletionDoesNotDeleteLocal(t *testing.T) {
	now := time.Now().UTC()
	// User deleted duplicate folder on Google Drive, so remote is empty.
	// Local still has the files. Baseline still remembers previous sync.
	entry := Entry{Path: "folder/doc.txt", ModTime: now, Size: 100}
	local := []Entry{entry}
	remote := []Entry{}
	baseline := Snapshot{Entries: map[string]Entry{"folder/doc.txt": entry}}

	actions := PlanMonodirectional(local, remote, baseline, Local, Remote)
	// Must copy local to remote, NEVER delete local!
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got: %#v", actions)
	}
	if actions[0].Path != "folder/doc.txt" || actions[0].Operation != Copy || actions[0].Target != Remote {
		t.Fatalf("expected Copy to Remote, got: %#v", actions[0])
	}
}

func TestPlanMonodirectionalLocalDeletionPropagatesToRemote(t *testing.T) {
	now := time.Now().UTC()
	// User deleted file locally. Remote still has it. Baseline remembers it.
	entry := Entry{Path: "del.txt", ModTime: now, Size: 50}
	local := []Entry{}
	remote := []Entry{entry}
	baseline := Snapshot{Entries: map[string]Entry{"del.txt": entry}}

	actions := PlanMonodirectional(local, remote, baseline, Local, Remote)
	if len(actions) != 1 {
		t.Fatalf("expected 1 delete action, got: %#v", actions)
	}
	if actions[0].Path != "del.txt" || actions[0].Operation != Delete || actions[0].Target != Remote {
		t.Fatalf("expected Delete on Remote, got: %#v", actions[0])
	}
}

func TestPlanMonodirectionalRemoteToLocal(t *testing.T) {
	now := time.Now().UTC()
	remote := []Entry{
		{Path: "cloud.txt", ModTime: now, Size: 10},
	}
	local := []Entry{}
	baseline := Snapshot{Entries: map[string]Entry{}}

	actions := PlanMonodirectional(remote, local, baseline, Remote, Local)
	if len(actions) != 1 || actions[0].Target != Local || actions[0].Operation != Copy {
		t.Fatalf("expected Copy to Local, got: %#v", actions)
	}
}
