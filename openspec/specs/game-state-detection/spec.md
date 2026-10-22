# game-state-detection Specification

## Purpose

Determines what the game is currently showing — world map, battle, or something else — using only screen captures of the game window, so the grinder can act on it.

## Requirements

### Requirement: Locate the game window
The bot SHALL find the running game window by its title and capture only that window's client area. If the window cannot be found, the bot MUST report this to the user and send no input.

#### Scenario: Game is running
- **WHEN** the bot starts while the game window is open
- **THEN** the bot reports that it found the game window and its client size

#### Scenario: Game is not running
- **WHEN** the bot starts and no game window matches
- **THEN** the bot prints an error naming the expected window title and sends no keystrokes

#### Scenario: Game window closes while running
- **WHEN** the game window disappears while the bot is running
- **THEN** the bot stops and reports that the window was lost

### Requirement: Calibrate reference images
The bot SHALL capture the map reference from the live game window each time the grinder is started, and SHALL provide a calibration command with which the user captures one reference of a battle. The battle reference MUST be saved to disk and reused on later runs without recalibrating. The map reference MUST also be saved to disk, but a run MUST NOT depend on a previously saved one. A map calibration command MAY remain for troubleshooting.

#### Scenario: Capture a map reference
- **WHEN** the user starts the grinder while standing on the world map
- **THEN** the bot saves the current screen as the map reference and classifies later frames against it

#### Scenario: Either window size
- **WHEN** the game runs at the default window size (960x720 on the user's display) or the wide size (1920x1440)
- **THEN** a map reference taken on start at that size classifies map frames as MAP, with no separate calibration step

#### Scenario: Capture a battle reference
- **WHEN** the user runs the battle calibration command and an encounter starts from the world map (map, then fade to black, then battle)
- **THEN** the bot saves a battle reference image once the battle screen has appeared and confirms it was saved

#### Scenario: Menu or town during battle calibration
- **WHEN** the screen leaves the world map without fading to black, such as when a menu opens or the party enters a town
- **THEN** the bot saves nothing and keeps waiting for an encounter

#### Scenario: Missing references
- **WHEN** the user starts grinding without a saved battle reference
- **THEN** the bot refuses to run and tells the user to run battle calibration first; a missing saved map reference does not stop it, because the map reference is taken when the user presses Space

### Requirement: Classify each frame
The bot SHALL classify every captured frame as exactly one of MAP, BATTLE, or NEITHER by comparing it with the saved references. A frame that does not match either reference closely enough MUST be classified NEITHER.

#### Scenario: World map frame
- **WHEN** the captured frame matches the map reference within the configured threshold
- **THEN** it is classified MAP

#### Scenario: Battle frame
- **WHEN** the captured frame matches the battle reference within the configured threshold
- **THEN** it is classified BATTLE

#### Scenario: Unrecognized frame
- **WHEN** the captured frame matches neither reference, such as the game-over screen, a town, or a black transition frame
- **THEN** it is classified NEITHER

### Requirement: Battle detection covers victory messages
The battle classification SHALL keep classifying a battle as BATTLE while post-battle victory, EXP, item drop, and level-up messages are on screen, so the grinder can clear them.

#### Scenario: Victory text after the last enemy falls
- **WHEN** the last enemy is defeated and the victory/EXP message is shown over the battle screen
- **THEN** the frame is still classified BATTLE
