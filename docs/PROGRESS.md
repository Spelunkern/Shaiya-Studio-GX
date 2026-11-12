# Port progress

Shaiya Studio GX ports shStudio 10 (commit `6b17af5`, 1,060 files, ~96.8k
lines of C++) to Go. A module is **done** when every feature of its C++ files
works in Go and is checked against real client data where possible.

Updated: 2026-11-12.

Sample files used by the tests: a client `data.saf` (`SHSTUDIO_SAF`) and a folder of sample
files laid out like Parsec's test data (`SHSTUDIO_TESTDATA`).

| shStudio module | C++ lines | Go package | Status |
|---|---:|---|---|
| `src/crypto` (SEED, data-file header) | 434 | `internal/crypto/seed` | done — round trip of a real DBItemData.SData is byte-identical |
| `src/archive` (sah/saf tree, read, write, holes, pack, extract) | 3,093 | `internal/archive`, `internal/content` | done — real EP5 data.sah (24,655 files) reads; write/grow/extract tested; workspace/archive providers; open/pack/extract dialogs in the app |
| `src/core` (streams, strings, containers, math, files, registry, threads) | 4,404 | `internal/binstream`, `internal/codepage`, `internal/natural` | streams and strings done; the rest is replaced by the Go standard library as each user is ported |
| `src/core/image` (DDS/TGA/BMP, GDI caches) | 2,963 | `internal/imageio`, `internal/icons` | DDS (DXT1-5, masked) and TGA decoders, tested on all 11,857 EP5 images; client icon sheets with shStudio's numbering and blank-cell test; GDI helpers replaced by ImGui textures |
| `src/data` (SData tables) | 11,285 | `internal/sdata` | data layer done: items, item sets, cash shop, monsters, NPC skills, player skills, NpcQuest (NPCs + quests), DB text tables, name lists — all formats (EP4/5/6/6.2 .SData, DB data, DB text) round-trip on sample client files; item-model/icon glue (`ClientData`) comes with the views |
| `src/formats/model` (.3dc/.3de/.3do/.smod/.mlt/.mlx/.itm/.mon, animations, effects) | 6,546 | — | pending |
| `src/formats/world` (.wld/.svmap/.dg/.vani/.wtr/.zon, heights, spots, portals, spawns, NPCs) | 8,941 | — | pending |
| `src/render3d` (OpenGL views, mouse controllers) | 2,938 | — | pending |
| `src/sql` (SQL Server, tables, schema, views, dialogs) | 2,621 | — | pending |
| `src/app` (main frame, settings, workspace/client/server tree, dark mode) | 4,465 | `internal/app`, `internal/prefs` | main window, menus, shortcuts, workspace tree (client, server repository), open/close/pack/new workspace, code page dialog, settings.ini, dark mode; settings dialog and SQL node pending |
| `src/ui/controls` | 10,427 | — | pending |
| `src/ui/views` | 8,722 | `internal/app` | SData document base (load, save, export, merge); Monster, NpcSkill and Skill views; others pending |
| `src/ui/lists` | 9,646 | `internal/ui/table`, `internal/app` | generic list control (ListCtrl/ListCtrlOD); monster and skill lists; item, NPC, cash and model lists pending |
| `src/ui/editors` | 16,914 | — | pending |
| `src/ui/dialogs` | 2,707 | — | pending |
| `src/compat` (decompiler glue) | 683 | — | not needed in Go |

## Next steps

1. Generic table view (lists with columns, sorting, filters, copy/paste, editors), then the SData views.
2. Images (DDS/TGA/BMP) and the text/image viewers.
3. Model and world formats, then the OpenGL views.
4. SQL Server.
