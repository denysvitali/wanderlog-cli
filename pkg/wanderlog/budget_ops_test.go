package wanderlog

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func newBudgetTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *string) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	oldBaseURL := BaseURL
	BaseURL = server.URL
	t.Cleanup(func() { BaseURL = oldBaseURL })

	client := NewClient()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	client.SetLogger(logger)
	client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf", UserID: "42"})
	return client, &server.URL
}

func TestSetTripBudgetAppliesBudgetAmountOperation(t *testing.T) {
	var opReq OperationRequest
	client, _ := newBudgetTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/tripPlans/test-trip"):
			_, _ = w.Write([]byte(`{"tripPlan":{"key":"test-trip","itinerary":{"budget":{"amount":{"amount":100,"currencyCode":"USD"},"expenses":[],"payments":[],"simplifyDebt":false},"sections":[]}},"resources":{"placeMetadata":[]}}`))
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/tripPlans/test-trip/applyOps"):
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &opReq); err != nil {
				t.Fatalf("unmarshal ops: %v", err)
			}
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	if err := client.SetTripBudget("test-trip", 2500, "eur"); err != nil {
		t.Fatalf("SetTripBudget: %v", err)
	}
	if len(opReq.Ops) != 1 {
		t.Fatalf("ops length = %d, want 1", len(opReq.Ops))
	}
	op := opReq.Ops[0]
	if got := strings.Trim(strings.Join([]string{op.P[0].(string), op.P[1].(string), op.P[2].(string)}, "."), "."); got != "itinerary.budget.amount" {
		t.Fatalf("path = %v", op.P)
	}
	newAmount, ok := op.OI.(map[string]any)
	if !ok {
		t.Fatalf("oi type = %T", op.OI)
	}
	if newAmount["currencyCode"] != "EUR" || newAmount["amount"] != float64(2500) {
		t.Fatalf("new amount = %#v", newAmount)
	}
}

func TestAddTripExpenseAppliesInsertOperation(t *testing.T) {
	var opReq OperationRequest
	client, _ := newBudgetTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/tripPlans/test-trip"):
			_, _ = w.Write([]byte(`{"tripPlan":{"key":"test-trip","itinerary":{"budget":{"amount":{"amount":0,"currencyCode":"USD"},"expenses":[],"payments":[],"simplifyDebt":false},"sections":[]}},"resources":{"placeMetadata":[]}}`))
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/tripPlans/test-trip/applyOps"):
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &opReq); err != nil {
				t.Fatalf("unmarshal ops: %v", err)
			}
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	blockID := 99
	expense, err := client.AddTripExpense("test-trip", AddExpenseRequest{
		Description:      "Museum tickets",
		Category:         "activities",
		Amount:           80,
		CurrencyCode:     "usd",
		Date:             "2026-05-10",
		BlockID:          &blockID,
		SplitWithUserIDs: []int{42, 77},
	})
	if err != nil {
		t.Fatalf("AddTripExpense: %v", err)
	}
	if expense.PaidByUserID != 42 {
		t.Fatalf("paidByUserID = %d, want 42", expense.PaidByUserID)
	}
	if len(opReq.Ops) != 1 {
		t.Fatalf("ops length = %d, want 1", len(opReq.Ops))
	}
	op := opReq.Ops[0]
	if len(op.P) != 4 || op.P[0] != "itinerary" || op.P[1] != "budget" || op.P[2] != "expenses" || op.P[3] != float64(0) {
		t.Fatalf("path = %#v", op.P)
	}
	inserted, ok := op.LI.(map[string]any)
	if !ok {
		t.Fatalf("li type = %T", op.LI)
	}
	if inserted["description"] != "Museum tickets" || inserted["category"] != "activities" {
		t.Fatalf("inserted expense = %#v", inserted)
	}
}

