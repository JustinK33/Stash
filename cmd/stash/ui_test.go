package main

import (
	"strings"
	"testing"
	"time"

	"stash/internal/store"
)

func TestFilterSnippets(t *testing.T) {
	snippets := []store.Snippet{{Text: "Write tests"}, {Text: "fix the README"}, {Text: "ship it"}}

	if got := filterSnippets(snippets, "  "); len(got) != 3 {
		t.Fatalf("blank query should keep everything, got %d", len(got))
	}
	got := filterSnippets(snippets, " readme ")
	if len(got) != 1 || got[0].Text != "fix the README" {
		t.Fatalf("expected case-insensitive match on README, got %+v", got)
	}
	if got := filterSnippets(snippets, "nope"); len(got) != 0 {
		t.Fatalf("expected no matches, got %+v", got)
	}
}

func TestTruncateSnippet(t *testing.T) {
	if got, cut := truncateSnippet("short", 6, 400); cut || got != "short" {
		t.Fatalf("short text should be untouched, got %q %v", got, cut)
	}
	got, cut := truncateSnippet("a\nb\nc\nd", 2, 400)
	if !cut || got != "a\nb…" {
		t.Fatalf("expected two lines, got %q %v", got, cut)
	}
	got, cut = truncateSnippet(strings.Repeat("é", 10), 6, 4)
	if !cut || got != "éééé…" {
		t.Fatalf("expected rune-safe cut, got %q %v", got, cut)
	}
}

func TestFormatTimestamp(t *testing.T) {
	if got := formatTimestamp(time.Time{}); got != "" {
		t.Fatalf("zero time should be blank, got %q", got)
	}
	when := time.Date(2026, 10, 3, 2, 14, 0, 0, time.Local)
	if got := formatTimestamp(when); got != "Oct 3, 2026, 2:14 AM" {
		t.Fatalf("got %q", got)
	}
}

func TestCharacterCount(t *testing.T) {
	cases := map[string]string{"": "0 characters", "a": "1 character", strings.Repeat("x", 12345): "12,345 characters"}
	for in, want := range cases {
		if got := characterCount(in); got != want {
			t.Fatalf("characterCount(%d runes) = %q, want %q", len(in), got, want)
		}
	}
}
