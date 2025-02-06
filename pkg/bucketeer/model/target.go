package model

type Target struct {
	Variation string   `json:"variation"`
	Users     []string `json:"users"`
}
