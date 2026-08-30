package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
	"github.com/spf13/viper"
)

func newClientE(requireAuth bool) (*wanderlog.Client, error) {
	client := wanderlog.NewClient()
	client.SetLogger(logger)
	if requireAuth {
		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return nil, fmt.Errorf("authentication required: %w", err)
		}
		return client, nil
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
	return client, nil
}

func parseRequiredIntE(value, name string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: must be a number", name, value)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("invalid %s %q: must be greater than zero", name, value)
	}
	return parsed, nil
}

func parseIntCSVE(value, name string) ([]int, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("%s is required", name)
	}

	parts := strings.Split(value, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		parsed, err := parseRequiredIntE(strings.TrimSpace(part), name)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}
	return result, nil
}

func parseChecklistItemsE(values []string) ([]wanderlog.ChecklistItem, error) {
	items := make([]wanderlog.ChecklistItem, 0, len(values))
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text == "" {
			continue
		}
		items = append(items, wanderlog.ChecklistItem{Text: text})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one non-empty --item value is required")
	}
	return items, nil
}

func validateDateFlagE(value, name string) error {
	if value == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return fmt.Errorf("invalid %s date %q: use YYYY-MM-DD", name, value)
	}
	return nil
}

func parseJSONBody(data string) ([]byte, error) {
	if strings.TrimSpace(data) == "" {
		return nil, nil
	}

	var raw json.RawMessage
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON body: %w", err)
	}
	return []byte(data), nil
}

func printSuccess(format string, message string, data interface{}) error {
	if format == "json" {
		return ui.PrintJSON(data)
	}
	_, err := fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ %s", message)))
	return err
}

// confirmAction keeps prompts off stdout (which may be consumed as JSON) and
// accepts --yes for scripts. EOF is treated as a canceled operation.
func confirmAction(cmd interface {
	InOrStdin() io.Reader
	ErrOrStderr() io.Writer
}, prompt string, assumeYes bool) (bool, error) {
	if assumeYes {
		return true, nil
	}
	if outputFormat == "json" {
		return false, fmt.Errorf("--yes is required with --output json")
	}

	if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "%s Type 'yes' to confirm: ", prompt); err != nil {
		return false, err
	}
	response, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	return strings.EqualFold(strings.TrimSpace(response), "yes"), nil
}
