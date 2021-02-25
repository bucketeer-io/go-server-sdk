package bucketeer

import "github.com/ca-dp/bucketeer-go-server-sdk/proto/user"

// User is the Bucketeer user.
//
// User contains a mandatory user id and optional attributes (Data) for the user targeting and the analysis.
type User struct {
	*user.User
}

// NewUser is the constructor for User.
//
// id is mandatory and attributes is optional.
func NewUser(id string, attributes map[string]string) *User {
	return &User{User: &user.User{
		Id:   id,
		Data: attributes,
	}}
}
