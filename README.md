# Shaiya Studio GX

An editor for Shaiya client and server data — SData tables, 3D models,
worlds, maps, data.sah/data.saf archives and SQL Server data — for **Windows
and Linux**.

Shaiya Studio GX is a Go port of [shStudio 10](https://github.com/1ly4337/shStudio)
(C++/Win32, by ilya). The goal is the same program with the same features:
nothing is added or removed. Only the language (Go instead of C++) and the
platform layer change — the UI is drawn with Dear ImGui over GLFW/OpenGL so
one code base runs natively on both systems.

> **Status: early port.** See [docs/PROGRESS.md](docs/PROGRESS.md) for what is
> done, module by module. Linux is the only tested platform for now.

## Build

Requirements: Go 1.27+, a C compiler (cgo is used by GLFW and Dear ImGui),
and on Linux the X11/Wayland and OpenGL development packages.

```sh
go build -o build/shstudiogx ./cmd/shstudiogx
go test ./...
```

Some tests read real client files when these variables are set:

| Variable | File |
|---|---|
| `SHSTUDIO_SAF` | a client `data.saf` (its `data.sah` next to it) |
| `SHSTUDIO_SDATA` | an encrypted `DB*Data.SData` |

## Layout

| Path | Contents | shStudio source |
|---|---|---|
| `internal/crypto/seed` | SEED cipher for BinarySData files | `src/crypto` |
| `internal/codepage` | Windows code pages (strings in files) | Win32 `WideCharToMultiByte` |
| `internal/binstream` | binary reader/writer | `src/core` iStream/oStream |
| `internal/natural` | logical (natural) name order | `StrCmpLogicalW` |
| `internal/archive` | data.sah/data.saf tree, read/write, pack, extract | `src/archive` |

## Differences from shStudio

Behavior differences are kept to platform necessities and to fixing memory
bugs of the original; every one is listed in
[docs/DIFFERENCES.md](docs/DIFFERENCES.md).

## License

MIT — see [LICENSE](LICENSE). Copyright (c) 2026 ilya (shStudio 10) and
Spelunkern (this port). Copies and forks must keep the copyright notices.
