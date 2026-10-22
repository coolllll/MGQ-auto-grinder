# Tasks

## 1. Regression fixtures (before changing the classifier)

- [ ] 1.1 Copy the second spot's map reference (`refs/map.png`) to `internal/detect/testdata/map_spot2.png`, record one battle at that spot with `grinder debug record`, and copy into `testdata/battle_spot2.png` a frame that the current classifier, using `map_spot2.png` as the map reference, calls NEITHER with a window score >= 0.55; verify by logging that frame's scores (window >= 0.55, above-band >= 0.55, verdict NEITHER)
- [ ] 1.2 Add a test that builds a classifier with `map_spot2.png` as the map reference and asserts `battle_spot2.png` is BATTLE and `map_spot2.png` is MAP; verify it fails on the current code

## 2. Classifier (`internal/detect`)

- [ ] 2.1 Remove the above-band rule from `StateFor`, remove `Scores.AboveWindow` and its computation, and update the `Params`/`StateFor` comments; verify the 1.2 test now passes and every existing fixture and synthetic test still passes (`go test ./...`)

## 3. CLI and docs

- [ ] 3.1 Drop `aboveWindow` from the `debug classify` output and from the fixture test's log line; verify with `go vet ./...`, `go build`, and a `debug classify` run that prints only window and map scores
- [ ] 3.2 Check the README for mentions of the above-band rule or dialogue detection and update them if any; verify with a search for "dialogue" and "above" in README.md

## 4. Integration check

- [ ] 4.1 With the real game at the second spot: start with Space and run through at least three encounters; verify the bot clears each battle without pausing
