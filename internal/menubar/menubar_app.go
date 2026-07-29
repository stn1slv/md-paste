//go:build darwin || windows

package menubar

import (
	"log/slog"
	"sync"
	"time"

	"fyne.io/systray"

	"github.com/stn1slv/md-paste/internal/clipboard"
	"github.com/stn1slv/md-paste/internal/config"
	"github.com/stn1slv/md-paste/internal/service"
)

const (
	convertLabel      = "Convert clipboard to Markdown"
	flashDuration     = 1 * time.Second
	errorRestoreDelay = 1500 * time.Millisecond
)

// app holds the mutable menu state shared between the click loop and timers.
type app struct {
	cfg Config // build metadata (About item)

	settings config.Config // user config: shortcut + launch-at-login
	cfgPath  string

	mConvert  *systray.MenuItem
	mShortcut *systray.MenuItem

	// unbindHotkey unregisters the current global shortcut; nil when none is
	// active (registration failed).
	unbindHotkey func()

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

	// Load user config first: it decides the shortcut and the launch-at-login
	// state (migrating any existing OS autostart into the file on first run).
	a.loadSettings()

	a.mConvert = systray.AddMenuItem(convertLabel, "Convert the current clipboard content to Markdown")

	systray.AddSeparator()
	a.mShortcut = systray.AddMenuItem(shortcutLabel(a.settings.Hotkey), "The global shortcut that triggers a conversion")
	a.mShortcut.Disable()
	mEditShortcut := systray.AddMenuItem("Edit shortcut...", "Open the config file to change the shortcut")
	mReload := systray.AddMenuItem("Reload config", "Re-read the config file and re-register the shortcut")
	mLogin := systray.AddMenuItemCheckbox("Launch at login", "Start md-paste automatically after you log in", a.settings.LaunchAtLogin)

	systray.AddSeparator()
	mAbout := systray.AddMenuItem("md-paste "+a.cfg.Version, "")
	mAbout.Disable()

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit md-paste")

	a.bindHotkey()

	go a.loop(mEditShortcut, mReload, mLogin, mQuit)
}

// loadSettings reads the config, creating it on first run with the current
// launch-at-login state migrated in, and reconciles autostart to the config.
func (a *app) loadSettings() {
	path, err := config.Path()
	if err != nil {
		slog.Error("failed to resolve config path", "error", err)
	}
	a.cfgPath = path

	defaults := config.Config{Hotkey: defaultHotkey, LaunchAtLogin: loginItemPresent()}
	cfg, existed, err := config.Load(path, defaults)
	if err != nil {
		slog.Error("failed to load config, using defaults", "error", err)
		cfg = defaults
	}
	a.settings = cfg

	if !existed && path != "" {
		if err := config.Save(path, cfg); err != nil {
			slog.Error("failed to write initial config", "error", err)
		}
	}

	a.reconcileLogin()
}

// reconcileLogin makes the OS autostart mechanism match the config. enable and
// disable are idempotent; enable also self-heals a stale executable path.
func (a *app) reconcileLogin() {
	var err error
	if a.settings.LaunchAtLogin {
		err = enableLoginItem()
	} else {
		err = disableLoginItem()
	}
	if err != nil {
		slog.Error("failed to reconcile launch at login", "enabled", a.settings.LaunchAtLogin, "error", err)
	}
}

// bindHotkey registers the configured shortcut, replacing any current one. On
// failure it keeps the app running (the menu Convert item still works) and
// marks the shortcut label unavailable.
func (a *app) bindHotkey() {
	if a.unbindHotkey != nil {
		a.unbindHotkey()
		a.unbindHotkey = nil
	}
	unbind, err := registerHotkey(a.settings.Hotkey, a.convert)
	if err != nil {
		slog.Error("failed to register global shortcut", "hotkey", a.settings.Hotkey, "error", err)
		a.mShortcut.SetTitle(shortcutLabel(a.settings.Hotkey) + " (unavailable)")
		return
	}
	a.unbindHotkey = unbind
	a.mShortcut.SetTitle(shortcutLabel(a.settings.Hotkey))
}

func (a *app) loop(mEditShortcut, mReload, mLogin, mQuit *systray.MenuItem) {
	for {
		select {
		case <-a.mConvert.ClickedCh:
			a.convert()
		case <-mEditShortcut.ClickedCh:
			if a.cfgPath == "" {
				slog.Error("cannot open config file: config path is unavailable")
				break
			}
			if err := openConfigFile(a.cfgPath); err != nil {
				slog.Error("failed to open config file", "path", a.cfgPath, "error", err)
			}
		case <-mReload.ClickedCh:
			a.reloadConfig(mLogin)
		case <-mLogin.ClickedCh:
			a.toggleLogin(mLogin)
		case <-mQuit.ClickedCh:
			if a.unbindHotkey != nil {
				a.unbindHotkey()
			}
			systray.Quit()
			return
		}
	}
}

// reloadConfig re-reads the file and applies changes live: it re-registers the
// shortcut only if it changed and reconciles launch-at-login to the new value.
func (a *app) reloadConfig(mLogin *systray.MenuItem) {
	defaults := config.Config{Hotkey: defaultHotkey, LaunchAtLogin: loginItemPresent()}
	cfg, _, err := config.Load(a.cfgPath, defaults)
	if err != nil {
		slog.Error("failed to reload config", "error", err)
		return
	}
	hotkeyChanged := cfg.Hotkey != a.settings.Hotkey
	a.settings = cfg

	// Rebind when the hotkey changed, or when no shortcut is currently active
	// (a previous registration failed), so Reload can recover from a conflict
	// without the user having to edit the hotkey value.
	if hotkeyChanged || a.unbindHotkey == nil {
		a.bindHotkey()
	} else {
		a.mShortcut.SetTitle(shortcutLabel(a.settings.Hotkey))
	}

	a.reconcileLogin()
	a.reflectLogin(mLogin)
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

// toggleLogin flips launch-at-login, persists it to the config (the source of
// truth), applies it to the OS mechanism, and reflects the resulting state.
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

	// Persist the actual on-disk autostart state so the config (the source of
	// truth) never diverges from reality when the OS call fails.
	a.settings.LaunchAtLogin = loginItemEnabled()
	if a.cfgPath != "" {
		if err := config.Save(a.cfgPath, a.settings); err != nil {
			slog.Error("failed to save config", "error", err)
		}
	}
	a.reflectLogin(mLogin)
}

// reflectLogin sets the checkbox to the actual on-disk autostart state.
func (a *app) reflectLogin(mLogin *systray.MenuItem) {
	if loginItemEnabled() {
		mLogin.Check()
	} else {
		mLogin.Uncheck()
	}
}
