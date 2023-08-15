package handlers

import (
	hubv1 "github.com/StanzaSystems/sdk-go/gen/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/otel"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	tracer trace.Tracer
	meter  *Meter
	attr   []attribute.KeyValue
}

func NewHandler() (*Handler, error) {
	m, err := GetStanzaMeter()
	return &Handler{
		meter: m,
		tracer: otel.GetTracerProvider().Tracer(
			global.InstrumentationName(),
			global.InstrumentationTraceVersion(),
		),
		attr: []attribute.KeyValue{
			clientIdKey.String(global.GetClientID()),
			environmentKey.String(global.GetServiceEnvironment()),
			serviceKey.String(global.GetServiceName()),
		},
	}, err
}

func (h *Handler) Meter() *Meter {
	return h.meter
}

func (h *Handler) Tracer() trace.Tracer {
	return h.tracer
}

func (h *Handler) Propagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

// OTEL Attribute Helper Functions //
func (h *Handler) Attributes() []attribute.KeyValue {
	return append(h.attr, customerIdKey.String(global.GetCustomerID()))
}

func (h *Handler) GuardKey(guard string) attribute.KeyValue {
	return guardKey.String(guard)
}

func (h *Handler) FeatureKey(feat string) attribute.KeyValue {
	return featureKey.String(feat)
}

func (h *Handler) ReasonKey(reason string) attribute.KeyValue {
	return reasonKey.String(reason)
}

func (h *Handler) ReasonFailOpen() attribute.KeyValue {
	return reasonKey.String("fail_open")
}

func (h *Handler) ReasonInvalidToken() attribute.KeyValue {
	return reasonKey.String("invalid_token")
}

func (h *Handler) ReasonQuota() attribute.KeyValue {
	return reasonKey.String("quota")
}

func (h *Handler) ReasonToken() attribute.KeyValue {
	return reasonKey.String("token")
}

// Global Helper Functions //
func (h *Handler) APIKey() string {
	return global.GetServiceKey()
}

func (h *Handler) ClientID() string {
	return global.GetClientID()
}

func (h *Handler) GuardConfig(guard string) *hubv1.GuardConfig {
	return global.GuardConfig(guard)
}

func (h *Handler) Environment() string {
	return global.GetServiceEnvironment()
}

func (h *Handler) OTELEnabled() bool {
	return global.OtelEnabled()
}

func (h *Handler) SentinelEnabled() bool {
	return global.SentinelEnabled()
}

func (h *Handler) QuotaServiceClient() hubv1.QuotaServiceClient {
	return global.QuotaServiceClient()
}
