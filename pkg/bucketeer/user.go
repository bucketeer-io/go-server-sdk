package bucketeer

import (
	"errors"

	protouser "github.com/ca-dp/bucketeer-go-server-sdk/proto/user"
)

// User is the Bucketeer user.
//
// User contains a mandatory user id and optional attributes (Data) for the user targeting and the analysis.
type User struct {
	*protouser.User
}

// NewUser creates a new User.
//
// id is mandatory and attributes is optional.
func NewUser(id string, attributes map[string]string) (*User, error) {
	if id == "" {
		return nil, errors.New("bucketeer: user id is empty")
	}
	return &User{User: &protouser.User{
		Id:   id,
		Data: attributes,
	}}, nil
}
