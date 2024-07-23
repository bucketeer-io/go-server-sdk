package model

type Evaluation struct {
	ID             string  `json:"id,omitempty"`
	FeatureID      string  `json:"featureId,omitempty"`
	FeatureVersion int32   `json:"featureVersion,omitempty"`
	UserID         string  `json:"userId,omitempty"`
	VariationID    string  `json:"variationId,omitempty"`
	Reason         *Reason `json:"reason,omitempty"`
	VariationValue string  `json:"variationValue,omitempty"`
	VariationName  string  `json:"variationName,omitempty"`
}
