package googleauth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCallbackAcceptsOnePickedFolder(t *testing.T) {
	result := make(chan callbackResult, 1)
	handler := callbackHandler("state", result)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/oauth/callback?state=state&code=code&picked_file_ids=folder-1", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	got := <-result
	if got.code != "code" || got.folderID != "folder-1" || got.err != nil {
		t.Fatalf("callback = %#v", got)
	}
}

func TestCallbackRejectsNoPickedFolder(t *testing.T) {
	result := make(chan callbackResult, 1)
	handler := callbackHandler("state", result)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/oauth/callback?state=state&code=code", nil))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
	if got := <-result; got.err == nil {
		t.Fatal("callback without a folder should fail")
	}
}

func TestUsesDriveFileScope(t *testing.T) {
	if driveScope != "https://www.googleapis.com/auth/drive.file" {
		t.Fatalf("unexpected scope: %s", driveScope)
	}
}

func TestDefaultClientSecret(t *testing.T) {
	secret := DefaultClientSecret()
	if !strings.HasPrefix(secret, "GOCSPX-") {
		t.Fatalf("DefaultClientSecret() should start with GOCSPX-, got: %s", secret)
	}
}
