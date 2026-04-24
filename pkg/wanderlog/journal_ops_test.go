package wanderlog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetViewOnlyJournal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || !strings.HasPrefix(r.URL.Path, "/tripPlans/viewOnlyJournal/") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"journal":{"id":1}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetViewOnlyJournal("abc")
	if err != nil {
		t.Fatalf("GetViewOnlyJournal: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestGetTripExpensesCSV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/expensesAsCSV") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("date,amount,description\n2026-04-24,100,lunch"))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	csv, err := client.GetTripExpensesCSV("mytrip")
	if err != nil {
		t.Fatalf("GetTripExpensesCSV: %v", err)
	}
	if !strings.Contains(string(csv), "lunch") {
		t.Errorf("unexpected csv: %s", csv)
	}
}

func TestGetTripUpdateRequired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("clientSchemaVersion") == "" {
			t.Error("expected clientSchemaVersion query param")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"updateRequired":false}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetTripUpdateRequired("abc")
	if err != nil {
		t.Fatalf("GetTripUpdateRequired: %v", err)
	}
	if resp.UpdateRequired {
		t.Error("expected updateRequired=false")
	}
}

func TestRegisterTripView(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/registerView") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.RegisterTripView("abc"); err != nil {
		t.Fatalf("RegisterTripView: %v", err)
	}
}
