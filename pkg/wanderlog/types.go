package wanderlog

// TripData represents the essential trip information for display purposes
type TripData struct {
	Title      string
	StartDate  string
	EndDate    string
	Days       int
	PlaceCount int
	ViewCount  int
	LikeCount  int
}

// ExtractTripData extracts essential trip data from the full API response
func ExtractTripData(response *TripResponse) *TripData {
	if response == nil || !response.Success {
		return nil
	}

	plan := response.TripPlan
	return &TripData{
		Title:      plan.Title,
		StartDate:  plan.StartDate,
		EndDate:    plan.EndDate,
		Days:       plan.Days,
		PlaceCount: plan.PlaceCount,
		ViewCount:  plan.ViewCount,
		LikeCount:  plan.LikeCount,
	}
}