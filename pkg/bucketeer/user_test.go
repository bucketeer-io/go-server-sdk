package bucketeer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	protouser "github.com/ca-dp/bucketeer-go-server-sdk/proto/user"
)

func TestNewUser(t *testing.T) {
	t.Parallel()
	userID := "user-id"
	userAttrs := map[string]string{"foo": "bar"}
	tests := []struct {
		desc       string
		id         string
		attributes map[string]string
		expected   *User
		isErr      bool
	}{
		{
			desc:       "error: empty id",
			id:         "",
			attributes: nil,
			expected:   nil,
			isErr:      true,
		},
		{
			desc:       "user without attributes",
			id:         userID,
			attributes: nil,
			expected: &User{User: &protouser.User{
				Id: userID,
			}},
			isErr: false,
		},
		{
			desc:       "user with attributes",
			id:         userID,
			attributes: userAttrs,
			expected: &User{User: &protouser.User{
				Id:   userID,
				Data: userAttrs,
			}},
			isErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			u, err := NewUser(tt.id, tt.attributes)
			if tt.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, u)
		})
	}
}
