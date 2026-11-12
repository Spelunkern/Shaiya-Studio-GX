//go:build windows

package app

import "golang.org/x/sys/windows/registry"

const darkKey = `Software\shStudio 10`

func readDarkModeSetting() (bool, bool) {
	for _, key := range []string{darkKey, `Software\ilya.sh.studio`} {
		k, err := registry.OpenKey(registry.CURRENT_USER, key, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		v, _, err := k.GetIntegerValue("DarkMode")
		k.Close()
		if err == nil {
			return v != 0, true
		}
	}
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err == nil {
		v, _, err := k.GetIntegerValue("AppsUseLightTheme")
		k.Close()
		if err == nil {
			return v == 0, true
		}
	}
	return false, false
}

func writeDarkModeSetting(on bool) {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, darkKey, registry.SET_VALUE)
	if err != nil {
		return
	}
	v := uint32(0)
	if on {
		v = 1
	}
	k.SetDWordValue("DarkMode", v)
	k.Close()
}
