package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/StanzaSystems/sdk-go/keys"

	"go.opentelemetry.io/otel/baggage"
)

func getFeature(ctx context.Context, feat string) (context.Context, *string) {
	if feat == "" { // If Feature is already supplied, use it
		featFromBaggage := baggage.FromContext(ctx).Member(keys.StzFeat).Value()
		if featFromBaggage != "" { // Otherwise inspect OTEL baggage
			feat = featFromBaggage
		} else if ctx.Value(keys.UberctxStzFeatKey) != nil { // Otherwise inspect Jaeger uberctx
			feat = ctx.Value(keys.UberctxStzFeatKey).(string)
		} else if ctx.Value(keys.OtStzFeatKey) != nil { // Otherwise inspect Datadog ot-baggage
			feat = ctx.Value(keys.OtStzFeatKey).(string)
		}
	}
	if feat != "" {
		if stzFeat, err := baggage.NewMember(keys.StzFeat, feat); err == nil {
			if bag, err := baggage.FromContext(ctx).SetMember(stzFeat); err == nil {
				ctx = baggage.ContextWithBaggage(ctx, bag)
			}
		}
		oh := make(http.Header)
		if ctx.Value(keys.OutboundHeadersKey) != nil {
			oh = ctx.Value(keys.OutboundHeadersKey).(http.Header)
		}
		oh.Set(string(keys.UberctxStzFeatKey), feat) // uberctx (jaeger)
		oh.Set(string(keys.OtStzFeatKey), feat)      // ot-baggage (datadog)
		ctx = context.WithValue(ctx, keys.OutboundHeadersKey, oh)
	}
	return ctx, &feat
}

func getPriorityBoost(ctx context.Context, boost int32) (context.Context, *int32) {
	// Handle additional PriorityBoost (from OTEL baggage or known headers)
	boostFromBaggage := baggage.FromContext(ctx).Member(keys.StzBoost).Value()
	if boostFromBaggage != "" {
		if boostInt, err := strconv.Atoi(boostFromBaggage); err == nil { // Inspect OTEL baggage
			boost = boost + int32(boostInt)
		}
	} else if ctx.Value(keys.UberctxStzBoostKey) != nil { // Otherwise inspect Jaeger uberctx
		if boostInt, err := strconv.Atoi(ctx.Value(keys.UberctxStzBoostKey).(string)); err == nil {
			boost = boost + int32(boostInt)
		}
	} else if ctx.Value(keys.OtStzBoostKey) != nil { // Otherwise inspect Datadog ot-baggage
		if boostInt, err := strconv.Atoi(ctx.Value(keys.OtStzBoostKey).(string)); err == nil {
			boost = boost + int32(boostInt)
		}
	}
	if boost != 0 {
		boostStr := strconv.Itoa(int(boost))
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
	return ctx, &boost
}
