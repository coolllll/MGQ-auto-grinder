# Spec Delta

## MODIFIED Requirements

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
