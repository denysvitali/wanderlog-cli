package cmd

import "testing"

func TestDecodeCommandJSON(t *testing.T) {
	payload, err := decodeCommandJSON(`{"places":[1,2]}`, "")
	if err != nil {
		t.Fatalf("decodeCommandJSON: %v", err)
	}
	if payload == nil {
		t.Fatal("expected payload")
	}
	if _, err := decodeCommandJSON("", ""); err == nil {
		t.Fatal("expected missing input error")
	}
	if _, err := decodeCommandJSON(`{}`, "also.json"); err == nil {
		t.Fatal("expected mutually exclusive input error")
	}
	if _, err := decodeCommandJSON(`{`, ""); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestInsightsCommandsRegistered(t *testing.T) {
	for _, path := range [][]string{{"trips", "analytics"}, {"trips", "optimize-route"}, {"trips", "recommendations"}} {
		if _, _, err := rootCmd.Find(path); err != nil {
			t.Fatalf("command %v not registered: %v", path, err)
		}
	}
}
