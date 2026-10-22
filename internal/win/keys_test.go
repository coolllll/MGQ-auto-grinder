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
