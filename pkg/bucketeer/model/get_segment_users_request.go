package model

type GetSegmentUsersRequest struct {
	SegmentIDs  []string     `json:"segmentIds"`
	RequestedAt int64        `json:"requestedAt"`
	SourceID    SourceIDType `json:"sourceId"`
	SDKVersion  string       `json:"sdkVersion"`
}

func NewGetSegmentUsersRequest(
	segmentIDs []string,
	requestedAt int64,
	sdkVersion string,
	sourceID SourceIDType,
) *GetSegmentUsersRequest {
	return &GetSegmentUsersRequest{
		SegmentIDs:  segmentIDs,
		RequestedAt: requestedAt,
		SourceID:    sourceID,
		SDKVersion:  sdkVersion,
	}
}
