// Package config loads the grinder's settings from config.json.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io/fs"
	"os"
	"time"

	"github.com/coolllll/MGQ-auto-grinder/internal/detect"
)

// Rect is a rectangle in the game's native 544x416 coordinates.
type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// Config is everything a user can tune. Durations are in milliseconds.
type Config struct {
	WindowTitle         string  `json:"windowTitle"`
	ToggleKey           string  `json:"toggleKey"`
	TickMs              int     `json:"tickMs"`
	WalkMs              int     `json:"walkMs"`
	ZIntervalMs         int     `json:"zIntervalMs"`
	GraceMs             int     `json:"graceMs"`
	WindowBand          Rect    `json:"windowBand"`
	WindowMinSimilarity float64 `json:"windowMinSimilarity"`
	MapMinSimilarity    float64 `json:"mapMinSimilarity"`
}

// Default returns the built-in settings.
func Default() Config {
	p := detect.DefaultParams()
	r := p.WindowBand
	return Config{
		WindowTitle:         "Monster Girl Quest",
		ToggleKey:           "space",
		TickMs:              50,
		WalkMs:              600,
		ZIntervalMs:         100,
		GraceMs:             3000,
		WindowBand:          Rect{X: r.Min.X, Y: r.Min.Y, W: r.Dx(), H: r.Dy()},
		WindowMinSimilarity: p.WindowMinSimilarity,
		MapMinSimilarity:    p.MapMinSimilarity,
	}
}

// Load reads path. Settings missing from the file keep their defaults. If the
// file doesn't exist it is created with the defaults, and created is true.
func Load(path string) (cfg Config, created bool, err error) {
	cfg = Default()
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return cfg, true, Save(path, cfg)
	}
	if err != nil {
		return cfg, false, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, false, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, false, cfg.Validate()
}

// Save writes cfg to path as indented JSON.
func Save(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// Validate rejects settings the bot can't run with.
func (c Config) Validate() error {
	positive := map[string]int{
		"tickMs": c.TickMs, "walkMs": c.WalkMs, "zIntervalMs": c.ZIntervalMs, "graceMs": c.GraceMs,
		"windowBand.w": c.WindowBand.W, "windowBand.h": c.WindowBand.H,
	}
	for name, v := range positive {
		if v <= 0 {
			return fmt.Errorf("config: %s must be greater than 0, got %d", name, v)
		}
	}
	return nil
}

// DetectParams converts the detection settings for the classifier.
func (c Config) DetectParams() detect.Params {
	r := c.WindowBand
	return detect.Params{
		WindowBand:          image.Rect(r.X, r.Y, r.X+r.W, r.Y+r.H),
		WindowMinSimilarity: c.WindowMinSimilarity,
		MapMinSimilarity:    c.MapMinSimilarity,
	}
}

func ms(n int) time.Duration { return time.Duration(n) * time.Millisecond }

func (c Config) Tick() time.Duration      { return ms(c.TickMs) }
func (c Config) Walk() time.Duration      { return ms(c.WalkMs) }
func (c Config) ZInterval() time.Duration { return ms(c.ZIntervalMs) }
func (c Config) Grace() time.Duration     { return ms(c.GraceMs) }
