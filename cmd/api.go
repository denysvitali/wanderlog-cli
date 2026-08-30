package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
)

var (
	apiMethod        string
	apiBody          string
	apiBodyFile      string
	apiHeaderValues  []string
	apiAuthenticated bool
)

var apiCmd = &cobra.Command{
	Use:   "api [path-or-url]",
	Short: "Call a raw Wanderlog API endpoint",
	Long: `Call any Wanderlog API endpoint discovered from the Android/web bundle.

The path can be /api/..., tripPlans/..., or a full URL. Authentication is never
attached unless --auth is set. Authenticated requests are restricted to the
configured Wanderlog API origin.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := parseJSONBody(apiBody)
		if err != nil {
			return err
		}
		if apiBodyFile != "" {
			fileBody, err := os.ReadFile(apiBodyFile)
			if err != nil {
				return fmt.Errorf("read body file: %w", err)
			}
			body, err = parseJSONBody(string(fileBody))
			if err != nil {
				return fmt.Errorf("parse body file: %w", err)
			}
		}

		headers := map[string]string{}
		for _, header := range apiHeaderValues {
			key, value, ok := strings.Cut(header, ":")
			if !ok {
				return fmt.Errorf("invalid header %q: use Name: value", header)
			}
			headers[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}

		client, err := newClientE(apiAuthenticated)
		if err != nil {
			return err
		}
		status, respBody, err := client.DoAPIContext(cmd.Context(), apiMethod, args[0], body, headers, apiAuthenticated)
		if err != nil {
			if len(respBody) > 0 {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), string(respBody))
			}
			return fmt.Errorf("API request failed: %w", err)
		}

		if outputFormat == "raw" {
			if _, err := cmd.OutOrStdout().Write(respBody); err != nil {
				return err
			}
			if len(respBody) > 0 && respBody[len(respBody)-1] != '\n' {
				_, err = fmt.Fprintln(cmd.OutOrStdout())
			}
			return err
		}

		if outputFormat == "json" {
			if _, err := cmd.OutOrStdout().Write(respBody); err != nil {
				return err
			}
			if len(respBody) > 0 && respBody[len(respBody)-1] != '\n' {
				_, err = fmt.Fprintln(cmd.OutOrStdout())
			}
			return err
		}

		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "HTTP %s\n", ui.HighlightStyle.Render(fmt.Sprintf("%d", status))); err != nil {
			return err
		}
		if _, err := io.Copy(cmd.OutOrStdout(), strings.NewReader(string(respBody))); err != nil {
			return err
		}
		if len(respBody) > 0 && respBody[len(respBody)-1] != '\n' {
			_, err = fmt.Fprintln(cmd.OutOrStdout())
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)

	apiCmd.Flags().StringVarP(&apiMethod, "method", "X", http.MethodGet, "HTTP method")
	apiCmd.Flags().StringVar(&apiBody, "body", "", "JSON request body")
	apiCmd.Flags().StringVar(&apiBodyFile, "body-file", "", "File containing a JSON request body")
	apiCmd.Flags().StringArrayVarP(&apiHeaderValues, "header", "H", nil, "HTTP header as 'Name: value'")
	apiCmd.Flags().BoolVar(&apiAuthenticated, "auth", false, "Require stored or supplied authentication")
	apiCmd.Flags().StringVarP(&outputFormat, "output", "o", "raw", "Output format (raw, json, pretty)")
	apiCmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
	apiCmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
}