func TestUpdateAndDeleteTripExpenseOperations(t *testing.T) {
	requests := []OperationRequest{}
	client, _ := newBudgetTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/tripPlans/test-trip"):
			_, _ = w.Write([]byte(`{"tripPlan":{"key":"test-trip","itinerary":{"budget":{"amount":{"amount":0,"currencyCode":"USD"},"expenses":[{"id":123,"amount":{"amount":25,"currencyCode":"USD"},"category":"food","description":"Lunch","date":"2026-05-10","paidByUserId":42,"paidByUser":{"type":"registered","id":42},"splitWith":{"type":"individuals","users":[]}}],"payments":[],"simplifyDebt":false},"sections":[]}},"resources":{"placeMetadata":[]}}`))
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/tripPlans/test-trip/applyOps"):
			body, _ := io.ReadAll(r.Body)
			var opReq OperationRequest
			if err := json.Unmarshal(body, &opReq); err != nil {
				t.Fatalf("unmarshal ops: %v", err)
			}
			requests = append(requests, opReq)
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	amount := 30.0
	if _, err := client.UpdateTripExpense("test-trip", 123, UpdateExpenseRequest{Amount: &amount}); err != nil {
		t.Fatalf("UpdateTripExpense: %v", err)
	}
	if err := client.DeleteTripExpense("test-trip", 123); err != nil {
		t.Fatalf("DeleteTripExpense: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("apply requests = %d, want 2", len(requests))
	}
	if requests[0].Ops[0].LI == nil || requests[0].Ops[0].LD == nil {
		t.Fatalf("update op = %#v", requests[0].Ops[0])
	}
	if requests[1].Ops[0].LD == nil || requests[1].Ops[0].LI != nil {
		t.Fatalf("delete op = %#v", requests[1].Ops[0])
	}
}

func TestBudgetOperationsPreserveRawServerFields(t *testing.T) {
	requests := []OperationRequest{}
	client, _ := newBudgetTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/tripPlans/test-trip"):
			_, _ = w.Write([]byte(`{"tripPlan":{"key":"test-trip","itinerary":{"budget":{"amount":{"amount":100,"currencyCode":"USD","serverCurrencyId":9007199254740993},"expenses":[{"id":123,"amount":{"amount":25,"currencyCode":"USD","precision":4},"category":"food","description":"Lunch","date":"2026-05-10","paidByUserId":42,"paidByUser":{"type":"registered","id":42},"splitWith":{"type":"individuals","users":[]},"serverOnly":{"revision":9007199254740993}}],"payments":[],"simplifyDebt":false},"sections":[]}},"resources":{"placeMetadata":[]}}`))
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/tripPlans/test-trip/applyOps"):
			var request OperationRequest
			decoder := json.NewDecoder(r.Body)
			decoder.UseNumber()
			if err := decoder.Decode(&request); err != nil {
				t.Fatalf("decode ops: %v", err)
			}
			requests = append(requests, request)
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	if err := client.SetTripBudget("test-trip", 250, "eur"); err != nil {
		t.Fatal(err)
	}
	amount := 30.0
	if _, err := client.UpdateTripExpense("test-trip", 123, UpdateExpenseRequest{Amount: &amount}); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteTripExpense("test-trip", 123); err != nil {
		t.Fatal(err)
	}
	if len(requests) != 3 {
		t.Fatalf("apply requests = %d, want 3", len(requests))
	}

	oldAmount := requests[0].Ops[0].OD.(map[string]any)
	newAmount := requests[0].Ops[0].OI.(map[string]any)
	assertExactJSONNumber(t, oldAmount["serverCurrencyId"], "9007199254740993")
	assertExactJSONNumber(t, newAmount["serverCurrencyId"], "9007199254740993")

	oldExpense := requests[1].Ops[0].LD.(map[string]any)
	newExpense := requests[1].Ops[0].LI.(map[string]any)
	assertExactJSONNumber(t, oldExpense["serverOnly"].(map[string]any)["revision"], "9007199254740993")
	assertExactJSONNumber(t, newExpense["serverOnly"].(map[string]any)["revision"], "9007199254740993")
	if newExpense["amount"].(map[string]any)["precision"] != json.Number("4") {
		t.Fatalf("updated expense dropped nested amount fields: %#v", newExpense)
	}
	deletedExpense := requests[2].Ops[0].LD.(map[string]any)
	assertExactJSONNumber(t, deletedExpense["serverOnly"].(map[string]any)["revision"], "9007199254740993")
}

func assertExactJSONNumber(t *testing.T, value any, want string) {
	t.Helper()
	number, ok := value.(json.Number)
	if !ok || number.String() != want {
		t.Fatalf("number = %#v (%T), want json.Number(%s)", value, value, want)
	}
}
