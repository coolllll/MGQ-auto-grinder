# Tasks

## 1. Detection (`internal/detect`)

- [x] 1.1 Add `(*Classifier).SetMapReference(image.Image)` and let `NewClassifier` accept a nil map reference (empty map histogram, nothing classified MAP); verify with unit tests that a nil-map classifier never returns MAP, and that after `SetMapReference(map fixture)` the map fixtures classify as MAP and the battle fixtures still classify as BATTLE
- [x] 1.2 Make `LoadRefs` require only `battle.png` (map optional, `Refs.Map` nil when absent); verify with a unit test for "battle only" (loads, no error), "map only" (`ErrMissingRefs`) and "both" (loads both)
- [x] 1.3 Verify a same-spot reference restores the map margin: unit test that `map_scrolled.png` scored against `map.png` taken as the reference stays well above the threshold, and log the score

## 2. Start path (`internal/bot`)

- [x] 2.1 Add `SetMapReference(image.Image)` to the bot `Classifier` interface and `Runner.Toggle(now)` with an optional `OnMapReference` callback: when PAUSED, capture, set the reference, call the callback, then start; when RUNNING, pause; verify with unit tests using the fake window and a recording classifier (start sets the reference before the first step, callback receives the frame, pause does not capture)
- [x] 2.2 Keep the bot PAUSED and report when the start capture fails; verify with a unit test using a fake window whose capture errors (mode stays PAUSED, an alert or info message names the error, no keys sent)
- [x] 2.3 Restarting retakes the reference; verify with a unit test that start, pause, start captures twice and the classifier ends up with the second frame

## 3. CLI (`cmd/grinder`)

- [x] 3.1 Route toggle presses in `run` through `Runner.Toggle`, save the captured frame to `refs/map.png` in `OnMapReference`, report "Map reference captured", and change the missing-references message to name only `calibrate battle`; verify with `go build` and by running `grinder run` with only `refs/battle.png` present (starts, waits for Space)
- [x] 3.2 Make `calibrate battle` capture its own map reference after the game comes to the front (and save it to `refs/map.png`) instead of loading one; verify by deleting `refs/map.png` and running `calibrate battle` through an encounter in the game

## 4. Docs

- [x] 4.1 Update the README: setup is `calibrate battle` once; the map reference is taken whenever you press Space, so start on open world map; works at either window size; `calibrate map` is for troubleshooting only. Verify every command in the README matches `grinder` usage output

## 5. Integration check

- [x] 5.1 With the real game at the default window size (960x720): delete `refs/map.png`, start with Space on the world map, and run through at least three encounters; verify the bot walks, clears each battle and returns to walking
- [x] 5.2 With the real game at the wide window size (1920x1440): start with Space and run through at least three encounters; verify the same
