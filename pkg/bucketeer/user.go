package bucketeer

import (
	"errors"
	"fmt"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/internal/uuid"
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
		return nil, errors.New("bucketeer: id is empty")
	}
	return &User{User: &protouser.User{
		Id:   id,
		Data: attributes,
	}}, nil
}

// NewRandomUser creates a new User with a random id.
//
// attributes is optional.
func NewRandomUser(attributes map[string]string) (*User, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("bucketeer: failed to create uuid: %w", err)
	}
	return &User{User: &protouser.User{
		Id:   id.String(),
		Data: attributes,
	}}, nil
}
