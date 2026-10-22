package bot

import (
	"errors"
	"image"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/coolllll/MGQ-auto-grinder/internal/detect"
)

type recorder struct {
	keys   []string
	infos  []string
	alerts []string
}

func (r *recorder) Down(k Key)         { r.keys = append(r.keys, "down "+k.String()) }
func (r *recorder) Up(k Key)           { r.keys = append(r.keys, "up "+k.String()) }
func (r *recorder) Info(msg string)    { r.infos = append(r.infos, msg) }
func (r *recorder) Alert(msg string)   { r.alerts = append(r.alerts, msg) }
func (r *recorder) takeKeys() []string { k := r.keys; r.keys = nil; return k }

var cfg = Config{Walk: 600 * time.Millisecond, ZInterval: 100 * time.Millisecond, Grace: 3 * time.Second}

var t0 = time.Unix(1_000_000, 0)

func at(ms int) time.Time { return t0.Add(time.Duration(ms) * time.Millisecond) }

func seen(s detect.State) Observation { return Observation{State: s, Focused: true} }

func newRunning(t *testing.T) (*Bot, *recorder) {
	t.Helper()
	r := &recorder{}
	b := New(cfg, r, r)
	b.Toggle()
	return b, r
}

func wantKeys(t *testing.T, r *recorder, want ...string) {
	t.Helper()
	got := r.takeKeys()
	if len(got) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("keys = %q, want %q", got, want)
	}
}

func TestStartsPausedAndSendsNothing(t *testing.T) {
	r := &recorder{}
	b := New(cfg, r, r)
	if b.Mode() != Paused {
		t.Fatalf("mode = %v, want PAUSED", b.Mode())
	}
	b.Step(at(0), seen(detect.Map))
	b.Step(at(1000), seen(detect.Battle))
	wantKeys(t, r)
}

func TestToggleStartsAndPauseReleasesHeldKeys(t *testing.T) {
	b, r := newRunning(t)
	if b.Mode() != Running {
		t.Fatalf("mode = %v after Toggle, want RUNNING", b.Mode())
	}
	if len(r.infos) == 0 || r.infos[0] != "Running." {
		t.Errorf("infos = %q, want a Running. message", r.infos)
	}

	b.Step(at(0), seen(detect.Map))
	wantKeys(t, r, "down Left")

	b.Toggle()
	if b.Mode() != Paused {
		t.Fatalf("mode = %v after second Toggle, want PAUSED", b.Mode())
	}
	wantKeys(t, r, "up Left")
	if last := r.infos[len(r.infos)-1]; last != "Paused." {
		t.Errorf("last info = %q, want Paused.", last)
	}

	b.Step(at(100), seen(detect.Map))
	wantKeys(t, r)
}

func TestWalkAlternatesLeftAndRight(t *testing.T) {
	b, r := newRunning(t)
	b.Step(at(0), seen(detect.Map))
	wantKeys(t, r, "down Left")
	b.Step(at(300), seen(detect.Map))
	wantKeys(t, r)
	b.Step(at(600), seen(detect.Map))
	wantKeys(t, r, "up Left", "down Right")
	b.Step(at(1200), seen(detect.Map))
	wantKeys(t, r, "up Right", "down Left")
}

func TestBattleTapsZAtInterval(t *testing.T) {
	b, r := newRunning(t)
	b.Step(at(0), seen(detect.Battle))
	wantKeys(t, r, "down Z")
	b.Step(at(50), seen(detect.Battle))
	wantKeys(t, r, "up Z")
	b.Step(at(75), seen(detect.Battle))
	wantKeys(t, r)
	b.Step(at(100), seen(detect.Battle))
	wantKeys(t, r, "down Z")
	b.Step(at(150), seen(detect.Battle))
	wantKeys(t, r, "up Z")
}

func TestEncounterMidStepReleasesArrowFirst(t *testing.T) {
	b, r := newRunning(t)
	b.Step(at(0), seen(detect.Map))
	wantKeys(t, r, "down Left")
	b.Step(at(300), seen(detect.Battle))
	wantKeys(t, r, "up Left", "down Z")
}

func TestBattleEndsAndWalkingResumes(t *testing.T) {
	b, r := newRunning(t)
	b.Step(at(0), seen(detect.Battle))
	wantKeys(t, r, "down Z")
	b.Step(at(20), seen(detect.Map))
	wantKeys(t, r, "up Z", "down Left")
}

