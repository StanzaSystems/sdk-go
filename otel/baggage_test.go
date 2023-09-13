package otel

import (
	"context"
	"strconv"
	"testing"

	"github.com/StanzaSystems/sdk-go/keys"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/baggage"
	"google.golang.org/protobuf/proto"
)

func TestGetFeature(t *testing.T) {
	type test struct {
		testName    string
		featureName *string
		otelBaggage string
		uberCtx     string
		otBaggage   string
		want        *string
	}

	tests := []test{
		{
			testName:    "TEST 1",
			featureName: proto.String("MyFeature"),
			otelBaggage: "",
			uberCtx:     "",
			otBaggage:   "",
			want:        proto.String("MyFeature"),
		},
		{
			testName:    "TEST 2",
			featureName: proto.String("MyFeature"),
			otelBaggage: "FeatureFromOtel",
			uberCtx:     "",
			otBaggage:   "",
			want:        proto.String("MyFeature"),
		},
		{
			testName:    "TEST 3",
			featureName: nil,
			otelBaggage: "FeatureFromOtel",
			uberCtx:     "",
			otBaggage:   "",
			want:        proto.String("FeatureFromOtel"),
		},
		{
			testName:    "TEST 4",
			featureName: nil,
			otelBaggage: "FeatureFromOtel",
			uberCtx:     "FeatureFromUberCtx",
			otBaggage:   "",
			want:        proto.String("FeatureFromOtel"),
		},
		{
			testName:    "TEST 5",
			featureName: nil,
			otelBaggage: "",
			uberCtx:     "FeatureFromUberCtx",
			otBaggage:   "",
			want:        proto.String("FeatureFromUberCtx"),
		},
		{
			testName:    "TEST 6",
			featureName: nil,
			otelBaggage: "",
			uberCtx:     "FeatureFromUberCtx",
			otBaggage:   "FeatureFromOTbaggage",
			want:        proto.String("FeatureFromUberCtx"),
		},
		{
			testName:    "TEST 7",
			featureName: nil,
			otelBaggage: "",
			uberCtx:     "",
			otBaggage:   "FeatureFromOTbaggage",
			want:        proto.String("FeatureFromOTbaggage"),
		},
		{
			testName:    "TEST 8",
			featureName: nil,
			otelBaggage: "",
			uberCtx:     "",
			otBaggage:   "",
			want:        nil,
		},
		{
			testName:    "TEST 9",
			featureName: proto.String("MyFeature"),
			otelBaggage: "FeatureFromOtel",
			uberCtx:     "FeatureFromUberCtx",
			otBaggage:   "FeatureFromOTbaggage",
			want:        proto.String("MyFeature"),
		},
	}

	for _, tc := range tests {
		ctx := context.Background()
		if tc.otelBaggage != "" {
			if stzFeat, err := baggage.NewMember(keys.StzFeat, tc.otelBaggage); err == nil {
				if bag, err := baggage.FromContext(ctx).SetMember(stzFeat); err == nil {
					ctx = baggage.ContextWithBaggage(ctx, bag)
				}
			}

		}
		if tc.uberCtx != "" {
			ctx = context.WithValue(ctx, keys.UberctxStzFeatKey, tc.uberCtx)
		}
		if tc.otBaggage != "" {
			ctx = context.WithValue(ctx, keys.OtStzFeatKey, tc.otBaggage)
		}
		_, got := GetFeature(ctx, tc.featureName)
		assert.Equal(t, got, tc.want)
		// TODO: also verify context!
	}
}

func TestGetPriorityBoost(t *testing.T) {
	type test struct {
		testName      string
		priorityBoost *int32
		otelBaggage   int32
		uberCtx       int32
		otBaggage     int32
		want          *int32
	}

	tests := []test{
		{
			testName:      "TEST 1",
			priorityBoost: proto.Int32(5),
			otelBaggage:   0,
			uberCtx:       0,
			otBaggage:     0,
			want:          proto.Int32(5),
		},
		{
			testName:      "TEST 2",
			priorityBoost: proto.Int32(5),
			otelBaggage:   5,
			uberCtx:       0,
			otBaggage:     0,
			want:          proto.Int32(10),
		},
		{
			testName:      "TEST 3",
			priorityBoost: proto.Int32(5),
			otelBaggage:   5,
			uberCtx:       5,
			otBaggage:     0,
			want:          proto.Int32(10),
		},
		{
			testName:      "TEST 4",
			priorityBoost: proto.Int32(5),
			otelBaggage:   5,
			uberCtx:       5,
			otBaggage:     5,
			want:          proto.Int32(10),
		},
		{
			testName:      "TEST 5",
			priorityBoost: nil,
			otelBaggage:   5,
			uberCtx:       5,
			otBaggage:     5,
			want:          proto.Int32(5),
		},
		{
			testName:      "TEST 6",
			priorityBoost: nil,
			otelBaggage:   0,
			uberCtx:       5,
			otBaggage:     5,
			want:          proto.Int32(5),
		},
		{
			testName:      "TEST 7",
			priorityBoost: nil,
			otelBaggage:   0,
			uberCtx:       0,
			otBaggage:     5,
			want:          proto.Int32(5),
		},
		{
			testName:      "TEST 8",
			priorityBoost: nil,
			otelBaggage:   0,
			uberCtx:       0,
			otBaggage:     0,
			want:          nil,
		},
	}

	for _, tc := range tests {
		ctx := context.Background()
		if tc.otelBaggage != 0 {
			boostStr := strconv.Itoa(int(tc.otelBaggage))
			if stzBoost, err := baggage.NewMember(keys.StzBoost, boostStr); err == nil {
				if bag, err := baggage.FromContext(ctx).SetMember(stzBoost); err == nil {
					ctx = baggage.ContextWithBaggage(ctx, bag)
				}
			}
		}
		if tc.uberCtx != 0 {
			ctx = context.WithValue(ctx, keys.UberctxStzBoostKey, strconv.Itoa(int(tc.uberCtx)))
		}
		if tc.otBaggage != 0 {
			ctx = context.WithValue(ctx, keys.OtStzBoostKey, strconv.Itoa(int(tc.otBaggage)))
		}
		_, got := GetPriorityBoost(ctx, tc.priorityBoost)
		assert.Equal(t, got, tc.want)
		// TODO: also verify context!
	}
}
