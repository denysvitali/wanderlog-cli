package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
	"github.com/spf13/viper"
)

func newClient(requireAuth bool) *wanderlog.Client {
	client := wanderlog.NewClient()
	client.SetLogger(logger)
	if requireAuth {
		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}
		return client
	}

	switch {
	case sessionCookie != "" || xsrfToken != "":
		client.SetAuth(&wanderlog.AuthCredentials{SessionCookie: sessionCookie, XSRFToken: xsrfToken})
	case viper.GetString("auth.session.cookie") != "" && viper.GetString("auth.session.xsrf_token") != "":
		client.SetAuth(&wanderlog.AuthCredentials{
			SessionCookie: viper.GetString("auth.session.cookie"),
			XSRFToken:     viper.GetString("auth.session.xsrf_token"),
			UserID:        viper.GetString("auth.session.user_id"),
		})
	default:
		if creds, err := wanderlog.LoadCredentials(); err == nil {
			client.SetAuth(creds)
		}
	}
	return client
}

func parseRequiredInt(value, name string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		logger.WithError(err).Errorf("Invalid %s - must be a number", name)
		os.Exit(1)
	}
	return parsed
}

func parseIntCSV(value, name string) []int {
	if strings.TrimSpace(value) == "" {
		logger.Errorf("%s is required", name)
		os.Exit(1)
	}

	parts := strings.Split(value, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		result = append(result, parseRequiredInt(strings.TrimSpace(part), name))
	}
	return result
}

func parseChecklistItems(values []string) []wanderlog.ChecklistItem {
	items := make([]wanderlog.ChecklistItem, 0, len(values))
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text == "" {
			continue
		}
		items = append(items, wanderlog.ChecklistItem{Text: text})
	}
	if len(items) == 0 {
		logger.Error("At least one --item value is required")
		os.Exit(1)
	}
	return items
}

func validateDateFlag(value, name string) {
	if value == "" {
		return
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		logger.WithError(err).Errorf("Invalid %s date format. Use YYYY-MM-DD", name)
		os.Exit(1)
	}
}

func mustJSON(data string) []byte {
	if strings.TrimSpace(data) == "" {
		return nil
	}

	var raw json.RawMessage
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		logger.WithError(err).Error("Invalid JSON body")
		os.Exit(1)
	}
	return []byte(data)
}

func printSuccess(format string, message string, data interface{}) {
	if format == "json" {
		ui.PrintJSON(data)
		return
	}
	fmt.Println(message)
}
