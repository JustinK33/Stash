package hotkey

/*
#cgo LDFLAGS: -framework Carbon
#include <Carbon/Carbon.h>

extern void stashHotkeyPressed();

static EventHotKeyRef stashHotKeyRef = NULL;
static EventHandlerRef stashHandlerRef = NULL;

static OSStatus stashHotKeyHandler(EventHandlerCallRef nextHandler, EventRef event, void *userData) {
	if (GetEventClass(event) == kEventClassKeyboard && GetEventKind(event) == kEventHotKeyPressed) {
		stashHotkeyPressed();
	}
	return noErr;
}

static OSStatus registerStashHotkey(UInt32 keyCode, UInt32 modifiers) {
	EventTypeSpec eventType;
	eventType.eventClass = kEventClassKeyboard;
	eventType.eventKind = kEventHotKeyPressed;

	EventHotKeyID hotKeyID;
	hotKeyID.signature = 'STSH';
	hotKeyID.id = 1;

	OSStatus status = InstallEventHandler(GetApplicationEventTarget(), stashHotKeyHandler, 1, &eventType, NULL, &stashHandlerRef);
	if (status != noErr) {
		return status;
	}

	status = RegisterEventHotKey(keyCode, modifiers, hotKeyID, GetApplicationEventTarget(), 0, &stashHotKeyRef);
	if (status != noErr) {
		if (stashHandlerRef != NULL) {
			RemoveEventHandler(stashHandlerRef);
			stashHandlerRef = NULL;
		}
		return status;
	}

	return noErr;
}

static void unregisterStashHotkey() {
	if (stashHotKeyRef != NULL) {
		UnregisterEventHotKey(stashHotKeyRef);
		stashHotKeyRef = NULL;
	}
	if (stashHandlerRef != NULL) {
		RemoveEventHandler(stashHandlerRef);
		stashHandlerRef = NULL;
	}
}
*/
import "C"

import (
	"fmt"
	"sync"

	"stash/internal/keybind"
)

var (
	callbackMu sync.RWMutex
	callback   func()
)

func Register(binding keybind.Binding, onPressed func()) error {
	keyCode, err := keyCode(binding.Key)
	if err != nil {
		return err
	}

	modifiers := modifiers(binding)
	if modifiers == 0 {
		return fmt.Errorf("register global shortcut: at least one modifier is required")
	}

	Unregister()

	callbackMu.Lock()
	callback = onPressed
	callbackMu.Unlock()

	if status := C.registerStashHotkey(C.UInt32(keyCode), C.UInt32(modifiers)); status != C.noErr {
		callbackMu.Lock()
		callback = nil
		callbackMu.Unlock()
		return fmt.Errorf("register global shortcut: status %d", int(status))
	}
	return nil
}

func Unregister() {
	C.unregisterStashHotkey()

	callbackMu.Lock()
	callback = nil
	callbackMu.Unlock()
}

//export stashHotkeyPressed
func stashHotkeyPressed() {
	callbackMu.RLock()
	onPressed := callback
	callbackMu.RUnlock()
	if onPressed != nil {
		onPressed()
	}
}

func keyCode(key string) (uint32, error) {
	switch (keybind.Binding{Key: key}).Normalize().Key {
	case "0":
		return C.kVK_ANSI_0, nil
	case "1":
		return C.kVK_ANSI_1, nil
	case "2":
		return C.kVK_ANSI_2, nil
	case "3":
		return C.kVK_ANSI_3, nil
	case "4":
		return C.kVK_ANSI_4, nil
	case "5":
		return C.kVK_ANSI_5, nil
	case "6":
		return C.kVK_ANSI_6, nil
	case "7":
		return C.kVK_ANSI_7, nil
	case "8":
		return C.kVK_ANSI_8, nil
	case "9":
		return C.kVK_ANSI_9, nil
	case "A":
		return C.kVK_ANSI_A, nil
	case "B":
		return C.kVK_ANSI_B, nil
	case "C":
		return C.kVK_ANSI_C, nil
	case "D":
		return C.kVK_ANSI_D, nil
	case "E":
		return C.kVK_ANSI_E, nil
	case "F":
		return C.kVK_ANSI_F, nil
	case "G":
		return C.kVK_ANSI_G, nil
	case "H":
		return C.kVK_ANSI_H, nil
	case "I":
		return C.kVK_ANSI_I, nil
	case "J":
		return C.kVK_ANSI_J, nil
	case "K":
		return C.kVK_ANSI_K, nil
	case "L":
		return C.kVK_ANSI_L, nil
	case "M":
		return C.kVK_ANSI_M, nil
	case "N":
		return C.kVK_ANSI_N, nil
	case "O":
		return C.kVK_ANSI_O, nil
	case "P":
		return C.kVK_ANSI_P, nil
	case "Q":
		return C.kVK_ANSI_Q, nil
	case "R":
		return C.kVK_ANSI_R, nil
	case "S":
		return C.kVK_ANSI_S, nil
	case "T":
		return C.kVK_ANSI_T, nil
	case "U":
		return C.kVK_ANSI_U, nil
	case "V":
		return C.kVK_ANSI_V, nil
	case "W":
		return C.kVK_ANSI_W, nil
	case "X":
		return C.kVK_ANSI_X, nil
	case "Y":
		return C.kVK_ANSI_Y, nil
	case "Z":
		return C.kVK_ANSI_Z, nil
	default:
		return 0, fmt.Errorf("unsupported shortcut key %q", key)
	}
}

func modifiers(binding keybind.Binding) uint32 {
	var modifiers uint32
	if binding.Command {
		modifiers |= C.cmdKey
	}
	if binding.Control {
		modifiers |= C.controlKey
	}
	if binding.Option {
		modifiers |= C.optionKey
	}
	if binding.Shift {
		modifiers |= C.shiftKey
	}
	return modifiers
}
