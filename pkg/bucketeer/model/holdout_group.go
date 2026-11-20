package model

import (
	"strconv"

	ftproto "github.com/bucketeer-io/bucketeer/v2/proto/feature"
)

// HoldoutGroup is a holdout group that is serving in the environment.
//
// The server sends only the fields that local evaluation reads.
// See https://github.com/bucketeer-io/bucketeer/blob/main/proto/feature/holdout.proto
type HoldoutGroup struct {
	ID             string                `json:"id"`
	EnvironmentID  string                `json:"environmentId"`
	AssignmentType string                `json:"assignmentType"`
	SamplingSeed   string                `json:"samplingSeed"`
	BucketStart    float64               `json:"bucketStart"`
	Percentage     float64               `json:"percentage"`
	SegmentID      string                `json:"segmentId"`
	Status         string                `json:"status"`
	UpdatedAt      string                `json:"updatedAt"`
	Features       []HoldoutGroupFeature `json:"features"`
}

// HoldoutGroupFeature is a feature flag targeted by a holdout group,
// with the variation served to the held-out users.
type HoldoutGroupFeature struct {
	ID                 string `json:"id"`
	HoldoutGroupID     string `json:"holdoutGroupId"`
	FeatureID          string `json:"featureId"`
	EnvironmentID      string `json:"environmentId"`
	HoldoutVariationID string `json:"holdoutVariationId"`
	AddedAt            string `json:"addedAt"`
}

func convertHoldoutGroupModel(g HoldoutGroup) *ftproto.HoldoutGroup {
	updatedAt, _ := strconv.ParseInt(g.UpdatedAt, 10, 64)
	return &ftproto.HoldoutGroup{
		Id:             g.ID,
		EnvironmentId:  g.EnvironmentID,
		AssignmentType: convertHoldoutAssignmentType(g.AssignmentType),
		SamplingSeed:   g.SamplingSeed,
		BucketStart:    g.BucketStart,
		Percentage:     g.Percentage,
		SegmentId:      g.SegmentID,
		Status:         ftproto.HoldoutGroup_Status(ftproto.HoldoutGroup_Status_value[g.Status]),
		UpdatedAt:      updatedAt,
		Features:       mapFields(g.Features, convertHoldoutGroupFeatureModel),
	}
}

// unknownHoldoutAssignmentType is not a value of the proto enum, so the
// evaluation module holds nobody out for it.
const unknownHoldoutAssignmentType = ftproto.HoldoutGroup_AssignmentType(-1)

// convertHoldoutAssignmentType maps the enum name sent by the server.
//
// A name this SDK version does not know must not fall back to the zero value:
// that is HASH, and the group would be bucketed by a rule it was not created
// with. An absent name is the proto3 default, which is HASH.
func convertHoldoutAssignmentType(name string) ftproto.HoldoutGroup_AssignmentType {
	if name == "" {
		return ftproto.HoldoutGroup_HASH
	}
	v, ok := ftproto.HoldoutGroup_AssignmentType_value[name]
	if !ok {
		return unknownHoldoutAssignmentType
	}
	return ftproto.HoldoutGroup_AssignmentType(v)
}

func convertHoldoutGroupFeatureModel(f HoldoutGroupFeature) *ftproto.HoldoutGroupFeature {
	addedAt, _ := strconv.ParseInt(f.AddedAt, 10, 64)
	return &ftproto.HoldoutGroupFeature{
		Id:                 f.ID,
		HoldoutGroupId:     f.HoldoutGroupID,
		FeatureId:          f.FeatureID,
		EnvironmentId:      f.EnvironmentID,
		HoldoutVariationId: f.HoldoutVariationID,
		AddedAt:            addedAt,
	}
}
