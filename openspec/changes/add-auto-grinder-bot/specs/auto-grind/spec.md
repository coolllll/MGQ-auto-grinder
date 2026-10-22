# Spec Delta

## Purpose

Plays the EXP grinding loop for the user: walks on the world map to find random encounters, clears battles by pressing Z, and stops as soon as the game shows anything else.

## ADDED Requirements

### Requirement: Start and stop with Space
The bot SHALL start paused and toggle between RUNNING and PAUSED each time the user presses Space. While PAUSED it MUST send no input to the game. Each toggle MUST be reported to the user.

#### Scenario: Start grinding
- **WHEN** the bot is PAUSED and the user presses Space
- **THEN** the bot switches to RUNNING and reports it

#### Scenario: Pause grinding
- **WHEN** the bot is RUNNING and the user presses Space
- **THEN** the bot releases any held keys, switches to PAUSED, and reports it

#### Scenario: Holding Space
- **WHEN** the user holds Space down
- **THEN** the bot toggles once, not repeatedly

### Requirement: Walk on the world map
While RUNNING and the state is MAP, the bot SHALL walk the character by alternately holding the Left and Right arrow keys for a configured duration each, until the state changes.

#### Scenario: Walking to find an encounter
- **WHEN** the bot is RUNNING and the state is MAP
- **THEN** the bot holds Left, then Right, alternating, for the configured walk duration each

#### Scenario: Encounter starts mid-step
- **WHEN** the state changes away from MAP while an arrow key is held
- **THEN** the bot releases the arrow key before doing anything else

### Requirement: Clear battles with Z
While RUNNING and the state is BATTLE, the bot SHALL press and release Z at a configured interval until the state is no longer BATTLE.

#### Scenario: Fighting a battle
- **WHEN** the bot is RUNNING and the state is BATTLE
- **THEN** the bot taps Z repeatedly at the configured interval

#### Scenario: Battle ends and returns to map
- **WHEN** the state changes from BATTLE to MAP
- **THEN** the bot stops pressing Z and resumes walking

### Requirement: Stop on any other screen
When the state has been NEITHER continuously for longer than a configured grace period (default 3 seconds), the bot SHALL switch to PAUSED, release all keys, alert the user, and stay PAUSED until the user presses Space. Shorter NEITHER periods MUST NOT stop the bot. The bot MUST send no input while the state is NEITHER.

#### Scenario: Game over
- **WHEN** the party is defeated and the game-over screen stays on screen past the grace period
- **THEN** the bot pauses, alerts the user, and sends no further input

#### Scenario: Encounter transition
- **WHEN** the screen shows the encounter flash or a fade lasting less than the grace period, then becomes BATTLE
- **THEN** the bot keeps running without pausing

#### Scenario: Dialogue or event appears
- **WHEN** a dialogue, menu, or event scene stays on screen past the grace period
- **THEN** the bot pauses and alerts the user

#### Scenario: Game-over screen is not pressed through
- **WHEN** the state changes from BATTLE to NEITHER because the game-over screen appeared
- **THEN** the bot sends no Z presses, so the game does not advance to the title menu

### Requirement: No input when the game is not in front
The bot SHALL send no keystrokes while the game window is not the foreground window, and MUST release any held keys when focus is lost. Losing focus MUST NOT change RUNNING/PAUSED; input resumes when the game is in front again.

#### Scenario: User switches to another window
- **WHEN** the bot is RUNNING and the user switches to a different window
- **THEN** the bot releases held keys, sends no input, and reports that it is waiting for the game window

#### Scenario: User returns to the game
- **WHEN** the game window becomes the foreground window again while RUNNING
- **THEN** the bot resumes acting on the detected state
