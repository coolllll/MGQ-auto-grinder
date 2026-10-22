package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileWritesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg, created, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Error("created = false for a missing file")
	}
	if cfg != Default() {
		t.Errorf("cfg = %+v, want defaults", cfg)
	}

	again, created, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if created || again != Default() {
		t.Errorf("reloading the written file: created=%v cfg=%+v, want false and defaults", created, again)
	}
}

func TestLoadPartialFileKeepsGivenValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	partial := `{"walkMs": 900, "toggleKey": "f8", "windowBand": {"y": 300}}`
	if err := os.WriteFile(path, []byte(partial), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, created, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Error("created = true for an existing file")
	}

	want := Default()
	want.WalkMs = 900
	want.ToggleKey = "f8"
	want.WindowBand.Y = 300
	if cfg != want {
		t.Errorf("cfg = %+v\nwant  %+v", cfg, want)
	}
}

func TestLoadRejectsNonPositiveDurations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"graceMs": 0}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(path); err == nil {
		t.Error("Load accepted graceMs = 0")
	}
}
