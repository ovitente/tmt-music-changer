# Architecture

TUI tool for editing themes.xml in Terminator: Dark Fate mods. Go + Bubble Tea + Lip Gloss.

See also [[delivery]], [[environments]], [[tests]].

## Source files

Five Go files, each with a single responsibility.

| File | Responsibility |
|------|---------------|
| [[main.go]] | Entry point, file loading, profile definitions |
| [[model.go]] | Bubble Tea Model — state, Update, View. Modes: View, Add, Reorder, Edit |
| [[xml.go]] | XML parser, serializer, mission sorting. Byte-offset precision |
| [[keys.go]] | Key bindings (`keyMap` struct) |
| [[styles.go]] | Lip Gloss styles — colors, borders, selectors |

## Data model

Three core structs hold all application state.

- `FileState` — per-profile: label, filePath, missions, xmlBytes, dirty, cursors, sort state
- `Mission` — from XML: Name, Themes, OriginalThemes, Comment, byte offsets, OrigOrder
- `Model` — global: files, activeFile, mode, filter state, edit state, confirmations

## Key architectural decisions

Design choices that shape the codebase.

- **Byte-offset editing**: parser tracks byte positions, serializer applies edits in reverse order
- **Filter over search**: filter overlays mission list, all modes work on filtered results
- **Per-file state**: each profile has independent cursor, scroll, dirty state
- **Dirty calculation**: `recalcDirty()` compares themes with originals per mission

## Modes and transitions

The TUI has four modes plus an inline filter overlay.

```
View <-> Toggle (t)
View <-> Reorder (o)
View -> Filter (f) -> View (Esc)
Reorder <-> Toggle (t/o)
Reorder -> Edit (e) -> Reorder (Enter/Esc)
Reorder -> Add (a) -> Reorder (Enter/Esc)
Toggle -> Add (a) -> Edit mode
```

## Rendering

Two-panel layout with status bar and title.

Left panel (1/4 width): mission list with filter overlay. Right panel (3/4): tracks/toggle/order/edit. Status bar: help keys left, profile tabs right. Title above panels.

## Constraints

Hard rules from the game's XML format.

- Excel XML Spreadsheet format. Row tags may have attributes.
- Tracks stored as `themes/filename.ogg`, displayed without prefix.
- Mute tracks allow duplicates, USA tracks are toggles.
