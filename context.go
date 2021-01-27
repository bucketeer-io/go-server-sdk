package bucketeer_go_sdk

import "context"

type key int

const (
	keyUserID key = iota
	keyAttributes
)

func NewContext(ctx context.Context, userID string, attrs map[string]string) context.Context {
	ctx = context.WithValue(ctx, keyUserID, userID)
	return context.WithValue(ctx, keyAttributes, attrs)
}

func userID(ctx context.Context) string {
	id, ok := ctx.Value(keyUserID).(string)
	if !ok {
		return ""
	}
	return id
}

func attributes(ctx context.Context) map[string]string {
	attrs, ok := ctx.Value(keyAttributes).(map[string]string)
	if !ok {
		return nil
	}
	return attrs
}
