// Package detect classifies game frames as world map, battle, or neither by
// comparing them with calibration references.
package detect

import (
	"image"
	"image/draw"
)

// NativeW and NativeH are the game's internal resolution (MGQ Paradox runs
// VX Ace at 640x480 rather than the stock 544x416). Every frame and reference
// is normalised to this size, so window scaling doesn't matter.
const (
	NativeW = 640
	NativeH = 480
)

// State is what the game is showing.
type State int

const (
	Neither State = iota
	Map
	Battle
	Menu
)

func (s State) String() string {
	switch s {
	case Map:
		return "MAP"
	case Battle:
		return "BATTLE"
	case Menu:
		return "MENU"
	default:
		return "NEITHER"
	}
}

// Normalize resizes img to NativeW x NativeH with nearest-neighbour sampling.
// The game's window ratios are integer multiples, so this is exact.
func Normalize(img image.Image) *image.RGBA {
	b := img.Bounds()
	if rgba, ok := img.(*image.RGBA); ok && b == image.Rect(0, 0, NativeW, NativeH) {
		return rgba
	}
	src, ok := img.(*image.RGBA)
	if !ok {
		src = image.NewRGBA(b)
		draw.Draw(src, b, img, b.Min, draw.Src)
	}
	dst := image.NewRGBA(image.Rect(0, 0, NativeW, NativeH))
	for y := 0; y < NativeH; y++ {
		sy := b.Min.Y + y*b.Dy()/NativeH
		for x := 0; x < NativeW; x++ {
			sx := b.Min.X + x*b.Dx()/NativeW
			i := src.PixOffset(sx, sy)
			copy(dst.Pix[dst.PixOffset(x, y):][:4], src.Pix[i:i+4])
		}
	}
	return dst
}

// Params are the classifier's tunables, all in native-resolution terms.
// Similarities are colour-histogram overlaps from 0 (nothing in common) to 1
// (identical colour mix).
type Params struct {
	// WindowBand is the strip at the bottom of the screen where every battle
	// screen shows its blue message or command windows.
	WindowBand image.Rectangle
	// WindowMinSimilarity is how closely the band must match the battle
	// reference's band for the frame to count as a battle.
	WindowMinSimilarity float64
	// MapMinSimilarity is how closely a frame must match the map reference
	// to count as the map.
	MapMinSimilarity float64
}

// DefaultParams are the starting values for the config file.
func DefaultParams() Params {
	return Params{
		WindowBand:          image.Rect(0, 368, NativeW, NativeH),
		WindowMinSimilarity: 0.40,
		MapMinSimilarity:    0.70,
	}
}

// menuBorderMin is the MenuBorder score from which a frame counts as the game
// menu. The menu scores 0.87-0.89 and everything else at most 0.16.
const menuBorderMin = 0.5

// Refs are the calibration images.
type Refs struct {
	Map    image.Image
	Battle image.Image
}

// Scores are the raw comparison results for one frame.
type Scores struct {
	// Window is the bottom band's similarity to the battle reference's band.
	Window float64
	// Map is the whole frame's similarity to the map reference.
	Map float64
	// MenuBorder is how strongly the frame shows the game menu's left panel
	// border. It doesn't depend on the references.
	MenuBorder float64
}

// Classifier compares frames with the calibration references.
type Classifier struct {
	p          Params
	windowHist histogram
	mapHist    histogram
}

// NewClassifier prepares the references for comparison. refs.Map may be nil:
// until SetMapReference is called, nothing is classified MAP.
func NewClassifier(refs Refs, p Params) *Classifier {
	c := &Classifier{
		p:          p,
		windowHist: histogramOf(Normalize(refs.Battle), p.WindowBand),
	}
	if refs.Map != nil {
		c.SetMapReference(refs.Map)
	}
	return c
}

// SetMapReference replaces the map reference, for example with the screen the
// user was on when starting the bot.
func (c *Classifier) SetMapReference(img image.Image) {
	m := Normalize(img)
	c.mapHist = histogramOf(m, m.Bounds())
}

