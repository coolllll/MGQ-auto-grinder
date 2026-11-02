# Design

## Context

See proposal.md for motivation. Current code:
- `detect.StateFor`: window band >= `WindowMinSimilarity` -> BATTLE; whole frame >= `MapMinSimilarity` -> MAP; otherwise NEITHER. Both default to 0.55.
- `bot.Step` switches on the state; only NEITHER can pause, after the grace period.
- `Runner.Toggle` captures and installs the map reference whenever it starts the bot.
- `config.Load` lets values in `config.json` override defaults, and the file is written with all defaults on first run, so the user's copy pins 0.55.

The window and map columns below come from the first 134 frames of a snowfield recording, against the snowfield references. The menu border column covers all 240 available frames: the 180-frame snowfield recording, a 48-frame grassland recording, and the 12 committed fixtures.

| Frames | window band | whole frame vs map | menu border |
|---|---|---|---|
| Real map | <= 0.117 | >= 0.818 | <= 0.16 |
| Battle, including white flashes | >= 0.485 | <= 0.615 | <= 0.16 |
| Game menu (65 frames) | 0.61-0.67 | ~0.15 | 0.87-0.89 |

The window score bottoms out in battle because the windows are translucent, so a white flash lightens them.

## Goals / Non-Goals

**Goals:**
- No arrow keys during a battle at bright grinding spots.
- The game menu pauses the bot at once.

**Non-Goals:**
- Recognising menu sub-screens (save screen and so on).
- Per-spot battle references or background-independent window matching. Those would be larger changes, to consider only if thresholds stop being enough.

## Decisions

### New thresholds: window 0.40, map 0.70
Each sits between the measured groups. Window 0.40: battle is >= 0.485; real map <= 0.117; transition frames seen at other spots reach 0.36 and 0.446, and either verdict is harmless there, since those frames are fades into or out of a battle. Map 0.70: real map >= 0.818, battle <= 0.615. A fresh map reference on every start is what makes 0.70 safe; the old 0.63-0.69 scores came from a stale reference taken elsewhere. If the party drifts far enough from the start point to score below 0.70, the state becomes NEITHER and the bot pauses, which fails safe. Both remain config values.

### Menu detection: a thin vertical line at a fixed position
The main menu's left panel has a tall white border near x=162 (in 640x480 coordinates) through both the upper screen and the bottom band. The score is the smaller of the "top" (y 10..280) and "band" (y 370..470) values, each computed as: the fraction of bright pixels (all channels >= 190) on the line, minus the larger such fraction 14 pixels to either side. Subtracting the neighbours is what rejects all-white snow and full-screen flashes; requiring both ranges rejects battle status rows. MENU if the best score across x 160..164 is >= 0.5. The small column range covers a border that lands a pixel away at another window scale.

The threshold and geometry are constants in `internal/detect`, not config: they describe the game's fixed layout, not something a user tunes.

Alternatives: a calibrated menu reference (one more manual step, and a translucent menu's colours change with the map behind it); matching the menu text glyphs (more code for the same result).

### MENU is checked first
A menu scores as a battle on the window band, so MENU must win over BATTLE. `StateFor` becomes: menu border -> MENU; window band -> BATTLE; whole frame -> MAP; otherwise NEITHER.

### Pausing and refusing to start
`bot.Step` handles `detect.Menu` by calling the existing `pause` with an alert, like the grace-period stop but with no timer. `Runner.Toggle` classifies the captured frame before installing it; on `Menu` it alerts and leaves the bot PAUSED with its reference unchanged. A classifier with no map reference can still report MENU, because the border check doesn't use references.

### Existing config.json
The new defaults reach only new config files. No migration code for a single local user: updating `config.json` is a task, and the README states the values.

## Risks / Trade-offs

- [Thresholds come from one recording at one spot] -> Regression fixtures pin the snowfield frames, and the existing fixtures cover two other spots; the playtest includes a second spot.
- [The border could land outside x 160..164 at the wide window size; all menu frames so far come from 960x720] -> A playtest at the wide size covers it, and the columns are constants that are easy to widen.
- [A frame captured while the menu is still opening, before its border is drawn, reads as BATTLE, so one Z press is possible] -> Accepted for now; check the opening frames during the playtest. The separate debounce idea would remove it.
- [A sub-screen such as the save screen reads as BATTLE] -> Only reachable through the main menu, where the bot has already paused. Pressing Space there would start the bot on that screen; accepted, since the user started it.
- [Lowering the window threshold makes other blue-window screens read as BATTLE] -> Same as today in practice: town text boxes and event dialogue are already out of scope.
