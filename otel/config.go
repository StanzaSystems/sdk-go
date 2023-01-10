package otel

type Config struct {
	traceRatio float64 // Percentage of traces to sample (default: 0.001)
}
