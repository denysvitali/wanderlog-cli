package wanderlog

import (
	"fmt"
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
	if amount < 0 {
		return fmt.Errorf("budget amount must be greater than or equal to 0")
	}
	currencyCode = normalizeCurrencyCode(currencyCode)
	if currencyCode == "" {
		return fmt.Errorf("currency code is required")
	}

	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}

	newAmount := CurrencyAmount{Amount: amount, CurrencyCode: currencyCode}
	op := ReplaceInObject(
		[]any{"itinerary", "budget", "amount"},
		trip.TripPlan.Itinerary.Budget.Amount,
		newAmount,
	)
	if err := c.ApplyOperations(tripKey, []Operation{op}); err != nil {
		return fmt.Errorf("setting trip budget: %w", err)
	}
	return nil
}

func (c *Client) AddTripExpense(tripKey string, req AddExpenseRequest) (*BudgetExpense, error) {
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

	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return nil, fmt.Errorf("getting current trip: %w", err)
	}
	userID, err := c.defaultBudgetUserID(req.PaidByUserID)
	if err != nil {
		return nil, err
	}

	expense := BudgetExpense{
		ID:           makeBudgetNumericID(),
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
		expense.AssociatedDate = &req.AssociatedDate
	}

	expenses := trip.TripPlan.Itinerary.Budget.Expenses
	op := InsertInList([]any{"itinerary", "budget", "expenses"}, len(expenses), expense)
	if err := c.ApplyOperations(tripKey, []Operation{op}); err != nil {
		return nil, fmt.Errorf("adding trip expense: %w", err)
	}
	return &expense, nil
}

func (c *Client) UpdateTripExpense(tripKey string, expenseID int, req UpdateExpenseRequest) (*BudgetExpense, error) {
	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return nil, fmt.Errorf("getting current trip: %w", err)
	}
	expenses := trip.TripPlan.Itinerary.Budget.Expenses
	index := FindBudgetExpenseIndex(expenses, expenseID)
	if index < 0 {
		return nil, fmt.Errorf("expense %d not found", expenseID)
	}

	oldExpense := expenses[index]
	newExpense := oldExpense
	if req.Description != nil {
		if strings.TrimSpace(*req.Description) == "" {
			return nil, fmt.Errorf("expense description cannot be empty")
		}
		newExpense.Description = strings.TrimSpace(*req.Description)
	}
	if req.Category != nil {
		category := normalizeExpenseCategory(*req.Category)
		if !validExpenseCategories[category] {
			return nil, fmt.Errorf("invalid expense category %q", *req.Category)
		}
		newExpense.Category = category
	}
	if req.Amount != nil {
		if *req.Amount <= 0 {
			return nil, fmt.Errorf("expense amount must be greater than 0")
		}
		newExpense.Amount.Amount = *req.Amount
	}
	if req.CurrencyCode != nil {
		currencyCode := normalizeCurrencyCode(*req.CurrencyCode)
		if currencyCode == "" {
			return nil, fmt.Errorf("currency code cannot be empty")
		}
		newExpense.Amount.CurrencyCode = currencyCode
	}
	if req.Date != nil {
		if err := validateBudgetDate("date", *req.Date); err != nil {
			return nil, err
		}
		newExpense.Date = *req.Date
	}
	if req.ClearBlockID {
		newExpense.BlockID = nil
	} else if req.BlockID != nil {
		newExpense.BlockID = req.BlockID
	}
	if req.PaidByUserID != nil {
		if *req.PaidByUserID <= 0 {
			return nil, fmt.Errorf("paid by user ID must be greater than 0")
		}
		newExpense.PaidByUserID = *req.PaidByUserID
		newExpense.PaidByUser = BudgetUser{Type: "registered", ID: *req.PaidByUserID}
	}
	if req.SetSplitWith {
		newExpense.SplitWith = budgetSplitWith(req.SplitWithUserIDs)
	}
	if req.ClearAssociatedDate {
		newExpense.AssociatedDate = nil
	} else if req.AssociatedDate != nil {
		if *req.AssociatedDate == "" {
			newExpense.AssociatedDate = nil
		} else {
			if err := validateBudgetDate("associated date", *req.AssociatedDate); err != nil {
				return nil, err
			}
			newExpense.AssociatedDate = req.AssociatedDate
		}
	}

	op := ReplaceInList([]any{"itinerary", "budget", "expenses"}, index, oldExpense, newExpense)
	if err := c.ApplyOperations(tripKey, []Operation{op}); err != nil {
		return nil, fmt.Errorf("updating trip expense: %w", err)
	}
	return &newExpense, nil
}

func (c *Client) DeleteTripExpense(tripKey string, expenseID int) error {
	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}
	expenses := trip.TripPlan.Itinerary.Budget.Expenses
	index := FindBudgetExpenseIndex(expenses, expenseID)
	if index < 0 {
		return fmt.Errorf("expense %d not found", expenseID)
	}

	op := DeleteFromList([]any{"itinerary", "budget", "expenses"}, index, expenses[index])
	if err := c.ApplyOperations(tripKey, []Operation{op}); err != nil {
		return fmt.Errorf("deleting trip expense: %w", err)
	}
	return nil
}

func FindBudgetExpenseIndex(expenses []BudgetExpense, expenseID int) int {
	for i, expense := range expenses {
		if expense.ID == expenseID {
			return i
		}
	}
	return -1
}

func (c *Client) defaultBudgetUserID(explicit int) (int, error) {
	if explicit > 0 {
		return explicit, nil
	}
	if c.auth != nil && c.auth.UserID != "" {
		id, err := strconv.Atoi(c.auth.UserID)
		if err == nil && id > 0 {
			return id, nil
		}
	}
	me, err := c.GetMe()
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

func makeBudgetNumericID() int {
	now := time.Now()
	return int(now.UnixMilli()*1000 + int64(now.Nanosecond()%1000))
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
