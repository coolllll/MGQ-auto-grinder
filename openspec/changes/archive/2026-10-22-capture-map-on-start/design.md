# Design

## Context

See proposal.md for motivation. Current code:
- `detect.NewClassifier(refs, params)` builds the map histogram once from `refs.Map`; there is no way to replace it.
- `detect.LoadRefs` requires both `refs/map.png` and `refs/battle.png` and returns `ErrMissingRefs` otherwise. `grinder run` refuses to start on that error.
- `cmd/grinder run` calls `b.Toggle()` directly when the toggle key is pressed (already gated on the game window being in front by `win.WatchKey`).
- `bot.Runner` owns the per-tick capture/classify/step and depends on the `Window` and `Classifier` interfaces, so it is testable with fakes.
- `calibrate battle` loads `refs/map.png` to recognise the map before an encounter.

Measured at the default size (960x720) with references from the wide size: battle 0.78-0.86, map 0.63-0.69, threshold 0.55. Both sizes render 640x480, so `detect.Normalize` already handles them.

## Goals / Non-Goals

**Goals:**
- A fresh, matching map reference on every start, at either window size.
- One-time setup reduced to `calibrate battle`.
- Start behaviour unit-testable with the existing fakes.

**Non-Goals:**
- Fullscreen mode, mid-run window-size changes, automatic battle-reference refresh.
- Checking that the screen really is the world map when the user starts the bot.

## Decisions

### Start goes through the runner
Add `Runner.Toggle(now)`. When the bot is PAUSED it captures a frame from `Win`, calls `Cls.SetMapReference(frame)`, hands the frame to an optional `OnMapReference func(image.Image)` callback (used by `cmd` to save `refs/map.png`), and only then calls `Bot.Toggle()`. On a capture error it reports through the bot's notifier and leaves the bot PAUSED. When RUNNING it just calls `Bot.Toggle()`. The `Classifier` interface in `internal/bot` gains `SetMapReference(image.Image)`.

Alternative: capture in `cmd/grinder` before calling `Bot.Toggle()`. Rejected because it would leave the new behaviour untestable without Windows.

### Classifier map reference is replaceable
`(*detect.Classifier).SetMapReference(img)` recomputes the map histogram from the normalised image. `NewClassifier` accepts a nil map reference; until one is set, the map histogram is empty, so map similarity is 0 and no frame is classified MAP. That cannot happen while RUNNING, because starting always sets one first.

### References load with the map optional
`detect.LoadRefs` returns `ErrMissingRefs` only when `battle.png` is missing; `Refs.Map` is nil when `map.png` is absent. `run`'s error message names only `calibrate battle`.

### Battle calibration takes its own map reference
`calibrate battle` no longer reads `refs/map.png`. After the user clicks into the game, it captures the current screen as its map reference (the user is told to stand on the world map first), then runs the existing map -> black -> battle detection against it. It also saves that capture to `refs/map.png`, so `debug classify` has a reference.

### Saved map reference
Every start overwrites `refs/map.png` with the new capture, so `debug classify` and `calibrate battle` reflect the latest spot. `calibrate map` stays as a troubleshooting command.

## Risks / Trade-offs

- [The user presses Space somewhere other than the world map, such as in a town] -> That screen becomes the "map" and the bot walks there until stopped. Accepted by the user; they stop it manually.
- [The start capture happens mid-animation, or while a toast or effect is on screen] -> Histograms tolerate small overlays; a bad start shows up immediately as NEITHER, and the bot stops after the grace period. Pausing and restarting retakes it.
- [Starting overwrites a hand-made `refs/map.png`] -> Intended; the file is now a cache of the last start.
- [Old configs and references] -> No config changes. Existing `refs/battle.png` keeps working; existing `refs/map.png` is ignored by `run` and replaced on the first start.
