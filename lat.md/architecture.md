# Architecture

See also [[delivery]], [[environments]], [[tests]].

## System shape

Standalone TUI tool for editing `themes.xml` files in Terminator: Dark Fate — Defiance mods. Built with Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Source files

| File | Responsibility |
|------|---------------|
| [[main.go]] | Entry point. Loads theme files from mod directory, builds `[]FileState`, launches TUI |
| [[model.go]] | Bubble Tea Model — all state, Update logic, View rendering. Modes: View, Add (toggle), Reorder, Edit |
| [[xml.go]] | XML parser (`ParseThemesXML`), serializer (`SerializeThemesXML`), mission sorting. Byte-offset precision for surgical edits |
| [[keys.go]] | Key bindings (`keyMap` struct). All hotkeys defined here |
| [[styles.go]] | Lip Gloss styles — colors, borders, selectors, tab styles |

## Data model

- `FileState` — per-profile state: label, filePath, missions slice, xmlBytes, dirty flag, cursor positions, sort state
- `Mission` — parsed from XML Row: Name, Themes (track list), OriginalThemes (for restore), Comment, byte offsets for serialization, OrigOrder (for unsorted view)
- `Model` — global: files slice, activeFile index, mode, filter state, edit state, confirmations

## Key architectural decisions

- **Byte-offset editing**: Parser tracks exact byte positions of theme data in XML. Serializer applies edits in reverse order to preserve offsets. No full XML re-serialization — only theme cells are touched.
- **Filter over search**: No separate search mode. Filter overlays the mission list — all modes (toggle, order, edit) work on filtered results. `missionCursor` always points to real index in `f.missions`.
- **Per-file state**: Each profile (Original/Resistance/Legion) has independent cursor, scroll, dirty state. Switching profiles preserves position.
- **Dirty calculation**: `recalcDirty()` compares current themes with `OriginalThemes` per mission. Restore correctly clears dirty flag.

## Modes and transitions

```
View ←→ Toggle (t)
View ←→ Reorder (o)
View → Filter (f) → View (Esc)
Reorder ←→ Toggle (t/o)
Reorder → Edit (e) → Reorder (Enter/Esc)
Reorder → Add (a) → Reorder (Enter/Esc)
Toggle → Add (a) → Edit mode
```

## Rendering

Two-panel layout: left (missions, 1/4 width) + right (tracks/toggle/order/edit, 3/4 width). Status bar at bottom in rounded border with help keys left, profile tabs right. Title "Music Themes Changer" above panels.

## Constraints

- XML format: Excel XML Spreadsheet (`urn:schemas-microsoft-com:office:spreadsheet`). Row tags may have attributes (`ss:AutoFitHeight`).
- Track paths stored as `themes/filename.ogg` in XML, displayed without `themes/` prefix.
- Mute tracks can have duplicates, USA tracks are toggles.
