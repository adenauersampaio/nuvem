package rcloneconfig

import "testing"

func TestParseDriveProfile(t *testing.T) {
	profile, err := ParseDriveProfile(`[GoogleDrive]
type = drive
client_id = client
client_secret = secret
root_folder_id = root
token = {"access_token":"access","refresh_token":"refresh","expiry":"2030-01-01T00:00:00Z"}
`, "GoogleDrive")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Token.AccessToken != "access" || profile.ClientID != "client" || profile.RootFolderID != "root" {
		t.Fatalf("unexpected profile: %#v", profile)
	}
}
