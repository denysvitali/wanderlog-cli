package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

type fakeTravelWriteClient struct {
	operation string
	request   any
	result    *wanderlog.TravelMutationResult
}

func (f *fakeTravelWriteClient) response(operation string, request any) (*wanderlog.TravelMutationResult, error) {
	f.operation, f.request = operation, request
	if f.result == nil {
		f.result = &wanderlog.TravelMutationResult{Success: true, Message: "done"}
	}
	return f.result, nil
}
func (f *fakeTravelWriteClient) AddFlightReservation(r wanderlog.AddFlightReservationRequest) (*wanderlog.TravelMutationResult, error) {
	return f.response("add_flight", r)
}
func (f *fakeTravelWriteClient) UpdateFlightReservation(r wanderlog.UpdateFlightReservationRequest) (*wanderlog.TravelMutationResult, error) {
	return f.response("update_flight", r)
}
func (f *fakeTravelWriteClient) DeleteFlightReservation(r wanderlog.DeleteTravelReservationRequest) (*wanderlog.TravelMutationResult, error) {
	return f.response("delete_flight", r)
}
func (f *fakeTravelWriteClient) AddLodgingReservation(r wanderlog.AddLodgingReservationRequest) (*wanderlog.TravelMutationResult, error) {
	return f.response("add_lodging", r)
}
func (f *fakeTravelWriteClient) UpdateLodgingReservation(r wanderlog.UpdateLodgingReservationRequest) (*wanderlog.TravelMutationResult, error) {
	return f.response("update_lodging", r)
}
func (f *fakeTravelWriteClient) DeleteLodgingReservation(r wanderlog.DeleteTravelReservationRequest) (*wanderlog.TravelMutationResult, error) {
	return f.response("delete_lodging", r)
}
func (f *fakeTravelWriteClient) AddTrainReservation(r wanderlog.AddTrainReservationRequest) (*wanderlog.TravelMutationResult, error) {
	return f.response("add_train", r)
}
func (f *fakeTravelWriteClient) DeleteTrainReservation(r wanderlog.DeleteTravelReservationRequest) (*wanderlog.TravelMutationResult, error) {
	return f.response("delete_train", r)
}
func fakeTravelFactory(fake *fakeTravelWriteClient) travelWriteClientFactory {
	return func() (travelWriteClient, error) { return fake, nil }
}

func TestTravelWriteCommandsAreRegistered(t *testing.T) {
	for _, path := range [][]string{{"flight", "add"}, {"flight", "update"}, {"flight", "delete"}, {"lodging", "add"}, {"lodging", "update"}, {"lodging", "delete"}, {"train", "add"}, {"train", "delete"}} {
		command, _, err := tripsCmd.Find(path)
		require.NoError(t, err)
		assert.Equal(t, path[len(path)-1], command.Name())
		assert.NotNil(t, command.RunE)
	}
}

func TestFlightAddValidatesBeforeCallingService(t *testing.T) {
	fake := &fakeTravelWriteClient{}
	command := newTripsFlightWriteCmd(fakeTravelFactory(fake))
	command.SetArgs([]string{"add", "trip-1", "--flight-number", "MU244", "--departure-date", "tomorrow"})
	assert.EqualError(t, command.Execute(), `invalid departure date "tomorrow": use YYYY-MM-DD`)
	assert.Empty(t, fake.operation)
}

func TestFlightAddCallsServiceAndFormatsJSON(t *testing.T) {
	fake := &fakeTravelWriteClient{result: &wanderlog.TravelMutationResult{Success: true, BlockID: 42, Message: "added"}}
	command := newTripsFlightWriteCmd(fakeTravelFactory(fake))
	var stdout bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&bytes.Buffer{})
	command.SetArgs([]string{"add", "trip-1", "--flight-number", "mu244", "--departure-date", "2026-06-02", "--departure-time", "09:30", "--departure-airport", "pvg", "--output", "json"})
	require.NoError(t, command.Execute())
	request := fake.request.(wanderlog.AddFlightReservationRequest)
	assert.Equal(t, "add_flight", fake.operation)
	assert.Equal(t, "MU244", request.FlightNumber)
	assert.Equal(t, "PVG", request.DepartureAirport)
	var output map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &output))
	assert.Equal(t, float64(42), output["block_id"])
}

func TestFlightUpdateRequiresChangedField(t *testing.T) {
	fake := &fakeTravelWriteClient{}
	command := newTripsFlightWriteCmd(fakeTravelFactory(fake))
	command.SetArgs([]string{"update", "trip-1", "42"})
	assert.EqualError(t, command.Execute(), "at least one flight field flag is required")
	assert.Empty(t, fake.operation)
}

func TestLodgingAndTrainCommandsCallServices(t *testing.T) {
	fake := &fakeTravelWriteClient{}
	lodging := newTripsLodgingWriteCmd(fakeTravelFactory(fake))
	lodging.SetOut(&bytes.Buffer{})
	lodging.SetArgs([]string{"add", "trip-1", "--name", "Hotel", "--lat", "1", "--lng", "2", "--check-in", "2026-06-03", "--check-out", "2026-06-04", "--traveler", "Ada"})
	require.NoError(t, lodging.Execute())
	assert.Equal(t, "add_lodging", fake.operation)
	assert.Equal(t, []string{"Ada"}, fake.request.(wanderlog.AddLodgingReservationRequest).TravelerNames)

	fake = &fakeTravelWriteClient{}
	train := newTripsTrainWriteCmd(fakeTravelFactory(fake))
	train.SetOut(&bytes.Buffer{})
	train.SetArgs([]string{"add", "trip-1", "--carrier", "SBB", "--departure-date", "2026-06-03", "--departure-name", "Zurich", "--departure-lat", "47.3", "--departure-lng", "8.5", "--arrival-place-id", "ChIJ123"})
	require.NoError(t, train.Execute())
	assert.Equal(t, "add_train", fake.operation)
	assert.Equal(t, 47.3, fake.request.(wanderlog.AddTrainReservationRequest).Departure.Latitude)

	fake = &fakeTravelWriteClient{}
	train = newTripsTrainWriteCmd(fakeTravelFactory(fake))
	train.SetOut(&bytes.Buffer{})
	train.SetArgs([]string{"delete", "trip-1", "77"})
	require.NoError(t, train.Execute())
	assert.Equal(t, "delete_train", fake.operation)
}

func TestTravelCommandAuthenticationError(t *testing.T) {
	command := newTripsFlightWriteCmd(func() (travelWriteClient, error) { return nil, errors.New("not logged in") })
	command.SetArgs([]string{"delete", "trip-1", "42"})
	assert.EqualError(t, command.Execute(), "authentication required: not logged in")
}
