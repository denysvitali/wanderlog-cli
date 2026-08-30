package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var tripsBudgetCmd = &cobra.Command{
	Use:   "budget",
	Short: "Manage trip budget",
}

var tripsBudgetSetCmd = &cobra.Command{
	Use:   "set [trip-key]",
	Short: "Set a trip's total budget",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if tripsBudgetAmount < 0 {
			return fmt.Errorf("amount must be greater than or equal to 0")
		}
		if tripsBudgetCurrency == "" {
			return fmt.Errorf("currency is required (--currency)")
		}
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		if err := client.SetTripBudget(args[0], tripsBudgetAmount, tripsBudgetCurrency); err != nil {
			return fmt.Errorf("set budget: %w", err)
		}
		return printSuccess(outputFormat, "Set trip budget", map[string]interface{}{
			"tripKey":  args[0],
			"amount":   tripsBudgetAmount,
			"currency": strings.ToUpper(tripsBudgetCurrency),
		})
	},
}

var tripsExpensesAddCmd = &cobra.Command{
	Use:   "add [trip-key]",
	Short: "Add an expense to a trip budget",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(tripsExpenseDescription) == "" {
			return fmt.Errorf("description is required (--description)")
		}
		if tripsExpenseAmount <= 0 {
			return fmt.Errorf("amount must be greater than 0")
		}
		if strings.TrimSpace(tripsExpenseCurrency) == "" {
			return fmt.Errorf("currency is required (--currency)")
		}
		splitWith, err := parseOptionalIntCSVE(tripsExpenseSplitWith)
		if err != nil {
			return err
		}
		req := wanderlog.AddExpenseRequest{
			Description:      tripsExpenseDescription,
			Category:         tripsExpenseCategory,
			Amount:           tripsExpenseAmount,
			CurrencyCode:     tripsExpenseCurrency,
			Date:             tripsExpenseDate,
			PaidByUserID:     tripsExpensePaidBy,
			SplitWithUserIDs: splitWith,
			AssociatedDate:   tripsExpenseAssociatedDate,
		}
		if cmd.Flags().Changed("block-id") {
			req.BlockID = &tripsExpenseBlockID
		}

		client, err := newClientE(true)
		if err != nil {
			return err
		}
		expense, err := client.AddTripExpense(args[0], req)
		if err != nil {
			return fmt.Errorf("add expense: %w", err)
		}
		return printSuccess(outputFormat, "Added expense", expense)
	},
}

var tripsExpensesUpdateCmd = &cobra.Command{
	Use:   "update [trip-key] [expense-id]",
	Short: "Update a trip budget expense",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		expenseID, err := parseRequiredIntE(args[1], "expense ID")
		if err != nil {
			return err
		}
		req := wanderlog.UpdateExpenseRequest{}
		if cmd.Flags().Changed("description") {
			req.Description = &tripsExpenseDescription
		}
		if cmd.Flags().Changed("desc") {
			req.Description = &tripsExpenseDescription
		}
		if cmd.Flags().Changed("category") {
			req.Category = &tripsExpenseCategory
		}
		if cmd.Flags().Changed("amount") {
			req.Amount = &tripsExpenseAmount
		}
		if cmd.Flags().Changed("currency") {
			req.CurrencyCode = &tripsExpenseCurrency
		}
		if cmd.Flags().Changed("date") {
			req.Date = &tripsExpenseDate
		}
		if cmd.Flags().Changed("block-id") {
			req.BlockID = &tripsExpenseBlockID
		}
		req.ClearBlockID = tripsExpenseClearBlockID
		if cmd.Flags().Changed("paid-by") {
			req.PaidByUserID = &tripsExpensePaidBy
		}
		if cmd.Flags().Changed("split-with") {
			req.SetSplitWith = true
			req.SplitWithUserIDs, err = parseOptionalIntCSVE(tripsExpenseSplitWith)
			if err != nil {
				return err
			}
		}
		if cmd.Flags().Changed("associated-date") {
			req.AssociatedDate = &tripsExpenseAssociatedDate
		}
		req.ClearAssociatedDate = tripsExpenseClearAssociatedDate

		client, err := newClientE(true)
		if err != nil {
			return err
		}
		expense, err := client.UpdateTripExpense(args[0], expenseID, req)
		if err != nil {
			return fmt.Errorf("update expense: %w", err)
		}
		return printSuccess(outputFormat, "Updated expense", expense)
	},
}

