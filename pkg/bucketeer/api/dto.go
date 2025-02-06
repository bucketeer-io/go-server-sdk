package api

import (
	"strconv"

	ftproto "github.com/bucketeer-io/bucketeer/proto/feature"
	gwproto "github.com/bucketeer-io/bucketeer/proto/gateway"
)

type GetFeatureFlagsResponse struct {
	FeatureFlagsID         string    `json:"featureFlagsId"`
	Features               []Feature `json:"features"`
	ArchivedFeatureFlagIDs []string  `json:"archivedFeatureFlagIds"`
	RequestedAt            string    `json:"requestedAt"`
	ForceUpdate            bool      `json:"forceUpdate"`
}

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

type Variation struct {
	ID          string `json:"id"`
	Value       string `json:"value"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Target struct {
	Variation string   `json:"variation"`
	Users     []string `json:"users"`
}

type Rule struct {
	ID       string    `json:"id"`
	Strategy *Strategy `json:"strategy"`
	Clauses  []Clause  `json:"clauses"`
}

type Strategy struct {
	Type            string           `json:"type"`
	FixedStrategy   *FixedStrategy   `json:"fixedStrategy"`
	RolloutStrategy *RolloutStrategy `json:"rolloutStrategy"`
}

type Clause struct {
	ID        string   `json:"id"`
	Attribute string   `json:"attribute"`
	Operator  string   `json:"operator"`
	Values    []string `json:"values"`
}

type FixedStrategy struct {
	Variation string `json:"variation"`
}

type RolloutStrategy struct {
	Variations []RolloutStrategyVariation `json:"variations"`
}

type RolloutStrategyVariation struct {
	Variation string `json:"variation"`
	Weight    int32  `json:"weight"`
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

func TransformFeatureFlagResponseDTO(response GetFeatureFlagsResponse) *gwproto.GetFeatureFlagsResponse {
	requestedAt, _ := strconv.ParseInt(response.RequestedAt, 10, 64)
	pbResp := &gwproto.GetFeatureFlagsResponse{
		FeatureFlagsId:         response.FeatureFlagsID,
		Features:               mapFields(response.Features, TransformFeatureDTO),
		ArchivedFeatureFlagIds: response.ArchivedFeatureFlagIDs,
		RequestedAt:            requestedAt,
		ForceUpdate:            response.ForceUpdate,
	}
	return pbResp
}

func TransformFeatureDTO(f Feature) *ftproto.Feature {
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
		Rules:           mapFields(f.Rules, TransformRuleDTO),
		DefaultStrategy: TransformStrategyDTO(f.DefaultStrategy),
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

func TransformStrategyDTO(s *Strategy) *ftproto.Strategy {
	pbStrategy := &ftproto.Strategy{
		FixedStrategy:   &ftproto.FixedStrategy{},
		RolloutStrategy: &ftproto.RolloutStrategy{},
	}
	if s == nil {
		return pbStrategy
	}

	pbStrategy.Type = ftproto.Strategy_Type(ftproto.Strategy_Type_value[s.Type])

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

func TransformRuleDTO(r Rule) *ftproto.Rule {
	rule := &ftproto.Rule{
		Id:       r.ID,
		Strategy: TransformStrategyDTO(r.Strategy),
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

func mapFields[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}
