package win

import (
	"errors"
	"fmt"
	"image"
	"unsafe"
)

type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

// Capture copies the window's client area from the screen. The window must be
// visible and unobstructed; copying from the screen (rather than the window's
// own DC) is what works for Direct3D-backed windows like RGSS.
func (w Window) Capture() (*image.RGBA, error) {
	const (
		srcCopy     = 0x00CC0020
		captureBlt  = 0x40000000
		dibRGBColor = 0
	)

	r, err := w.ClientRect()
	if err != nil {
		return nil, err
	}
	width, height := r.Dx(), r.Dy()
	if width <= 0 || height <= 0 {
		return nil, errors.New("game window has an empty client area (is it minimized?)")
	}

	screen, _, _ := procGetDC.Call(0)
	if screen == 0 {
		return nil, errors.New("GetDC failed")
	}
	defer procReleaseDC.Call(0, screen)

	mem, _, _ := procCreateCompatibleDC.Call(screen)
	if mem == 0 {
		return nil, errors.New("CreateCompatibleDC failed")
	}
	defer procDeleteDC.Call(mem)

	bmp, _, _ := procCreateCompatibleBitmap.Call(screen, uintptr(width), uintptr(height))
	if bmp == 0 {
		return nil, errors.New("CreateCompatibleBitmap failed")
	}
	defer procDeleteObject.Call(bmp)

	old, _, _ := procSelectObject.Call(mem, bmp)
	x, y := r.Min.X, r.Min.Y
	ok, _, bltErr := procBitBlt.Call(mem, 0, 0, uintptr(width), uintptr(height), screen, uintptr(x), uintptr(y), srcCopy|captureBlt)
	procSelectObject.Call(mem, old)
	if ok == 0 {
		return nil, fmt.Errorf("BitBlt: %w", bltErr)
	}

	hdr := bitmapInfoHeader{
		Width:    int32(width),
		Height:   -int32(height), // negative height = top-down rows
		Planes:   1,
		BitCount: 32,
	}
	hdr.Size = uint32(unsafe.Sizeof(hdr))
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	lines, _, dibErr := procGetDIBits.Call(mem, bmp, 0, uintptr(height),
		uintptr(unsafe.Pointer(&img.Pix[0])), uintptr(unsafe.Pointer(&hdr)), dibRGBColor)
	if lines == 0 {
		return nil, fmt.Errorf("GetDIBits: %w", dibErr)
	}

	// GDI gives BGRX; image.RGBA wants RGBA.
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+2] = img.Pix[i+2], img.Pix[i]
		img.Pix[i+3] = 0xFF
	}
	return img, nil
}
