package store

import "testing"

func TestAddMovesDuplicateToTop(t *testing.T) {
	snippets := Add(nil, "first")
	snippets = Add(snippets, "second")
	snippets = Add(snippets, "first")

	if len(snippets) != 2 {
		t.Fatalf("got %d snippets, want 2: %#v", len(snippets), snippets)
	}
	if snippets[0].Text != "first" {
		t.Errorf("got top snippet %q, want %q", snippets[0].Text, "first")
	}
	if snippets[1].Text != "second" {
		t.Errorf("got second snippet %q, want %q", snippets[1].Text, "second")
	}
}

func TestAddTrimsSurroundingWhitespace(t *testing.T) {
	snippets := Add(nil, "  padded  ")

	if len(snippets) != 1 {
		t.Fatalf("got %d snippets, want 1", len(snippets))
	}
	if snippets[0].Text != "padded" {
		t.Errorf("got %q, want %q", snippets[0].Text, "padded")
	}
}

func TestAddIgnoresBlankText(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{name: "empty", text: ""},
		{name: "spaces", text: "   "},
		{name: "newline and tab", text: "\n\t"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			existing := Add(nil, "keep me")

			snippets := Add(existing, testCase.text)

			if len(snippets) != 1 || snippets[0].Text != "keep me" {
				t.Errorf("got %#v, want the input unchanged", snippets)
			}
		})
	}
}

func TestDeleteRemovesMatchingText(t *testing.T) {
	snippets := Add(Add(nil, "second"), "first")

	snippets = Delete(snippets, "second")

	if len(snippets) != 1 || snippets[0].Text != "first" {
		t.Fatalf("got %#v, want only %q", snippets, "first")
	}
}

func TestDeleteMissingTextIsNoOp(t *testing.T) {
	snippets := Add(nil, "first")

	if remaining := Delete(snippets, "absent"); len(remaining) != 1 {
		t.Errorf("got %d snippets, want 1", len(remaining))
	}
}
