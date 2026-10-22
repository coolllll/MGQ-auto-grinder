// Command grinder plays the EXP grinding loop in Monster Girl Quest! Paradox:
// walk left and right on the map, press Z in battle, stop on anything else.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/coolllll/MGQ-auto-grinder/internal/bot"
	"github.com/coolllll/MGQ-auto-grinder/internal/config"
	"github.com/coolllll/MGQ-auto-grinder/internal/detect"
	"github.com/coolllll/MGQ-auto-grinder/internal/win"
)

const usage = `Usage:
  grinder run                     grind (the default); the toggle key starts and stops
  grinder calibrate map           save the world map reference
  grinder calibrate battle        save the battle reference
  grinder debug window            show the game window the bot would use
  grinder debug capture <file>    save one full-size screenshot of the game
  grinder debug fixture <name>    save a 640x480 test fixture to internal/detect/testdata
  grinder debug record <dir> [s]  save a frame every 0.5 s for s seconds (default 90)
  grinder debug classify        print live detection scores (Ctrl+C to quit)
  grinder debug walk              walk left and right twice, to test input

config.json and refs/ live next to grinder.exe.
`

// Commands that capture or press keys wait up to focusTimeout for the user to
// click into the game, then settle before acting.
const (
	focusTimeout = 30 * time.Second
	settle       = time.Second
)

func main() {
	win.SetDPIAware()

	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"run"}
	}

	code := 0
	var err error
	switch {
	case args[0] == "run":
		err = run()
	case args[0] == "calibrate" && len(args) == 2:
		err = calibrate(args[1])
	case args[0] == "debug" && len(args) >= 2:
		err = debug(args[1], args[2:])
	default:
		fmt.Fprint(os.Stderr, usage)
		code = 2
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		code = 1
	}
	exit(code)
}

