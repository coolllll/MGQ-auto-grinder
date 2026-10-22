package win

import (
	"testing"
	"unsafe"
)

func TestEdgeDetectorHeldKeyTogglesOnce(t *testing.T) {
	samples := []bool{false, true, true, true, false, false, true, false}
	want := []bool{false, true, false, false, false, false, true, false}

	var e EdgeDetector
	for i, down := range samples {
		if got := e.Pressed(down); got != want[i] {
			t.Errorf("sample %d (down=%v): Pressed = %v, want %v", i, down, got, want[i])
		}
	}
}

func TestToggleGateIgnoresPressesOutsideTheGame(t *testing.T) {
	samples := []struct{ down, inFront, want bool }{
		{false, false, false},
		{true, false, false}, // space typed in another window
		{false, false, false},
		{true, false, false}, // held while switching to the game...
		{true, true, false},  // ...still the same press, not a new one
		{false, true, false},
		{true, true, true}, // a fresh press in the game
		{true, true, false},
	}
	var g ToggleGate
	for i, s := range samples {
		if got := g.Pressed(s.down, s.inFront); got != s.want {
			t.Errorf("sample %d (down=%v inFront=%v): Pressed = %v, want %v", i, s.down, s.inFront, got, s.want)
		}
	}
}

func TestVKByName(t *testing.T) {
	if vk, err := VKByName("Space"); err != nil || vk != 0x20 {
		t.Errorf("VKByName(Space) = %#x, %v; want 0x20, nil", vk, err)
	}
	if _, err := VKByName("enter"); err == nil {
		t.Error("VKByName(enter) succeeded, want an error")
	}
}

func TestKeyboardInputMatchesWin64InputSize(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("layout is only checked on 64-bit")
	}
	if got := unsafe.Sizeof(keyboardInput{}); got != 40 {
		t.Errorf("sizeof(keyboardInput) = %d, want 40 (sizeof INPUT on Win64)", got)
	}
}
