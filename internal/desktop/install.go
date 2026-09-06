// Package desktop installs Nuvem's Linux desktop entry and icon.
package desktop

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ID          = "io.github.adenauersampaio.Nuvem"
	desktopName = ID + ".desktop"
)

//go:embed assets/icons/*.png
var iconFiles embed.FS

var iconSizes = [...]int{16, 32, 48, 64, 120, 128, 256, 512, 1024}

func Install(executable string) error {
	if !filepath.IsAbs(executable) || strings.ContainsAny(executable, "\n\r") {
		return fmt.Errorf("the desktop executable path must be absolute")
	}
	data, err := userDataDir()
	if err != nil {
		return err
	}
	applications := filepath.Join(data, "applications")
	if err := os.MkdirAll(applications, 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(applications, desktopName), []byte(entry(executable)), 0o644); err != nil {
		return err
	}
	return installIcons(data)
}

func Uninstall() error {
	data, err := userDataDir()
	if err != nil {
		return err
	}
	paths := []string{
		filepath.Join(data, "applications", desktopName),
		filepath.Join(data, "icons", "hicolor", "scalable", "apps", ID+".svg"),
	}
	for _, size := range iconSizes {
		paths = append(paths, iconPath(data, size))
	}
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func userDataDir() (string, error) {
	if directory := os.Getenv("XDG_DATA_HOME"); directory != "" {
		return directory, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

func installIcons(data string) error {
	for _, size := range iconSizes {
		contents, err := iconFiles.ReadFile(fmt.Sprintf("assets/icons/nuvem-%d.png", size))
		if err != nil {
			return err
		}
		destination := iconPath(data, size)
		if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(destination, contents, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func iconPath(data string, size int) string {
	directory := fmt.Sprintf("%dx%d", size, size)
	return filepath.Join(data, "icons", "hicolor", directory, "apps", ID+".png")
}

func entry(executable string) string {
	return "[Desktop Entry]\n" +
		"Type=Application\n" +
		"Name=Nuvem\n" +
		"Name[pt_BR]=Nuvem\n" +
		"Comment=Continuous cloud folder synchronization\n" +
		"Comment[pt_BR]=Sincronização contínua de pastas na nuvem\n" +
		"Exec=" + executable + "\n" +
		"Icon=" + ID + "\n" +
		"Categories=Utility;FileTransfer;\n" +
		"StartupNotify=true\n"
}
