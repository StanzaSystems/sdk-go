package sentinel

import (
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/core/system"
)

type Config struct {
	CircuitBreakerRules []*circuitbreaker.Rule
	FlowRules           []*flow.Rule
	IsolationRules      []*isolation.Rule
	SystemRules         []*system.Rule
}
