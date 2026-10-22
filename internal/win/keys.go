package win

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

var vkByName = map[string]int{
	"space":      0x20,
	"pause":      0x13,
	"scrolllock": 0x91,
	"f8":         0x77,
	"f9":         0x78,
}

// VKByName returns the virtual-key code for a toggle key name from the config.
func VKByName(name string) (int, error) {
	if vk, ok := vkByName[strings.ToLower(name)]; ok {
		return vk, nil
	}
	names := make([]string, 0, len(vkByName))
	for n := range vkByName {
		names = append(names, n)
	}
	sort.Strings(names)
	return 0, fmt.Errorf("unknown toggle key %q (use one of: %s)", name, strings.Join(names, ", "))
}

// KeyIsDown reports whether a key is physically held right now.
func KeyIsDown(vk int) bool {
	state, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return state&0x8000 != 0
}

// EdgeDetector turns a stream of "is the key down" samples into presses, so a
// held key counts once.
type EdgeDetector struct {
	prev bool
}

// Pressed reports whether this sample is the transition from up to down.
func (e *EdgeDetector) Pressed(down bool) bool {
	pressed := down && !e.prev
	e.prev = down
	return pressed
}

// WatchKey polls vk every interval and sends on presses each time the key goes
// down, until stop is closed.
func WatchKey(vk int, interval time.Duration, presses chan<- struct{}, stop <-chan struct{}) {
	var edge EdgeDetector
	edge.Pressed(KeyIsDown(vk)) // a key already held at start is not a press
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			if edge.Pressed(KeyIsDown(vk)) {
				select {
				case presses <- struct{}{}:
				default: // previous press not handled yet; drop this one
				}
			}
		}
	}
}
