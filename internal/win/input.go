package win

import (
	"fmt"
	"unsafe"
)

// Scan is a hardware keyboard scan code. Games that poll keyboard state
// accept scan-code input more reliably than virtual-key events.
type Scan struct {
	Code     uint16
	Extended bool
}

var (
	ScanZ     = Scan{Code: 0x2C}
	ScanLeft  = Scan{Code: 0x4B, Extended: true}
	ScanRight = Scan{Code: 0x4D, Extended: true}
)

const (
	inputKeyboard        = 1
	keyeventfExtendedKey = 0x0001
	keyeventfKeyUp       = 0x0002
	keyeventfScancode    = 0x0008
)

// keyboardInput mirrors the 64-bit INPUT struct with its KEYBDINPUT member.
// The trailing padding brings it up to the size of the union's largest
// member (MOUSEINPUT), which SendInput checks via cbSize.
type keyboardInput struct {
	typ   uint32
	_     uint32
	vk    uint16
	scan  uint16
	flags uint32
	time  uint32
	_     uint32
	extra uintptr
	_     [8]byte
}

// KeyDown presses a key.
func KeyDown(s Scan) error { return sendKey(s, false) }

// KeyUp releases a key.
func KeyUp(s Scan) error { return sendKey(s, true) }

func sendKey(s Scan, up bool) error {
	in := keyboardInput{typ: inputKeyboard, scan: s.Code, flags: keyeventfScancode}
	if s.Extended {
		in.flags |= keyeventfExtendedKey
	}
	if up {
		in.flags |= keyeventfKeyUp
	}
	n, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
	if n != 1 {
		return fmt.Errorf("SendInput: %w", err)
	}
	return nil
}
