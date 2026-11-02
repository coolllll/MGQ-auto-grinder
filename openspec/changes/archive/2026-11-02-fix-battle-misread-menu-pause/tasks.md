# Tasks

## 1. Regression fixtures

- [x] 1.1 Copy from `refs/` and the snowfield recording (`..\_rec_snow`) into `internal/detect/testdata/`: `map_snow.png` (`refs/map.png`), `battle_snow_flash.png` (frame 0000, window 0.485, map 0.611), `battle_snow_flash2.png` (frame 0054, window 0.491, map 0.598), `map_snow_walking.png` (a real map frame such as 0111), `menu_snow.png` (frame 0082) and `menu_snow2.png` (frame 0117); verify all six files exist and open as 640x480
- [x] 1.2 Add tests, using `map_snow.png` as the map reference and default params: the flash frames are BATTLE, the walking frame is MAP, the menu frames are MENU, and all-white and battle fixtures are not MENU; verify the flash-frame and menu assertions fail on the current code

## 2. Detection (`internal/detect`)

- [x] 2.1 Change the default thresholds to window 0.40 and map 0.70; verify with `go test ./internal/detect` that the flash frames now pass and all existing fixtures still classify correctly
- [x] 2.2 Add the `Menu` state, a `Scores.MenuBorder` score (thin-line check over x 160..164, top and band ranges) and the MENU-first rule in `StateFor`; verify the menu fixture tests pass, every non-menu fixture scores below 0.5, and a synthetic all-white frame is not MENU

## 3. Bot (`internal/bot`)

- [x] 3.1 Pause at once on `detect.Menu` in `Step`, releasing keys and alerting; verify with a unit test that held keys are released, the mode is PAUSED on the same tick, and no further input is sent
- [x] 3.2 Make `Runner.Toggle` refuse to start when the captured frame is `Menu`; verify with a unit test that the bot stays PAUSED, the map reference and `OnMapReference` are untouched, and the alert names the menu

## 4. CLI, config and docs

- [x] 4.1 Show the menu state and score in `debug classify`; verify with `go vet ./...`, `go build`, and a `debug classify` run with the menu open that prints MENU
- [x] 4.2 Update the README (settings table values, menu pausing in "When it stops by itself", refusing to start on the menu) and set the user's local `config.json` to window 0.40 and map 0.70; verify by reading both files, and that a `debug classify` run prints the new minimums

## 5. Integration check

- [x] 5.1 With the real game at the default window size at the snowfield: start with Space and run through at least three encounters including bright attack effects; verify no arrow keys in battle (no `State: MAP` during a battle in the log) and that each battle is cleared
- [x] 5.2 Open the game menu during a run at the default window size; verify the bot pauses at once, that Space with the menu open reports it won't start, and check the log for any BATTLE frame while the menu opens

## Not verified

Skipped by the user's decision; these were planned as playtests and did not run:

- The game menu at the wide window size (1920x1440). The border check has only seen menu frames from 960x720.
- The new thresholds in a real run at a second grinding spot. At other spots they are covered only by committed fixtures, not by play.
