package bucketeer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	const id = "userid"
	attrs := map[string]string{"foo": "bar"}

	ctx := context.Background()
	ctx = New(ctx, id, attrs)

	// test: success
	require.Equal(t, id, userID(ctx))
	require.Equal(t, attrs, attributes(ctx))

	// test: nil
	ctx = context.Background()
	require.Empty(t, userID(ctx))
	require.Empty(t, attributes(ctx))
}
