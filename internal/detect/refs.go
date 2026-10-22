package detect

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ErrMissingRefs is returned by LoadRefs when calibration hasn't been done.
var ErrMissingRefs = errors.New("calibration references missing")

// RefPath is where the reference for s is stored inside dir.
func RefPath(dir string, s State) string {
	return filepath.Join(dir, strings.ToLower(s.String())+".png")
}

// SaveRef normalises img and saves it as the reference for s.
func SaveRef(dir string, s State, img image.Image) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := RefPath(dir, s)
	return path, SavePNG(path, Normalize(img))
}

// LoadRefs loads the references from dir. The battle reference is required;
// the map reference is optional (Refs.Map is nil without it), because the bot
// takes a fresh one each time it starts.
func LoadRefs(dir string) (Refs, error) {
	battle, err := LoadPNG(RefPath(dir, Battle))
	if errors.Is(err, fs.ErrNotExist) {
		return Refs{}, fmt.Errorf("%w in %s: battle", ErrMissingRefs, dir)
	}
	if err != nil {
		return Refs{}, err
	}
	refs := Refs{Battle: battle}
	m, err := LoadPNG(RefPath(dir, Map))
	switch {
	case err == nil:
		refs.Map = m
	case !errors.Is(err, fs.ErrNotExist):
		return Refs{}, err
	}
	return refs, nil
}

// SavePNG writes img to path.
func SavePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// LoadPNG reads a PNG as RGBA.
func LoadPNG(path string) (*image.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba, nil
	}
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
	return rgba, nil
}
