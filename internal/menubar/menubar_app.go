//go:build darwin || windows

package menubar

import (
	"errors"
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

	// readyGrace bounds how long Run waits for systray's onReady goroutine to be
	// scheduled before concluding the status bar item was never created.
	readyGrace = 1 * time.Second
)

// statusUI is the slice of the status-bar surface the flash logic drives. It is
// an interface so the flash generation guards can be tested without a live tray.
type statusUI interface {
	setIcon(icon []byte)
	setConvertTitle(title string)
	setConvertEnabled(enabled bool)
}

// systrayUI drives the real status bar.
type systrayUI struct {
	convert *systray.MenuItem
}

func (u systrayUI) setIcon(icon []byte) { systray.SetTemplateIcon(icon, icon) }

func (u systrayUI) setConvertTitle(title string) { u.convert.SetTitle(title) }

func (u systrayUI) setConvertEnabled(enabled bool) {
	if enabled {
		u.convert.Enable()
		return
	}
	u.convert.Disable()
}

// app holds the mutable menu state shared between the click loop and timers.
type app struct {
	cfg Config // build metadata (About item)

	settings config.Config // user config: shortcut + launch-at-login
	cfgPath  string

	ui        statusUI
	mConvert  *systray.MenuItem
	mShortcut *systray.MenuItem

	// unbindHotkey unregisters the current global shortcut; nil when none is
	// active (registration failed).
	unbindHotkey func()

	mu sync.Mutex
	// Flash generation guards: only the most recent flash of a given surface
	// restores it. The icon and the Convert item's title are tracked separately
	// because they are independent surfaces; a shared counter would let a later
	// icon flash cancel a pending title restore and leave the item disabled.
	iconGen  uint64
	titleGen uint64
	// converting is true while a conversion is in flight.
	converting bool
}

// Run starts the menu bar event loop. It must execute on the main goroutine,
// which cobra's RunE already guarantees; do not wrap this call in `go`.
func Run(cfg Config) error {
	a := &app{cfg: cfg}

	started := make(chan struct{})
	var once sync.Once
	systray.Run(func() {
		once.Do(func() { close(started) })
		a.onReady()
	}, func() {})

	// systray.Run blocks until Quit. Returning without onReady ever having run
	// means the status bar item was never created, so report it instead of
	// exiting 0: this command is what the LaunchAgent and the Windows Run key
	// invoke, and a broken autostart must not look like a clean run.
	//
	// systray runs onReady on a goroutine of its own and neither the macOS nor
	// the Windows backend waits for it, so a fast quit can return from Run
	// before that goroutine is scheduled. Allow a short grace period: if the
	// tray did come up, the goroutine is already runnable and arrives at once;
	// if it never came up, nothing will ever release it and we fall through.
	select {
	case <-started:
		return nil
	case <-time.After(readyGrace):
		return errors.New("menu bar failed to start: the status bar item was never created")
	}
}

func (a *app) onReady() {
	systray.SetTemplateIcon(iconNormal, iconNormal)
	systray.SetTooltip("md-paste")

	// Load user config first: it decides the shortcut and the launch-at-login
	// state (migrating any existing OS autostart into the file on first run).
	a.loadSettings()

	a.mConvert = systray.AddMenuItem(convertLabel, "Convert the current clipboard content to Markdown")
	a.ui = systrayUI{convert: a.mConvert}

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

	var backup func() (string, error)
	if path != "" {
		backup = func() (string, error) { return config.Backup(path) }
	}

	settings, write := resolveLoad(cfg, existed, err, defaults, backup)
	a.settings = settings

	if write && path != "" {
		if err := config.Save(path, settings); err != nil {
			slog.Error("failed to write config", "error", err)
		}
	}

	a.reconcileLogin()
}

// resolveLoad decides which settings to use and whether the config file should
// be rewritten from them.
//
// On a load error the file is unreadable or corrupt. Rewriting it stops every
// later launch from silently discarding the user's settings, but the file is
// edited by hand through the "Edit shortcut..." item, so the original must be
// preserved first: a YAML typo must not destroy their settings. If it cannot be
// preserved, nothing is rewritten and this session simply runs on defaults —
// overwriting the only copy would be exactly the loss the backup exists to
// prevent.
func resolveLoad(
	cfg config.Config,
	existed bool,
	loadErr error,
	defaults config.Config,
	backup func() (string, error),
) (settings config.Config, write bool) {
	if loadErr == nil {
		return cfg, !existed
	}

	slog.Error("failed to load config", "error", loadErr)
	if backup == nil {
		return defaults, false
	}

	path, err := backup()
	if err != nil {
		slog.Error("leaving the unreadable config in place, it could not be preserved", "error", err)
		return defaults, false
	}

	slog.Warn("moved the unreadable config aside and restored defaults", "backup", path)
	return defaults, true
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
		if quit := a.handleNextClick(mEditShortcut, mReload, mLogin, mQuit); quit {
			return
		}
	}
}

