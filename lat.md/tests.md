# Tests

Test strategy and verification for this TUI tool.

See also [[architecture]], [[delivery]].

## Required checks

Automated gates that must pass before commit.

- `go build` must succeed
- `lat check` must pass (enforced by pre-commit hook)

## Manual verification checklist

No automated tests — verification is manual against the running TUI.

### Core flows

Essential operations that must work after any change.

- Launch: all 3 profiles load, mission list shows sorted
- Navigate: j/k moves cursor, tracks update in right panel
- Toggle (t): Space toggles USA tracks, adds mute tracks
- Order (o): J/K moves tracks, d deletes
- Edit (e in order): inline input, Enter saves, Esc cancels
- Add (a): empty input, Enter adds track
- Save (s): file written, dirty cleared
- Restore (R): confirm prompt, tracks restored, dirty recalculated

### Filter

Filter overlay must work with all modes.

- f activates, typing filters live
- Tab switches Track/Mission mode
- Enter locks, j/k navigates filtered list
- t/o/e/a/s work on filtered missions
- Esc clears filter
- Sort (r) recomputes filter

### Profiles

Profile switching must preserve per-file state.

- p cycles profiles, cursor preserved per-profile
- Dirty prompt on switch if unsaved changes
- Profile tabs in status bar update

### Edge cases

Boundary conditions that have caused bugs before.

- Modified indicator (middot) appears/disappears correctly
- [modified] clears after restoring all changed missions
- Empty mission themes handled gracefully
- Sort toggle (r) preserves cursor on same mission

## Failure signals

Conditions that must block a commit or release.

- `go build` failure
- `lat check` failure
- Launch crash with any of the 3 XML formats
- Save produces invalid XML
