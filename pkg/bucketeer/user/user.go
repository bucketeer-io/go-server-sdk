package user

// User is the Bucketeer user.
//
// User contains a mandatory user id and optional attributes (Data) for the user targeting and the analysis.

type User struct {
	ID   string            `json:"id,omitempty"`
	Data map[string]string `json:"data,omitempty"`
}
