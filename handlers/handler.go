package handlers

import (
	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/otel"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
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
			instrumentationName,
			trace.WithInstrumentationVersion(instrumentationVersion),
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
	return h.attr
}

func (h *Handler) CustomerKey() attribute.KeyValue {
	return customerIdKey.String(global.GetCustomerID())
}

func (h *Handler) DecoratorKey(dec string) attribute.KeyValue {
	return decoratorKey.String(dec)
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

// Global Helper Functions //
func (h *Handler) APIKey() string {
	return global.GetServiceKey()
}

func (h *Handler) ClientID() string {
	return global.GetClientID()
}

func (h *Handler) DecoratorConfig(decorator string) *hubv1.DecoratorConfig {
	return global.DecoratorConfig(decorator)
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
