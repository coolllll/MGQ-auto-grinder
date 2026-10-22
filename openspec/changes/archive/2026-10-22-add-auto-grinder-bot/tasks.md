# Tasks

## 1. Setup

- [x] 1.1 Install Go (`winget install GoLang.Go`) and verify `go version` prints 1.22 or newer
- [x] 1.2 Create the Go module (`go mod init github.com/coolllll/MGQ-auto-grinder`), add `golang.org/x/sys`, create the `cmd/grinder`, `internal/win`, `internal/detect`, `internal/bot` packages, and verify `go build ./...` succeeds
- [x] 1.3 Add a `.gitignore` for `grinder.exe`, `refs/` and `config.json`, and verify `git status` stays clean after a build

## 2. Windows layer (`internal/win`)

- [x] 2.1 Implement window lookup by class `RGSS Player` and a title substring, plus a foreground check; verify with a debug command that prints the found window handle, title and client size while the game is open, and an error when it is closed
- [x] 2.2 Implement client-area capture via screen-DC `BitBlt` into an `image.RGBA`; verify by saving a capture to PNG and checking it shows the game without borders
- [x] 2.3 Implement `SendInput` key down/up with scan codes (Z, Left, Right as extended keys) and a release-all helper; verify manually that a debug command walks the character left and right in the game
- [x] 2.4 Implement `GetAsyncKeyState` edge detection for the toggle key; verify with a unit test of the edge-detection logic over a scripted sequence of key states (held Space toggles once)

## 3. Config

- [x] 3.1 Implement `config.json` load/save with defaults (title substring, tick rate, walk ms, Z interval, grace period, window band, thresholds, toggle key); verify with unit tests that a missing file writes defaults and a partial file keeps the given values

## 4. State detection (`internal/detect`)

- [x] 4.1 Capture 640x480 fixture PNGs from the real game into `internal/detect/testdata/` (via `grinder debug fixture` / `debug record`): map, map scrolled, battle, battle command menu, battle skill menu, battle attack flash, victory/EXP text, level-up text, encounter swirl, black fade; verify all ten files exist (game over is covered by the `internal/bot` unit tests instead)
- [x] 4.2 Implement frame normalisation to 640x480; verify with a unit test that a 2x-scaled image normalises to the same pixels as its 1x version
- [x] 4.3 Implement the window-band, whole-frame and above-band histogram comparisons and the MAP / BATTLE / NEITHER rules (band + map above = NEITHER); verify with table tests that every fixture gets its expected class, then record the tuned thresholds and window band as config defaults
- [x] 4.4 Implement `grinder calibrate map|battle` (map: capture once the game comes to the front; battle: wait for an encounter from the map; save to `refs/`) and the "references missing" check on `grinder run`; verify both commands save PNGs and `run` refuses to start without them

## 5. Grinding state machine (`internal/bot`)

- [x] 5.1 Implement PAUSED/RUNNING with the toggle, starting PAUSED; verify with fake-clock unit tests covering start, pause (releases held keys), and a held toggle key toggling once
- [x] 5.2 Implement MAP behaviour (alternating Left/Right holds for the walk duration) and BATTLE behaviour (Z taps at the interval), releasing keys on every state change; verify with unit tests asserting the recorded key sequence for each, including an encounter starting mid-step
- [x] 5.3 Implement the NEITHER grace period: no input while NEITHER, pause plus beep and console alert after the grace period, a short fade does not pause; verify with unit tests for game over after battle (no Z sent), a short transition, and a long non-map, non-battle screen
- [x] 5.4 Implement focus safety (no input and keys released when the game is not in front, RUNNING kept); verify with unit tests for losing and regaining focus
- [x] 5.5 Implement the window-lost check (stop and report when the game window closes); verify with a unit test using a fake window source
- [x] 5.6 Ignore the toggle key unless the game window is in front; verify with a unit test and by typing a space in another window during a run

## 6. CLI and docs

- [x] 6.1 Wire `grinder run` (loop at the configured tick rate: capture, classify, act; console line on each state or mode change) and verify `go build -o grinder.exe ./cmd/grinder` produces a working exe
- [x] 6.2 Write README usage (install Go, build, calibrate, place the character on an open map, Space to toggle, the Space-is-also-confirm caveat, what makes it stop); verify by following the README from a clean checkout

## 7. Integration check

- [x] 7.1 With the real game: calibrate, start on the world map, and run through at least five encounters; verify the bot walks, clears each battle including victory text, and returns to walking
- [x] 7.2 With the real game: open the game menu and switch to another window during a run; verify the bot pauses after the grace period for the menu and sends no input while the other window is in front
