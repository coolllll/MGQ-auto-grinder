# Spec Delta

## MODIFIED Requirements

### Requirement: Classify each frame
The bot SHALL classify every captured frame as exactly one of MAP, BATTLE, or NEITHER by comparing it with the saved references. A frame that matches the battle reference MUST be classified BATTLE regardless of how the rest of the screen compares with the map reference. A frame that does not match either reference closely enough MUST be classified NEITHER.

#### Scenario: World map frame
- **WHEN** the captured frame matches the map reference within the configured threshold
- **THEN** it is classified MAP

#### Scenario: Battle frame
- **WHEN** the captured frame matches the battle reference within the configured threshold
- **THEN** it is classified BATTLE

#### Scenario: Battle backdrop that looks like the map
- **WHEN** a battle's backdrop shares its colours with the surrounding world map, so the screen partly resembles the map reference
- **THEN** the frame is still classified BATTLE, and the bot keeps pressing Z

#### Scenario: Unrecognized frame
- **WHEN** the captured frame matches neither reference, such as the game-over screen, a town, or a black transition frame
- **THEN** it is classified NEITHER
