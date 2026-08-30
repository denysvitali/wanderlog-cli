package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var (
	analyticsFile    string
	analyticsOutput  string
	optimizeBody     string
	optimizeFile     string
	recommendGeoID   int
	recommendInput   string
	recommendExclude []string
)

var tripsAnalyticsCmd = &cobra.Command{
	Use:   "analytics [trip-key]",
	Short: "Analyze itinerary density, block mix, and expenses",
	Args: func(cmd *cobra.Command, args []string) error {
		if analyticsFile != "" {
			if len(args) != 0 {
				return fmt.Errorf("cannot combine a trip key with --file")
			}
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			trip *wanderlog.TripResponse
			err  error
		)
		if analyticsFile != "" {
			trip, err = wanderlog.LoadTripFromFile(analyticsFile)
		} else {
			client := wanderlog.NewClient()
			client.SetLogger(logger)
			trip, err = client.GetTripContext(cmd.Context(), args[0])
		}
		if err != nil {
			return fmt.Errorf("load trip: %w", err)
		}
		analysis, err := wanderlog.AnalyzeTrip(trip)
		if err != nil {
			return fmt.Errorf("analyze trip: %w", err)
		}
		if analyticsOutput == "json" {
			return ui.PrintJSON(analysis)
		}
		printTripAnalytics(analysis)
		return nil
	},
}

var tripsOptimizeRouteCmd = &cobra.Command{
	Use:   "optimize-route",
	Short: "Ask Wanderlog to optimize a route payload",
	Long: `Optimize a route using Wanderlog's directions API.

The API payload is intentionally accepted as JSON because the unofficial
endpoint evolves independently of this CLI. Exactly one of --body or
--body-file is required.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := decodeCommandJSON(optimizeBody, optimizeFile)
		if err != nil {
			return err
		}
		client := wanderlog.NewClient()
		client.SetLogger(logger)
		response, err := client.OptimizeRouteContext(cmd.Context(), payload)
		if err != nil {
			return fmt.Errorf("optimize route: %w", err)
		}
		return ui.PrintJSON(response)
	},
}

var tripsRecommendationsCmd = &cobra.Command{
	Use:   "recommendations [trip-key]",
	Short: "Get place recommendations for a trip",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if recommendGeoID <= 0 {
			return fmt.Errorf("--geo-id must be a positive Wanderlog geo ID")
		}
		client := wanderlog.NewClient()
		client.SetLogger(logger)
		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}
		trip, err := client.GetTripContext(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("get trip: %w", err)
		}
		if trip.TripPlan.ID <= 0 {
			return fmt.Errorf("trip response did not include a numeric trip plan ID")
		}
		response, err := client.GetRecommendedPlacesContext(cmd.Context(), wanderlog.RecommendedPlacesRequest{
			TripPlanID:        trip.TripPlan.ID,
			GeoID:             recommendGeoID,
			Input:             recommendInput,
			ExcludingPlaceIDs: recommendExclude,
		})
		if err != nil {
			return fmt.Errorf("get recommendations: %w", err)
		}
		return ui.PrintJSON(response)
	},
}

func init() {
	tripsCmd.AddCommand(tripsAnalyticsCmd, tripsOptimizeRouteCmd, tripsRecommendationsCmd)

	tripsAnalyticsCmd.Flags().StringVar(&analyticsFile, "file", "", "Load a trip JSON file instead of fetching a trip")
	tripsAnalyticsCmd.Flags().StringVarP(&analyticsOutput, "output", "o", "pretty", "Output format (pretty or json)")
	_ = tripsAnalyticsCmd.RegisterFlagCompletionFunc("output", cobra.FixedCompletions([]string{"pretty", "json"}, cobra.ShellCompDirectiveNoFileComp))
	tripsAnalyticsCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if analyticsOutput != "pretty" && analyticsOutput != "json" {
			return fmt.Errorf("unsupported output format %q (use pretty or json)", analyticsOutput)
		}
		return nil
	}

	tripsOptimizeRouteCmd.Flags().StringVar(&optimizeBody, "body", "", "Route request payload as JSON")
	tripsOptimizeRouteCmd.Flags().StringVar(&optimizeFile, "body-file", "", "File containing the route request JSON")

	tripsRecommendationsCmd.Flags().IntVar(&recommendGeoID, "geo-id", 0, "Wanderlog destination geo ID")
	tripsRecommendationsCmd.Flags().StringVar(&recommendInput, "input", "", "Optional recommendation prompt or category")
	tripsRecommendationsCmd.Flags().StringSliceVar(&recommendExclude, "exclude-place-id", nil, "Google place ID to exclude (repeatable)")
	tripsRecommendationsCmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
	tripsRecommendationsCmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
}

func decodeCommandJSON(inline, filename string) (any, error) {
	if strings.TrimSpace(inline) == "" && filename == "" {
		return nil, fmt.Errorf("exactly one of --body or --body-file is required")
	}
	if strings.TrimSpace(inline) != "" && filename != "" {
		return nil, fmt.Errorf("cannot combine --body and --body-file")
	}
	data := []byte(inline)
	if filename != "" {
		var err error
		data, err = os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("read body file: %w", err)
		}
	}
	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON payload: %w", err)
	}
	return payload, nil
}

func printTripAnalytics(analysis *wanderlog.TripAnalytics) {
	fmt.Println(ui.TitleStyle.Render("Trip analytics: " + ui.SafeText(analysis.Title)))
	fmt.Printf("Key: %s\n", ui.SafeText(analysis.TripKey))
	if analysis.StartDate != "" || analysis.EndDate != "" {
		fmt.Printf("Dates: %s to %s (%d days)\n", analysis.StartDate, analysis.EndDate, analysis.Days)
	}
	fmt.Printf("Sections: %d (%d dated)\n", analysis.Sections, analysis.DatedSections)
	fmt.Printf("Blocks: %d places, %d notes, %d flights, %d lodging, %d transit, %d other\n",
		analysis.PlaceBlocks, analysis.Notes, analysis.Flights, analysis.Lodgings, analysis.Transit, analysis.OtherBlocks)
	if len(analysis.Expenses) > 0 {
		fmt.Println("Expenses:")
		for _, total := range analysis.Expenses {
			fmt.Printf("  %s %.2f (%d items)\n", ui.SafeText(total.Currency), total.Amount, total.Count)
		}
	}
	if len(analysis.DayLoads) > 0 {
		fmt.Println("Day load:")
		for _, day := range analysis.DayLoads {
			label := day.Date
			if label == "" {
				label = day.Heading
			}
			fmt.Printf("  %s: %d places, %d blocks\n", ui.SafeText(label), day.Places, day.Blocks)
		}
	}
	if len(analysis.Warnings) > 0 {
		fmt.Println(ui.WarningStyle.Render("Warnings:"))
		for _, warning := range analysis.Warnings {
			fmt.Printf("  - %s\n", ui.SafeText(warning))
		}
	}
}
