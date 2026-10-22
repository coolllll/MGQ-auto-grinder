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
)

func (s State) String() string {
	switch s {
	case Map:
		return "MAP"
	case Battle:
		return "BATTLE"
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
	// to count as the map. It also decides whether the area above a window
	// band is the map, which makes the frame a dialogue on the map.
	MapMinSimilarity float64
}

// DefaultParams are the starting values for the config file.
func DefaultParams() Params {
	return Params{
		WindowBand:          image.Rect(0, 368, NativeW, NativeH),
		WindowMinSimilarity: 0.55,
		MapMinSimilarity:    0.55,
	}
}

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
	// AboveWindow is the similarity of the area above the band to the map
	// reference.
	AboveWindow float64
}

// Classifier compares frames with the calibration references.
type Classifier struct {
	p          Params
	windowHist histogram
	mapHist    histogram
}

// NewClassifier prepares the references for comparison.
func NewClassifier(refs Refs, p Params) *Classifier {
	m := Normalize(refs.Map)
	return &Classifier{
		p:          p,
		windowHist: histogramOf(Normalize(refs.Battle), p.WindowBand),
		mapHist:    histogramOf(m, m.Bounds()),
	}
}

// Score compares a frame with both references.
func (c *Classifier) Score(frame image.Image) Scores {
	f := Normalize(frame)
	above := image.Rect(0, 0, NativeW, c.p.WindowBand.Min.Y)
	return Scores{
		Window:      histogramOf(f, c.p.WindowBand).intersect(c.windowHist),
		Map:         histogramOf(f, f.Bounds()).intersect(c.mapHist),
		AboveWindow: histogramOf(f, above).intersect(c.mapHist),
	}
}

// StateFor turns scores into a state.
//
// Battle screens always have the battle windows along the bottom, whatever is
// happening above them. The map's own text box looks the same, so a window
// band with the map above it is a dialogue on the map, which is NEITHER.
func (c *Classifier) StateFor(s Scores) State {
	if s.Window >= c.p.WindowMinSimilarity {
		if s.AboveWindow >= c.p.MapMinSimilarity {
			return Neither
		}
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
