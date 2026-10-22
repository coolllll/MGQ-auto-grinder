// Package detect classifies game frames as world map, battle, or neither by
// comparing them with calibration references.
package detect

import (
	"image"
	"image/draw"
)

// NativeW and NativeH are the game's internal resolution. Every frame and
// reference is normalised to this size, so window scaling doesn't matter.
const (
	NativeW = 544
	NativeH = 416
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
type Params struct {
	// BattleRegion is compared pixel by pixel with the battle reference.
	BattleRegion image.Rectangle
	// BattleMaxDiff is the largest mean per-channel difference (0-255) that
	// still counts as a battle.
	BattleMaxDiff float64
	// MapMinSimilarity is the smallest colour-histogram overlap (0-1) with
	// the map reference that still counts as the map.
	MapMinSimilarity float64
}

// DefaultParams are the starting values for the config file.
func DefaultParams() Params {
	return Params{
		BattleRegion:     image.Rect(0, 288, NativeW, NativeH),
		BattleMaxDiff:    20,
		MapMinSimilarity: 0.80,
	}
}

// Refs are the calibration images.
type Refs struct {
	Map    image.Image
	Battle image.Image
}

// Scores are the raw comparison results for one frame.
type Scores struct {
	BattleDiff    float64
	MapSimilarity float64
}

// Classifier compares frames with the calibration references.
type Classifier struct {
	p       Params
	battle  *image.RGBA
	mapHist histogram
}

// NewClassifier prepares the references for comparison.
func NewClassifier(refs Refs, p Params) *Classifier {
	return &Classifier{
		p:       p,
		battle:  Normalize(refs.Battle),
		mapHist: histogramOf(Normalize(refs.Map)),
	}
}

// Score compares a frame with both references.
func (c *Classifier) Score(frame image.Image) Scores {
	f := Normalize(frame)
	return Scores{
		BattleDiff:    regionDiff(f, c.battle, c.p.BattleRegion),
		MapSimilarity: histogramOf(f).intersect(c.mapHist),
	}
}

// StateFor turns scores into a state. Battle wins over map, because the
// battle region check is the more specific of the two.
func (c *Classifier) StateFor(s Scores) State {
	switch {
	case s.BattleDiff <= c.p.BattleMaxDiff:
		return Battle
	case s.MapSimilarity >= c.p.MapMinSimilarity:
		return Map
	default:
		return Neither
	}
}

// Classify scores a frame and returns its state.
func (c *Classifier) Classify(frame image.Image) State {
	return c.StateFor(c.Score(frame))
}

// sampleStep skips pixels in the region comparison; the UI is large and flat
// enough that every 4th pixel is plenty.
const sampleStep = 4

// regionDiff is the mean absolute per-channel difference of a and b inside r.
func regionDiff(a, b *image.RGBA, r image.Rectangle) float64 {
	r = r.Intersect(a.Bounds()).Intersect(b.Bounds())
	var sum, n int
	for y := r.Min.Y; y < r.Max.Y; y += sampleStep {
		for x := r.Min.X; x < r.Max.X; x += sampleStep {
			i, j := a.PixOffset(x, y), b.PixOffset(x, y)
			for c := 0; c < 3; c++ {
				d := int(a.Pix[i+c]) - int(b.Pix[j+c])
				if d < 0 {
					d = -d
				}
				sum += d
			}
			n += 3
		}
	}
	if n == 0 {
		return 255
	}
	return float64(sum) / float64(n)
}

// histogram is a normalised 4x4x4-bin RGB colour histogram. It hardly changes
// when the map scrolls a few tiles, unlike a pixel comparison.
type histogram [64]float64

func histogramOf(img *image.RGBA) histogram {
	var h histogram
	var n float64
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y += 2 {
		for x := b.Min.X; x < b.Max.X; x += 2 {
			i := img.PixOffset(x, y)
			bin := int(img.Pix[i]>>6)<<4 | int(img.Pix[i+1]>>6)<<2 | int(img.Pix[i+2]>>6)
			h[bin]++
			n++
		}
	}
	for i := range h {
		h[i] /= n
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
