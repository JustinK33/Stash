package hotkey

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"stash/internal/keybind"
)

const (
	modAlt      = 0x0001
	modControl  = 0x0002
	modShift    = 0x0004
	modWin      = 0x0008
	modNoRepeat = 0x4000
	wmHotkey    = 0x0312
	wmQuit      = 0x0012
	hotkeyID    = 1
)

var (
	user32                 = syscall.NewLazyDLL("user32.dll")
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procRegisterHotKey     = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey   = user32.NewProc("UnregisterHotKey")
	procGetMessageW        = user32.NewProc("GetMessageW")
	procPostThreadMessageW = user32.NewProc("PostThreadMessageW")
	procGetCurrentThreadId = kernel32.NewProc("GetCurrentThreadId")
)

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      [2]int32
	private uint32
}

// listener is the thread that owns the registration. Windows only lets that
// thread receive WM_HOTKEY and unregister the hotkey, so it lives until Unregister.
var listener struct {
	threadID uintptr
	done     chan struct{}
}

func Register(binding keybind.Binding, onPressed func()) error {
	binding = binding.Normalize()
	if !keybind.IsSupportedKey(binding.Key) {
		return fmt.Errorf("unsupported shortcut key %q", binding.Key)
	}
	modifiers := modifiers(binding)
	if modifiers == 0 {
		return fmt.Errorf("register global shortcut: at least one modifier is required")
	}
	// 0-9 and A-Z have virtual key codes equal to their ASCII codes.
	keyCode := uintptr(binding.Key[0])

	Unregister()

	result := make(chan error, 1)
	threadID := make(chan uintptr, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		tid, _, _ := procGetCurrentThreadId.Call()
		threadID <- tid
		if ok, _, err := procRegisterHotKey.Call(0, hotkeyID, uintptr(modifiers|modNoRepeat), keyCode); ok == 0 {
			result <- fmt.Errorf("register global shortcut: %w", err)
			return
		}
		result <- nil
		defer procUnregisterHotKey.Call(0, hotkeyID)

		var m msg
		for {
			// GetMessageW returns 0 on WM_QUIT and -1 on error.
			if r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0); int32(r) <= 0 {
				return
			}
			if m.message == wmHotkey {
				onPressed()
			}
		}
	}()

	listener.threadID = <-threadID
	if err := <-result; err != nil {
		<-done
		listener.threadID = 0
		return err
	}
	listener.done = done
	return nil
}

func Unregister() {
	if listener.done == nil {
		return
	}
	// The thread may not have a message queue yet; retry until the post lands.
	for {
		if ok, _, _ := procPostThreadMessageW.Call(listener.threadID, wmQuit, 0, 0); ok != 0 {
			break
		}
		runtime.Gosched()
	}
	<-listener.done
	listener.threadID = 0
	listener.done = nil
}

func modifiers(binding keybind.Binding) uint32 {
	var modifiers uint32
	if binding.Command {
		modifiers |= modWin
	}
	if binding.Control {
		modifiers |= modControl
	}
	if binding.Option {
		modifiers |= modAlt
	}
	if binding.Shift {
		modifiers |= modShift
	}
	return modifiers
}
