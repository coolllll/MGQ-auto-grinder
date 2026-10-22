# Proposal

## Why

At a second grinding spot the bot stopped in the middle of battles. `debug classify` there showed why: battle frames matched the battle window band well (0.61-0.81), but the battle backdrop above the band shares its colours with the surrounding world map, so "area above the band vs the map reference" scored 0.47-0.58. That is right on the 0.55 threshold, and every frame over it was classified NEITHER by the rule that treats "battle-style window band with the map above it" as a dialogue on the map. Once those NEITHER frames last 3 seconds, the bot pauses. The rule only exists to catch dialogue on the world map, which the user put out of scope, so it adds risk without serving a requirement.

## What Changes

- A frame whose bottom window band matches the battle reference is classified BATTLE, whatever is above the band. The "map above the window band means dialogue on the map" rule is removed.
- The above-the-band score is no longer computed or shown in `debug classify`.
- **Accepted consequence**: a text box shown over the world map now counts as BATTLE, so the bot presses Z through it. Dialogue on the world map stays out of scope.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `game-state-detection`: "Classify each frame" states that a frame matching the battle reference is BATTLE even when its background looks like the map.

## Impact

- `internal/detect/detect.go`: `StateFor`, `Scores`, `Score`, and the `Params` comments.
- `internal/detect/detect_test.go`: fixture test log line, plus a new fixture pair from the second spot (its map reference and a battle frame whose backdrop resembles it).
- `cmd/grinder/main.go`: `debug classify` output.
- No config, CLI or dependency changes.
