package ui

import "strings"

// SafeText removes terminal control characters from untrusted API text and
// flattens line breaks so a trip title or place name cannot inject terminal
// escape sequences or spoof additional output lines.
func SafeText(value string) string {
	var b strings.Builder
	b.Grow(len(value))
	spacePending := false
	for _, r := range value {
		if r == '\n' || r == '\r' || r == '\t' {
			spacePending = b.Len() > 0
			continue
		}
		if r < 0x20 || (r >= 0x7f && r <= 0x9f) {
			continue
		}
		if spacePending && r != ' ' {
			b.WriteByte(' ')
		}
		spacePending = false
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

// MarkdownInline additionally escapes structural Markdown characters for data
// inserted into headings and list values.
func MarkdownInline(value string) string {
	value = SafeText(value)
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
		"<", "&lt;",
		">", "&gt;",
		"#", "\\#",
		"|", "\\|",
	)
	return replacer.Replace(value)
}
