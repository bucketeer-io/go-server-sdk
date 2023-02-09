package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"
)

const (
	tag                  = "tag"
	value                = "1s"
	goalID               = "goalID"
	id                   = "id"
	featureVersion       = 7
	variationID          = "vid"
	variationValue       = "value"
	sizeByte       int32 = 1000
	featureID            = "fid"
)

func TestNewEvent(t *testing.T) {
	t.Parallel()
	id := "sample"
	encoded := []byte{}
	e := NewEvent(id, encoded)
	assert.IsType(t, &Event{}, e)
	assert.Equal(t, e.ID, id)
	assert.Equal(t, e.Event, json.RawMessage(encoded))
}

func newUser(t *testing.T, id string) *user.User {
	t.Helper()
	return &user.User{ID: id}
}
