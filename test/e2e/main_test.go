package e2e

import (
	"flag"
	"time"
)

const (
	timeout = 20 * time.Second

	tag                 = "go-server"
	sdkVersion          = "1.0.0"
	sourceID            = 5
	userID              = "bucketeer-go-server-user-id-1"
	featureID           = "feature-go-server-e2e-string"
	featureIDVariation1 = "value-1"
	featureIDVariation2 = "value-2"
	goalID              = "goal-go-server-e2e-1"

	// Sdk Test
	targetUserID                   = "bucketeer-go-server-user-id-1"
	targetSegmentUserID            = "bucketeer-go-server-user-id-2" // This ID is configured in the segment user on the console
	featureIDString                = "feature-go-server-e2e-string"
	featureIDStringTargetVariation = featureIDStringVariation2
	featureIDStringVariation1      = "value-1"
	featureIDStringVariation2      = "value-2"
	featureIDStringVariation3      = "value-3"

	featureIDBoolean                = "feature-go-server-e2e-boolean"
	featureIDBooleanTargetVariation = false

	featureIDInt                = "feature-go-server-e2e-int"
	featureIDIntTargetVariation = featureIDIntVariation2
	featureIDIntVariation1      = 10
	featureIDIntVariation2      = 20

	featureIDInt64                = "feature-go-server-e2e-int64"
	featureIDInt64TargetVariation = featureIDInt64Variation2
	featureIDInt64Variation1      = 3000000000
	featureIDInt64Variation2      = -3000000000

	featureIDFloat                = "feature-go-server-e2e-float"
	featureIDFloatTargetVariation = featureIDFloatVariation2
	featureIDFloatVariation1      = 2.1
	featureIDFloatVariation2      = 3.1

	featureIDJson = "feature-go-server-e2e-json"

	// Rule-based segment tests
	//
	// These tests require the following fixtures to be configured in the test environment
	// (Bucketeer server 2.3.0 or later):
	//
	// Segment "go-server-e2e-rule-based" (mixed: uploaded user list AND rules):
	//   - Uploaded included-user list: bucketeer-go-server-user-id-3
	//   - Rule 1 (clauses are AND-ed):
	//       country EQUALS "japan"
	//       AND age GREATER "19"
	//       AND age LESS "60"
	//   - Rule 2 (clauses are AND-ed):
	//       email STARTS_WITH "test@"
	//       AND plan IN ["premium", "enterprise"]
	//   (Rules are OR-ed: a user matching either rule is in the segment)
	//
	// Flag "feature-go-server-e2e-rule-based-segment" (string):
	//   - Variations: value-1, value-2
	//   - Targeting rule: user is included in segment "go-server-e2e-rule-based" -> value-2
	//   - Default strategy: value-1
	//
	// Flag "feature-go-server-e2e-segment-attribute" (string):
	//   - Variations: value-1, value-2
	//   - Targeting rule (single rule, two AND-ed clauses):
	//       user is included in segment "go-server-e2e-rule-based"
	//       AND region EQUALS "tokyo"
	//     -> value-2
	//   - Default strategy: value-1
	featureIDRuleBasedSegment        = "feature-go-server-e2e-rule-based-segment"
	featureIDSegmentAndAttribute     = "feature-go-server-e2e-segment-attribute"
	ruleBasedSegmentDefaultVariation = "value-1"
	ruleBasedSegmentMatchedVariation = "value-2"
	ruleBasedSegmentListedUserID     = "bucketeer-go-server-user-id-3" // Uploaded to the segment's user list on the console
)

var (
	apiKey       = flag.String("api-key", "", "API key for the Bucketeer service")
	apiKeyServer = flag.String("api-key-server", "", "API key for Server SDK")
	apiEndpoint  = flag.String("api-endpoint", "", "API Endpoint for the Bucketeer service, e.g. api.example.com")
	scheme       = flag.String("scheme", "https", "Scheme of the Bucketeer service, e.g. https")
)
