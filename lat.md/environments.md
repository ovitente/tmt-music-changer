# Environments

Game paths, mod structure, and runtime requirements.

See also [[architecture]], [[delivery]].

## Game and mod paths

Default mod root points to the GeneralsMusic mod on a Windows drive via WSL.

```
/mnt/d/SteamLibrary/steamapps/common/Terminator Dark Fate - Defiance/mods/GeneralsMusic
```

Override with `-mod` flag.

## themes.xml locations within a mod

Each campaign has its own themes.xml under the mod root.

```
ModRoot/
├── basis/scripts/themes.xml                    ← Original campaign
├── dlc/Resistance/basis/scripts/themes.xml     ← Resistance DLC
└── dlc/Legion/basis/scripts/themes.xml         ← Legion DLC
```

## XML format

Excel XML Spreadsheet with Row/Cell/Data elements.

Each Row is one mission context. Cell 1: context_name. Cell 2: comma-separated theme paths. Cell 3 (optional): comment. Row tags may have attributes like `ss:AutoFitHeight`.

## Track types

Three categories of tracks with different toggle behavior.

- **USA tracks** (10): USA_01.ogg through USA_11-01.ogg — toggle on/off
- **Mute tracks** (3): mute_10.ogg, mute_20.ogg, mute_30.ogg — can have duplicates
- **Custom tracks**: any path, added via edit/add in order mode

## Runtime requirements

Minimal requirements for building and running the tool.

- Terminal with 256-color support
- WSL2 with access to Windows filesystem via `/mnt/d/`
- Go 1.21+ for building