// exit keeps a double-clicked console window open so its output can be read.
func exit(code int) {
	if win.OwnsConsole() {
		fmt.Print("\nPress Enter to close this window.")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
	os.Exit(code)
}

func baseDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func refsDir() string { return filepath.Join(baseDir(), "refs") }

func loadConfig() (config.Config, error) {
	path := filepath.Join(baseDir(), "config.json")
	cfg, created, err := config.Load(path)
	if created && err == nil {
		logf("Wrote default settings to %s", path)
	}
	return cfg, err
}

func logf(format string, a ...any) {
	fmt.Printf("%s  %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, a...))
}

func findWindow(cfg config.Config) (win.Window, error) {
	w, err := win.FindGameWindow(cfg.WindowTitle)
	if err != nil {
		return w, fmt.Errorf("%w\nStart the game first, or set windowTitle in config.json", err)
	}
	r, err := w.ClientRect()
	if err != nil {
		return w, err
	}
	logf("Found game window %#x %q (client %dx%d)", w.HWND, w.Title, r.Dx(), r.Dy())
	return w, nil
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	vk, err := win.VKByName(cfg.ToggleKey)
	if err != nil {
		return err
	}
	refs, err := detect.LoadRefs(refsDir())
	if errors.Is(err, detect.ErrMissingRefs) {
		return fmt.Errorf("%w\nRun \"grinder calibrate map\" on the world map and \"grinder calibrate battle\" in a battle first", err)
	}
	if err != nil {
		return err
	}
	w, err := findWindow(cfg)
	if err != nil {
		return err
	}

	b := bot.New(bot.Config{Walk: cfg.Walk(), ZInterval: cfg.ZInterval(), Grace: cfg.Grace()}, keyboard{}, console{})
	runner := &bot.Runner{Bot: b, Win: gameWindow{w}, Cls: detect.NewClassifier(refs, cfg.DetectParams())}

	toggles := make(chan struct{}, 1)
	stop := make(chan struct{})
	defer close(stop)
	go win.WatchKey(vk, 10*time.Millisecond, toggles, stop)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ticker := time.NewTicker(cfg.Tick())
	defer ticker.Stop()

	logf("Paused. Put the character on an open stretch of the world map, then press %s in the game to start.", strings.ToUpper(cfg.ToggleKey))
	var lastErr string
	for {
		select {
		case <-interrupt:
			if b.Mode() == bot.Running {
				b.Toggle() // releases held keys
			}
			logf("Quit.")
			return nil
		case <-toggles:
			b.Toggle()
		case now := <-ticker.C:
			err := runner.Tick(now)
			if errors.Is(err, bot.ErrWindowLost) {
				return err
			}
			// Report capture errors once, not on every tick.
			msg := ""
			if err != nil {
				msg = err.Error()
			}
			if msg != lastErr && msg != "" {
				logf("Capture failed: %s", msg)
			}
			lastErr = msg
		}
	}
}

func calibrate(what string) error {
	var state detect.State
	switch what {
	case "map":
		state = detect.Map
	case "battle":
		state = detect.Battle
	default:
		return fmt.Errorf("calibrate takes \"map\" or \"battle\", not %q", what)
	}
	var img *image.RGBA
	var err error
	if state == detect.Battle {
		img, err = captureNextBattle()
	} else {
		img, err = captureAfterCountdown()
	}
	if err != nil {
		return err
	}
	path, err := detect.SaveRef(refsDir(), state, img)
	if err != nil {
		return err
	}
	logf("Saved the %s reference to %s", strings.ToLower(state.String()), path)
	return nil
}

func captureAfterCountdown() (*image.RGBA, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	w, err := findWindow(cfg)
	if err != nil {
		return nil, err
	}
	if err := waitForGame(w, "capturing"); err != nil {
		return nil, err
	}
	return w.Capture()
}

// captureNextBattle waits for the screen to change away from the calibrated
// map, lets the encounter transition finish, and captures the battle. Timing a
// click to land inside a battle proved unreliable.
func captureNextBattle() (*image.RGBA, error) {
	const (
		battleTimeout = 3 * time.Minute
		transition    = 3 * time.Second
	)
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	mapRef, err := detect.LoadPNG(detect.RefPath(refsDir(), detect.Map))
	if err != nil {
		return nil, fmt.Errorf("calibrate the map first: %w", err)
	}
	w, err := findWindow(cfg)
	if err != nil {
		return nil, err
	}
	if err := waitForGame(w, "watching for a battle"); err != nil {
		return nil, err
	}
	// Every encounter goes map -> fade to black -> battle. Menus and towns
	// don't fade to black from the map, so they can't be mistaken for it.
	const (
		waitMap = iota
		waitBlack
		waitBattle
	)
	// The encounter swirl is neither map nor black, so leaving the map only
	// counts as "not a battle" once it lasts longer than the swirl.
	const swirlMax = 2 * time.Second
	var offMapSince time.Time
	phase := waitMap
	logf("Close any menu and stand on the world map...")
	deadline := time.Now().Add(battleTimeout)
	for phase != waitBattle+1 {
		if time.Now().After(deadline) {
			return nil, errors.New("no battle started in time; try again")
		}
		time.Sleep(100 * time.Millisecond)
		if !w.IsForeground() {
			continue
		}
		img, err := w.Capture()
		if err != nil {
			return nil, err
		}
		onMap := detect.Similarity(img, mapRef) >= cfg.MapMinSimilarity
		dark := detect.DarkFraction(img) > 0.5
		if onMap || dark {
			offMapSince = time.Time{}
		} else if offMapSince.IsZero() {
			offMapSince = time.Now()
		}
		switch {
		case phase == waitMap && onMap:
			logf("On the map. Walk around until a battle starts...")
			phase = waitBlack
		case phase == waitBlack && dark:
			phase = waitBattle
		case phase == waitBlack && !offMapSince.IsZero() && time.Since(offMapSince) > swirlMax:
			logf("Left the map without a battle (menu or town?). Go back to the world map...")
			phase = waitMap
		case phase == waitBattle && !dark:
			phase = waitBattle + 1
		}
	}
	logf("Encounter! Waiting %v for the battle screen to settle...", transition)
	time.Sleep(transition)
	if !w.IsForeground() {
		return nil, errors.New("the game window left the front during the battle; try again")
	}
	img, err := w.Capture()
	if err != nil {
		return nil, err
	}
	if detect.Similarity(img, mapRef) >= cfg.MapMinSimilarity || detect.DarkFraction(img) > 0.5 {
		return nil, errors.New("the screen doesn't look like a battle (still the map, or a black transition); try again")
	}
	return img, nil
}

// waitForGame waits for the user to click into the game, then gives the
// window a moment to redraw before the caller captures or sends keys.
func waitForGame(w win.Window, action string) error {
	logf("Click into the game window; %s %v after it comes to the front...", action, settle)
	deadline := time.Now().Add(focusTimeout)
	for !w.IsForeground() {
		if time.Now().After(deadline) {
			return fmt.Errorf("the game window didn't come to the front within %v", focusTimeout)
		}
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(settle)
	if !w.IsForeground() {
		return errors.New("the game window left the front before " + action + "; try again")
	}
	return nil
}

func debug(what string, args []string) error {
	switch what {
	case "window":
		return debugWindow()
	case "capture":
		if len(args) != 1 {
			return errors.New("debug capture takes one output file name")
		}
		img, err := captureAfterCountdown()
		if err != nil {
			return err
		}
		if err := detect.SavePNG(args[0], img); err != nil {
			return err
		}
		logf("Saved %s", args[0])
		return nil
	case "fixture":
		if len(args) != 1 {
			return errors.New("debug fixture takes one fixture name, such as map_scrolled")
		}
		img, err := captureAfterCountdown()
		if err != nil {
			return err
		}
		path := filepath.Join("internal", "detect", "testdata", args[0]+".png")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := detect.SavePNG(path, detect.Normalize(img)); err != nil {
			return err
		}
		logf("Saved %s", path)
		return nil
	case "record":
		if len(args) < 1 || len(args) > 2 {
			return errors.New("debug record takes an output folder and optionally a number of seconds")
		}
		seconds := 90
		if len(args) == 2 {
			if _, err := fmt.Sscan(args[1], &seconds); err != nil || seconds <= 0 {
				return fmt.Errorf("invalid number of seconds %q", args[1])
			}
		}
		return debugRecord(args[0], time.Duration(seconds)*time.Second)
	case "classify":
		return debugClassify()
	case "walk":
		return debugWalk()
	}
	return fmt.Errorf("unknown debug command %q", what)
}

func debugWindow() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if _, err := findWindow(cfg); err != nil {
		fmt.Println("Visible windows that might be the game:")
		for _, w := range win.ListWindows() {
			if w.Class == win.GameClass || strings.Contains(strings.ToLower(w.Title), strings.ToLower(cfg.WindowTitle)) {
				fmt.Printf("  %#x  class=%q  title=%q\n", w.HWND, w.Class, w.Title)
			}
		}
		return err
	}
	return nil
}

// debugRecord saves a normalised frame every half second while the game is in
// front, so a whole encounter can be studied afterwards.
func debugRecord(dir string, length time.Duration) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	w, err := findWindow(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := waitForGame(w, "recording"); err != nil {
		return err
	}
	logf("Recording for %v. Play normally; frames are skipped while the game isn't in front.", length)
	start := time.Now()
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	saved := 0
	for now := range t.C {
		if now.Sub(start) >= length {
			break
		}
		if !w.Exists() {
			return errors.New("the game window was closed")
		}
		if !w.IsForeground() {
			continue
		}
		img, err := w.Capture()
		if err != nil {
			logf("capture failed: %v", err)
			continue
		}
		name := filepath.Join(dir, fmt.Sprintf("frame_%04d_%06dms.png", saved, now.Sub(start).Milliseconds()))
		if err := detect.SavePNG(name, detect.Normalize(img)); err != nil {
			return err
		}
		saved++
	}
	logf("Saved %d frames to %s", saved, dir)
	return nil
}

func debugClassify() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	refs, err := detect.LoadRefs(refsDir())
	if err != nil {
		return err
	}
	w, err := findWindow(cfg)
	if err != nil {
		return err
	}
	cls := detect.NewClassifier(refs, cfg.DetectParams())
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-interrupt:
			return nil
		case <-t.C:
			img, err := w.Capture()
			if err != nil {
				logf("capture failed: %v", err)
				continue
			}
			s := cls.Score(img)
			logf("%-7v window=%.3f (min %.2f)  map=%.3f  aboveWindow=%.3f (min %.2f)",
				cls.StateFor(s), s.Window, cfg.WindowMinSimilarity, s.Map, s.AboveWindow, cfg.MapMinSimilarity)
		}
	}
}

