# Differences from shStudio 10

Shaiya Studio GX keeps shStudio's features and file formats. This page lists
every place where the behavior is not byte-for-byte the same, and why.

## Platform

| Area | shStudio | Shaiya Studio GX |
|---|---|---|
| Settings | registry `HKCU\Software\shStudio 10` | the same keys in a config file (see below) |
| `CP_ACP` / `CP_OEMCP` code pages | the Windows system code pages | the Windows system code pages on Windows; 1252 / 437 elsewhere |
| Code page 1361 (Johab) | full | ASCII and the 11,172 modern Hangul syllables; Johab symbols and Hanja become `?` |
| Logical name order | `StrCmpLogicalW` | case-insensitive order with numeric digit runs (`internal/natural`); punctuation is ordered by code point |
| Hidden files when scanning a workspace | `FILE_ATTRIBUTE_HIDDEN` | the hidden attribute on Windows, dot files elsewhere |
| Archive paths | `folder\file` | unchanged inside the archive; disk paths use the OS separator |

## Bugs of the original that are fixed

| Where | shStudio | Shaiya Studio GX |
|---|---|---|
| `SahArchive::Holes::add` | clears memory past the end of the grown array | grows the list without the overflow |
| `SahArchive::Holes::consume` | when the consumed hole is the second to last, the last hole is dropped and the consumed one stays | removes the consumed hole |
| `SahArchive::saveHoles` | when holes.bin changes size, the new table is appended after the old one in the same buffer (the old table is what gets read back) and the index keeps the old holes.bin size | writes the updated table and records its size |
| `sub_418C00` (decrypt) | a stored plain size larger than the body reads past the buffer | rejected as an error |
| `Monster::read` (EP4/EP5 `Monster.SData`) | the 5th attack byte is ignored and the third attack type is taken from the 3rd byte; saving then writes that type over the 5th byte, changing about 2,000 bytes of an untouched file | the third attack type is read from the 5th byte, so an untouched file saves identical |
| `Monster::read` (`MonsterEx.SData`) | a drop count above 9 writes past the drop arrays | extra drops are read and ignored |
| `Monster::assignFrom` (paste) | appends the source name to the current name | replaces the name |
| `MobsSkills::read` (`DBNpcSkillText_*.SData`) | reads the row count without skipping the four column names it writes itself, so the language file can not be loaded | checks and skips the column names |
| `MobsSkills::read` (`DBNpcSkillData.SData`) | creates the rows with the skill id as version and 0x1620 as id (swapped arguments); the id is then read back, but the version stays wrong and CSV/SQL exports of DB rows lose abilities 4 to 10 | rows get version 0x1620 |
| `ToonSkills::read` (`DBSkillText_*.SData`) | same as for NPC skills: the column names are not skipped and the file can not be loaded | checks and skips the column names |
| `NpcData::write` (`NpcQuest.SData`) | writes 14 NPC lists starting with an empty list 0, but reads 13 starting with list 1, so a saved file has an extra count at the start and can not be read back | writes lists 1 to 13 |
| `StandardNPC::writeMain` | writes the NPC type as 4 bytes where it reads 1 byte, so every non-merchant, non-gate-keeper NPC grows by 3 bytes on save and the file can not be read back | writes 1 byte |
| `QuestsDetails::read` (`NpcQuestTrans_*.SData`) | after reading the texts it goes on to the next format, finds none and throws, so the language file is never loaded | returns after reading the texts |
| `QuestDetails::read` (`NpcQuestTrans_*.SData`) | reads the 12 texts in an order different from the one it writes and from the files (replies end up as description, welcome, ...) | reads them in the file order: name, summary, 6 replies, description, welcome, reminder, alternate message |
| `writeFixedString256` (256-byte name lists) | writes the 4-byte length plus the first 252 bytes of the field, then the text again, which `readFixedString256` can not read back | writes the NUL-padded 256-byte field that is read |
