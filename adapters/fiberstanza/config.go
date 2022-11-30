package fiberstanza

// Config defines the config for fiberstanza middleware.
type Config struct {
	// Stanza

	// Sentinel
	FlowRules           string `yaml:"flowRules"`
	CircuitBreakerRules string `yaml:"circuitBreakerRules"`
	IsolationRules      string `yaml:"isolationRules"`
	SystemRules         string `yaml:"systemRules"`

	// OpenTelemetry
}
