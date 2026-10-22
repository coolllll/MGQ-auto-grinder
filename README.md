# MGQ-auto-grinder

Auto player for grinding EXP in *Monster Girl Quest! Paradox RPG* (RPG Maker VX Ace) on Windows.

It watches the game window and:

- **on the world map**, walks the character left and right to trigger random encounters;
- **in battle**, presses Z until the battle (including the victory/EXP text) is over;
- **on anything else** (game over, a town, menus, a crash dialog) for more than 3 seconds, stops and beeps.

It only reads the screen and presses keys. It doesn't touch the game's files, memory or saves.

## Build

1. Install Go 1.22 or newer: `winget install GoLang.Go`
2. In this folder:

   ```
   go build -o grinder.exe ./cmd/grinder
   ```

`grinder.exe` keeps its settings (`config.json`) and calibration images (`refs/`) next to itself.

## Calibrate (once)

The bot recognises battles by comparing them with a screenshot of one battle, which you take once. With the game running in a window:

1. Stand on the world map with no menu open and run `.\grinder.exe calibrate battle`.
2. Click into the game (within 30 seconds), then walk around until a battle starts. The command saves the battle screen by itself once the encounter has faded in.

You don't calibrate the map: each time you press Space to start the bot, the screen you're on becomes the map reference. So start the bot where you want to grind, and it adapts to any area and to either window size (`WindowedRatio=1` or `2` in `Game.ini`).

Run `calibrate battle` again only if battles stop being recognised. `calibrate map` still exists for troubleshooting with `debug classify`.

## Grind

1. Put the character on an open stretch of the world map, with room to walk left and right.
2. Run `.\grinder.exe run`. It starts **paused**.
3. Click into the game and press **Space** to start; the bot takes the current screen as the map reference. Press **Space** again to pause, and again to start, which retakes the map reference wherever you are. Space only counts while the game window is in front, so typing in other windows won't start or stop the bot.

Keep the game window in front: Windows only delivers the keys to the window you're using. If you switch to another window, the bot stops sending keys until the game is back in front.

Note that Space is also the game's confirm key, so each start/stop also confirms whatever the game has selected. To use another key, set `toggleKey` in `config.json` to `f8`, `f9`, `pause` or `scrolllock`.

## When it stops by itself

The bot pauses, beeps and prints why when:

- the screen has been neither the map nor a battle for 3 seconds (game over, a town, a menu, an error dialog). It sends no keys on these screens, so it never presses through the game-over screen to the title menu;
- the game window is closed.

Press Space to resume once you've dealt with it. Ctrl+C in the console quits.

## Settings

`config.json` is created with defaults on first run:

| Setting | Default | Meaning |
|---|---|---|
| `windowTitle` | `Monster Girl Quest` | Part of the game window's title |
| `toggleKey` | `space` | Start/stop key: `space`, `f8`, `f9`, `pause`, `scrolllock` |
| `tickMs` | `50` | How often the screen is checked |
| `walkMs` | `600` | How long to walk in each direction |
| `zIntervalMs` | `100` | Time between Z presses in battle |
| `graceMs` | `3000` | How long an unrecognised screen may last before stopping |
| `windowBand` | bottom 112 px | Strip where battles always show their blue windows (640x480 game pixels) |
| `windowMinSimilarity` | `0.55` | How closely that strip must match the battle reference (0-1, higher = stricter) |
| `mapMinSimilarity` | `0.55` | How closely the screen must match the map reference (0-1, higher = stricter) |

## Troubleshooting

- `.\grinder.exe debug window` shows which game window the bot finds.
- `.\grinder.exe debug classify` prints the detection scores twice a second. Use it to see why a screen is or isn't recognised, and to tune `windowMinSimilarity` and `mapMinSimilarity`.
- `.\grinder.exe debug walk` walks left and right twice, to check that the game accepts the bot's keys.
- `.\grinder.exe debug capture shot.png` saves a screenshot the way the bot sees it.
