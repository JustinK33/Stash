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

static OSStatus registerStashHotkey() {
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

	status = RegisterEventHotKey(kVK_ANSI_0, controlKey | optionKey, hotKeyID, GetApplicationEventTarget(), 0, &stashHotKeyRef);
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
)

var (
	callbackMu sync.RWMutex
	callback   func()
)

func Register(onPressed func()) error {
	callbackMu.Lock()
	callback = onPressed
	callbackMu.Unlock()

	if status := C.registerStashHotkey(); status != C.noErr {
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
