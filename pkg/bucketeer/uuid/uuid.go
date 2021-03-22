package uuid

import (
	"crypto/rand"
	"fmt"
)

// UUID is a 128 bit (16 byte) Universal Unique IDentifier as defined in RFC 4122.
// https://tools.ietf.org/html/rfc4122
type UUID [16]byte

// String returns the string form of uuid.
func (uuid *UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

// NewV4 creates a new random (version 4) UUID.
func NewV4() (*UUID, error) {
	uuid := &UUID{}
	if _, err := rand.Read(uuid[:]); err != nil {
		return nil, fmt.Errorf("bucketeer/uuid: failed to read rand: %w", err)
	}
	uuid.setV4Variant()
	uuid.setV4Version()
	return uuid, nil
}

func (uuid *UUID) setV4Variant() {
	uuid[8] = (uuid[8] & 0x3f) | 0x80
}

func (uuid *UUID) setV4Version() {
	uuid[6] = (uuid[6] & 0x0f) | 0x40
}
