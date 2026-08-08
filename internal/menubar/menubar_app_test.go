//go:build darwin || windows

package menubar

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stn1slv/md-paste/internal/config"
)

// fakeUI records the status-bar state. Restore timers run on their own
// goroutines, so every field is mutex-guarded.
type fakeUI struct {
	mu           sync.Mutex
	title        string
	enabled      bool
	iconIsCheck  bool
	titleChanges int
}

func (u *fakeUI) setIcon(icon []byte) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.iconIsCheck = &icon[0] == &iconCheck[0]
}

func (u *fakeUI) setConvertTitle(title string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.title = title
	u.titleChanges++
}

func (u *fakeUI) setConvertEnabled(enabled bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.enabled = enabled
}

func (u *fakeUI) state() (title string, enabled bool, changes int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.title, u.enabled, u.titleChanges
}

// Timings for the flash tests. The two flashes are given deliberately lopsided
// durations so the observation point sits far from both deadlines: equal
// durations put the check at the very instant the second timer is due, which is
// a coin flip no matter how precise the clock is.
const (
	shortFlash   = 50 * time.Millisecond
	longFlash    = 10 * time.Second
	observeAfter = 750 * time.Millisecond
	settleWithin = 5 * time.Second
	settlePoll   = 10 * time.Millisecond
)

func TestFlashSupersededFlashDoesNotRestore(t *testing.T) {
	ui := &fakeUI{enabled: true}
	a := &app{ui: ui}

	apply := func(msg string) func() { return func() { ui.setConvertTitle(msg) } }
	restore := func() { ui.setConvertTitle(convertLabel) }

	a.flash(&a.titleGen, apply("first"), restore, shortFlash)
	a.flash(&a.titleGen, apply("second"), restore, longFlash)

	// Well past the first flash's deadline and nowhere near the second's, so
	// timer granularity on a loaded runner cannot move the check across either.
	time.Sleep(observeAfter)

	title, _, _ := ui.state()
	assert.Equal(t, "second", title, "a superseded flash restored while a later one was still showing")
}

func TestFlashLatestFlashRestores(t *testing.T) {
	ui := &fakeUI{enabled: true}
	a := &app{ui: ui}

	a.flash(&a.titleGen,
		func() { ui.setConvertTitle("flash") },
		func() { ui.setConvertTitle(convertLabel) },
		shortFlash,
	)

	require.Eventually(t, func() bool {
		title, _, _ := ui.state()
		return title == convertLabel
	}, settleWithin, settlePoll, "the flash never restored the Convert item")
}

func TestFlashErrorDisablesConvertAndRestoresOnce(t *testing.T) {
	ui := &fakeUI{enabled: true}
	a := &app{ui: ui}

	a.flashError("Nothing to convert")

	title, enabled, changes := ui.state()
	assert.Equal(t, "Nothing to convert", title)
	assert.False(t, enabled, "the Convert item was left enabled during an error flash")
	assert.Equal(t, 1, changes)
}

func TestFlashSuccessSwapsTheIcon(t *testing.T) {
	ui := &fakeUI{enabled: true}
	a := &app{ui: ui}

	a.flashSuccess()

	ui.mu.Lock()
	defer ui.mu.Unlock()
	assert.True(t, ui.iconIsCheck)
}

// The icon and the Convert title are independent surfaces. An icon flash must
// not cancel a pending title restore, or an error would leave the Convert item
// permanently disabled.
func TestIconFlashDoesNotCancelAPendingTitleRestore(t *testing.T) {
	ui := &fakeUI{enabled: true}
	a := &app{ui: ui}

	a.flash(&a.titleGen,
		func() { ui.setConvertTitle("error"); ui.setConvertEnabled(false) },
		func() { ui.setConvertTitle(convertLabel); ui.setConvertEnabled(true) },
		shortFlash,
	)
	a.flash(&a.iconGen, func() { ui.setIcon(iconCheck) }, func() { ui.setIcon(iconNormal) }, shortFlash/2)

	require.Eventually(t, func() bool {
		title, enabled, _ := ui.state()
		return title == convertLabel && enabled
	}, settleWithin, settlePoll, "the icon flash cancelled the pending title restore")
}

