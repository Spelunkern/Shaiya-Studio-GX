// Package prefs stores the settings of the program in settings.ini next to
// the executable, like shStudio's IniFile/Preferences (which use
// GetPrivateProfileString on "settings.ini" in the program folder).
package prefs

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// File is an INI file with the semantics of the Win32 private profile API:
// section and key names are case-insensitive, values are kept as written.
type File struct {
	mu       sync.Mutex
	path     string
	sections []*section
}

type section struct {
	name string
	keys []kv
}

type kv struct{ key, value string }

// AppSection is the section of the main window placement: the executable
// file name, as in shStudio (m_appName).
var AppSection = "shStudio 10.exe"

// Default is the settings file of the program.
var Default = Open(defaultPath())

func defaultPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "settings.ini"
	}
	AppSection = filepath.Base(exe)
	return filepath.Join(filepath.Dir(exe), "settings.ini")
}

// Open loads path (a missing file is an empty one).
func Open(path string) *File {
	f := &File{path: path}
	f.load()
	return f
}

func (f *File) load() {
	fh, err := os.Open(f.path)
	if err != nil {
		return
	}
	defer fh.Close()
	var cur *section
	sc := bufio.NewScanner(fh)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == ';' {
			continue
		}
		if line[0] == '[' {
			name := strings.TrimSuffix(line[1:], "]")
			cur = f.section(name, true)
			continue
		}
		if cur == nil {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		cur.keys = append(cur.keys, kv{strings.TrimSpace(k), strings.TrimSpace(v)})
	}
}

func (f *File) section(name string, create bool) *section {
	for _, s := range f.sections {
		if strings.EqualFold(s.name, name) {
			return s
		}
	}
	if !create {
		return nil
	}
	s := &section{name: name}
	f.sections = append(f.sections, s)
	return s
}

func (f *File) save() {
	var b strings.Builder
	for _, s := range f.sections {
		b.WriteString("[" + s.name + "]\r\n")
		for _, e := range s.keys {
			b.WriteString(e.key + "=" + e.value + "\r\n")
		}
	}
	os.WriteFile(f.path, []byte(b.String()), 0o644)
}

func (f *File) lookup(sec, key string) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s := f.section(sec, false)
	if s == nil {
		return "", false
	}
	for _, e := range s.keys {
		if strings.EqualFold(e.key, key) {
			return e.value, true
		}
	}
	return "", false
}

// String returns the value of key in sec, or def (GetPrivateProfileString).
func (f *File) String(sec, key, def string) string {
	if v, ok := f.lookup(sec, key); ok && v != "" {
		return unquote(v)
	}
	return def
}

// Int returns the value of key in sec as an int, or def
// (GetPrivateProfileInt: leading digits only, optional minus sign).
func (f *File) Int(sec, key string, def int) int {
	v, ok := f.lookup(sec, key)
	if !ok {
		return def
	}
	v = unquote(v)
	n := 0
	for n < len(v) && (v[n] >= '0' && v[n] <= '9' || n == 0 && v[n] == '-') {
		n++
	}
	if n == 0 {
		return def
	}
	i, err := strconv.Atoi(v[:n])
	if err != nil {
		return def
	}
	return i
}

// SetString writes key=value in sec and saves the file
// (WritePrivateProfileString).
func (f *File) SetString(sec, key, value string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s := f.section(sec, true)
	for i := range s.keys {
		if strings.EqualFold(s.keys[i].key, key) {
			s.keys[i].value = value
			f.save()
			return
		}
	}
	s.keys = append(s.keys, kv{key, value})
	f.save()
}

// SetInt writes an int value.
func (f *File) SetInt(sec, key string, value int) { f.SetString(sec, key, strconv.Itoa(value)) }

func unquote(v string) string {
	if len(v) >= 2 && (v[0] == '"' && v[len(v)-1] == '"' || v[0] == '\'' && v[len(v)-1] == '\'') {
		return v[1 : len(v)-1]
	}
	return v
}
