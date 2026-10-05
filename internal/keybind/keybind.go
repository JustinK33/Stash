package keybind

import (
	"runtime"
	"strings"
)

// ModifierLabels names the modifier keys the way the current platform does.
// Shortcut is the in-app shortcut key, Command on macOS and Ctrl elsewhere.
type ModifierLabels struct {
	Command, Control, Option, Shift, Shortcut string
}

func Labels() ModifierLabels {
	return labelsFor(runtime.GOOS)
}

func labelsFor(goos string) ModifierLabels {
	if goos == "windows" {
		return ModifierLabels{Command: "Win", Control: "Ctrl", Option: "Alt", Shift: "Shift", Shortcut: "Ctrl"}
	}
	return ModifierLabels{Command: "Command", Control: "Control", Option: "Option", Shift: "Shift", Shortcut: "⌘"}
}

type Binding struct {
	Key     string `json:"key"`
	Command bool   `json:"command"`
	Control bool   `json:"control"`
	Option  bool   `json:"option"`
	Shift   bool   `json:"shift"`
}

func Default() Binding {
	return Binding{
		Key:     "0",
		Control: true,
		Option:  true,
	}
}

func (binding Binding) Normalize() Binding {
	binding.Key = strings.ToUpper(strings.TrimSpace(binding.Key))
	if binding.Key == "" || !IsSupportedKey(binding.Key) {
		binding.Key = Default().Key
	}
	return binding
}

func (binding Binding) HasModifier() bool {
	return binding.Command || binding.Control || binding.Option || binding.Shift
}

func (binding Binding) Display() string {
	return binding.display(Labels())
}

func (binding Binding) display(labels ModifierLabels) string {
	binding = binding.Normalize()

	parts := make([]string, 0, 5)
	if binding.Command {
		parts = append(parts, labels.Command)
	}
	if binding.Control {
		parts = append(parts, labels.Control)
	}
	if binding.Option {
		parts = append(parts, labels.Option)
	}
	if binding.Shift {
		parts = append(parts, labels.Shift)
	}
	parts = append(parts, binding.Key)
	return strings.Join(parts, " + ")
}

func Keys() []string {
	keys := []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	}
	for letter := 'A'; letter <= 'Z'; letter++ {
		keys = append(keys, string(letter))
	}
	return keys
}

func IsSupportedKey(key string) bool {
	key = strings.ToUpper(strings.TrimSpace(key))
	for _, supported := range Keys() {
		if key == supported {
			return true
		}
	}
	return false
}
