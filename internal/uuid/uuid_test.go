package uuid

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var v4Regex = regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")

func TestNewV4(t *testing.T) {
	for cnt := 0; cnt < 32; cnt++ {
		id, err := NewV4()
		require.NoError(t, err)
		assert.Regexp(t, v4Regex, id.String())
	}
}
