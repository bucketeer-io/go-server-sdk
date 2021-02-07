package bucketeer

import (
	"context"

	"github.com/google/uuid"
)

type key int

const (
	keyUserID key = iota
	keyAttributes
)

var DefaultUserID = newDefaultUserID()

func newDefaultUserID() string {
	return uuid.New().String()
}

func New(ctx context.Context, userID string, attrs map[string]string) context.Context {
	ctx = context.WithValue(ctx, keyUserID, userID)
	return context.WithValue(ctx, keyAttributes, attrs)
}

func NewDefault(ctx context.Context) context.Context {
	return context.WithValue(ctx, keyUserID, DefaultUserID)
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
