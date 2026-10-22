# Design

## Context

- The repository is empty apart from a README, so this is a greenfield Go program. See proposal.md for motivation.
- Target game: `Monster Girl Quest! Paradox RPG v3.01 bobMTL v1.943`, RPG Maker VX Ace (`RGSS301.dll`). Window class is `RGSS Player`. The game renders at 640x480 (not VX Ace's stock 544x416); on the dev machine the window's client area is 1920x1440, an exact 3x. The game uses the Fullscreen++ script, whose F5/F6 keys change the window size at runtime, so the captured size can change.
- A recorded encounter (map, transition, fight, victory, back to map) showed what every battle screen has in common: blue message/command windows along the bottom ~112 px. The party status row above them disappears during command input and after victory, and the battle background changes with every attack animation.
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
- Handling dialogue or events on the world map or in towns.

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

### Normalise frames to native 640x480
Every capture and reference is resized to 640x480 before comparison (nearest-neighbour; window scales are exact integers). This makes references independent of Fullscreen++ window size changes.

### Classifier: colour histograms of three areas
All comparisons use coarse colour histograms (4x4x4 RGB bins, histogram intersection, 0-1). A histogram barely changes when the map scrolls, the text in a window changes, or a battle animation plays, whereas a pixel-by-pixel comparison breaks on all three.
- **Window band** (bottom 112 px, `y 368..480`) against the same band of the battle reference. Every battle screen shows its blue windows here: plain fight view, command and skill menus, intro text, victory/EXP, item drops and level-up.
- **Whole frame** against the map reference.
- **Area above the band** against the map reference. The map's own text box uses the same window style, so a matching band with the map above it is a dialogue on the map.

Rules: band matches and the area above is the map -> NEITHER (dialogue on the map); band matches -> BATTLE; whole frame matches the map -> MAP; otherwise NEITHER. Defaults: both thresholds 0.55, tuned on the recorded encounter, where battle bands scored 0.69-1.00 (map frames <= 0.12), map frames scored 0.75-1.00 against the map (battle frames <= 0.23), and transitions were NEITHER for 0.5-1.5 s. The band and thresholds live in the config file.

Rejected: comparing the party-status row pixel by pixel (the first design), because the row disappears during command input and after victory; template matching, as slower and more fragile to tune.

### Calibration from the CLI
`grinder calibrate map` / `grinder calibrate battle` waits up to 30 seconds for the user to click into the game, captures one frame 1 second after it comes to the front, and saves it as PNG under `refs/` next to the exe. No in-game hotkeys are used, which avoids F1/F2/F5/F6/F12 conflicts. The same wait is used by the debug commands that capture or press keys (`capture`, `fixture`, `record`, `walk`).

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
`config.json` next to the exe, created with defaults on first run: window title substring, toggle key, tick rate, walk duration per direction (default 600 ms), Z tap interval (default 100 ms), grace period (default 3 s), window band, and both thresholds.

### Testing
- `internal/detect`: table tests over 640x480 PNG fixtures from the real game, asserting the expected class: map, map scrolled, battle, battle command menu, battle skill menu, battle attack flash, victory/EXP text, level-up text, encounter swirl, black fade. There is no game-over fixture (it would mean losing on purpose); the game-over behaviour is covered by the bot tests, and any screen unlike both references is NEITHER. Most battle fixtures come from a `grinder debug record` session of one full encounter.
- `internal/bot`: tests with a fake clock, a scripted classifier, and a recording keyboard, covering every spec scenario (toggle, walk alternation, Z taps, grace period, no input during NEITHER, focus loss).
- `internal/win`: thin wrappers verified manually against the running game.

## Risks / Trade-offs

- [Space is also the game's confirm key, so each toggle sends one confirm to the game] -> Accepted by the user; documented in the README. The key is a config value if it ever needs to change.
- [A dialogue box on the map is the same window style as a battle's, so it could be taken for a battle and pressed through with Z] -> Dialogue on the world map is out of scope. The area-above-the-band rule is kept as a cheap safety net that turns it into NEITHER, but it is not covered by a fixture. Dialogue inside towns is not caught by it (the area above is not the world map) and is also out of scope.
- [A battle background that looks like the map (e.g. a grassland photo) could make a battle look like a map dialogue] -> On the recorded grassland battle the area above scored <= 0.23 against the map, far below 0.55. Revisit if another area's battles stop the bot.
- [A boss or strong enemy beats the party] -> Game over is NEITHER, so the bot sends no input and pauses after the grace period; the save is not touched.
- [The game crashes and shows an error dialog] -> The crash dialog becomes the foreground window, so focus safety stops input immediately and the grace period then pauses the bot.
- [Antivirus may flag a program that sends synthetic keystrokes] -> Build locally from source; no packing or obfuscation.
- [A wall blocks walking in both directions] -> Out of scope; the user picks an open grinding spot.

## Open Questions

- Exact default battle region and thresholds: to be measured from the fixture screenshots during implementation. This doesn't change the approach or the tasks.
- Default walk duration per direction: 600 ms is a guess, to be tuned in play. Also a config value.
