# Proposal

## Why

The game can be launched at two window sizes: wide (`WindowedRatio=2`, a 1920x1440 window on the user's 150 % display) and default (`WindowedRatio=1`, 960x720). Both render the game at 640x480, and the bot already classifies both correctly. But at the default size, map scores against a map reference taken at the wide size, somewhere else on the map, dropped to 0.63-0.69 against a 0.55 threshold (wide: 0.75-1.00). Battle scores stayed high (0.78-0.86). The user is happy to recalibrate on every run, and pressing Space to start already means "I am standing on the map where I want to grind". So taking the map reference at that moment gives a matching reference every time, at either size and in any area, without a separate step.

## What Changes

- Each press of the toggle key that starts the bot first captures the current game screen as the map reference, replaces the classifier's map reference with it, and saves it to `refs/map.png`. Pauses keep the existing reference.
- If that capture fails, the bot stays paused and reports why.
- `grinder run` only requires the battle reference. A missing `refs/map.png` no longer stops it from starting.
- `grinder calibrate map` stays available, for `debug classify` and troubleshooting, but is no longer part of setup. Setup becomes "run `calibrate battle` once".
- README: calibration section and a note that the bot works at either window size.

Out of scope: fullscreen mode (black bars), detecting a mid-run window-size change, refreshing the battle reference automatically, and guarding against starting the bot somewhere other than the world map (the user stops it manually).

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `auto-grind`: "Start and stop with Space" now captures a fresh map reference each time the bot starts.
- `game-state-detection`: "Calibrate reference images" now takes the map reference from the start press; only the battle reference needs the calibration command, and only it is required to run.

## Impact

- `internal/detect`: classifier gains a way to replace its map reference; references load with the map optional.
- `internal/bot`: the runner's start path captures and installs the map reference before switching to RUNNING.
- `cmd/grinder`: `run` routes toggle presses through the runner and saves the captured reference; the missing-references check only covers the battle reference.
- README and tests. No new dependencies, and no config changes.