func TestGameOverAfterBattleSendsNoZAndStops(t *testing.T) {
	b, r := newRunning(t)
	b.Step(at(0), seen(detect.Battle))
	b.Step(at(50), seen(detect.Battle))
	r.takeKeys()

	for ms := 100; ms < 3100; ms += 50 {
		b.Step(at(ms), seen(detect.Neither))
	}
	wantKeys(t, r)
	if b.Mode() != Running {
		t.Fatal("paused before the grace period ended")
	}

	b.Step(at(3100), seen(detect.Neither))
	if b.Mode() != Paused {
		t.Fatalf("mode = %v after the grace period, want PAUSED", b.Mode())
	}
	if len(r.alerts) != 1 {
		t.Errorf("alerts = %q, want one", r.alerts)
	}

	b.Step(at(3200), seen(detect.Battle))
	wantKeys(t, r)
}

func TestNeitherReleasesHeldKeyImmediately(t *testing.T) {
	b, r := newRunning(t)
	b.Step(at(0), seen(detect.Battle))
	wantKeys(t, r, "down Z")
	b.Step(at(10), seen(detect.Neither))
	wantKeys(t, r, "up Z")
}

func TestShortTransitionDoesNotStop(t *testing.T) {
	b, r := newRunning(t)
	b.Step(at(0), seen(detect.Map))
	b.Step(at(100), seen(detect.Neither))
	b.Step(at(900), seen(detect.Neither))
	b.Step(at(1000), seen(detect.Battle))
	if b.Mode() != Running {
		t.Fatalf("mode = %v after a 0.9 s fade, want RUNNING", b.Mode())
	}
	if len(r.alerts) != 0 {
		t.Errorf("alerts = %q, want none", r.alerts)
	}
	wantKeys(t, r, "down Left", "up Left", "down Z")
}

func TestLongOtherScreenStops(t *testing.T) {
	b, r := newRunning(t)
	b.Step(at(0), seen(detect.Map))
	b.Step(at(500), seen(detect.Neither))
	b.Step(at(3499), seen(detect.Neither))
	if b.Mode() != Running {
		t.Fatal("paused before the grace period ended")
	}
	b.Step(at(3500), seen(detect.Neither))
	if b.Mode() != Paused {
		t.Fatalf("mode = %v, want PAUSED", b.Mode())
	}
	if len(r.alerts) != 1 || !strings.Contains(r.alerts[0], "neither map nor battle") {
		t.Errorf("alerts = %q, want the neither-map-nor-battle alert", r.alerts)
	}
}

func TestFocusLossReleasesKeysAndKeepsRunning(t *testing.T) {
	b, r := newRunning(t)
	b.Step(at(0), seen(detect.Map))
	wantKeys(t, r, "down Left")

	b.Step(at(100), Observation{Focused: false})
	wantKeys(t, r, "up Left")
	for ms := 200; ms < 5000; ms += 100 {
		b.Step(at(ms), Observation{Focused: false})
	}
	wantKeys(t, r)
	if b.Mode() != Running {
		t.Fatalf("mode = %v while unfocused, want RUNNING", b.Mode())
	}

	b.Step(at(5000), seen(detect.Map))
	wantKeys(t, r, "down Left")
}

type fakeWindow struct {
	exists, foreground bool
	captureErr         error
	captures           int
}

func (w *fakeWindow) Exists() bool       { return w.exists }
func (w *fakeWindow) IsForeground() bool { return w.foreground }

// Capture returns a 1x1 frame whose red channel is the capture count, so tests
// can tell frames apart.
func (w *fakeWindow) Capture() (image.Image, error) {
	if w.captureErr != nil {
		return nil, w.captureErr
	}
	w.captures++
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Pix[0] = uint8(w.captures)
	return img, nil
}

func frameID(img image.Image) uint8 { return img.(*image.RGBA).Pix[0] }

type fixedClassifier detect.State

func (c fixedClassifier) Classify(image.Image) detect.State { return detect.State(c) }
func (c fixedClassifier) SetMapReference(image.Image)       {}

// recordingClassifier remembers its map reference and reports MAP once it has
// one, and NEITHER before that.
type recordingClassifier struct {
	mapRef image.Image
	sets   int
}

func (c *recordingClassifier) SetMapReference(img image.Image) { c.mapRef = img; c.sets++ }
func (c *recordingClassifier) Classify(image.Image) detect.State {
	if c.mapRef == nil {
		return detect.Neither
	}
	return detect.Map
}

