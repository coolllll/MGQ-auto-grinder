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

// LoadRefs loads both references from dir.
func LoadRefs(dir string) (Refs, error) {
	var refs Refs
	var missing []string
	for _, s := range []State{Map, Battle} {
		img, err := LoadPNG(RefPath(dir, s))
		if errors.Is(err, fs.ErrNotExist) {
			missing = append(missing, strings.ToLower(s.String()))
			continue
		}
		if err != nil {
			return Refs{}, err
		}
		if s == Map {
			refs.Map = img
		} else {
			refs.Battle = img
		}
	}
	if len(missing) > 0 {
		return Refs{}, fmt.Errorf("%w in %s: %s", ErrMissingRefs, dir, strings.Join(missing, ", "))
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
