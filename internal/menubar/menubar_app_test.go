//go:build darwin || windows

package menubar

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestFlashOnlyTheLatestFlashRestores(t *testing.T) {
	ui := &fakeUI{enabled: true}
	a := &app{ui: ui}

	const d = 60 * time.Millisecond
	apply := func(msg string) func() { return func() { ui.setConvertTitle(msg) } }
	restore := func() { ui.setConvertTitle(convertLabel) }

	a.flash(&a.titleGen, apply("first"), restore, d)
	time.Sleep(d / 3)
	a.flash(&a.titleGen, apply("second"), restore, d)

	// The first flash's timer fires here. It must not restore, because the
	// second flash is still meant to be showing.
	time.Sleep(d)
	title, _, _ := ui.state()
	assert.Equal(t, "second", title, "the earlier flash restored while a later one was showing")

	// The second flash's timer restores.
	time.Sleep(d)
	title, _, _ = ui.state()
	assert.Equal(t, convertLabel, title)
}

func TestFlashErrorDisablesConvertAndRestoresOnce(t *testing.T) {
	ui := &fakeUI{enabled: true}
	a := &app{ui: ui}

	a.flashError("Nothing to convert")

	title, enabled, changes := ui.state()
	assert.Equal(t, "Nothing to convert", title)
	assert.False(t, enabled, "the Convert item stays enabled during an error flash")
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

	const d = 60 * time.Millisecond
	a.flash(&a.titleGen,
		func() { ui.setConvertTitle("error"); ui.setConvertEnabled(false) },
		func() { ui.setConvertTitle(convertLabel); ui.setConvertEnabled(true) },
		d,
	)
	a.flash(&a.iconGen, func() { ui.setIcon(iconCheck) }, func() { ui.setIcon(iconNormal) }, d/2)

	time.Sleep(2 * d)
	title, enabled, _ := ui.state()
	assert.Equal(t, convertLabel, title)
	assert.True(t, enabled, "the Convert item was left disabled")
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