// handleNextClick waits for one menu click, handles it, and reports whether the
// app should quit. A panic in a handler is contained here rather than killing
// the resident app: reloadConfig unmarshals a file the user edits by hand, and
// the launch-at-login helpers call into the OS. Recovering per click rather than
// around the whole loop keeps the menu responsive afterwards.
func (a *app) handleNextClick(mEditShortcut, mReload, mLogin, mQuit *systray.MenuItem) (quit bool) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("menu bar action panicked", "panic", r)
		}
	}()

	select {
	case <-a.mConvert.ClickedCh:
		a.convert()
	case <-mEditShortcut.ClickedCh:
		if a.cfgPath == "" {
			slog.Error("cannot open config file: config path is unavailable")
			return false
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
		return true
	}
	return false
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

// beginConvert reports whether the caller may start a conversion. Conversions
// are dropped rather than queued while one is in flight: the global shortcut and
// the menu item both reach convert, and running two at once would drive the
// system clipboard concurrently.
func (a *app) beginConvert() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.converting {
		return false
	}
	a.converting = true
	return true
}

func (a *app) endConvert() {
	a.mu.Lock()
	a.converting = false
	a.mu.Unlock()
}

func (a *app) convert() {
	if !a.beginConvert() {
		slog.Warn("a conversion is already in progress, ignoring this request")
		return
	}
	defer a.endConvert()

	// The global shortcut runs conversions on a background goroutine, where an
	// unrecovered panic anywhere in the HTML parser or the Markdown converter
	// would take down the whole resident app. Degrade to a failed conversion.
	defer func() {
		if r := recover(); r != nil {
			slog.Error("menu bar conversion panicked", "panic", r)
			a.flashError("Conversion failed")
		}
	}()

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

// flash applies a transient status-bar state and schedules its restore. gen is
// the generation counter for the surface being changed: every flash bumps it, so
// only the most recent one restores. Without this an earlier flash's timer would
// restore while a later flash is still meant to be showing.
func (a *app) flash(gen *uint64, apply, restore func(), d time.Duration) {
	a.mu.Lock()
	*gen++
	mine := *gen
	a.mu.Unlock()

	// Neither apply nor restore runs under a.mu, deliberately. Every systray call
	// on macOS blocks on a main-thread round-trip that AppKit does not service
	// while a status menu is open, so holding a.mu across one would block
	// beginConvert for as long as the user keeps the menu open. A panic in apply
	// would be worse still: convert's recover handler calls flashError, which
	// would block on the same mutex forever and wedge the app permanently.
	//
	// The cost is a narrow window in which a flash starting just as an older
	// restore fires can have its message wiped early. That is cosmetic: the
	// newer flash's own timer still runs, so the resting state is always right.
	// Ordering between two flashes of the same surface is guaranteed anyway by
	// beginConvert, which lets only one conversion run at a time.
	apply()

	time.AfterFunc(d, func() {
		a.mu.Lock()
		latest := *gen == mine
		a.mu.Unlock()

		if latest {
			restore()
		}
	})
}

// flashSuccess swaps the status icon to the checkmark for a moment.
func (a *app) flashSuccess() {
	a.flash(&a.iconGen,
		func() { a.ui.setIcon(iconCheck) },
		func() { a.ui.setIcon(iconNormal) },
		flashDuration,
	)
}

// flashError shows a transient, non-modal error by retitling and disabling the
// Convert item, then restoring it. The global shortcut reaches convert without
// touching the menu item, so error flashes really can overlap; the generation
// guard is what keeps the earlier one from restoring too soon.
func (a *app) flashError(msg string) {
	a.flash(&a.titleGen,
		func() {
			a.ui.setConvertTitle(msg)
			a.ui.setConvertEnabled(false)
		},
		func() {
			a.ui.setConvertTitle(convertLabel)
			a.ui.setConvertEnabled(true)
		},
		errorRestoreDelay,
	)
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
