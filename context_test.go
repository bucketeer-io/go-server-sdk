package bucketeer_go_sdk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	const id = "userid"
	attrs := map[string]string{"foo": "bar"}

	ctx := context.Background()
	ctx = NewContext(ctx, id, attrs)

	require.Equal(t, id, userID(ctx))
	require.Equal(t, attrs, attributes(ctx))
}
