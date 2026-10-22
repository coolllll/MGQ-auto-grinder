package detect

import (
	"errors"
	"image"
	"image/color"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func randomNative(seed int64) *image.RGBA {
	r := rand.New(rand.NewSource(seed))
	img := image.NewRGBA(image.Rect(0, 0, NativeW, NativeH))
	r.Read(img.Pix)
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 0xFF
	}
	return img
}

func upscale2x(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx()*2, b.Dy()*2))
	for y := 0; y < dst.Bounds().Dy(); y++ {
		for x := 0; x < dst.Bounds().Dx(); x++ {
			dst.Set(x, y, src.At(x/2, y/2))
		}
	}
	return dst
}

func TestNormalize2xMatches1x(t *testing.T) {
	native := randomNative(1)
	got := Normalize(upscale2x(native))
	if got.Bounds() != native.Bounds() {
		t.Fatalf("bounds = %v, want %v", got.Bounds(), native.Bounds())
	}
	for i := range native.Pix {
		if got.Pix[i] != native.Pix[i] {
			t.Fatalf("pixel byte %d = %d, want %d", i, got.Pix[i], native.Pix[i])
		}
	}
}

func TestNormalizeNativeIsUnchanged(t *testing.T) {
	native := randomNative(2)
	if Normalize(native) != native {
		t.Error("Normalize copied an image that was already native size")
	}
}

func fill(img *image.RGBA, r image.Rectangle, c color.RGBA) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

// synthetic scenes: a map of grass and water, and a battle with a varying
// background over a fixed status window strip.
func syntheticMap(shift int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, NativeW, NativeH))
	fill(img, img.Bounds(), color.RGBA{40, 160, 60, 255})
	fill(img, image.Rect(100+shift, 50, 300+shift, 200), color.RGBA{30, 80, 200, 255})
	return img
}

func syntheticBattle(bg color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, NativeW, NativeH))
	fill(img, img.Bounds(), bg)
	band := DefaultParams().WindowBand
	fill(img, band, color.RGBA{40, 70, 150, 255})
	fill(img, image.Rect(16, band.Min.Y+12, 120, band.Min.Y+28), color.RGBA{240, 240, 240, 255})
	return img
}

func TestClassifySynthetic(t *testing.T) {
	c := NewClassifier(Refs{Map: syntheticMap(0), Battle: syntheticBattle(color.RGBA{120, 90, 60, 255})}, DefaultParams())

	black := image.NewRGBA(image.Rect(0, 0, NativeW, NativeH))
	fill(black, black.Bounds(), color.RGBA{0, 0, 0, 255})

	cases := []struct {
		name  string
		frame image.Image
		want  State
	}{
		{"map", syntheticMap(0), Map},
		{"map scrolled", syntheticMap(24), Map},
		{"map at 2x", upscale2x(syntheticMap(8)), Map},
		{"battle", syntheticBattle(color.RGBA{120, 90, 60, 255}), Battle},
		{"battle with another background", syntheticBattle(color.RGBA{200, 30, 30, 255}), Battle},
		{"black fade", black, Neither},
	}
	for _, tc := range cases {
		if got := c.Classify(tc.frame); got != tc.want {
			t.Errorf("%s: got %v (scores %+v), want %v", tc.name, got, c.Score(tc.frame), tc.want)
		}
	}
}

