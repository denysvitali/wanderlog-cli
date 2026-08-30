package wanderlog

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

var validExpenseCategories = map[string]bool{
	"flights":       true,
	"lodging":       true,
	"carRental":     true,
	"publicTransit": true,
	"food":          true,
	"drinks":        true,
	"sightseeing":   true,
	"activities":    true,
	"shopping":      true,
	"gas":           true,
	"groceries":     true,
	"other":         true,
}

type AddExpenseRequest struct {
	Description      string
	Category         string
	Amount           float64
	CurrencyCode     string
	Date             string
	BlockID          *int
	PaidByUserID     int
	SplitWithUserIDs []int
	AssociatedDate   string
}

type UpdateExpenseRequest struct {
	Description         *string
	Category            *string
	Amount              *float64
	CurrencyCode        *string
	Date                *string
	BlockID             *int
	ClearBlockID        bool
	PaidByUserID        *int
	SplitWithUserIDs    []int
	SetSplitWith        bool
	AssociatedDate      *string
	ClearAssociatedDate bool
}

func (c *Client) SetTripBudget(tripKey string, amount float64, currencyCode string) error {
	return c.SetTripBudgetContext(context.Background(), tripKey, amount, currencyCode)
}

// SetTripBudgetContext sets the budget and binds all network I/O to ctx.
func (c *Client) SetTripBudgetContext(ctx context.Context, tripKey string, amount float64, currencyCode string) error {
	if amount < 0 {
		return fmt.Errorf("budget amount must be greater than or equal to 0")
	}
	currencyCode = normalizeCurrencyCode(currencyCode)
	if currencyCode == "" {
		return fmt.Errorf("currency code is required")
	}

	err := c.retryJSON0MutationContext(ctx, tripKey, "SetTripBudget", func(ctx context.Context) ([]Operation, error) {
		trip, err := c.GetTripRawContext(ctx, tripKey)
		if err != nil {
			return nil, fmt.Errorf("getting current trip: %w", err)
		}
		budget, err := rawTripBudget(trip)
		if err != nil {
			return nil, err
		}
		oldAmount, ok := budget["amount"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("trip budget amount has unexpected type %T", budget["amount"])
		}
		newAmount, err := cloneRawMap(oldAmount)
		if err != nil {
			return nil, fmt.Errorf("copying trip budget amount: %w", err)
		}
		newAmount["amount"] = amount
		newAmount["currencyCode"] = currencyCode
		return []Operation{ReplaceInObject(
			[]any{"itinerary", "budget", "amount"},
			oldAmount,
			newAmount,
		)}, nil
	})
	if err != nil {
		return fmt.Errorf("setting trip budget: %w", err)
	}
	return nil
}

func (c *Client) AddTripExpense(tripKey string, req AddExpenseRequest) (*BudgetExpense, error) {
	return c.AddTripExpenseContext(context.Background(), tripKey, req)
}

// AddTripExpenseContext adds an expense and binds all network I/O to ctx.
func (c *Client) AddTripExpenseContext(ctx context.Context, tripKey string, req AddExpenseRequest) (*BudgetExpense, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("expense amount must be greater than 0")
	}
	if strings.TrimSpace(req.Description) == "" {
		return nil, fmt.Errorf("expense description is required")
	}
	category := normalizeExpenseCategory(req.Category)
	if !validExpenseCategories[category] {
		return nil, fmt.Errorf("invalid expense category %q", req.Category)
	}
	currencyCode := normalizeCurrencyCode(req.CurrencyCode)
	if currencyCode == "" {
		return nil, fmt.Errorf("currency code is required")
	}
	if req.Date == "" {
		req.Date = time.Now().Format("2006-01-02")
	}
	if err := validateBudgetDate("date", req.Date); err != nil {
		return nil, err
	}
	if req.AssociatedDate != "" {
		if err := validateBudgetDate("associated date", req.AssociatedDate); err != nil {
			return nil, err
		}
	}

	userID, err := c.defaultBudgetUserIDContext(ctx, req.PaidByUserID)
	if err != nil {
		return nil, err
	}

	var expense *BudgetExpense
	err = c.retryJSON0MutationContext(ctx, tripKey, "AddTripExpense", func(ctx context.Context) ([]Operation, error) {
		trip, err := c.GetTripRawContext(ctx, tripKey)
		if err != nil {
			return nil, fmt.Errorf("getting current trip: %w", err)
		}
		_, expenses, err := rawBudgetExpenses(trip)
		if err != nil {
			return nil, err
		}

		expenseID, err := makeBudgetNumericID()
		if err != nil {
			return nil, fmt.Errorf("generating expense ID: %w", err)
		}
		rebuilt := BudgetExpense{
			ID:           expenseID,
			Amount:       CurrencyAmount{Amount: req.Amount, CurrencyCode: currencyCode},
			Category:     category,
			Description:  strings.TrimSpace(req.Description),
			Date:         req.Date,
			BlockID:      req.BlockID,
			PaidByUserID: userID,
			PaidByUser:   BudgetUser{Type: "registered", ID: userID},
			SplitWith:    budgetSplitWith(req.SplitWithUserIDs),
		}
		if req.AssociatedDate != "" {
			rebuilt.AssociatedDate = &req.AssociatedDate
		}
		expense = &rebuilt

		return []Operation{InsertInList([]any{"itinerary", "budget", "expenses"}, len(expenses), rebuilt)}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("adding trip expense: %w", err)
	}
	return expense, nil
}