var tripsExpensesDeleteCmd = &cobra.Command{
	Use:   "delete [trip-key] [expense-id]",
	Short: "Delete a trip budget expense",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		expenseID, err := parseRequiredIntE(args[1], "expense ID")
		if err != nil {
			return err
		}
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		if err := client.DeleteTripExpense(args[0], expenseID); err != nil {
			return fmt.Errorf("delete expense: %w", err)
		}
		return printSuccess(outputFormat, "Deleted expense", map[string]interface{}{"tripKey": args[0], "expenseId": expenseID})
	},
}

var (
	tripsBudgetAmount   float64
	tripsBudgetCurrency string

	tripsExpenseDescription         string
	tripsExpenseCategory            string
	tripsExpenseAmount              float64
	tripsExpenseCurrency            string
	tripsExpenseDate                string
	tripsExpenseAssociatedDate      string
	tripsExpenseClearAssociatedDate bool
	tripsExpenseBlockID             int
	tripsExpenseClearBlockID        bool
	tripsExpensePaidBy              int
	tripsExpenseSplitWith           string
)

func init() {
	tripsCmd.AddCommand(tripsBudgetCmd)
	tripsBudgetCmd.AddCommand(tripsBudgetSetCmd)
	tripsExpensesCmd.AddCommand(tripsExpensesAddCmd, tripsExpensesUpdateCmd, tripsExpensesDeleteCmd)

	tripsBudgetSetCmd.Flags().Float64Var(&tripsBudgetAmount, "amount", 0, "Budget amount")
	tripsBudgetSetCmd.Flags().StringVar(&tripsBudgetCurrency, "currency", "USD", "Currency code")
	tripsBudgetSetCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format")
	tripsBudgetSetCmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
	tripsBudgetSetCmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	_ = tripsBudgetSetCmd.MarkFlagRequired("amount")

	addExpenseFlags(tripsExpensesAddCmd)
	addExpenseFlags(tripsExpensesUpdateCmd)
	tripsExpensesUpdateCmd.Flags().BoolVar(&tripsExpenseClearBlockID, "clear-block-id", false, "Remove linked block ID")
	tripsExpensesUpdateCmd.Flags().BoolVar(&tripsExpenseClearAssociatedDate, "clear-associated-date", false, "Remove associated date")

	for _, c := range []*cobra.Command{tripsExpensesAddCmd, tripsExpensesUpdateCmd, tripsExpensesDeleteCmd} {
		c.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format")
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}

func addExpenseFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&tripsExpenseDescription, "description", "", "Expense description")
	cmd.Flags().StringVar(&tripsExpenseDescription, "desc", "", "Expense description")
	cmd.Flags().StringVar(&tripsExpenseCategory, "category", "other", "Expense category")
	cmd.Flags().Float64Var(&tripsExpenseAmount, "amount", 0, "Expense amount")
	cmd.Flags().StringVar(&tripsExpenseCurrency, "currency", "USD", "Currency code")
	cmd.Flags().StringVar(&tripsExpenseDate, "date", "", "Expense date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&tripsExpenseAssociatedDate, "associated-date", "", "Associated trip date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&tripsExpenseBlockID, "block-id", 0, "Optional itinerary block ID to link")
	cmd.Flags().IntVar(&tripsExpensePaidBy, "paid-by", 0, "User ID who paid (defaults to authenticated user)")
	cmd.Flags().StringVar(&tripsExpenseSplitWith, "split-with", "", "Comma-separated user IDs to split with")
}

func parseOptionalIntCSVE(value string) ([]int, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID %q in comma-separated list: %w", part, err)
		}
		if id <= 0 {
			return nil, fmt.Errorf("invalid user ID %q in comma-separated list: must be greater than zero", part)
		}
		result = append(result, id)
	}
	return result, nil
}
