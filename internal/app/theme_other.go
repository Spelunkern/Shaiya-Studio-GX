//go:build !windows

package app

import "github.com/Spelunkern/Shaiya-Studio-GX/internal/prefs"

func readDarkModeSetting() (bool, bool) { return iniDarkMode() }

func writeDarkModeSetting(on bool) {
	v := 0
	if on {
		v = 1
	}
	prefs.Default.SetInt("Views", "DarkMode", v)
}
