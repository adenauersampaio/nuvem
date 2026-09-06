package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type recordedRunner struct{ commands []string }

func (r *recordedRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	r.commands = append(r.commands, name+" "+strings.Join(args, " "))
	return "active\n", nil
}

func TestInstallWritesAndEnablesUnit(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "nuvem")
	if err := os.WriteFile(binary, []byte("test"), 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "systemd", "user", UnitName)
	runner := &recordedRunner{}
	manager := New(path, runner)
	if err := manager.Install(context.Background(), binary); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "ExecStart="+binary+" daemon") {
		t.Fatalf("unidade inesperada: %s", contents)
	}
	if len(runner.commands) != 2 || !strings.Contains(runner.commands[1], "enable --now "+UnitName) {
		t.Fatalf("comandos = %#v", runner.commands)
	}
}
