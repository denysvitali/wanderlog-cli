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

func TestGetJournalStopPolylines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/tripPlans/journalStopPolylines" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"polylines":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	_, err := client.GetJournalStopPolylines(JournalStopPolylinesRequest{})
	if err != nil {
		t.Fatalf("GetJournalStopPolylines: %v", err)
	}
}

func TestGetTripDistinction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/distinction") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"distinction":"community-pick"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetTripDistinction("mytrip")
	if err != nil {
		t.Fatalf("GetTripDistinction: %v", err)
	}
	if resp.Distinction != "community-pick" {
		t.Errorf("unexpected distinction: %s", resp.Distinction)
	}
}

func TestSetTripDistinction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/distinction") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.SetTripDistinction("mytrip", "community-pick"); err != nil {
		t.Fatalf("SetTripDistinction: %v", err)
	}
}

func TestCreateGuideFromTripPlan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/createGuideFromTripPlan") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"guide":{"id":1}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.CreateGuideFromTripPlan("mytrip")
	if err != nil {
		t.Fatalf("CreateGuideFromTripPlan: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
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