func (c *Client) UpdateTripExpense(tripKey string, expenseID int, req UpdateExpenseRequest) (*BudgetExpense, error) {
	return c.UpdateTripExpenseContext(context.Background(), tripKey, expenseID, req)
}

// UpdateTripExpenseContext updates an expense and binds all network I/O to ctx.
func (c *Client) UpdateTripExpenseContext(ctx context.Context, tripKey string, expenseID int, req UpdateExpenseRequest) (*BudgetExpense, error) {
	var result *BudgetExpense
	err := c.retryJSON0MutationContext(ctx, tripKey, "UpdateTripExpense", func(ctx context.Context) ([]Operation, error) {
		trip, err := c.GetTripRawContext(ctx, tripKey)
		if err != nil {
			return nil, fmt.Errorf("getting current trip: %w", err)
		}
		_, expenses, err := rawBudgetExpenses(trip)
		if err != nil {
			return nil, err
		}
		index := findRawBudgetExpenseIndex(expenses, expenseID)
		if index < 0 {
			return nil, fmt.Errorf("expense %d not found", expenseID)
		}
		oldExpense, ok := expenses[index].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expense %d has unexpected type %T", expenseID, expenses[index])
		}
		newExpense, err := cloneRawMap(oldExpense)
		if err != nil {
			return nil, fmt.Errorf("copying expense %d: %w", expenseID, err)
		}
		if req.Description != nil {
			if strings.TrimSpace(*req.Description) == "" {
				return nil, fmt.Errorf("expense description cannot be empty")
			}
			newExpense["description"] = strings.TrimSpace(*req.Description)
		}
		if req.Category != nil {
			category := normalizeExpenseCategory(*req.Category)
			if !validExpenseCategories[category] {
				return nil, fmt.Errorf("invalid expense category %q", *req.Category)
			}
			newExpense["category"] = category
		}
		if req.Amount != nil {
			if *req.Amount <= 0 {
				return nil, fmt.Errorf("expense amount must be greater than 0")
			}
			amountValue, ok := newExpense["amount"].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("expense %d amount has unexpected type %T", expenseID, newExpense["amount"])
			}
			amountValue["amount"] = *req.Amount
		}
		if req.CurrencyCode != nil {
			currencyCode := normalizeCurrencyCode(*req.CurrencyCode)
			if currencyCode == "" {
				return nil, fmt.Errorf("currency code cannot be empty")
			}
			amountValue, ok := newExpense["amount"].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("expense %d amount has unexpected type %T", expenseID, newExpense["amount"])
			}
			amountValue["currencyCode"] = currencyCode
		}
		if req.Date != nil {
			if err := validateBudgetDate("date", *req.Date); err != nil {
				return nil, err
			}
			newExpense["date"] = *req.Date
		}
		if req.ClearBlockID {
			delete(newExpense, "blockId")
		} else if req.BlockID != nil {
			newExpense["blockId"] = *req.BlockID
		}
		if req.PaidByUserID != nil {
			if *req.PaidByUserID <= 0 {
				return nil, fmt.Errorf("paid by user ID must be greater than 0")
			}
			newExpense["paidByUserId"] = *req.PaidByUserID
			newExpense["paidByUser"] = map[string]any{"type": "registered", "id": *req.PaidByUserID}
		}
		if req.SetSplitWith {
			newExpense["splitWith"] = budgetSplitWith(req.SplitWithUserIDs)
		}
		if req.ClearAssociatedDate {
			delete(newExpense, "associatedDate")
		} else if req.AssociatedDate != nil {
			if *req.AssociatedDate == "" {
				delete(newExpense, "associatedDate")
			} else {
				if err := validateBudgetDate("associated date", *req.AssociatedDate); err != nil {
					return nil, err
				}
				newExpense["associatedDate"] = *req.AssociatedDate
			}
		}

		result, err = budgetExpenseFromRaw(newExpense)
		if err != nil {
			return nil, fmt.Errorf("decoding updated expense %d: %w", expenseID, err)
		}
		return []Operation{ReplaceInList([]any{"itinerary", "budget", "expenses"}, index, oldExpense, newExpense)}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("updating trip expense: %w", err)
	}
	return result, nil
}

