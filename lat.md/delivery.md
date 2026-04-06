# Delivery

See also [[architecture]], [[tests]].

## Build

```bash
go build -o theme-changer .
```

No Nix, no Docker — plain Go build. Binary is gitignored.

## Run

```bash
./theme-changer              # uses default mod path (GeneralsMusic)
./theme-changer -mod /path   # override mod root
```

## Profile system

Three theme files loaded from mod directory:

| Profile | Relative path |
|---------|--------------|
| Original | `basis/scripts/themes.xml` |
| Resistance | `dlc/Resistance/basis/scripts/themes.xml` |
| Legion | `dlc/Legion/basis/scripts/themes.xml` |

Missing profiles are skipped with a warning. At least one must load.

## Save behavior

- `s` writes current profile's themes.xml
- Serializer applies byte-level edits to original XML bytes
- After save, xmlBytes updated for correct repeated saves
- No auto-save, no backup files

## Submodule in TMT

This repo is a git submodule of [ovitente/tmt](https://github.com/ovitente/tmt) at `tools/theme-changer/`. Commits go to this repo; TMT updates the submodule ref separately.

## lat.md policy

Update lat.md files when:
- New mode or feature added
- Key bindings changed
- File structure changed
- Architectural decision made

Don't update for: cosmetic tweaks, bug fixes within existing patterns, style changes.
