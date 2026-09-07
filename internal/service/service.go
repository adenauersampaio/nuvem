// Package service installs the Nuvem daemon as a user service.
package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const UnitName = "nuvem.service"

type CommandRunner interface {
	Run(context.Context, string, ...string) (string, error)
}

type OSRunner struct{}

func (OSRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return string(output), nil
}

type Manager struct {
	unitPath string
	runner   CommandRunner
}

func DefaultUnitPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "systemd", "user", UnitName), nil
}

func New(unitPath string, runner CommandRunner) Manager {
	return Manager{unitPath: unitPath, runner: runner}
}

func (m Manager) Install(ctx context.Context, binary string) error {
	if !filepath.IsAbs(binary) || strings.ContainsAny(binary, " \t\n") {
		return errors.New("o executável deve ter caminho absoluto sem espaços")
	}
	info, err := os.Stat(binary)
	if err != nil {
		return fmt.Errorf("acessar executável: %w", err)
	}
	if info.IsDir() || info.Mode()&0o111 == 0 {
		return errors.New("o caminho informado não é um executável")
	}
	if err := os.MkdirAll(filepath.Dir(m.unitPath), 0o700); err != nil {
		return fmt.Errorf("criar diretório do serviço: %w", err)
	}
	if err := os.WriteFile(m.unitPath, []byte(unit(binary)), 0o600); err != nil {
		return fmt.Errorf("salvar serviço: %w", err)
	}
	if _, err := m.runner.Run(ctx, "systemctl", "--user", "daemon-reload"); err != nil {
		return err
	}
	if _, err := m.runner.Run(ctx, "systemctl", "--user", "enable", "--now", UnitName); err != nil {
		return err
	}
	return nil
}

func (m Manager) Uninstall(ctx context.Context) error {
	if _, err := m.runner.Run(ctx, "systemctl", "--user", "disable", "--now", UnitName); err != nil {
		return err
	}
	if err := os.Remove(m.unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remover serviço: %w", err)
	}
	_, err := m.runner.Run(ctx, "systemctl", "--user", "daemon-reload")
	return err
}

func (m Manager) Status(ctx context.Context) (string, error) {
	output, err := m.runner.Run(ctx, "systemctl", "--user", "show", UnitName, "--property", "ActiveState", "--value")
	return strings.TrimSpace(output), err
}

func (m Manager) Stop(ctx context.Context) error {
	_, err := m.runner.Run(ctx, "systemctl", "--user", "stop", UnitName)
	return err
}

func (m Manager) Start(ctx context.Context) error {
	_, err := m.runner.Run(ctx, "systemctl", "--user", "start", UnitName)
	return err
}

func (m Manager) Restart(ctx context.Context) error {
	_, err := m.runner.Run(ctx, "systemctl", "--user", "restart", UnitName)
	return err
}

func unit(binary string) string {
	return "[Unit]\n" +
		"Description=Nuvem continuous synchronization\n" +
		"After=network-online.target\n" +
		"Wants=network-online.target\n\n" +
		"[Service]\n" +
		"Type=simple\n" +
		"ExecStart=" + binary + " daemon\n" +
		"Restart=on-failure\n" +
		"RestartSec=15s\n\n" +
		"[Install]\n" +
		"WantedBy=default.target\n"
}
