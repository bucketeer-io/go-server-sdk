package bucketeer

import (
	"testing"

	protouser "github.com/ca-dp/bucketeer-go-server-sdk/proto/user"
	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	t.Parallel()
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
			id:         "id",
			attributes: nil,
			expected: &User{User: &protouser.User{
				Id: "id",
			}},
			isErr: false,
		},
		{
			desc:       "user with attributes",
			id:         "id",
			attributes: map[string]string{"foo": "bar"},
			expected: &User{User: &protouser.User{
				Id:   "id",
				Data: map[string]string{"foo": "bar"},
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
