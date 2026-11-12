package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/prefs"
)

// darkModeEnabled reads the Dark mode setting; when it is unset, the system
// theme decides (dark_mode.cpp).
func darkModeEnabled() bool {
	if v, ok := readDarkModeSetting(); ok {
		return v
	}
	return systemPrefersDark()
}

func setDarkMode(on bool) { writeDarkModeSetting(on) }

func applyTheme(dark bool) {
	if dark {
		imgui.StyleColorsDark()
	} else {
		imgui.StyleColorsLight()
	}
}

// Non-Windows storage of the setting: settings.ini [Views] DarkMode (the
// Windows build keeps HKCU\Software\shStudio 10\DarkMode).
func iniDarkMode() (bool, bool) {
	v := prefs.Default.Int("Views", "DarkMode", -1)
	if v < 0 {
		return false, false
	}
	return v != 0, true
}

func systemPrefersDark() bool {
	if t := os.Getenv("GTK_THEME"); strings.Contains(strings.ToLower(t), "dark") {
		return true
	}
	if out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme").Output(); err == nil {
		if strings.Contains(string(out), "dark") {
			return true
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		if data, err := os.ReadFile(filepath.Join(home, ".config", "kdeglobals")); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "ColorScheme=") && strings.Contains(strings.ToLower(line), "dark") {
					return true
				}
			}
		}
	}
	return false
}
