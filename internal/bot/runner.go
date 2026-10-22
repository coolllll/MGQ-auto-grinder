package bot

import (
	"errors"
	"image"
	"time"

	"github.com/coolllll/MGQ-auto-grinder/internal/detect"
)

// ErrWindowLost is returned by Runner.Tick once the game window is gone.
var ErrWindowLost = errors.New("the game window was closed")

// Window is the game window as the runner needs it.
type Window interface {
	Exists() bool
	IsForeground() bool
	Capture() (image.Image, error)
}

// Classifier turns a captured frame into a state.
type Classifier interface {
	Classify(image.Image) detect.State
	SetMapReference(image.Image)
}

// Runner feeds the bot one observation per tick.
type Runner struct {
	Bot *Bot
	Win Window
	Cls Classifier
	// OnMapReference, if set, receives each map reference taken on start,
	// for example to save it to disk.
	OnMapReference func(image.Image)
}

// Toggle starts or pauses the bot. Starting first captures the current screen
// as the map reference, since the user starts the bot while standing on the
// map; if the capture fails, the bot stays paused.
func (r *Runner) Toggle() {
	if r.Bot.Mode() == Running {
		r.Bot.Toggle()
		return
	}
	img, err := r.Win.Capture()
	if err != nil {
		r.Bot.n.Alert("Not started: couldn't capture the map reference: " + err.Error())
		return
	}
	r.Cls.SetMapReference(img)
	if r.OnMapReference != nil {
		r.OnMapReference(img)
	}
	r.Bot.n.Info("Map reference captured.")
	r.Bot.Toggle()
}

// Tick checks the window, captures and classifies a frame, and steps the bot.
// A failed capture counts as NEITHER, so it can never cause input; the error
// is returned for the caller to report.
func (r *Runner) Tick(now time.Time) error {
	if !r.Win.Exists() {
		r.Bot.WindowLost()
		return ErrWindowLost
	}
	if r.Bot.Mode() != Running {
		return nil
	}
	if !r.Win.IsForeground() {
		r.Bot.Step(now, Observation{Focused: false})
		return nil
	}
	img, err := r.Win.Capture()
	state := detect.Neither
	if err == nil {
		state = r.Cls.Classify(img)
	}
	r.Bot.Step(now, Observation{State: state, Focused: true})
	return err
}
