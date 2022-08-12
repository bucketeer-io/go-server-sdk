package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValid(t *testing.T) {
	t.Parallel()
	UserID := "user-id"
	userAttrs := map[string]string{"foo": "bar"}
	tests := []struct {
		desc  string
		user  *User
		valid bool
	}{
		{
			desc:  "return false when user is nil",
			user:  nil,
			valid: false,
		},
		{
			desc:  "return false when user id is empty",
			user:  NewUser("", nil),
			valid: false,
		},
		{
			desc:  "return true when user attributes is nil",
			user:  NewUser(UserID, nil),
			valid: true,
		},
		{
			desc:  "return true",
			user:  NewUser(UserID, userAttrs),
			valid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.user.Valid())
		})
	}
}
