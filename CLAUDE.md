# Music Theme Changer

TUI tool for editing themes.xml in Terminator: Dark Fate — Defiance mods.

## Architecture
Go + Bubble Tea. See `lat.md/` for full knowledge base.

## Key files
- `main.go` — entry, file loading, profile definitions
- `model.go` — TUI model, all modes, rendering
- `xml.go` — XML parser/serializer, sorting
- `keys.go` — keybindings
- `styles.go` — Lip Gloss styles

## Build & Run
```bash
go build -o theme-changer .
./theme-changer
```

## Keybindings
| Key | Mode | Action |
|-----|------|--------|
| j/k | All | Navigate |
| t | View | Toggle tracks (checkbox) |
| o | View | Order mode (reorder/delete) |
| f | View | Filter missions |
| r | View | Sort toggle (sorted/original) |
| p | View | Switch profile |
| R | View | Restore mission |
| s | All | Save |
| e | Order | Edit track name |
| a | Order/Toggle | Add custom track |
| d | Order | Delete track |
| J/K | Order | Move track up/down |
| Space | Toggle | Toggle track on/off |
| Tab | Filter | Switch track/mission filter |
| Esc | Any | Back / clear filter |
| q | View | Quit |

## Git Workflow
Direct to main for all changes. This is a standalone tool, not production code.

## lat.md policy
Update `lat.md/` when adding features, changing keybindings, or making architectural decisions. Skip for cosmetic fixes.
