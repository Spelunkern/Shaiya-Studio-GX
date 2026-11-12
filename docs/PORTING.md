# Porting notes

How Shaiya Studio GX is ported from shStudio 10, for whoever continues the
work.

## Rules

- Same features as shStudio 10. Nothing added, nothing removed.
- File formats are ported exactly. Every reader/writer is checked on real
  client files: load + save must give the same bytes, or, where shStudio
  itself changes a file on save, the change must be the same one.
- Bugs of the C++ reconstruction that corrupt files or crash are fixed and
  listed in [DIFFERENCES.md](DIFFERENCES.md). Odd but harmless behavior
  (for example the "clean CR" text trimming) is kept.
- Platform pieces are replaced by portable ones (settings.ini instead of
  registry where the original uses the INI file, ImGui dialogs instead of
  common dialogs) and listed in DIFFERENCES.md too.

## Conventions

- One Go package per shStudio area (`internal/archive`, `internal/sdata`,
  ...). Each Go file names the C++ classes it ports in its comments.
- Field identifiers: shStudio uses four-character tags (`'type'`, `'lvl '`)
  for columns, sorting and generic access. Go keeps them as
  `sdata.Tag` strings with the same four characters, so column tables port
  as they are.
- Serialization formats: `'file'` (client .SData), `'flnt'` (BinarySData
  DB*Data), `'ltxt'` (DB*Text and NpcQuestTrans language files) map to
  `sdata.File`, `sdata.Flnt`, `sdata.Ltxt`.
- Streams: `binstream.Reader` panics with `*binstream.Error` on malformed
  input like iStream throws; wrap parsing in `binstream.Try`.
- Strings in files are converted with the code page of the stream
  (`internal/codepage`); `binstream.Writer.WriteNul` mirrors
  `oStream::m_writeNul`.
- Sorting uses `sdata.QuickSort`, a copy of shStudio's quicksort, so rows
  with equal keys end in the same order.
- Long C++ switch tables (tag -> field) are converted with small scripts
  from the C++ source instead of by hand; see the `ptr(tag)` methods.

## Tests

`go test ./...` runs everywhere; tests that need client files skip unless
these variables are set:

| Variable | Content |
|---|---|
| `SHSTUDIO_SAF` | a client `data.saf` |
| `SHSTUDIO_SDATA` | an encrypted `DB*Data.SData` |
| `SHSTUDIO_TESTDATA` | a folder of sample files (`Item/`, `Monster/`, `NpcQuest/`, `Skill/`, ...) |

The UI can be checked without a person at the screen:
`SHSTUDIO_SCREENSHOT=out.png SHSTUDIO_OPEN=<workspace or data.saf>
SHSTUDIO_CP=<code page> SHSTUDIO_DOC=<archive path> build/shstudiogx` opens
a workspace and a document, saves a capture of the window and exits.

## UI mapping

| shStudio | Shaiya Studio GX |
|---|---|
| main window (StudioFrame + StudioList tree) | `internal/app` main viewport: menu bar, tree, status bar |
| document popups (Doc/DocView) | ImGui windows with their own menu bar; they can be docked or dragged out as OS windows |
| modal dialogs (Modal, AlertDial, FileDialogs) | ImGui popup modals queued with `App.pushModal` |
| ListCtrl / ListCtrlOD | `internal/ui/table` (columns with tags, sort by clicked + previous column, multi-selection, insert/delete/clear, copy/paste) |
| menus from `res/shstudio.res` | written out in `docs/reference/resources.txt` and rebuilt with ImGui |
