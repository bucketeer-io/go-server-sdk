package bucketeer

import "context"

// SDK is the Bucketeer SDK.
//
// SDK represents the ability to get the value of a feature flag and to track goal events
// by communicating with the Bucketeer servers.
//
// SDK must be safe for concurrent use by multiple goroutines.
type SDK interface {
	// BoolVariation returns the value of a feature flag (whose variations are booleans) for the given user.
	//
	// BoolVariation returns defaultValue if an error occurs.
	BoolVariation(ctx context.Context, featureID string, user *User, defaultValue bool) bool

	// IntVariation returns the value of a feature flag (whose variations are ints) for the given user.
	//
	// IntVariation returns defaultValue if an error occurs.
	IntVariation(ctx context.Context, featureID string, user *User, defaultValue int) int

	// Int64Variation returns the value of a feature flag (whose variations are int64s) for the given user.
	//
	// Int64Variation returns defaultValue if an error occurs.
	Int64Variation(ctx context.Context, featureID string, user *User, defaultValue int64) int64

	// Float64Variation returns the value of a feature flag (whose variations are float64s) for the given user.
	//
	// Float64Variation returns defaultValue if an error occurs.
	Float64Variation(ctx context.Context, featureID string, user *User, defaultValue float64) float64

	// StringVariation returns the value of a feature flag (whose variations are strings) for the given user.
	//
	// StringVariation returns defaultValue if an error occurs.
	StringVariation(ctx context.Context, featureID, user *User, defaultValue string) string

	// JSONVariation parses the value of a feature flag (whose variations are jsons) for the given user,
	// and stores the result in dst.
	JSONVariation(ctx context.Context, featureID string, user *User, dst interface{})

	// Track reports that a user has performed a goal event.
	Track(ctx context.Context, user *User, goalID string)

	// TrackValue reports that a user has performed a goal event, and associates it with a custom value.
	TrackValue(ctx context.Context, user *User, goalID string, value float64)
}

// TODO: implement below

type sdk struct{}

type options struct{}

type Option func(*options)

func NewSDK(ctx context.Context, opts ...Option) (SDK, error) {
	return &sdk{}, nil
}

func (s *sdk) BoolVariation(ctx context.Context, featureID string, user *User, defaultValue bool) bool {
	return false
}

func (s *sdk) IntVariation(ctx context.Context, featureID string, user *User, defaultValue int) int {
	return 0
}

func (s *sdk) Int64Variation(ctx context.Context, featureID string, user *User, defaultValue int64) int64 {
	return 0
}

func (s *sdk) Float64Variation(ctx context.Context, featureID string, user *User, defaultValue float64) float64 {
	return 0.0
}

func (s *sdk) StringVariation(ctx context.Context, featureID, user *User, defaultValue string) string {
	return ""
}

func (s *sdk) JSONVariation(ctx context.Context, featureID string, user *User, dst interface{}) {
}

func (s *sdk) Track(ctx context.Context, user *User, goalID string) {
}

func (s *sdk) TrackValue(ctx context.Context, user *User, goalID string, value float64) {
}