func debugWalk() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	w, err := findWindow(cfg)
	if err != nil {
		return err
	}
	if err := waitForGame(w, "walking"); err != nil {
		return err
	}
	kb := keyboard{}
	for i := 0; i < 2; i++ {
		for _, k := range []bot.Key{bot.KeyLeft, bot.KeyRight} {
			if !w.IsForeground() {
				return errors.New("the game window is not in front; stopped")
			}
			logf("Holding %v", k)
			kb.Down(k)
			time.Sleep(cfg.Walk())
			kb.Up(k)
		}
	}
	logf("Done.")
	return nil
}

// keyboard sends bot keys to the game as scan codes.
type keyboard struct{}

var scans = map[bot.Key]win.Scan{bot.KeyZ: win.ScanZ, bot.KeyLeft: win.ScanLeft, bot.KeyRight: win.ScanRight}

func (keyboard) Down(k bot.Key) { report(win.KeyDown(scans[k])) }
func (keyboard) Up(k bot.Key)   { report(win.KeyUp(scans[k])) }

func report(err error) {
	if err != nil {
		logf("Input failed: %v", err)
	}
}

// console prints bot messages; alerts also beep.
type console struct{}

func (console) Info(msg string) { logf("%s", msg) }
func (console) Alert(msg string) {
	logf("!! %s", msg)
	win.Beep()
}

// gameWindow adapts win.Window to bot.Window.
type gameWindow struct{ win.Window }

func (g gameWindow) Capture() (image.Image, error) {
	img, err := g.Window.Capture()
	if err != nil {
		return nil, err
	}
	return img, nil
}
