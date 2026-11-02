# Proposal

## Why

At a snowfield grinding spot, the bot sometimes reads a battle as the world map and presses arrow keys, which move the battle menu cursor. A recording of that spot shows the cause: during bright white attack effects, the translucent message window stops matching the battle reference (window score down to 0.485, threshold 0.55) while the white screen resembles the white map (map score up to 0.615, threshold 0.55). Of 128 frames, 24 battle frames were read as MAP and 28 as NEITHER.

The same recording shows that the game menu now reads as BATTLE (window score 0.61-0.67), so the bot would press Z in the menu, where it can select things such as Quit. The user's decision: the game menu means pause.

## What Changes

- Default thresholds change from 0.55 to `windowMinSimilarity` 0.40 and `mapMinSimilarity` 0.70. In the recording, real map frames scored window <= 0.117 and map >= 0.818, and battle frames scored window >= 0.485 and map <= 0.615.
- The bot recognises the game menu by the tall white border of its left panel, and pauses immediately, without the grace period, releasing keys and alerting the user.
- Pressing Space while the game menu is showing does not start the bot, so the menu can't become the map reference.
- `debug classify` shows the menu state and its score.

Out of scope: menu sub-screens such as the save screen, which have a different layout and still read as BATTLE. They can only be reached through the main menu, by which point the bot has paused.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `game-state-detection`: "Classify each frame" adds a MENU state, and battles under bright effects stay BATTLE.
- `auto-grind`: new requirement "Pause on the game menu"; the "Another screen appears" scenario of "Stop on any other screen" no longer covers menus.

## Impact

- `internal/detect`: new `Menu` state, menu border score, new default thresholds.
- `internal/bot`: `Step` pauses on `Menu`; `Runner.Toggle` refuses to start on it.
- `cmd/grinder`: `debug classify` output.
- README settings table and "when it stops" section.
- The user's existing `config.json` stores the old 0.55 values and overrides the new defaults, so it must be updated as well. It is a local, git-ignored file.
- New regression fixtures from the snowfield recording. No new dependencies.
