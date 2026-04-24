package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
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

The path can be /api/..., tripPlans/..., or a full URL. Auth is optional by
default and is attached when credentials are available; use --auth to require it.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		body := mustJSON(apiBody)
		if apiBodyFile != "" {
			fileBody, err := os.ReadFile(apiBodyFile)
			if err != nil {
				logger.WithError(err).Error("Failed to read body file")
				os.Exit(1)
			}
			body = mustJSON(string(fileBody))
		}

		headers := map[string]string{}
		for _, header := range apiHeaderValues {
			key, value, ok := strings.Cut(header, ":")
			if !ok {
				logger.Errorf("Invalid header %q. Use Name: value", header)
				os.Exit(1)
			}
			headers[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}

		client := newClient(apiAuthenticated)
		status, respBody, err := client.DoAPI(apiMethod, args[0], body, headers, apiAuthenticated)
		if err != nil {
			logger.WithError(err).Error("API request failed")
			if len(respBody) > 0 {
				fmt.Fprintln(os.Stderr, string(respBody))
			}
			os.Exit(1)
		}

		if outputFormat == "raw" {
			_, _ = os.Stdout.Write(respBody)
			if len(respBody) > 0 && respBody[len(respBody)-1] != '\n' {
				fmt.Println()
			}
			return
		}

		if outputFormat == "json" {
			_, _ = os.Stdout.Write(respBody)
			if len(respBody) > 0 && respBody[len(respBody)-1] != '\n' {
				fmt.Println()
			}
			return
		}

		fmt.Printf("HTTP %d\n", status)
		_, _ = io.Copy(os.Stdout, strings.NewReader(string(respBody)))
		if len(respBody) > 0 && respBody[len(respBody)-1] != '\n' {
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)

	apiCmd.Flags().StringVarP(&apiMethod, "method", "X", http.MethodGet, "HTTP method")
	apiCmd.Flags().StringVar(&apiBody, "body", "", "JSON request body")
	apiCmd.Flags().StringVar(&apiBodyFile, "body-file", "", "File containing a JSON request body")
	apiCmd.Flags().StringArrayVarP(&apiHeaderValues, "header", "H", nil, "HTTP header as 'Name: value'")
	apiCmd.Flags().BoolVar(&apiAuthenticated, "auth", false, "Require stored or supplied authentication")
	apiCmd.Flags().StringVarP(&outputFormat, "format", "f", "raw", "Output format (raw, json, pretty)")
	apiCmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
	apiCmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
}
