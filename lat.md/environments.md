# Environments

See also [[architecture]], [[delivery]].

## Game and mod paths

Default mod root:
```
/mnt/d/SteamLibrary/steamapps/common/Terminator Dark Fate - Defiance/mods/GeneralsMusic
```

Override with `-mod` flag.

## themes.xml locations within a mod

```
ModRoot/
├── basis/scripts/themes.xml                    ← Original campaign
├── dlc/Resistance/basis/scripts/themes.xml     ← Resistance DLC
└── dlc/Legion/basis/scripts/themes.xml         ← Legion DLC
```

Legion structure was manually created by copying from `basis-legion` mod.

## XML format

Excel XML Spreadsheet. Each Row = one mission context:
- Cell 1: context_name (e.g. `mission_1st_mission`)
- Cell 2: comma-separated theme paths (e.g. `themes/USA_01.ogg,themes/mute_10.ogg`)
- Cell 3 (optional): comment

Row tags may have attributes (`<Row ss:AutoFitHeight="0">`). Parser handles both `<Row>` and `<Row ...>`.

## Track types

- **USA tracks** (10): USA_01.ogg through USA_11-01.ogg — toggle on/off
- **Mute tracks** (3): mute_10.ogg, mute_20.ogg, mute_30.ogg — can have duplicates
- **Custom tracks**: any path, added via edit/add in order mode

## Runtime requirements

- Terminal with 256-color support
- WSL2 with access to Windows filesystem via `/mnt/d/`
- Go 1.21+ for building
