package model

import (
	"strconv"

	ftproto "github.com/bucketeer-io/bucketeer/v2/proto/feature"
	gwproto "github.com/bucketeer-io/bucketeer/v2/proto/gateway"
)

type GetFeatureFlagsResponse struct {
	FeatureFlagsID         string    `json:"featureFlagsId"`
	Features               []Feature `json:"features"`
	ArchivedFeatureFlagIDs []string  `json:"archivedFeatureFlagIds"`
	RequestedAt            string    `json:"requestedAt"`
	ForceUpdate            bool      `json:"forceUpdate"`
}

func ConvertFeatureFlagsResponse(response *GetFeatureFlagsResponse) *gwproto.GetFeatureFlagsResponse {
	requestedAt, _ := strconv.ParseInt(response.RequestedAt, 10, 64)
	pbResp := &gwproto.GetFeatureFlagsResponse{
		FeatureFlagsId:         response.FeatureFlagsID,
		Features:               mapFields(response.Features, convertFeatureModel),
		ArchivedFeatureFlagIds: response.ArchivedFeatureFlagIDs,
		RequestedAt:            requestedAt,
		ForceUpdate:            response.ForceUpdate,
	}
	return pbResp
}

func convertFeatureModel(f Feature) *ftproto.Feature {
	createdAt, _ := strconv.ParseInt(f.CreatedAt, 10, 64)
	updatedAt, _ := strconv.ParseInt(f.UpdatedAt, 10, 64)
	pbFeature := &ftproto.Feature{
		Id:          f.ID,
		Name:        f.Name,
		Description: f.Description,
		Enabled:     f.Enabled,
		Deleted:     f.Deleted,
		Ttl:         f.TTL,
		Version:     f.Version,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Variations: mapFields(f.Variations, func(v Variation) *ftproto.Variation {
			return &ftproto.Variation{
				Id:          v.ID,
				Value:       v.Value,
				Name:        v.Name,
				Description: v.Description,
			}
		}),
		Targets: mapFields(f.Targets, func(t Target) *ftproto.Target {
			return &ftproto.Target{
				Variation: t.Variation,
				Users:     t.Users,
			}
		}),
		Rules:           mapFields(f.Rules, convertRuleModel),
		DefaultStrategy: convertStrategyModel(f.DefaultStrategy),
		OffVariation:    f.OffVariation,
		Tags:            f.Tags,
		Maintainer:      f.Maintainer,
		VariationType:   ftproto.Feature_VariationType(ftproto.Feature_VariationType_value[f.VariationType]),
		Archived:        f.Archived,
		Prerequisites: mapFields(f.Prerequisites, func(p *Prerequisite) *ftproto.Prerequisite {
			return &ftproto.Prerequisite{
				FeatureId:   p.FeatureID,
				VariationId: p.VariationID,
			}
		}),
		SamplingSeed: f.SamplingSeed,
	}
	if f.LastUsedInfo != nil {
		lastUsedAt, _ := strconv.ParseInt(f.LastUsedInfo.LastUsedAt, 10, 64)
		createdAt, _ := strconv.ParseInt(f.LastUsedInfo.CreatedAt, 10, 64)
		pbFeature.LastUsedInfo = &ftproto.FeatureLastUsedInfo{
			FeatureId:           f.LastUsedInfo.FeatureID,
			Version:             f.LastUsedInfo.Version,
			LastUsedAt:          lastUsedAt,
			CreatedAt:           createdAt,
			ClientOldestVersion: f.LastUsedInfo.ClientOldestVersion,
			ClientLatestVersion: f.LastUsedInfo.ClientLatestVersion,
		}
	}
	return pbFeature
}

func convertStrategyModel(s *Strategy) *ftproto.Strategy {
	if s == nil {
		return nil
	}

	pbStrategy := &ftproto.Strategy{
		Type: ftproto.Strategy_Type(ftproto.Strategy_Type_value[s.Type]),
	}

	if s.FixedStrategy != nil {
		pbStrategy.FixedStrategy = &ftproto.FixedStrategy{
			Variation: s.FixedStrategy.Variation,
		}
	}
	if s.RolloutStrategy != nil {
		pbStrategy.RolloutStrategy = &ftproto.RolloutStrategy{
			Variations: mapFields(
				s.RolloutStrategy.Variations,
				func(v RolloutStrategyVariation) *ftproto.RolloutStrategy_Variation {
					return &ftproto.RolloutStrategy_Variation{
						Variation: v.Variation,
						Weight:    v.Weight,
					}
				},
			),
		}
	}

	return pbStrategy
}

func convertRuleModel(r Rule) *ftproto.Rule {
	rule := &ftproto.Rule{
		Id:       r.ID,
		Strategy: convertStrategyModel(r.Strategy),
		Clauses: mapFields(r.Clauses, func(c Clause) *ftproto.Clause {
			return &ftproto.Clause{
				Id:        c.ID,
				Attribute: c.Attribute,
				Operator:  ftproto.Clause_Operator(ftproto.Clause_Operator_value[c.Operator]),
				Values:    c.Values,
			}
		}),
	}
	return rule
}
