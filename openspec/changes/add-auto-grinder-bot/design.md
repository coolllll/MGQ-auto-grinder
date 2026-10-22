# Design

## Context

- The repository is empty apart from a README, so this is a greenfield Go program. See proposal.md for motivation.
- Target game: `Monster Girl Quest! Paradox RPG v3.01 bobMTL v1.943`, RPG Maker VX Ace (`RGSS301.dll`). Window class is `RGSS Player`. Native resolution 544x416, shown at 2x (`WindowedRatio=2`). The game uses the Fullscreen++ script, whose F5/F6 keys change the window size at runtime, so the captured size can change.
- RGSS reads the real keyboard state each frame. Keys posted to a background window are ignored, so the game must be in front for input to work. Space, Enter and Z are all "confirm" in VX Ace; F1 opens the RGSS settings dialog, F2 shows FPS, and F12 resets the game.
- `errors.txt` shows the game sometimes crashes on missing images mid-battle; a crash dialog must be treated like any other unknown screen.
- Go is not installed on the dev machine yet.

## Goals / Non-Goals

**Goals:**
- One self-contained `grinder.exe` with no runtime dependencies besides Windows.
- State detection that tolerates window scaling and map scrolling.
- Behaviour logic testable without the game running.

**Non-Goals:**
- Choosing battle commands, using skills/items, healing, or saving.
- Reading game memory, editing game files, or injecting RGSS scripts.
- Running with the game in the background or minimized.
- Navigating to a grinding spot; the user places the character first.

## Decisions

### Go with `golang.org/x/sys/windows`, no cgo
All needed Win32 calls (`EnumWindows`, `GetClientRect`, `ClientToScreen`, `GetForegroundWindow`, `BitBlt`, `GetDIBits`, `SendInput`, `GetAsyncKeyState`, `MessageBeep`) are reachable through `x/sys/windows` lazy DLL procs. This keeps the build to `go build` with no C toolchain. Alternatives: `robotgo` (needs cgo/MinGW) and `kbinani/screenshot` (fine, but captures the whole screen, so we'd still need window math). Rejected to keep dependencies at one module.

### Package layout
```
cmd/grinder/        main: CLI (run / calibrate), wiring, console output
internal/win/       window lookup, focus check, capture, SendInput, key polling
internal/detect/    references, frame normalisation, MAP/BATTLE/NEITHER classifier
internal/bot/       state machine, driven by a clock + Frame source + Input sink
```
`internal/bot` depends only on small interfaces (`Classifier`, `Keyboard`, `Clock`), so it can be unit-tested with fakes.

### Capture the window's client area from the screen DC
Find the window with `EnumWindows`, matching class `RGSS Player` and a title containing `Monster Girl Quest` (configurable). Capture with `BitBlt` from the screen DC at the client rectangle. `PrintWindow`/window-DC capture can return black for Direct3D-backed windows. Screen capture is reliable because the game is required to be in front anyway.

### Normalise frames to native 544x416
Every capture and reference is resized to 544x416 before comparison (nearest-neighbour; the 2x scale is an exact integer). This makes references independent of Fullscreen++ window size changes.

### Classifier: two different signals
- **BATTLE**: compare a fixed region (default: the bottom party-status window strip, `y 288..416`) against the same region of the battle reference, using mean absolute per-pixel difference on a 4x-downsampled image. That window stays visible through victory/EXP/level-up messages, which satisfies the "battle detection covers victory messages" requirement. The command window was rejected because it disappears when the battle is won.
- **MAP**: compare a coarse colour histogram (e.g. 4x4x4 RGB bins) of the whole frame against the map reference's histogram (histogram intersection). A histogram barely changes when the map scrolls a few tiles left/right or the character animates, whereas a pixel diff would fail as soon as the camera moves. Battle, game-over, black fades, and dialogue scenes produce very different histograms.
- BATTLE is checked first, then MAP; otherwise NEITHER. Both thresholds and the battle region live in the config file. They are tuned from the captured fixtures.

Alternative considered: template matching for UI elements. Rejected as slower and more fragile to tune than two reference comparisons.

### Calibration from the CLI
`grinder calibrate map` / `grinder calibrate battle` waits 3 seconds (so the user can click into the game), captures one frame, and saves it as PNG under `refs/` next to the exe. No in-game hotkeys are used, which avoids F1/F2/F5/F6/F12 conflicts.

### State machine
```
                 Space
     +--------+ <-----> +-----------------------------------------+
     | PAUSED |         | RUNNING                                 |
     +--------+         |   MAP     -> alternate Left/Right holds |
         ^              |   BATTLE  -> tap Z every interval       |
         |              |   NEITHER -> no input; start timer      |
         |              +-----------------------------------------+
         |                               |
         +---- NEITHER > grace period ---+  (release keys, beep, log)
```
- Tick at ~10 Hz: capture, classify, act.
- Every state change releases all held keys first.
- Focus check happens before every input; when the game is not in front the bot sends nothing, but RUNNING/PAUSED is kept.
- Space is detected with `GetAsyncKeyState` edge detection (press transitions only), so holding it toggles once.

### Input via `SendInput` with scan codes
Use hardware scan codes (`KEYEVENTF_SCANCODE`, plus `KEYEVENTF_EXTENDEDKEY` for the arrow keys) rather than virtual-key events. Games that poll keyboard state accept these more reliably.

### Configuration
`config.json` next to the exe, created with defaults on first run: window title substring, tick rate, walk duration per direction (default 600 ms), Z tap interval (default 100 ms), grace period (default 3 s), battle region, and both thresholds.

### Testing
- `internal/detect`: table tests over PNG fixtures captured from the real game (map, map scrolled, battle, victory text, game over, black fade, dialogue), asserting the expected class.
- `internal/bot`: tests with a fake clock, a scripted classifier, and a recording keyboard, covering every spec scenario (toggle, walk alternation, Z taps, grace period, no input during NEITHER, focus loss).
- `internal/win`: thin wrappers verified manually against the running game.

## Risks / Trade-offs

- [Space is also the game's confirm key, so each toggle sends one confirm to the game] -> Accepted by the user; documented in the README. The key is a config value if it ever needs to change.
- [A dialogue box drawn over the map might still match the map histogram, so the bot would keep walking instead of pausing] -> Message windows add a large block of dark window colour, which shifts the histogram. Include a dialogue-over-map fixture in the tests and set the map threshold tight enough to reject it.
- [A boss or strong enemy beats the party] -> Game over is NEITHER, so the bot sends no input and pauses after the grace period; the save is not touched.
- [The game crashes and shows an error dialog] -> The crash dialog becomes the foreground window, so focus safety stops input immediately and the grace period then pauses the bot.
- [Antivirus may flag a program that sends synthetic keystrokes] -> Build locally from source; no packing or obfuscation.
- [A wall blocks walking in both directions] -> Out of scope; the user picks an open grinding spot.

## Open Questions

- Exact default battle region and thresholds: to be measured from the fixture screenshots during implementation. This doesn't change the approach or the tasks.
- Default walk duration per direction: 600 ms is a guess, to be tuned in play. Also a config value.
