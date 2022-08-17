package user

// User is the Bucketeer user.
//
// User contains a mandatory user id and optional attributes (Data) for the user targeting and the analysis.

type User struct {
	ID   string            `json:"id,omitempty"`
	Data map[string]string `json:"data,omitempty"`
}

func NewUser(id string, attributes map[string]string) *User {
	return &User{
		ID:   id,
		Data: attributes,
	}
}

// Valid returns true if valid user, otherwise returns false.
func (u *User) Valid() bool {
	if u == nil {
		return false
	}
	if u.ID == "" {
		return false
	}
	return true
}
