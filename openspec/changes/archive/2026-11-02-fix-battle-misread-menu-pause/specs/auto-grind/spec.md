# Spec Delta

## ADDED Requirements

### Requirement: Pause on the game menu
When the state is MENU, the bot SHALL immediately switch to PAUSED, release all keys, alert the user, and stay PAUSED until the user presses Space, without waiting for the grace period. The bot MUST NOT start while the game menu is showing.

#### Scenario: Menu opened while running
- **WHEN** the bot is RUNNING and the user opens the game menu
- **THEN** the bot releases held keys, pauses at once, alerts the user, and sends no further input

#### Scenario: Space pressed while the menu is showing
- **WHEN** the bot is PAUSED, the game menu is showing, and the user presses Space
- **THEN** the bot stays PAUSED, keeps its current map reference, and reports that the menu must be closed first

## MODIFIED Requirements

### Requirement: Stop on any other screen
When the state has been NEITHER continuously for longer than a configured grace period (default 3 seconds), the bot SHALL switch to PAUSED, release all keys, alert the user, and stay PAUSED until the user presses Space. Shorter NEITHER periods MUST NOT stop the bot. The bot MUST send no input while the state is NEITHER.

#### Scenario: Game over
- **WHEN** the party is defeated and the game-over screen stays on screen past the grace period
- **THEN** the bot pauses, alerts the user, and sends no further input

#### Scenario: Encounter transition
- **WHEN** the screen shows the encounter flash or a fade lasting less than the grace period, then becomes BATTLE
- **THEN** the bot keeps running without pausing

#### Scenario: Another screen appears
- **WHEN** a screen that is neither the world map, a battle, nor the game menu, such as a town, stays on screen past the grace period
- **THEN** the bot pauses and alerts the user

#### Scenario: Game-over screen is not pressed through
- **WHEN** the state changes from BATTLE to NEITHER because the game-over screen appeared
- **THEN** the bot sends no Z presses, so the game does not advance to the title menu
