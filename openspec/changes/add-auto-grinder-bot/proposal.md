# Proposal

## Why

Grinding EXP in *Monster Girl Quest! Paradox RPG* (RPG Maker VX Ace) means hours of walking back and forth and mashing confirm through random battles. A small bot that watches the game window and plays this loop for the user removes the tedium, while stopping safely the moment anything unexpected appears (game over, a town or menu, a crash dialog).

## What Changes

- New Windows command-line program, written in Go, that runs alongside the game.
- Reads the game's state only from screen captures of the game window (no memory reading, no game file or script changes).
- One-time calibration step: the user captures a reference of the world map and of a battle; the bot classifies live frames against these references.
- On the map, the bot walks the character left and right to trigger random encounters.
- In battle, the bot presses Z repeatedly until the battle ends.
- Whenever the screen is neither map nor battle for longer than a short grace period (default ~3 s, covers encounter transitions and victory text), the bot stops. This includes the game-over screen.
- The Space key toggles the bot between running and paused.
- The bot sends no input while the game window is not the foreground window.

## Capabilities

### New Capabilities
- `game-state-detection`: Locating the game window, calibrating reference images, and classifying each captured frame as map, battle, or neither.
- `auto-grind`: The grinding behavior driven by the detected state: walking on the map, pressing Z in battle, stopping on anything else, the Space start/stop toggle, and focus safety.

### Modified Capabilities
- None (the project has no existing specs).

## Impact

- New Go module in this repository (currently README-only); produces a single `.exe`.
- Requires Go to be installed to build (`winget install GoLang.Go`); not currently installed on the dev machine.
- Windows-only: depends on Win32 APIs for window lookup, screen capture, keyboard input, and key-state polling.
- Target game: `Monster Girl Quest Part 3 - Copy` (RGSS3 / VX Ace, `WindowedRatio=2`). The game itself is not modified.
- Assumption to confirm during review: the 3-second grace period was proposed but not explicitly confirmed by the user; it is configurable.
