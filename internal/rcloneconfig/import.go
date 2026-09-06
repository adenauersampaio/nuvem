// Package rcloneconfig migrates one existing remote into Nuvem's private state.
package rcloneconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ImportRemote(source, destination, remote string) error {
	contents, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("ler configuração de origem: %w", err)
	}
	section, ok := findSection(string(contents), remote)
	if !ok {
		return fmt.Errorf("remote %q não encontrado", remote)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return fmt.Errorf("criar diretório privado: %w", err)
	}
	if err := os.WriteFile(destination, []byte(section), 0o600); err != nil {
		return fmt.Errorf("salvar configuração privada: %w", err)
	}
	return nil
}

func findSection(contents, remote string) (string, bool) {
	lines := strings.Split(contents, "\n")
	header := "[" + remote + "]"
	start := -1
	for index, line := range lines {
		if strings.TrimSpace(line) == header {
			start = index
			continue
		}
		if start >= 0 && strings.HasPrefix(strings.TrimSpace(line), "[") && strings.HasSuffix(strings.TrimSpace(line), "]") {
			return strings.Join(lines[start:index], "\n") + "\n", true
		}
	}
	if start >= 0 {
		return strings.Join(lines[start:], "\n"), true
	}
	return "", false
}
