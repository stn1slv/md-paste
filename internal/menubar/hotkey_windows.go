//go:build windows

package menubar

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

// defaultHotkey is the shortcut used when the config has none.
const defaultHotkey = "ctrl+alt+m"

// Win32 modifier flags for RegisterHotKey (winuser.h).
const (
	modAlt      = 0x0001
	modControl  = 0x0002
	modShift    = 0x0004
	modWin      = 0x0008
	modNoRepeat = 0x4000

	wmHotkey = 0x0312 // WM_HOTKEY
	wmApp    = 0x8000 // WM_APP: our "unregister and stop" signal
	hotkeyID = 1
)

var winModifiers = map[string]uint32{
	"ctrl": modControl, "control": modControl,
	"alt": modAlt, "opt": modAlt, "option": modAlt,
	"shift": modShift,
	"win":   modWin, "super": modWin, "meta": modWin, "cmd": modWin,
}

// winKeyCodes maps key names to Win32 virtual-key codes (letters, digits, and
// function keys are contiguous, so build them programmatically).
var winKeyCodes = buildWinKeyCodes()

func buildWinKeyCodes() map[string]uint32 {
	m := map[string]uint32{"space": 0x20}
	for c := 'a'; c <= 'z'; c++ {
		m[string(c)] = uint32(0x41 + (c - 'a')) // VK_A..VK_Z
	}
	for d := '0'; d <= '9'; d++ {
		m[string(d)] = uint32(0x30 + (d - '0')) // VK_0..VK_9
	}
	for i := 1; i <= 12; i++ {
		m[fmt.Sprintf("f%d", i)] = uint32(0x70 + (i - 1)) // VK_F1..VK_F12
	}
	return m
}

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterHotKey     = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey   = user32.NewProc("UnregisterHotKey")
	procGetMessageW        = user32.NewProc("GetMessageW")
	procPostThreadMessageW = user32.NewProc("PostThreadMessageW")
	procGetCurrentThreadID = kernel32.NewProc("GetCurrentThreadId")
)

// msg mirrors the Win32 MSG structure for GetMessageW.
type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      struct{ x, y int32 }
}

// registerHotkey registers a global shortcut that calls onFire when pressed.
// RegisterHotKey with a nil window posts WM_HOTKEY to the calling thread, so a
// dedicated OS-locked goroutine owns the registration and its message loop. The
// returned closure signals that goroutine to unregister and exit.
func registerHotkey(spec string, onFire func()) (func(), error) {
	mods, key, err := splitHotkey(spec)
	if err != nil {
		return nil, err
	}
	mask, err := winModifiersMask(mods)
	if err != nil {
		return nil, err
	}
	vk, ok := winKeyCodes[key]
	if !ok {
		return nil, fmt.Errorf("unknown key %q in hotkey %q", key, spec)
	}

	type ready struct {
		tid uint32
		err error
	}
	readyCh := make(chan ready, 1)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		tid, _, _ := procGetCurrentThreadID.Call()
		r, _, callErr := procRegisterHotKey.Call(0, hotkeyID, uintptr(mask|modNoRepeat), uintptr(vk))
		if r == 0 {
			readyCh <- ready{err: fmt.Errorf("failed to register hotkey %q, the combination may already be in use: %w", spec, callErr)}
			return
		}
		// Release the hotkey on every exit path (the wmApp stop signal, WM_QUIT,
		// or a GetMessage error), not only the explicit stop.
		defer func() { _, _, _ = procUnregisterHotKey.Call(0, hotkeyID) }()
		readyCh <- ready{tid: uint32(tid)} //nolint:gosec // GetCurrentThreadId returns a DWORD (fits uint32)

		var m msg
		for {
			//nolint:gosec // MSG pointer for GetMessageW; ret is a BOOL (-1/0/1)
			r1, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
			if int32(r1) <= 0 { //nolint:gosec // GetMessage BOOL: 0 = WM_QUIT, -1 = error
				return
			}
			switch m.message {
			case wmHotkey:
				// Run off the message loop so conversion never stalls it.
				go onFire()
			case wmApp:
				return
			}
		}
	}()

	res := <-readyCh
	if res.err != nil {
		return nil, res.err
	}
	tid := res.tid
	return func() {
		_, _, _ = procPostThreadMessageW.Call(uintptr(tid), wmApp, 0, 0)
	}, nil
}

// winModifiersMask ORs the modifier tokens into a Win32 modifier mask.
func winModifiersMask(mods []string) (uint32, error) {
	var mask uint32
	for _, m := range mods {
		bit, ok := winModifiers[m]
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
