package desktop

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"strings"
	"testing"
)

func TestEntryContainsExecutableAndLocalizedName(t *testing.T) {
	entry := entry("/opt/nuvem/nuvem-desktop")
	if !strings.Contains(entry, "Exec=/opt/nuvem/nuvem-desktop\n") || !strings.Contains(entry, "Name[pt_BR]=Nuvem") {
		t.Fatalf("unexpected desktop entry: %s", entry)
	}
}

func TestEmbeddedIconsHaveExpectedDimensions(t *testing.T) {
	for _, size := range iconSizes {
		contents, err := iconFiles.ReadFile(fmt.Sprintf("assets/icons/nuvem-%d.png", size))
		if err != nil {
			t.Fatal(err)
		}
		config, err := png.DecodeConfig(bytes.NewReader(contents))
		if err != nil {
			t.Fatal(err)
		}
		if config.Width != size || config.Height != size {
			t.Fatalf("icon %d has dimensions %dx%d", size, config.Width, config.Height)
		}
	}
}

func TestInstallWrites120PixelIcon(t *testing.T) {
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	if err := Install("/opt/nuvem/nuvem-desktop"); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(iconPath(data, 120))
	if err != nil {
		t.Fatal(err)
	}
	config, err := png.DecodeConfig(bytes.NewReader(contents))
	if err != nil {
		t.Fatal(err)
	}
	if config.Width != 120 || config.Height != 120 {
		t.Fatalf("installed icon has dimensions %dx%d", config.Width, config.Height)
	}
}
