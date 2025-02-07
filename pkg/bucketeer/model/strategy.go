package model

type Strategy struct {
	Type            string           `json:"type"`
	FixedStrategy   *FixedStrategy   `json:"fixedStrategy"`
	RolloutStrategy *RolloutStrategy `json:"rolloutStrategy"`
}
