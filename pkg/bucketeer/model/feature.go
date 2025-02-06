package model

type Feature struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	Enabled         bool                 `json:"enabled"`
	Deleted         bool                 `json:"deleted"`
	TTL             int32                `json:"ttl"`
	Version         int32                `json:"version"`
	CreatedAt       string               `json:"createdAt"`
	UpdatedAt       string               `json:"updatedAt"`
	Variations      []Variation          `json:"variations"`
	Targets         []Target             `json:"targets"`
	Rules           []Rule               `json:"rules"`
	DefaultStrategy *Strategy            `json:"defaultStrategy"`
	OffVariation    string               `json:"offVariation"`
	Tags            []string             `json:"tags"`
	LastUsedInfo    *FeatureLastUsedInfo `json:"lastUsedInfo"`
	Maintainer      string               `json:"maintainer"`
	VariationType   string               `json:"variationType"`
	Archived        bool                 `json:"archived"`
	Prerequisites   []*Prerequisite      `json:"prerequisites"`
	SamplingSeed    string               `json:"samplingSeed"`
}

type FeatureLastUsedInfo struct {
	FeatureID           string `json:"featureId"`
	Version             int32  `json:"version"`
	LastUsedAt          string `json:"lastUsedAt"`
	CreatedAt           string `json:"createdAt"`
	ClientOldestVersion string `json:"clientOldestVersion"`
	ClientLatestVersion string `json:"clientLatestVersion"`
}

type Prerequisite struct {
	FeatureID   string `json:"featureId"`
	VariationID string `json:"variationId"`
}