func newPausedRunner() (*Runner, *recorder, *fakeWindow, *recordingClassifier) {
	r := &recorder{}
	w := &fakeWindow{exists: true, foreground: true}
	cls := &recordingClassifier{}
	return &Runner{Bot: New(cfg, r, r), Win: w, Cls: cls}, r, w, cls
}

func TestRunnerStartTakesMapReferenceFirst(t *testing.T) {
	run, r, w, cls := newPausedRunner()
	var saved image.Image
	run.OnMapReference = func(img image.Image) { saved = img }

	run.Toggle()
	if run.Bot.Mode() != Running {
		t.Fatalf("mode = %v after start, want RUNNING", run.Bot.Mode())
	}
	if cls.sets != 1 || frameID(cls.mapRef) != 1 {
		t.Fatalf("map reference set %d times (frame %v), want once with the first capture", cls.sets, cls.mapRef)
	}
	if saved == nil || frameID(saved) != 1 {
		t.Errorf("OnMapReference got %v, want the captured frame", saved)
	}

	// The first step already sees MAP, because the reference was set first.
	if err := run.Tick(at(0)); err != nil {
		t.Fatal(err)
	}
	wantKeys(t, r, "down Left")

	captures := w.captures
	run.Toggle()
	if run.Bot.Mode() != Paused {
		t.Fatalf("mode = %v after pause, want PAUSED", run.Bot.Mode())
	}
	if w.captures != captures || cls.sets != 1 {
		t.Errorf("pausing captured or replaced the map reference")
	}
	wantKeys(t, r, "up Left")
}

func TestRunnerStartCaptureFailureStaysPaused(t *testing.T) {
	run, r, w, cls := newPausedRunner()
	w.captureErr = errors.New("screen busy")

	run.Toggle()
	if run.Bot.Mode() != Paused {
		t.Fatalf("mode = %v, want PAUSED after a failed start capture", run.Bot.Mode())
	}
	if cls.sets != 0 {
		t.Error("map reference was set from a failed capture")
	}
	if len(r.alerts) != 1 || !strings.Contains(r.alerts[0], "screen busy") {
		t.Errorf("alerts = %q, want one naming the error", r.alerts)
	}
	if err := run.Tick(at(0)); err != nil {
		t.Fatal(err)
	}
	wantKeys(t, r)
}

func TestRunnerRestartRetakesMapReference(t *testing.T) {
	run, _, w, cls := newPausedRunner()
	run.Toggle() // start: capture 1
	run.Toggle() // pause
	run.Toggle() // start: capture 2
	if w.captures != 2 || cls.sets != 2 {
		t.Fatalf("captures=%d sets=%d, want 2 and 2", w.captures, cls.sets)
	}
	if frameID(cls.mapRef) != 2 {
		t.Errorf("map reference is frame %d, want the second capture", frameID(cls.mapRef))
	}
}

func TestRunnerStopsWhenWindowCloses(t *testing.T) {
	b, r := newRunning(t)
	w := &fakeWindow{exists: true, foreground: true}
	run := &Runner{Bot: b, Win: w, Cls: fixedClassifier(detect.Map)}

	if err := run.Tick(at(0)); err != nil {
		t.Fatal(err)
	}
	wantKeys(t, r, "down Left")

	w.exists = false
	if err := run.Tick(at(100)); !errors.Is(err, ErrWindowLost) {
		t.Fatalf("Tick err = %v, want ErrWindowLost", err)
	}
	wantKeys(t, r, "up Left")
	if b.Mode() != Paused {
		t.Errorf("mode = %v, want PAUSED", b.Mode())
	}
	if len(r.alerts) != 1 {
		t.Errorf("alerts = %q, want one", r.alerts)
	}
}

func TestRunnerTreatsCaptureFailureAsNeither(t *testing.T) {
	b, r := newRunning(t)
	w := &fakeWindow{exists: true, foreground: true, captureErr: errors.New("boom")}
	run := &Runner{Bot: b, Win: w, Cls: fixedClassifier(detect.Battle)}

	if err := run.Tick(at(0)); err == nil {
		t.Fatal("Tick hid the capture error")
	}
	wantKeys(t, r)
}

func TestRunnerSendsNothingWhenNotForeground(t *testing.T) {
	b, r := newRunning(t)
	w := &fakeWindow{exists: true, foreground: false}
	run := &Runner{Bot: b, Win: w, Cls: fixedClassifier(detect.Battle)}
	for ms := 0; ms < 1000; ms += 50 {
		if err := run.Tick(at(ms)); err != nil {
			t.Fatal(err)
		}
	}
	wantKeys(t, r)
}
