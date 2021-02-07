package bucketeer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewError(t *testing.T) {
	const msg = "message"
	original := errors.New("error")
	err := &Error{
		Message:     msg,
		originalErr: original,
	}

	require.Equal(t, "bucketeer: "+msg, err.Error())
	require.Equal(t, original, err.Unwrap())
}
