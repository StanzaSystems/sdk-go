package handlers

import (
	"github.com/StanzaSystems/sdk-go/otel"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	apikey          string
	clientId        string
	customerId      string
	environment     string
	otelEnabled     bool
	sentinelEnabled bool

	decoratorConfig map[string]*hubv1.DecoratorConfig
	qsc             hubv1.QuotaServiceClient
	propagators     propagation.TextMapPropagator
	tracer          trace.Tracer
	meter           *Meter
	attr            []attribute.KeyValue
}

func NewHandler(apikey, clientId, environment, service string, otelEnabled, sentinelEnabled bool) (*Handler, error) {
	m, err := GetStanzaMeter()
	return &Handler{
		apikey:          apikey,
		clientId:        clientId,
		environment:     environment,
		otelEnabled:     otelEnabled,
		sentinelEnabled: sentinelEnabled,
		decoratorConfig: make(map[string]*hubv1.DecoratorConfig),
		qsc:             nil,
		propagators:     otel.GetTextMapPropagator(),
		meter:           m,
		tracer: otel.GetTracerProvider().Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(instrumentationVersion),
		),
		attr: []attribute.KeyValue{
			clientIdKey.String(clientId),
			environmentKey.String(environment),
			serviceKey.String(service),
		},
	}, err
}

func (h *Handler) APIKey() string {
	return h.apikey
}

func (h *Handler) Attributes() []attribute.KeyValue {
	return h.attr
}

func (h *Handler) ClientID() string {
	return h.clientId
}

func (h *Handler) DecoratorConfig(dec string) *hubv1.DecoratorConfig {
	return h.decoratorConfig[dec]
}

func (h *Handler) DecoratorKey(dec string) attribute.KeyValue {
	return decoratorKey.String(dec)
}

func (h *Handler) Environment() string {
	return h.environment
}

func (h *Handler) FeatureKey(feat string) attribute.KeyValue {
	return featureKey.String(feat)
}

func (h *Handler) Meter() *Meter {
	return h.meter
}

func (h *Handler) Propagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

func (h *Handler) ReasonKey(reason string) attribute.KeyValue {
	return reasonKey.String(reason)
}

func (h *Handler) SetCustomerId(id string) {
	if h.customerId == "" {
		h.customerId = id
		h.attr = append(h.attr, customerIdKey.String(id))
	}
}

func (h *Handler) SetDecoratorConfig(d string, dc *hubv1.DecoratorConfig) {
	if h.decoratorConfig[d] == nil {
		h.decoratorConfig[d] = dc
	}
}

func (h *Handler) SetQuotaServiceClient(quotaServiceClient hubv1.QuotaServiceClient) {
	if h.qsc == nil {
		h.qsc = quotaServiceClient
	}
}

func (h *Handler) SentinelEnabled() bool {
	return h.sentinelEnabled
}

func (h *Handler) Tracer() trace.Tracer {
	return h.tracer
}

func (h *Handler) QuotaServiceClient() hubv1.QuotaServiceClient {
	return h.qsc
}
