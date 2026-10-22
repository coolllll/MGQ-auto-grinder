# Tasks

## 1. Setup

- [ ] 1.1 Install Go (`winget install GoLang.Go`) and verify `go version` prints 1.22 or newer
- [ ] 1.2 Create the Go module (`go mod init github.com/coolllll/MGQ-auto-grinder`), add `golang.org/x/sys`, create the `cmd/grinder`, `internal/win`, `internal/detect`, `internal/bot` packages, and verify `go build ./...` succeeds
- [ ] 1.3 Add a `.gitignore` for `grinder.exe`, `refs/` and `config.json`, and verify `git status` stays clean after a build

## 2. Windows layer (`internal/win`)

- [ ] 2.1 Implement window lookup by class `RGSS Player` and a title substring, plus a foreground check; verify with a debug command that prints the found window handle, title and client size while the game is open, and an error when it is closed
- [ ] 2.2 Implement client-area capture via screen-DC `BitBlt` into an `image.RGBA`; verify by saving a capture to PNG and checking it shows the game without borders
- [ ] 2.3 Implement `SendInput` key down/up with scan codes (Z, Left, Right as extended keys) and a release-all helper; verify manually that a debug command walks the character left and right in the game
- [ ] 2.4 Implement `GetAsyncKeyState` edge detection for the toggle key; verify with a unit test of the edge-detection logic over a scripted sequence of key states (held Space toggles once)

## 3. Config

- [ ] 3.1 Implement `config.json` load/save with defaults (title substring, tick rate, walk ms, Z interval, grace period, battle region, thresholds, toggle key); verify with unit tests that a missing file writes defaults and a partial file keeps the given values

## 4. State detection (`internal/detect`)

- [ ] 4.1 Capture fixture PNGs from the real game into `internal/detect/testdata/`: map, map scrolled, battle, battle victory/EXP text, game over, black fade, dialogue over the map; verify all seven files exist
- [ ] 4.2 Implement frame normalisation to 544x416; verify with a unit test that a 1088x832 fixture normalises to the same pixels as its 1x version
- [ ] 4.3 Implement the battle-region comparison and the map histogram comparison, and the MAP / BATTLE / NEITHER classifier; verify with table tests that every fixture gets its expected class, then record the tuned thresholds and battle region as config defaults
- [ ] 4.4 Implement `grinder calibrate map|battle` (3 s countdown, capture, save to `refs/`) and the "references missing" check on `grinder run`; verify both commands save PNGs and `run` refuses to start without them

## 5. Grinding state machine (`internal/bot`)

- [ ] 5.1 Implement PAUSED/RUNNING with the toggle, starting PAUSED; verify with fake-clock unit tests covering start, pause (releases held keys), and a held toggle key toggling once
- [ ] 5.2 Implement MAP behaviour (alternating Left/Right holds for the walk duration) and BATTLE behaviour (Z taps at the interval), releasing keys on every state change; verify with unit tests asserting the recorded key sequence for each, including an encounter starting mid-step
- [ ] 5.3 Implement the NEITHER grace period: no input while NEITHER, pause plus beep and console alert after the grace period, a short fade does not pause; verify with unit tests for game over after battle (no Z sent), a short transition, and a long dialogue
- [ ] 5.4 Implement focus safety (no input and keys released when the game is not in front, RUNNING kept); verify with unit tests for losing and regaining focus
- [ ] 5.5 Implement the window-lost check (stop and report when the game window closes); verify with a unit test using a fake window source

## 6. CLI and docs

- [ ] 6.1 Wire `grinder run` (10 Hz loop: capture, classify, act; console line on each state or mode change) and verify `go build -o grinder.exe ./cmd/grinder` produces a working exe
- [ ] 6.2 Write README usage (install Go, build, calibrate, place the character on an open map, Space to toggle, the Space-is-also-confirm caveat, what makes it stop); verify by following the README from a clean checkout

## 7. Integration check

- [ ] 7.1 With the real game: calibrate, start on the world map, and run through at least five encounters; verify the bot walks, clears each battle including victory text, and returns to walking
- [ ] 7.2 With the real game: trigger a dialogue and switch to another window during a run; verify the bot pauses after the grace period for the dialogue and sends no input while the other window is in front
