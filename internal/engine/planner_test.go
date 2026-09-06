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
