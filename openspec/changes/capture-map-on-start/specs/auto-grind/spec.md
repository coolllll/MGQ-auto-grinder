# Spec Delta

## MODIFIED Requirements

### Requirement: Start and stop with Space
The bot SHALL start paused and toggle between RUNNING and PAUSED each time the user presses Space while the game window is in front. Presses while another window is in front MUST be ignored. Each switch to RUNNING MUST first capture the current game screen as the map reference; if that capture fails, the bot MUST stay PAUSED. While PAUSED it MUST send no input to the game. Each toggle MUST be reported to the user.

#### Scenario: Start grinding
- **WHEN** the bot is PAUSED and the user presses Space
- **THEN** the bot captures the current screen as the map reference, switches to RUNNING, and reports both

#### Scenario: Pause grinding
- **WHEN** the bot is RUNNING and the user presses Space
- **THEN** the bot releases any held keys, switches to PAUSED, and reports it, keeping the current map reference

#### Scenario: Restart at another spot
- **WHEN** the user pauses the bot, moves to another part of the world map, and presses Space again
- **THEN** the bot uses a map reference captured at the new spot

#### Scenario: Capture fails on start
- **WHEN** the user presses Space to start and the game screen cannot be captured
- **THEN** the bot stays PAUSED, sends no input, and reports why it did not start

#### Scenario: Holding Space
- **WHEN** the user holds Space down
- **THEN** the bot toggles once, not repeatedly

#### Scenario: Space typed in another window
- **WHEN** the user presses Space while another window is in front, for example while typing in a chat
- **THEN** the bot ignores it and stays RUNNING or PAUSED as it was
