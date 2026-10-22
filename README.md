# MGQ-auto-grinder

Auto player for grinding EXP in *Monster Girl Quest! Paradox RPG* (RPG Maker VX Ace) on Windows.

It watches the game window and:

- **on the world map**, walks the character left and right to trigger random encounters;
- **in battle**, presses Z until the battle (including the victory/EXP text) is over;
- **on anything else** (game over, dialogue, menus, a crash dialog) for more than 3 seconds, stops and beeps.

It only reads the screen and presses keys. It doesn't touch the game's files, memory or saves.

## Build

1. Install Go 1.22 or newer: `winget install GoLang.Go`
2. In this folder:

   ```
   go build -o grinder.exe ./cmd/grinder
   ```

`grinder.exe` keeps its settings (`config.json`) and calibration images (`refs/`) next to itself.

## Calibrate (once)

The bot recognises the map and battles by comparing them with two screenshots you take. With the game running in a window:

1. Stand on the world map where you want to grind, run `.\grinder.exe calibrate map`, then click into the game within 3 seconds.
2. Start a battle and wait for the command menu, run `.\grinder.exe calibrate battle`, then click into the game within 3 seconds.

Recalibrate if you grind on a different map, or if the bot stops recognising a screen.

## Grind

1. Put the character on an open stretch of the world map, with room to walk left and right.
2. Run `.\grinder.exe run`. It starts **paused**.
3. Click into the game and press **Space** to start. Press **Space** again to pause.

Keep the game window in front: Windows only delivers the keys to the window you're using. If you switch to another window, the bot stops sending keys until the game is back in front.

Note that Space is also the game's confirm key, so each start/stop also confirms whatever the game has selected. To use another key, set `toggleKey` in `config.json` to `f8`, `f9`, `pause` or `scrolllock`.

## When it stops by itself

The bot pauses, beeps and prints why when:

- the screen has been neither the map nor a battle for 3 seconds (game over, a dialogue or event, a menu, an error dialog). It sends no keys on these screens, so it never presses through the game-over screen to the title menu;
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
| `battleRegion` | bottom 128 px | Part of the battle screen that must match the battle reference |
| `battleMaxDiff` | `20` | Lower = stricter battle match |
| `mapMinSimilarity` | `0.8` | Higher = stricter map match |

## Troubleshooting

- `.\grinder.exe debug window` shows which game window the bot finds.
- `.\grinder.exe debug classify` prints the detection scores twice a second. Use it to see why a screen is or isn't recognised, and to tune `battleMaxDiff` and `mapMinSimilarity`.
- `.\grinder.exe debug walk` walks left and right twice, to check that the game accepts the bot's keys.
- `.\grinder.exe debug capture shot.png` saves a screenshot the way the bot sees it.
