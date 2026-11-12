# Port progress

Shaiya Studio GX ports shStudio 10 (commit `6b17af5`, 1,060 files, ~96.8k
lines of C++) to Go. A module is **done** when every feature of its C++ files
works in Go and is checked against real client data where possible.

Updated: 2026-11-12.

| shStudio module | C++ lines | Go package | Status |
|---|---:|---|---|
| `src/crypto` (SEED, data-file header) | 434 | `internal/crypto/seed` | done — round trip of a real DBItemData.SData is byte-identical |
| `src/archive` (sah/saf tree, read, write, holes, pack, extract) | 3,093 | `internal/archive` | core done — real EP5 data.sah (24,655 files) reads; write/grow/extract tested; providers and dialogs pending |
| `src/core` (streams, strings, containers, math, files, registry, threads) | 4,404 | `internal/binstream`, `internal/codepage`, `internal/natural` | streams and strings done; the rest is replaced by the Go standard library as each user is ported |
| `src/core/image` (DDS/TGA/BMP, GDI caches) | 2,963 | — | pending |
| `src/data` (SData tables) | 11,285 | — | pending |
| `src/formats/model` (.3dc/.3de/.3do/.smod/.mlt/.mlx/.itm/.mon, animations, effects) | 6,546 | — | pending |
| `src/formats/world` (.wld/.svmap/.dg/.vani/.wtr/.zon, heights, spots, portals, spawns, NPCs) | 8,941 | — | pending |
| `src/render3d` (OpenGL views, mouse controllers) | 2,938 | — | pending |
| `src/sql` (SQL Server, tables, schema, views, dialogs) | 2,621 | — | pending |
| `src/app` (main frame, settings, workspace/client/server tree, dark mode) | 4,465 | — | pending |
| `src/ui/controls` | 10,427 | — | pending |
| `src/ui/views` | 8,722 | — | pending |
| `src/ui/lists` | 9,646 | — | pending |
| `src/ui/editors` | 16,914 | — | pending |
| `src/ui/dialogs` | 2,707 | — | pending |
| `src/compat` (decompiler glue) | 683 | — | not needed in Go |

## Next steps

1. Content providers (`ContentProvider`, `DataFileProvider`, `ArchiveProvider`).
2. SData tables (`src/data`), each verified by load + save = identical bytes.
3. Images, then model and world formats.
4. Application shell (GLFW + Dear ImGui), then views, lists and editors.
5. OpenGL views, then SQL Server.
