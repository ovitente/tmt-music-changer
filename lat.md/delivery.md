# Delivery

How to build, run, and ship this tool.

See also [[architecture]], [[tests]].

## Build

Plain Go build, no external toolchain required.

```bash
go build -o theme-changer .
```

Binary is gitignored.

## Run

Launch with default mod path or override with `-mod` flag.

```bash
./theme-changer              # uses default mod path (GeneralsMusic)
./theme-changer -mod /path   # override mod root
```

## Profile system

Three theme files loaded from the mod directory structure.

| Profile | Relative path |
|---------|--------------|
| Original | `basis/scripts/themes.xml` |
| Resistance | `dlc/Resistance/basis/scripts/themes.xml` |
| Legion | `dlc/Legion/basis/scripts/themes.xml` |

Missing profiles are skipped with a warning. At least one must load.

## Save behavior

Save writes the current profile's themes.xml using byte-level edits.

- `s` triggers save for active profile
- Serializer applies edits to original XML bytes, not full re-serialization
- After save, xmlBytes updated for correct repeated saves
- No auto-save, no backup files

## Submodule in TMT

This repo is a submodule of ovitente/tmt at `tools/theme-changer/`.

Commits go to this repo; TMT updates the submodule ref separately.

## lat.md policy

Update lat.md when features, keybindings, or architecture change. Skip for cosmetic fixes.
