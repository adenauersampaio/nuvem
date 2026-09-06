package rcloneconfig

import "testing"

func TestFindSection(t *testing.T) {
	contents := "[Other]\ntype = drive\n\n[GoogleDrive]\ntype = drive\ntoken = secret\n\n[Last]\ntype = s3\n"
	got, ok := findSection(contents, "GoogleDrive")
	if !ok || got != "[GoogleDrive]\ntype = drive\ntoken = secret\n\n" {
		t.Fatalf("section = %q, %t", got, ok)
	}
}
