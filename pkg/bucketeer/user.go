package bucketeer

import (
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"
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
func NewUser(id string, attributes map[string]string) *User {
	return &User{User: &protouser.User{
		Id:   id,
		Data: attributes,
	}}
}

// Valid returns true if valid user, otherwise returns false.
func (u *User) Valid() bool {
	if u == nil {
		return false
	}
	if u.Id == "" {
		return false
	}
	return true
}

// NewUser creates a new User.
//
// id is mandatory and attributes is optional.
func NewRestUser(id string, attributes map[string]string) *user.User {
	return &user.User{
		ID:   id,
		Data: attributes,
	}
}
