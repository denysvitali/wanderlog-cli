package wanderlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGetUserTrips(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/tripPlans") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[{"id":1,"key":"abc","title":"My Trip"}]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.GetUserTrips()
	if err != nil {
		t.Fatalf("GetUserTrips: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if len(result.Data) != 1 || result.Data[0].Key != "abc" {
		t.Errorf("unexpected trips: %+v", result.Data)
	}
}

func TestGetTripImages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || !strings.Contains(r.URL.Path, "/images") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"images":[{"id":1,"key":"img-key","url":"https://example.com/img.jpg"}]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.GetTripImages("test-key")
	if err != nil {
		t.Fatalf("GetTripImages: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if len(result.Images) != 1 || result.Images[0].URL != "https://example.com/img.jpg" {
		t.Errorf("unexpected images: %+v", result.Images)
	}
}

func TestGetTripPlaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || !strings.Contains(r.URL.Path, "/places") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		b, _ := json.Marshal(TripResponse{
			Success: true,
			TripPlan: Plan{
				Title: "Test Trip",
			},
		})
		_, _ = w.Write(b)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.GetTripPlaces("test-key")
	if err != nil {
		t.Fatalf("GetTripPlaces: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.TripPlan.Title != "Test Trip" {
		t.Errorf("unexpected title: %s", result.TripPlan.Title)
	}
}

func TestLikeTrip(t *testing.T) {
	t.Run("like without auth returns error", func(t *testing.T) {
		client := NewClient()
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		client.SetLogger(logger)

		err := client.LikeTrip("test-key", true)
		if err == nil {
			t.Fatal("expected auth error")
		}
	})

	t.Run("successful like", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			// Verify auth headers are sent
			if _, err := r.Cookie("connect.sid"); err != nil {
				t.Errorf("expected session cookie: %v", err)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newTestClient(t, server)
		err := client.LikeTrip("test-key", true)
		if err != nil {
			t.Fatalf("LikeTrip: %v", err)
		}
	})

	t.Run("unlike", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Liked bool `json:"liked"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Liked {
				t.Error("expected liked=false for unlike")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newTestClient(t, server)
		err := client.LikeTrip("test-key", false)
		if err != nil {
			t.Fatalf("LikeTrip(false): %v", err)
		}
	})
}

func TestRegisterView(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || !strings.Contains(r.URL.Path, "/registerView") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	err := client.RegisterView("test-key")
	if err != nil {
		t.Fatalf("RegisterView: %v", err)
	}
}
