// Package wanderlog provides a Go client for interacting with the Wanderlog API.
//
// Wanderlog is a trip planning application that allows users to create detailed
// itineraries, add places, manage flights, and share their travel plans.
//
// Basic usage:
//
//	client := wanderlog.NewClient()
//	trip, err := client.GetTrip("your-trip-id")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Trip: %s (%s to %s)\n",
//		trip.TripPlan.Title,
//		trip.TripPlan.StartDate,
//		trip.TripPlan.EndDate)
//
// The client supports fetching trip data with detailed information including:
//   - Trip metadata (title, dates, view count, etc.)
//   - Daily itineraries with places
//   - Flight information
//   - Budget and expense tracking
//   - Place search and details through Wanderlog APIs
//
// This package can be used both as a standalone library and as part of the
// wanderlog-cli command-line application.
package wanderlog

// Version represents the current version of the wanderlog package
const Version = "1.0.0"