// TestClassifyFixtures runs the classifier over screenshots of the real game
// in testdata/. map.png and battle.png double as the references.
func TestClassifyFixtures(t *testing.T) {
	dir := "testdata"
	refs, err := LoadRefs(dir)
	if errors.Is(err, ErrMissingRefs) {
		t.Skip("no game fixtures in testdata/ yet")
	}
	if err != nil {
		t.Fatal(err)
	}
	c := NewClassifier(refs, DefaultParams())

	cases := []struct {
		file string
		want State
	}{
		{"map.png", Map},
		{"map_scrolled.png", Map},
		{"battle.png", Battle},
		{"battle_command.png", Battle},
		{"battle_skill_menu.png", Battle},
		{"battle_flash.png", Battle},
		{"battle_victory.png", Battle},
		{"battle_level_up.png", Battle},
		{"encounter_swirl.png", Neither},
		{"black_fade.png", Neither},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			img, err := LoadPNG(filepath.Join(dir, tc.file))
			if errors.Is(err, os.ErrNotExist) {
				t.Skip("fixture not captured yet")
			}
			if err != nil {
				t.Fatal(err)
			}
			s := c.Score(img)
			t.Logf("window=%.3f map=%.3f", s.Window, s.Map)
			if got := c.StateFor(s); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func loadFixture(t *testing.T, name string) *image.RGBA {
	t.Helper()
	img, err := LoadPNG(filepath.Join("testdata", name))
	if errors.Is(err, os.ErrNotExist) {
		t.Skipf("fixture %s not captured", name)
	}
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func TestSetMapReference(t *testing.T) {
	battleRef := loadFixture(t, "battle.png")
	maps := []string{"map.png", "map_scrolled.png"}
	battles := []string{"battle.png", "battle_command.png", "battle_victory.png"}

	c := NewClassifier(Refs{Battle: battleRef}, DefaultParams())
	for _, f := range maps {
		if got := c.Classify(loadFixture(t, f)); got == Map {
			t.Errorf("without a map reference, %s classified MAP", f)
		}
	}

	c.SetMapReference(loadFixture(t, "map.png"))
	for _, f := range maps {
		if got := c.Classify(loadFixture(t, f)); got != Map {
			t.Errorf("after SetMapReference, %s = %v, want MAP", f, got)
		}
	}
	for _, f := range battles {
		if got := c.Classify(loadFixture(t, f)); got != Battle {
			t.Errorf("after SetMapReference, %s = %v, want BATTLE", f, got)
		}
	}
}

// TestSameSpotReferenceMargin checks that a map reference taken where the user
// grinds leaves a wide margin: the default-size run scored only 0.63-0.69
// against a reference taken elsewhere.
func TestSameSpotReferenceMargin(t *testing.T) {
	c := NewClassifier(Refs{Battle: loadFixture(t, "battle.png")}, DefaultParams())
	c.SetMapReference(loadFixture(t, "map.png"))
	s := c.Score(loadFixture(t, "map_scrolled.png"))
	t.Logf("map_scrolled vs same-spot reference: map=%.3f (threshold %.2f)", s.Map, DefaultParams().MapMinSimilarity)
	if s.Map < 0.75 {
		t.Errorf("map similarity %.3f, want >= 0.75", s.Map)
	}
}

// TestBattleBackdropLikeTheMap is the regression test for the second grinding
// spot, where the battle backdrop shares its colours with the world map. There
// the area above the window band scored up to 0.58 against the map reference,
// and a rule that called "battle windows with the map above them" a dialogue
// on the map turned battle frames into NEITHER. The recorded frame peaks at
// 0.53, so the map threshold is lowered to 0.50 to stand in for a slightly
// more map-like spot; a battle frame must be BATTLE whatever the threshold.
func TestBattleBackdropLikeTheMap(t *testing.T) {
	p := DefaultParams()
	p.MapMinSimilarity = 0.50
	c := NewClassifier(Refs{Battle: loadFixture(t, "battle.png"), Map: loadFixture(t, "map_spot2.png")}, p)

	for _, tc := range []struct {
		file string
		want State
	}{
		{"battle_spot2.png", Battle},
		{"map_spot2.png", Map},
	} {
		img := loadFixture(t, tc.file)
		s := c.Score(img)
		t.Logf("%s: window=%.3f map=%.3f", tc.file, s.Window, s.Map)
		if got := c.StateFor(s); got != tc.want {
			t.Errorf("%s = %v, want %v", tc.file, got, tc.want)
		}
	}
}

func TestRefsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadRefs(dir); !errors.Is(err, ErrMissingRefs) {
		t.Fatalf("LoadRefs on empty dir: err = %v, want ErrMissingRefs", err)
	}
	if _, err := SaveRef(dir, Map, upscale2x(syntheticMap(0))); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRefs(dir); !errors.Is(err, ErrMissingRefs) {
		t.Fatalf("LoadRefs with only map: err = %v, want ErrMissingRefs", err)
	}

	battleOnly := t.TempDir()
	if _, err := SaveRef(battleOnly, Battle, syntheticBattle(color.RGBA{1, 2, 3, 255})); err != nil {
		t.Fatal(err)
	}
	if refs, err := LoadRefs(battleOnly); err != nil || refs.Battle == nil || refs.Map != nil {
		t.Fatalf("LoadRefs with only battle: refs=%+v err=%v, want battle and no map", refs, err)
	}

	if _, err := SaveRef(dir, Battle, syntheticBattle(color.RGBA{1, 2, 3, 255})); err != nil {
		t.Fatal(err)
	}
	refs, err := LoadRefs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if refs.Map.Bounds() != image.Rect(0, 0, NativeW, NativeH) {
		t.Errorf("saved map reference is %v, want native size", refs.Map.Bounds())
	}
	if _, err := os.Stat(filepath.Join(dir, "battle.png")); err != nil {
		t.Error(err)
	}
}
