package keybind

import "strings"

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
	binding = binding.Normalize()

	parts := make([]string, 0, 5)
	if binding.Command {
		parts = append(parts, "Command")
	}
	if binding.Control {
		parts = append(parts, "Control")
	}
	if binding.Option {
		parts = append(parts, "Option")
	}
	if binding.Shift {
		parts = append(parts, "Shift")
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
