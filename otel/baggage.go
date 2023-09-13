package otel

import (
	"context"
	"net/http"
	"strconv"

	"github.com/StanzaSystems/sdk-go/keys"
	"google.golang.org/protobuf/proto"

	"go.opentelemetry.io/otel/baggage"
)

func GetFeature(ctx context.Context, fn *string) (context.Context, *string) {
	var feat *string
	if fn == nil { // If Feature Name is nil, inspect baggage
		featFromBaggage := baggage.FromContext(ctx).Member(keys.StzFeat).Value()
		if featFromBaggage != "" { // Otherwise inspect OTEL baggage
			feat = proto.String(featFromBaggage)
		} else if ctx.Value(keys.UberctxStzFeatKey) != nil { // Otherwise inspect Jaeger uberctx
			feat = proto.String(ctx.Value(keys.UberctxStzFeatKey).(string))
		} else if ctx.Value(keys.OtStzFeatKey) != nil { // Otherwise inspect Datadog ot-baggage
			feat = proto.String(ctx.Value(keys.OtStzFeatKey).(string))
		}
	}
	if feat != nil {
		if stzFeat, err := baggage.NewMember(keys.StzFeat, *fn); err == nil {
			if bag, err := baggage.FromContext(ctx).SetMember(stzFeat); err == nil {
				ctx = baggage.ContextWithBaggage(ctx, bag)
			}
		}
		oh := make(http.Header)
		if ctx.Value(keys.OutboundHeadersKey) != nil {
			oh = ctx.Value(keys.OutboundHeadersKey).(http.Header)
		}
		oh.Set(string(keys.UberctxStzFeatKey), *fn) // uberctx (jaeger)
		oh.Set(string(keys.OtStzFeatKey), *fn)      // ot-baggage (datadog)
		ctx = context.WithValue(ctx, keys.OutboundHeadersKey, oh)
	}
	return ctx, feat
}

func GetPriorityBoost(ctx context.Context, pb *int32) (context.Context, *int32) {
	var boost *int32
	boostFromBaggage := baggage.FromContext(ctx).Member(keys.StzBoost).Value()
	if boostFromBaggage != "" {
		if boostInt, err := strconv.Atoi(boostFromBaggage); err == nil { // Inspect OTEL baggage
			boost = totalBoost(boost, boostInt)
		}
	} else if ctx.Value(keys.UberctxStzBoostKey) != nil { // Otherwise inspect Jaeger uberctx
		if boostInt, err := strconv.Atoi(ctx.Value(keys.UberctxStzBoostKey).(string)); err == nil {
			boost = totalBoost(boost, boostInt)
		}
	} else if ctx.Value(keys.OtStzBoostKey) != nil { // Otherwise inspect Datadog ot-baggage
		if boostInt, err := strconv.Atoi(ctx.Value(keys.OtStzBoostKey).(string)); err == nil {
			boost = totalBoost(boost, boostInt)
		}
	} else if pb != nil {
		boost = pb
	}
	if boost != nil {
		boostStr := strconv.Itoa(int(*boost))
		if stzBoost, err := baggage.NewMember(keys.StzBoost, boostStr); err == nil {
			if bag, err := baggage.FromContext(ctx).SetMember(stzBoost); err == nil {
				ctx = baggage.ContextWithBaggage(ctx, bag)
			}
		}
		oh := make(http.Header)
		if ctx.Value(keys.OutboundHeadersKey) != nil {
			oh = ctx.Value(keys.OutboundHeadersKey).(http.Header)
		}
		oh.Set(string(keys.UberctxStzBoostKey), boostStr) // uberctx (jaeger)
		oh.Set(string(keys.OtStzBoostKey), boostStr)      // ot-baggage (datadog)
		ctx = context.WithValue(ctx, keys.OutboundHeadersKey, oh)
	}
	return ctx, boost
}

func totalBoost(b1 *int32, b2 int) *int32 {
	var totalBoost int32
	if b1 == nil {
		totalBoost = int32(b2)
	} else {
		totalBoost = *b1 + int32(b2)
	}
	return &totalBoost
}