// Score compares a frame with both references.
func (c *Classifier) Score(frame image.Image) Scores {
	f := Normalize(frame)
	return Scores{
		Window:     histogramOf(f, c.p.WindowBand).intersect(c.windowHist),
		Map:        histogramOf(f, f.Bounds()).intersect(c.mapHist),
		MenuBorder: menuBorderScore(f),
	}
}

// StateFor turns scores into a state.
//
// The game menu uses the same blue windows as a battle, so it is checked first.
// Battle screens always have the battle windows along the bottom, whatever is
// happening above them, so a matching window band means BATTLE even when the
// battle backdrop looks like the map. Checking it before the map also keeps
// such battles out of MAP.
func (c *Classifier) StateFor(s Scores) State {
	if s.MenuBorder >= menuBorderMin {
		return Menu
	}
	if s.Window >= c.p.WindowMinSimilarity {
		return Battle
	}
	if s.Map >= c.p.MapMinSimilarity {
		return Map
	}
	return Neither
}

// Classify scores a frame and returns its state.
func (c *Classifier) Classify(frame image.Image) State {
	return c.StateFor(c.Score(frame))
}

// menuBorderScore looks for the tall white border of the game menu's left
// panel, which runs down the screen near x=162. It must show both above and
// below the status row height, and stand out from the pixels beside it, so that
// white maps, white flashes and the battle status row don't count.
func menuBorderScore(img *image.RGBA) float64 {
	const (
		firstX, lastX = 160, 164
		beside        = 14
		bright        = 190
	)
	brightShare := func(x, y0, y1 int) float64 {
		n := 0
		for y := y0; y < y1; y++ {
			p := img.Pix[img.PixOffset(x, y):]
			if p[0] >= bright && p[1] >= bright && p[2] >= bright {
				n++
			}
		}
		return float64(n) / float64(y1-y0)
	}
	line := func(x, y0, y1 int) float64 {
		return brightShare(x, y0, y1) - max(brightShare(x-beside, y0, y1), brightShare(x+beside, y0, y1))
	}
	var best float64
	for x := firstX; x <= lastX; x++ {
		best = max(best, min(line(x, 10, 280), line(x, 370, 470)))
	}
	return best
}

// Similarity is the colour-histogram overlap of two whole frames, 0-1.
func Similarity(a, b image.Image) float64 {
	na, nb := Normalize(a), Normalize(b)
	return histogramOf(na, na.Bounds()).intersect(histogramOf(nb, nb.Bounds()))
}

// DarkFraction is the share of a frame's pixels in the darkest colour bin,
// which is close to 1 during fades to black.
func DarkFraction(img image.Image) float64 {
	n := Normalize(img)
	return histogramOf(n, n.Bounds())[0]
}

// histogram is a normalised 4x4x4-bin RGB colour histogram. It barely changes
// when the map scrolls or the text in a window changes, unlike a pixel-by-pixel
// comparison.
type histogram [64]float64

// histogramOf builds the histogram of r, sampling every other pixel.
func histogramOf(img *image.RGBA, r image.Rectangle) histogram {
	var h histogram
	var n float64
	r = r.Intersect(img.Bounds())
	for y := r.Min.Y; y < r.Max.Y; y += 2 {
		for x := r.Min.X; x < r.Max.X; x += 2 {
			i := img.PixOffset(x, y)
			bin := int(img.Pix[i]>>6)<<4 | int(img.Pix[i+1]>>6)<<2 | int(img.Pix[i+2]>>6)
			h[bin]++
			n++
		}
	}
	if n > 0 {
		for i := range h {
			h[i] /= n
		}
	}
	return h
}

// intersect returns the overlap of two normalised histograms: 1 for identical
// colour distributions, 0 for completely different ones.
func (h histogram) intersect(o histogram) float64 {
	var sum float64
	for i := range h {
		sum += min(h[i], o[i])
	}
	return sum
}
