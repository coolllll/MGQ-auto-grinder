// Package win wraps the few Win32 calls the grinder needs: finding the game
// window, capturing its client area, sending keys and polling the keyboard.
package win

import (
	"errors"
	"fmt"
	"image"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// GameClass is the window class used by every RPG Maker VX Ace game.
const GameClass = "RGSS Player"

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procGetConsoleProcessList = kernel32.NewProc("GetConsoleProcessList")

	procEnumWindows         = user32.NewProc("EnumWindows")
	procGetClassNameW       = user32.NewProc("GetClassNameW")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	procIsWindow            = user32.NewProc("IsWindow")
	procIsWindowVisible     = user32.NewProc("IsWindowVisible")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetClientRect       = user32.NewProc("GetClientRect")
	procClientToScreen      = user32.NewProc("ClientToScreen")
	procGetDC               = user32.NewProc("GetDC")
	procReleaseDC           = user32.NewProc("ReleaseDC")
	procSendInput           = user32.NewProc("SendInput")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
	procMessageBeep         = user32.NewProc("MessageBeep")
	procSetProcessDPIAware  = user32.NewProc("SetProcessDPIAware")

	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procGetDIBits              = gdi32.NewProc("GetDIBits")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
)

// ErrWindowNotFound is returned when no game window matches.
var ErrWindowNotFound = errors.New("game window not found")

// Window is a top-level window found on the desktop.
type Window struct {
	HWND  uintptr
	Title string
	Class string
}

// EnumWindows callbacks are a limited resource, so one is created up front
// and shared through enumOut under enumMu.
var (
	enumMu  sync.Mutex
	enumOut []Window
	enumCB  = windows.NewCallback(func(hwnd, _ uintptr) uintptr {
		if visible, _, _ := procIsWindowVisible.Call(hwnd); visible != 0 {
			enumOut = append(enumOut, Window{HWND: hwnd, Class: className(hwnd), Title: windowText(hwnd)})
		}
		return 1
	})
)

// ListWindows returns every visible top-level window.
func ListWindows() []Window {
	enumMu.Lock()
	defer enumMu.Unlock()
	enumOut = nil
	procEnumWindows.Call(enumCB, 0)
	out := enumOut
	enumOut = nil
	return out
}

// FindGameWindow returns the visible RGSS Player window whose title contains
// titleSub.
func FindGameWindow(titleSub string) (Window, error) {
	for _, w := range ListWindows() {
		if w.Class == GameClass && strings.Contains(w.Title, titleSub) {
			return w, nil
		}
	}
	return Window{}, fmt.Errorf("%w: no %q window with a title containing %q", ErrWindowNotFound, GameClass, titleSub)
}

// Exists reports whether the window still exists.
func (w Window) Exists() bool {
	ok, _, _ := procIsWindow.Call(w.HWND)
	return ok != 0
}

// IsForeground reports whether the window is the one receiving keyboard input.
func (w Window) IsForeground() bool {
	fg, _, _ := procGetForegroundWindow.Call()
	return fg == w.HWND
}

type rect struct{ Left, Top, Right, Bottom int32 }

type point struct{ X, Y int32 }

// ClientRect returns the window's client area in screen coordinates.
func (w Window) ClientRect() (image.Rectangle, error) {
	var r rect
	if ok, _, err := procGetClientRect.Call(w.HWND, uintptr(unsafe.Pointer(&r))); ok == 0 {
		return image.Rectangle{}, fmt.Errorf("GetClientRect: %w", err)
	}
	var p point
	if ok, _, err := procClientToScreen.Call(w.HWND, uintptr(unsafe.Pointer(&p))); ok == 0 {
		return image.Rectangle{}, fmt.Errorf("ClientToScreen: %w", err)
	}
	return image.Rect(int(p.X), int(p.Y), int(p.X+r.Right), int(p.Y+r.Bottom)), nil
}

func className(hwnd uintptr) string {
	buf := make([]uint16, 256)
	n, _, _ := procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return windows.UTF16ToString(buf[:n])
}

func windowText(hwnd uintptr) string {
	buf := make([]uint16, 512)
	n, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return windows.UTF16ToString(buf[:n])
}

// SetDPIAware makes window and screen coordinates agree on scaled displays.
// Call it once at startup, before any other call in this package.
func SetDPIAware() {
	procSetProcessDPIAware.Call()
}

// OwnsConsole reports whether this process is alone in its console window,
// which means it was started by double-clicking rather than from a terminal,
// and the window will close as soon as the process exits.
func OwnsConsole() bool {
	pids := make([]uint32, 2)
	n, _, _ := procGetConsoleProcessList.Call(uintptr(unsafe.Pointer(&pids[0])), uintptr(len(pids)))
	return n == 1
}

// Beep plays the system warning sound.
func Beep() {
	const mbIconExclamation = 0x30
	procMessageBeep.Call(mbIconExclamation)
}
