package model

type GetFeatureFlagsRequest struct {
	Tag            string       `json:"tag"`
	FeatureFlagsID string       `json:"featureFlagsId"`
	RequestedAt    int64        `json:"requestedAt"`
	SourceID       SourceIDType `json:"sourceId"`
	SDKVersion     string       `json:"sdkVersion"`
}

func NewGetFeatureFlagsRequest(
	tag, featureFlagsID, sdkVersion string,
	requestedAt int64,
) *GetFeatureFlagsRequest {
	return &GetFeatureFlagsRequest{
		Tag:            tag,
		FeatureFlagsID: featureFlagsID,
		RequestedAt:    requestedAt,
		SourceID:       SourceIDGoServer,
		SDKVersion:     sdkVersion,
	}
}
