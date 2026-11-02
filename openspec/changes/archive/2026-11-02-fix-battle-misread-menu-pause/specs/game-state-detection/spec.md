# Spec Delta

## MODIFIED Requirements

### Requirement: Classify each frame
The bot SHALL classify every captured frame as exactly one of MAP, BATTLE, MENU, or NEITHER by comparing it with the saved references and checking for the game menu's layout. A frame showing the game menu MUST be classified MENU, whatever else it matches. Any other frame that matches the battle reference MUST be classified BATTLE regardless of how the rest of the screen compares with the map reference. A frame that matches nothing closely enough MUST be classified NEITHER.

#### Scenario: World map frame
- **WHEN** the captured frame matches the map reference within the configured threshold
- **THEN** it is classified MAP

#### Scenario: Battle frame
- **WHEN** the captured frame matches the battle reference within the configured threshold
- **THEN** it is classified BATTLE

#### Scenario: Battle backdrop that looks like the map
- **WHEN** a battle's backdrop shares its colours with the surrounding world map, so the screen partly resembles the map reference
- **THEN** the frame is still classified BATTLE, and the bot keeps pressing Z

#### Scenario: Bright attack effect on a bright map
- **WHEN** a bright attack effect whitens a battle screen at a grinding spot whose world map is also bright, such as a snowfield
- **THEN** the frame is classified BATTLE, not MAP

#### Scenario: Game menu frame
- **WHEN** the game menu is on screen, whichever map it is opened over
- **THEN** the frame is classified MENU

#### Scenario: White map is not the game menu
- **WHEN** the captured frame is an all-white area of the world map or a full-screen white flash
- **THEN** it is not classified MENU

#### Scenario: Unrecognized frame
- **WHEN** the captured frame matches neither reference, such as the game-over screen, a town, or a black transition frame
- **THEN** it is classified NEITHER
