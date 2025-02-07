package model

type RolloutStrategy struct {
	Variations []RolloutStrategyVariation `json:"variations"`
}

type RolloutStrategyVariation struct {
	Variation string `json:"variation"`
	Weight    int32  `json:"weight"`
}
