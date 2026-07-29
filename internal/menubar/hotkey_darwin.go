//go:build darwin

package menubar

/*
#cgo LDFLAGS: -framework Carbon

#include <Carbon/Carbon.h>
#include <dispatch/dispatch.h>
#include <pthread.h>
#include <stdint.h>

// Implemented in Go and exported via cgo.
extern void goHotkeyFired(void);

static EventHandlerRef gHandlerRef = NULL;
static EventHotKeyRef  gHotKeyRef  = NULL;

static OSStatus hotkeyHandler(EventHandlerCallRef next, EventRef evt, void *ud) {
	(void)next; (void)evt; (void)ud;
	goHotkeyFired();
	return noErr;
}

// ensureHandler installs the application-level hot key handler once.
static void ensureHandler(void) {
	if (gHandlerRef != NULL) return;
	EventTypeSpec spec;
	spec.eventClass = kEventClassKeyboard;
	spec.eventKind  = kEventHotKeyPressed;
	InstallEventHandler(GetApplicationEventTarget(), NewEventHandlerUPP(hotkeyHandler),
	                    1, &spec, NULL, &gHandlerRef);
}

// doRegister replaces the current hot key. Must run on the main thread.
static OSStatus doRegister(uint32_t keycode, uint32_t modifiers) {
	ensureHandler();
	if (gHotKeyRef != NULL) {
		UnregisterEventHotKey(gHotKeyRef);
		gHotKeyRef = NULL;
	}
	EventHotKeyID hkID;
	hkID.signature = 0x6d647073; // 'mdps'
	hkID.id = 1;
	return RegisterEventHotKey(keycode, modifiers, hkID,
	                           GetApplicationEventTarget(), 0, &gHotKeyRef);
}

static void doUnregister(void) {
	if (gHotKeyRef != NULL) {
		UnregisterEventHotKey(gHotKeyRef);
		gHotKeyRef = NULL;
	}
}

// registerHotKey registers the hot key on the main thread. It runs inline when
// already on the main thread (the systray onReady callback) and otherwise hops
// to the main queue (the reload path, called from a goroutine), avoiding a
// dispatch_sync deadlock on the main thread.
static int registerHotKey(uint32_t keycode, uint32_t modifiers) {
	if (pthread_main_np() != 0) {
		return (int)doRegister(keycode, modifiers);
	}
	__block OSStatus status = -1;
	dispatch_sync(dispatch_get_main_queue(), ^{
		status = doRegister(keycode, modifiers);
	});
	return (int)status;
}

static void unregisterHotKey(void) {
	if (pthread_main_np() != 0) {
		doUnregister();
		return;
	}
	dispatch_sync(dispatch_get_main_queue(), ^{
		doUnregister();
	});
}
*/
import "C"

import (
	"fmt"
	"sync"
)

// defaultHotkey is the shortcut used when the config has none.
const defaultHotkey = "cmd+opt+m"

// Carbon modifier masks (from HIToolbox Events.h).
var macModifiers = map[string]uint32{
	"cmd": 0x0100, "command": 0x0100, "⌘": 0x0100,
	"shift": 0x0200, "⇧": 0x0200,
	"opt": 0x0800, "option": 0x0800, "alt": 0x0800, "⌥": 0x0800,
	"ctrl": 0x1000, "control": 0x1000, "⌃": 0x1000,
}

// Carbon virtual key codes (from HIToolbox Events.h, kVK_ANSI_*).
var macKeyCodes = map[string]uint32{
	"a": 0x00, "b": 0x0B, "c": 0x08, "d": 0x02, "e": 0x0E, "f": 0x03, "g": 0x05,
	"h": 0x04, "i": 0x22, "j": 0x26, "k": 0x28, "l": 0x25, "m": 0x2E, "n": 0x2D,
	"o": 0x1F, "p": 0x23, "q": 0x0C, "r": 0x0F, "s": 0x01, "t": 0x11, "u": 0x20,
	"v": 0x09, "w": 0x0D, "x": 0x07, "y": 0x10, "z": 0x06,
	"0": 0x1D, "1": 0x12, "2": 0x13, "3": 0x14, "4": 0x15, "5": 0x17, "6": 0x16,
	"7": 0x1A, "8": 0x1C, "9": 0x19,
	"space": 0x31,
	"f1":    0x7A, "f2": 0x78, "f3": 0x63, "f4": 0x76, "f5": 0x60, "f6": 0x61,
	"f7": 0x62, "f8": 0x64, "f9": 0x65, "f10": 0x6D, "f11": 0x67, "f12": 0x6F,
}

var (
	hotkeyMu sync.Mutex
	hotkeyFn func()
)

//export goHotkeyFired
func goHotkeyFired() {
	hotkeyMu.Lock()
	fn := hotkeyFn
	hotkeyMu.Unlock()
	if fn != nil {
		// Run off the Carbon handler (main thread) so conversion work never
		// blocks event delivery.
		go fn()
	}
}

// registerHotkey registers a global shortcut that calls onFire when pressed. It
// returns an unregister closure. Only one hot key is active per process.
func registerHotkey(spec string, onFire func()) (func(), error) {
	mods, key, err := splitHotkey(spec)
	if err != nil {
		return nil, err
	}
	mask, err := carbonModifiers(mods)
	if err != nil {
		return nil, err
	}
	code, ok := macKeyCodes[key]
	if !ok {
		return nil, fmt.Errorf("unknown key %q in hotkey %q", key, spec)
	}

	hotkeyMu.Lock()
	hotkeyFn = onFire
	hotkeyMu.Unlock()

	if status := C.registerHotKey(C.uint32_t(code), C.uint32_t(mask)); status != 0 {
		hotkeyMu.Lock()
		hotkeyFn = nil
		hotkeyMu.Unlock()
		return nil, fmt.Errorf("failed to register hotkey %q (RegisterEventHotKey status %d)", spec, int(status))
	}

	return func() {
		C.unregisterHotKey()
		hotkeyMu.Lock()
		hotkeyFn = nil
		hotkeyMu.Unlock()
	}, nil
}

// carbonModifiers ORs the modifier tokens into a Carbon modifier mask.
func carbonModifiers(mods []string) (uint32, error) {
	var mask uint32
	for _, m := range mods {
		bit, ok := macModifiers[m]
		if !ok {
			return 0, fmt.Errorf("unknown modifier %q", m)
		}
		mask |= bit
	}
	if mask == 0 {
		return 0, fmt.Errorf("hotkey needs at least one modifier")
	}
	return mask, nil
}