// panicUI fails the way a broken systray would: from inside apply.
type panicUI struct{ fakeUI }

func (u *panicUI) setIcon(_ []byte) { panic("systray blew up") }

// A panic inside apply must not leave a.mu held. convert's recover handler
// calls flashError, which takes the same mutex; holding it across apply would
// block that handler forever, leaving converting set so every later conversion
// is rejected while the icon still looks healthy.
func TestFlashSurvivesAPanicInApply(t *testing.T) {
	a := &app{ui: &panicUI{}}

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				// Mirrors convert's recover handler.
				a.flashError("Conversion failed")
			}
		}()
		a.flashSuccess()
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the recover handler blocked on a.mu held by the panicking apply")
	}
}

func TestBeginConvertDropsOverlappingConversions(t *testing.T) {
	a := &app{}

	require.True(t, a.beginConvert(), "the first conversion must be allowed to start")
	assert.False(t, a.beginConvert(), "a second conversion started while one was in flight")

	a.endConvert()
	assert.True(t, a.beginConvert(), "conversions were not allowed again after the first finished")
}

func TestBeginConvertAdmitsExactlyOneConcurrentCaller(t *testing.T) {
	a := &app{}

	const callers = 32
	var started sync.WaitGroup
	var admitted int64
	var mu sync.Mutex

	started.Add(callers)
	for range callers {
		go func() {
			defer started.Done()
			if a.beginConvert() {
				mu.Lock()
				admitted++
				mu.Unlock()
			}
		}()
	}
	started.Wait()

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, int64(1), admitted, "more than one conversion ran concurrently")
}

func TestResolveLoad(t *testing.T) {
	loaded := config.Config{Hotkey: "cmd+shift+v", LaunchAtLogin: true}
	defaults := config.Config{Hotkey: "ctrl+cmd+opt+m"}
	boom := errors.New("invalid YAML")

	okBackup := func() (string, error) { return "/tmp/config.yaml.bak", nil }
	failBackup := func() (string, error) { return "", errors.New("permission denied") }

	tests := []struct {
		name         string
		cfg          config.Config
		existed      bool
		loadErr      error
		backup       func() (string, error)
		wantSettings config.Config
		wantWrite    bool
	}{
		{
			name:         "first run seeds the file",
			cfg:          defaults,
			existed:      false,
			backup:       okBackup,
			wantSettings: defaults,
			wantWrite:    true,
		},
		{
			name:         "existing file is used and left alone",
			cfg:          loaded,
			existed:      true,
			backup:       okBackup,
			wantSettings: loaded,
			wantWrite:    false,
		},
		{
			name:         "corrupt file is preserved, then rewritten",
			existed:      true,
			loadErr:      boom,
			backup:       okBackup,
			wantSettings: defaults,
			wantWrite:    true,
		},
		{
			// The whole point of the backup is not to lose the user's settings.
			// If it cannot be taken, the file must be left exactly as it is.
			name:         "corrupt file is left alone when it cannot be preserved",
			existed:      true,
			loadErr:      boom,
			backup:       failBackup,
			wantSettings: defaults,
			wantWrite:    false,
		},
		{
			name:         "no config path means nothing to preserve or write",
			existed:      true,
			loadErr:      boom,
			backup:       nil,
			wantSettings: defaults,
			wantWrite:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, write := resolveLoad(tt.cfg, tt.existed, tt.loadErr, defaults, tt.backup)
			assert.Equal(t, tt.wantSettings, settings)
			assert.Equal(t, tt.wantWrite, write, "wrong decision about rewriting the config file")
		})
	}
}
