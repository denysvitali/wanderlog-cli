package ui

import "testing"

func TestSafeTextRemovesTerminalControls(t *testing.T) {
	got := SafeText("safe\x1b]52;c;secret\a\nnext")
	if got != "safe]52;c;secret next" {
		t.Fatalf("SafeText = %q", got)
	}
}

func TestMarkdownInlineEscapesStructure(t *testing.T) {
	got := MarkdownInline("# [Hotel] *great*")
	if got != `\# \[Hotel\] \*great\*` {
		t.Fatalf("MarkdownInline = %q", got)
	}
}
