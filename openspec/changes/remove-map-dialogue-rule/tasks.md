# Tasks

## 1. Regression fixtures (before changing the classifier)

- [x] 1.1 Copy the second spot's map reference (`refs/map.png`) to `internal/detect/testdata/map_spot2.png`, record battles at that spot with `grinder debug record`, and copy into `testdata/battle_spot2.png` the battle frame with the highest above-band score (four recorded battles peaked at 0.533, under 0.55; the two frames that crossed were seen only in `debug classify`); verify by logging that frame's scores (window >= 0.55, above-band the highest recorded)
- [x] 1.2 Add a test that builds a classifier with `map_spot2.png` as the map reference and a map threshold of 0.50 (as if the spot were slightly more map-like, so the current rule misfires on the recorded frame), and asserts `battle_spot2.png` is BATTLE and `map_spot2.png` is MAP; verify it fails on the current code

## 2. Classifier (`internal/detect`)

- [x] 2.1 Remove the above-band rule from `StateFor`, remove `Scores.AboveWindow` and its computation, and update the `Params`/`StateFor` comments; verify the 1.2 test now passes and every existing fixture and synthetic test still passes (`go test ./...`)

## 3. CLI and docs

- [x] 3.1 Drop `aboveWindow` from the `debug classify` output and from the fixture test's log line; verify with `go vet ./...`, `go build`, and a `debug classify` run that prints only window and map scores
- [x] 3.2 Check the README for mentions of the above-band rule or dialogue detection and update them if any; verify with a search for "dialogue" and "above" in README.md

## 4. Integration check

- [x] 4.1 With the real game at the second spot: start with Space and run through at least three encounters; verify the bot clears each battle without pausing
