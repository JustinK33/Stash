package keybind

import "testing"

func TestBindingDisplay(t *testing.T) {
	cases := []struct {
		name    string
		binding Binding
		want    string
	}{
		{name: "default", binding: Default(), want: "Control + Option + 0"},
		{name: "single modifier", binding: Binding{Key: "S", Command: true}, want: "Command + S"},
		{
			name:    "modifier order",
			binding: Binding{Key: "A", Shift: true, Option: true, Control: true, Command: true},
			want:    "Command + Control + Option + Shift + A",
		},
		{name: "unsupported key falls back", binding: Binding{Key: "F13", Command: true}, want: "Command + 0"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.binding.Display(); got != testCase.want {
				t.Errorf("got %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestBindingNormalize(t *testing.T) {
	cases := []struct {
		name string
		key  string
		want string
	}{
		{name: "upcases", key: "s", want: "S"},
		{name: "trims", key: "  s  ", want: "S"},
		{name: "keeps digits", key: "7", want: "7"},
		{name: "rejects multi character", key: "F13", want: Default().Key},
		{name: "rejects empty", key: "", want: Default().Key},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := (Binding{Key: testCase.key}).Normalize().Key; got != testCase.want {
				t.Errorf("got %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestBindingNormalizePreservesModifiers(t *testing.T) {
	binding := Binding{Key: "s", Command: true, Shift: true}.Normalize()

	if !binding.Command || !binding.Shift || binding.Control || binding.Option {
		t.Errorf("got %+v, want only Command and Shift set", binding)
	}
}

func TestBindingHasModifier(t *testing.T) {
	if (Binding{Key: "S"}).HasModifier() {
		t.Error("got true for a binding with no modifiers, want false")
	}
	if !(Binding{Key: "S", Option: true}).HasModifier() {
		t.Error("got false for a binding with Option set, want true")
	}
}

func TestIsSupportedKey(t *testing.T) {
	cases := []struct {
		key  string
		want bool
	}{
		{key: "0", want: true},
		{key: "Z", want: true},
		{key: "s", want: true},
		{key: " s ", want: true},
		{key: "F13", want: false},
		{key: "", want: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.key, func(t *testing.T) {
			if got := IsSupportedKey(testCase.key); got != testCase.want {
				t.Errorf("got %t, want %t", got, testCase.want)
			}
		})
	}
}
