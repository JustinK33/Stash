package hotkey

import (
	"testing"

	"stash/internal/keybind"
)

func TestRegisterUnregisterRegister(t *testing.T) {
	binding := keybind.Binding{Key: "9", Control: true, Option: true, Shift: true}
	for i := 0; i < 2; i++ {
		if err := Register(binding, func() {}); err != nil {
			t.Fatalf("register %d: %v", i, err)
		}
		Unregister()
	}
}

func TestRegisterRequiresModifier(t *testing.T) {
	if err := Register(keybind.Binding{Key: "9"}, func() {}); err == nil {
		Unregister()
		t.Fatal("expected error for binding without modifiers")
	}
}
