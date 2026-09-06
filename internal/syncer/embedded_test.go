package syncer

import (
	"encoding/json"
	"testing"
)

func TestSyncPayload(t *testing.T) {
	payload, err := json.Marshal(map[string]string{
		"path1":     "/home/user/Modelos",
		"path2":     "GoogleDrive:Modelos",
		"checkSync": "true",
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != `{"checkSync":"true","path1":"/home/user/Modelos","path2":"GoogleDrive:Modelos"}` {
		t.Fatalf("payload inesperado: %s", payload)
	}
}