func (c *Client) DeleteTripExpense(tripKey string, expenseID int) error {
	return c.DeleteTripExpenseContext(context.Background(), tripKey, expenseID)
}

// DeleteTripExpenseContext deletes an expense and binds all network I/O to ctx.
func (c *Client) DeleteTripExpenseContext(ctx context.Context, tripKey string, expenseID int) error {
	err := c.retryJSON0MutationContext(ctx, tripKey, "DeleteTripExpense", func(ctx context.Context) ([]Operation, error) {
		trip, err := c.GetTripRawContext(ctx, tripKey)
		if err != nil {
			return nil, fmt.Errorf("getting current trip: %w", err)
		}
		_, expenses, err := rawBudgetExpenses(trip)
		if err != nil {
			return nil, err
		}
		index := findRawBudgetExpenseIndex(expenses, expenseID)
		if index < 0 {
			return nil, fmt.Errorf("expense %d not found", expenseID)
		}

		return []Operation{DeleteFromList([]any{"itinerary", "budget", "expenses"}, index, expenses[index])}, nil
	})
	if err != nil {
		return fmt.Errorf("deleting trip expense: %w", err)
	}
	return nil
}

func rawTripBudget(trip map[string]any) (map[string]any, error) {
	tripPlan, ok := trip["tripPlan"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("trip response is missing tripPlan")
	}
	itinerary, ok := tripPlan["itinerary"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("trip response is missing tripPlan.itinerary")
	}
	budget, ok := itinerary["budget"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("trip response is missing tripPlan.itinerary.budget")
	}
	return budget, nil
}

func rawBudgetExpenses(trip map[string]any) (map[string]any, []any, error) {
	budget, err := rawTripBudget(trip)
	if err != nil {
		return nil, nil, err
	}
	expensesValue, exists := budget["expenses"]
	if !exists || expensesValue == nil {
		return budget, []any{}, nil
	}
	expenses, ok := expensesValue.([]any)
	if !ok {
		return nil, nil, fmt.Errorf("trip budget expenses has unexpected type %T", expensesValue)
	}
	return budget, expenses, nil
}

func findRawBudgetExpenseIndex(expenses []any, expenseID int) int {
	for i, value := range expenses {
		expense, ok := value.(map[string]any)
		if ok && rawInt(expense["id"]) == expenseID {
			return i
		}
	}
	return -1
}

func budgetExpenseFromRaw(value map[string]any) (*BudgetExpense, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var expense BudgetExpense
	if err := json.Unmarshal(data, &expense); err != nil {
		return nil, err
	}
	return &expense, nil
}

func FindBudgetExpenseIndex(expenses []BudgetExpense, expenseID int) int {
	for i, expense := range expenses {
		if expense.ID == expenseID {
			return i
		}
	}
	return -1
}

func (c *Client) defaultBudgetUserIDContext(ctx context.Context, explicit int) (int, error) {
	if explicit > 0 {
		return explicit, nil
	}
	if c.auth != nil && c.auth.UserID != "" {
		id, err := strconv.Atoi(c.auth.UserID)
		if err == nil && id > 0 {
			return id, nil
		}
	}
	me, err := c.GetMeContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("paid by user ID is required: %w", err)
	}
	if me.ID <= 0 {
		return 0, fmt.Errorf("paid by user ID is required")
	}
	return me.ID, nil
}

func budgetSplitWith(userIDs []int) BudgetSplitWith {
	users := make([]BudgetUser, 0, len(userIDs))
	for _, id := range userIDs {
		if id > 0 {
			users = append(users, BudgetUser{Type: "registered", ID: id})
		}
	}
	return BudgetSplitWith{Type: "individuals", Users: users}
}

func makeBudgetNumericID() (int, error) {
	// Keep IDs within JavaScript's exactly representable integer range while
	// avoiding timestamp collisions across concurrent CLI processes.
	value, err := rand.Int(rand.Reader, big.NewInt(1<<52-1))
	if err != nil {
		return 0, err
	}
	return int(value.Int64() + 1), nil
}

func normalizeCurrencyCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeExpenseCategory(value string) string {
	value = strings.TrimSpace(value)
	switch strings.ToLower(value) {
	case "", "other":
		return "other"
	case "car-rental", "car_rental", "carrental":
		return "carRental"
	case "public-transit", "public_transit", "publictransit", "transit":
		return "publicTransit"
	default:
		return value
	}
}

func validateBudgetDate(name, value string) error {
	if value == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return fmt.Errorf("invalid %s date format, use YYYY-MM-DD", name)
	}
	return nil
}
