# Tests

See also [[architecture]], [[delivery]].

## No automated tests

This is a TUI tool — no unit tests currently. Verification is manual.

## Manual verification checklist

### Core flows
- [ ] Launch: all 3 profiles load, mission list shows sorted
- [ ] Navigate: j/k moves cursor, tracks update in right panel
- [ ] Toggle (t): checkbox panel, Space toggles USA tracks, adds mute tracks
- [ ] Order (o): J/K moves tracks, d deletes, list updates
- [ ] Edit (e in order): inline input with current track name, Enter saves, Esc cancels
- [ ] Add (a in order/toggle): empty input, Enter adds track, Esc cancels
- [ ] Save (s): file written, "Saved!" message, dirty cleared
- [ ] Restore (R): confirm prompt, tracks restored to original, dirty recalculated

### Filter
- [ ] f activates filter, typing filters live
- [ ] Tab switches Track/Mission mode
- [ ] Enter locks filter, j/k navigates filtered list
- [ ] t/o/e/a/s work on filtered missions
- [ ] Esc clears filter, full list returns
- [ ] Sort (r) recomputes filter

### Profiles
- [ ] p cycles profiles, cursor preserved per-profile
- [ ] Dirty prompt on switch if unsaved changes
- [ ] Profile tabs in status bar update

### Edge cases
- [ ] Modified indicator (·) appears/disappears correctly
- [ ] [modified] flag clears after restore all changed missions
- [ ] Empty mission themes: toggle and order handle gracefully
- [ ] Sort toggle (r) preserves cursor on same mission

## Failure signals

- `go build` must succeed
- Launch must not crash with any of the 3 XML formats
- Save must produce valid XML that the game reads
