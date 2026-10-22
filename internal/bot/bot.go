// Package bot is the grinding state machine: walk on the map, press Z in
// battle, stop on anything else. It talks to the game only through the
// Keyboard and Notifier interfaces, so it can be tested without Windows.
package bot

import (
	"fmt"
	"time"

	"github.com/coolllll/MGQ-auto-grinder/internal/detect"
)

// Mode is whether the bot is allowed to send input.
type Mode int

const (
	Paused Mode = iota
	Running
)

func (m Mode) String() string {
	if m == Running {
		return "RUNNING"
	}
	return "PAUSED"
}

// Key is a game key the bot presses.
type Key int

const (
	KeyZ Key = iota
	KeyLeft
	KeyRight
)

func (k Key) String() string {
	return [...]string{"Z", "Left", "Right"}[k]
}

// Keyboard sends key events to the game.
type Keyboard interface {
	Down(Key)
	Up(Key)
}

// Notifier tells the user what the bot is doing. Alert is for anything that
// needs the user's attention, such as the bot stopping by itself.
type Notifier interface {
	Info(msg string)
	Alert(msg string)
}

// Config holds the bot's timings.
type Config struct {
	Walk      time.Duration // how long to hold each arrow key on the map
	ZInterval time.Duration // time between Z presses in battle
	Grace     time.Duration // how long NEITHER may last before stopping
}

// Observation is what the bot sees on one tick.
type Observation struct {
	State   detect.State
	Focused bool // the game window is in front
}

// Bot is the grinding state machine. Call Toggle on each press of the toggle
// key and Step on every tick.
type Bot struct {
	cfg Config
	kb  Keyboard
	n   Notifier

	mode    Mode
	held    [3]bool // indexed by Key
	focused bool

	hasState   bool
	state      detect.State
	stateSince time.Time

	walkDir   Key
	walkSince time.Time
	zDownAt   time.Time
}

// New returns a paused bot.
func New(cfg Config, kb Keyboard, n Notifier) *Bot {
	return &Bot{cfg: cfg, kb: kb, n: n, focused: true, walkDir: KeyLeft}
}

// Mode returns whether the bot is running.
func (b *Bot) Mode() Mode { return b.mode }

// Toggle switches between running and paused.
func (b *Bot) Toggle() {
	if b.mode == Running {
		b.pause("Paused.", false)
		return
	}
	b.mode = Running
	b.focused = true
	b.hasState = false
	b.n.Info("Running.")
}

// WindowLost stops the bot because the game window is gone.
func (b *Bot) WindowLost() {
	b.pause("Stopped: the game window was closed.", true)
}

// Step acts on one observation.
func (b *Bot) Step(now time.Time, obs Observation) {
	if b.mode != Running {
		return
	}

	if !obs.Focused {
		if b.focused {
			b.releaseAll()
			b.focused = false
			b.n.Info("The game window is not in front; waiting for it.")
		}
		return
	}
	if !b.focused {
		b.focused = true
		b.hasState = false
		b.n.Info("The game window is in front again.")
	}

	if !b.hasState || obs.State != b.state {
		b.releaseAll()
		b.hasState = true
		b.state = obs.State
		b.stateSince = now
		b.zDownAt = time.Time{}
		b.n.Info("State: " + obs.State.String())
	}

	switch b.state {
	case detect.Map:
		b.walk(now)
	case detect.Battle:
		b.mashZ(now)
	case detect.Neither:
		// No input at all: pressing Z on the game-over screen would go on
		// to the title menu.
		if now.Sub(b.stateSince) >= b.cfg.Grace {
			b.pause(fmt.Sprintf("Stopped: the screen was neither map nor battle for %v. Press the toggle key to resume.", b.cfg.Grace), true)
		}
	}
}

// walk alternates between holding Left and Right.
func (b *Bot) walk(now time.Time) {
	if !b.held[b.walkDir] {
		b.down(b.walkDir)
		b.walkSince = now
		return
	}
	if now.Sub(b.walkSince) >= b.cfg.Walk {
		b.up(b.walkDir)
		if b.walkDir == KeyLeft {
			b.walkDir = KeyRight
		} else {
			b.walkDir = KeyLeft
		}
		b.down(b.walkDir)
		b.walkSince = now
	}
}

// mashZ presses Z once per interval, holding it for half the interval so the
// game, which samples the keyboard once per frame, always sees the press.
func (b *Bot) mashZ(now time.Time) {
	if b.held[KeyZ] {
		if now.Sub(b.zDownAt) >= b.cfg.ZInterval/2 {
			b.up(KeyZ)
		}
		return
	}
	if b.zDownAt.IsZero() || now.Sub(b.zDownAt) >= b.cfg.ZInterval {
		b.down(KeyZ)
		b.zDownAt = now
	}
}

func (b *Bot) pause(msg string, alert bool) {
	b.releaseAll()
	b.mode = Paused
	b.hasState = false
	if alert {
		b.n.Alert(msg)
	} else {
		b.n.Info(msg)
	}
}

func (b *Bot) down(k Key) {
	b.kb.Down(k)
	b.held[k] = true
}

func (b *Bot) up(k Key) {
	b.kb.Up(k)
	b.held[k] = false
}

func (b *Bot) releaseAll() {
	for k := range b.held {
		if b.held[k] {
			b.up(Key(k))
		}
	}
}
