package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = original })
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = original
	var output bytes.Buffer
	if _, err := io.Copy(&output, r); err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func TestTripListOutputSanitizesUntrustedText(t *testing.T) {
	trips := &wanderlog.UserTripsResponse{}
	if err := json.Unmarshal([]byte(`{"data":[{"title":"unsafe\u001b[31m\nheading","key":"key\rspoof","type":"# type"}]}`), trips); err != nil {
		t.Fatal(err)
	}

	pretty := captureStdout(t, func() { tripsListPretty(trips) })
	if strings.Contains(pretty, "\x1b[31m") || strings.Contains(pretty, "\nheading") {
		t.Fatalf("pretty output contains injected control text: %q", pretty)
	}
	markdown := captureStdout(t, func() { tripsListMarkdown(trips) })
	if strings.Contains(markdown, "\x1b[31m") || strings.Contains(markdown, "## unsafe") && strings.Contains(markdown, "\nheading") {
		t.Fatalf("markdown output contains injected structure: %q", markdown)
	}
	if !strings.Contains(markdown, `\# type`) {
		t.Fatalf("markdown type was not escaped: %q", markdown)
	}
}

func TestAnalyticsOutputSanitizesUntrustedText(t *testing.T) {
	output := captureStdout(t, func() {
		printTripAnalytics(&wanderlog.TripAnalytics{
			Title:    "trip\x1b[2J\nspoof",
			TripKey:  "key\rspoof",
			Warnings: []string{"warning\nforged"},
		})
	})
	if strings.Contains(output, "\x1b[2J") || strings.Contains(output, "\nspoof") || strings.Contains(output, "\nforged") {
		t.Fatalf("analytics output contains injected control text: %q", output)
	}
}
