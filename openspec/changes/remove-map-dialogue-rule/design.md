# Design

## Context

See proposal.md for motivation. `detect.(*Classifier).StateFor` currently checks, in order: window band matches the battle reference -> if the area above the band also matches the map reference, NEITHER ("dialogue on the map"), else BATTLE; whole frame matches the map -> MAP; else NEITHER. `Scores.AboveWindow` exists only for that check and for `debug classify` output. `MapMinSimilarity` serves both the whole-frame map check and the above-band check.

Measured at the second spot (default window size, map reference taken on start there): battle band 0.61-0.81, above-band 0.47-0.58 in battle and 0.75-0.89 on the map, whole-frame map 0.85-1.00 on the map and 0.44-0.53 in battle.

## Goals / Non-Goals

**Goals:**
- Battles are BATTLE whenever the window band matches, at any grinding spot.

**Non-Goals:**
- Detecting dialogue on the world map or in towns.
- Re-tuning thresholds; both stay at 0.55.

## Decisions

### Delete the rule instead of tuning it
`StateFor` becomes: band matches -> BATTLE; whole frame matches the map -> MAP; otherwise NEITHER. `Scores.AboveWindow` and its computation are removed, and the `MapMinSimilarity` comment no longer mentions the area above the band. Alternative considered: a separate, stricter threshold (0.75) for the above-band check; rejected by the user because it protects only an out-of-scope case and its margin rests on one location.

### Order stays battle-first
At the second spot, battle frames scored up to 0.53 on the whole-frame map check, close to the 0.55 threshold. Checking the band first keeps those frames BATTLE even if a future spot pushes the whole-frame score over 0.55.

### Regression fixtures from the second spot
Add `map_spot2.png` (the map reference taken at the second spot, currently `refs/map.png`) and `battle_spot2.png` (a battle frame there whose above-band score is >= 0.55 under the old rule), captured with `grinder debug fixture`. A test builds a classifier with `map_spot2.png` as the map reference and asserts `battle_spot2.png` is BATTLE. Before the change it fails, which shows that the fixture reproduces the bug.

## Risks / Trade-offs

- [A text box over the world map is now BATTLE, so the bot presses Z through it] -> Out of scope by the user's decision. Pressing Z only advances or confirms the text; a choice prompt would take its default.
- [The game menu opened on the map already flickers to BATTLE for a frame] -> Unchanged by this change; the menu covers the screen, so the removed rule never applied to it. Covered by the separate debounce follow-up.
