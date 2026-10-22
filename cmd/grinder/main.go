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
  grinder debug capture <file>    save one normalised screenshot of the game
  grinder debug classify          print live detection scores (Ctrl+C to quit)
  grinder debug walk              walk left and right twice, to test input

config.json and refs/ live next to grinder.exe.
`

// countdown is how long capture commands wait so the user can click into the
// game first.
const countdown = 3 * time.Second

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
	img, err := captureAfterCountdown()
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
	logf("Click into the game now; capturing in %v...", countdown)
	time.Sleep(countdown)
	if !w.IsForeground() {
		return nil, errors.New("the game window was not in front, so the capture could show something else; try again")
	}
	return w.Capture()
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
		if err := detect.SavePNG(args[0], detect.Normalize(img)); err != nil {
			return err
		}
		logf("Saved %s", args[0])
		return nil
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
			logf("%-7v battleDiff=%6.2f (max %.2f)  mapSimilarity=%.3f (min %.3f)",
				cls.StateFor(s), s.BattleDiff, cfg.BattleMaxDiff, s.MapSimilarity, cfg.MapMinSimilarity)
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
	logf("Click into the game now; walking in %v...", countdown)
	time.Sleep(countdown)
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
