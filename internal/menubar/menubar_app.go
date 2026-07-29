//go:build darwin || windows

package menubar

import (
	"log/slog"
	"sync"
	"time"

	"fyne.io/systray"

	"github.com/stn1slv/md-paste/internal/clipboard"
	"github.com/stn1slv/md-paste/internal/service"
)

const (
	convertLabel      = "Convert clipboard to Markdown"
	flashDuration     = 1 * time.Second
	errorRestoreDelay = 1500 * time.Millisecond
)

// app holds the mutable menu state shared between the click loop and timers.
type app struct {
	cfg Config

	mConvert *systray.MenuItem

	mu  sync.Mutex
	gen uint64 // flash generation guard: only the latest flash reverts the icon
}

// Run starts the menu bar event loop. It must execute on the main goroutine,
// which cobra's RunE already guarantees; do not wrap this call in `go`.
func Run(cfg Config) error {
	a := &app{cfg: cfg}
	systray.Run(a.onReady, func() {})
	return nil
}

func (a *app) onReady() {
	systray.SetTemplateIcon(iconNormal, iconNormal)
	systray.SetTooltip("md-paste")

	// Repair a stale login-item entry (e.g. after an update moved the binary)
	// before seeding the checkbox, so autostart survives updates.
	reconcileLoginItem()

	a.mConvert = systray.AddMenuItem(convertLabel, "Convert the current clipboard content to Markdown")
	mLogin := systray.AddMenuItemCheckbox("Launch at login", "Start md-paste automatically after you log in", loginItemEnabled())

	systray.AddSeparator()
	mAbout := systray.AddMenuItem("md-paste "+a.cfg.Version, "")
	mAbout.Disable()

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit md-paste")

	go a.loop(mLogin, mQuit)
}

func (a *app) loop(mLogin, mQuit *systray.MenuItem) {
	for {
		select {
		case <-a.mConvert.ClickedCh:
			a.convert()
		case <-mLogin.ClickedCh:
			a.toggleLogin(mLogin)
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (a *app) convert() {
	outcome, _, err := service.Convert(clipboard.Read, clipboard.WriteMarkdown, service.Hooks{})
	switch {
	case err != nil:
		slog.Error("menu bar conversion failed", "error", err)
		a.flashError("Conversion failed")
	case outcome == service.OutcomeEmpty:
		a.flashError("Nothing to convert")
	default:
		a.flashSuccess()
	}
}

// flashSuccess swaps the status icon to the checkmark for a moment. A generation
// guard ensures rapid repeat clicks only revert once, after the last flash.
func (a *app) flashSuccess() {
	a.mu.Lock()
	a.gen++
	gen := a.gen
	a.mu.Unlock()

	systray.SetTemplateIcon(iconCheck, iconCheck)
	time.AfterFunc(flashDuration, func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		if a.gen == gen {
			systray.SetTemplateIcon(iconNormal, iconNormal)
		}
	})
}

// flashError shows a transient, non-modal error by retitling and disabling the
// Convert item, then restoring it. While disabled the item emits no clicks, so
// overlapping error flashes cannot stack.
func (a *app) flashError(msg string) {
	a.mConvert.SetTitle(msg)
	a.mConvert.Disable()
	time.AfterFunc(errorRestoreDelay, func() {
		a.mConvert.SetTitle(convertLabel)
		a.mConvert.Enable()
	})
}

func (a *app) toggleLogin(mLogin *systray.MenuItem) {
	enabling := !mLogin.Checked()
	var err error
	if enabling {
		err = enableLoginItem()
	} else {
		err = disableLoginItem()
	}
	if err != nil {
		slog.Error("failed to toggle launch at login", "enabling", enabling, "error", err)
	}
	// Reflect the actual on-disk state, whether or not the toggle succeeded.
	if loginItemEnabled() {
		mLogin.Check()
	} else {
		mLogin.Uncheck()
	}
}
