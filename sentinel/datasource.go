package sentinel

// A sentinel DataSource watches for sentinel config changes and converts
// config JSON into structured sentinel Rules. We use the refreshable file
// DataSource, but with JSON sent to us from Stanza Hub.

import (
	"os"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/ext/datasource/file"
)

// Initialize new refreshable file datasource
func InitFileDataSource(rules map[string]string) error {
	// circuitbreaker rules
	if fn, ok := rules["circuitbreaker"]; ok {
		if _, err := os.Stat(fn); err == nil {
			cbDataSource := file.NewFileDataSource(fn,
				datasource.NewCircuitBreakerRulesHandler(datasource.CircuitBreakerRuleJsonArrayParser))
			if err := cbDataSource.Initialize(); err != nil {
				return err
			}
		}
	}

	// flow control rules
	if fn, ok := rules["flow"]; ok {
		if _, err := os.Stat(fn); err == nil {
			flowDataSource := file.NewFileDataSource(fn,
				datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser))
			if err := flowDataSource.Initialize(); err != nil {
				return err
			}
		}
	}

	// isolation rules
	if fn, ok := rules["isolation"]; ok {
		if _, err := os.Stat(fn); err == nil {
			isolationDataSource := file.NewFileDataSource(fn,
				datasource.NewIsolationRulesHandler(datasource.IsolationRuleJsonArrayParser))
			if err := isolationDataSource.Initialize(); err != nil {
				return err
			}
		}
	}

	// system rules
	if fn, ok := rules["system"]; ok {
		if _, err := os.Stat(fn); err == nil {
			systemDataSource := file.NewFileDataSource(fn,
				datasource.NewSystemRulesHandler(datasource.SystemRuleJsonArrayParser))
			if err := systemDataSource.Initialize(); err != nil {
				return err
			}
		}
	}

	return nil
}
