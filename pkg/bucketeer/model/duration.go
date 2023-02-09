package model

type wellKnownTypes string

const DurationType wellKnownTypes = "type.googleapis.com/google.protobuf.Duration"

type Duration struct {
	Type  wellKnownTypes `json:"@type,omitempty"`
	Value string         `json:"value,omitempty"`
}
