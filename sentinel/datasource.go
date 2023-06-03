package sentinel

// A sentinel DataSource watches for sentinel config changes and converts
// config JSON into structured sentinel Rules.

import (
	"os"
	"path/filepath"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/ext/datasource/file"
)

const (
	cb_rules        string = "circuitbreaker_rules.json"
	flow_rules      string = "flow_rules.json"
	isolation_rules string = "isolation_rules.json"
	system_rules    string = "system_rules.json"
)

// Initialize new refreshable file datasource
func InitFileDataSource(ConfigPath string) error {
	// TODO: create empty rules files if they don't exist

	// circuitbreaker rules
	if _, err := os.Stat(filepath.Join(ConfigPath, cb_rules)); err == nil {
		cbDataSource := file.NewFileDataSource(
			filepath.Join(ConfigPath, cb_rules),
			datasource.NewCircuitBreakerRulesHandler(datasource.CircuitBreakerRuleJsonArrayParser))
		if err := cbDataSource.Initialize(); err != nil {
			return err
		}
	}

	// flow control rules
	if _, err := os.Stat(filepath.Join(ConfigPath, flow_rules)); err == nil {
		flowDataSource := file.NewFileDataSource(
			filepath.Join(ConfigPath, flow_rules),
			datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser))
		if err := flowDataSource.Initialize(); err != nil {
			return err
		}
	}

	// isolation rules
	if _, err := os.Stat(filepath.Join(ConfigPath, isolation_rules)); err == nil {
		isolationDataSource := file.NewFileDataSource(
			filepath.Join(ConfigPath, isolation_rules),
			datasource.NewIsolationRulesHandler(datasource.IsolationRuleJsonArrayParser))
		if err := isolationDataSource.Initialize(); err != nil {
			return err
		}
	}

	// system rules
	if _, err := os.Stat(filepath.Join(ConfigPath, system_rules)); err == nil {
		systemDataSource := file.NewFileDataSource(
			filepath.Join(ConfigPath, system_rules),
			datasource.NewSystemRulesHandler(datasource.SystemRuleJsonArrayParser))
		if err := systemDataSource.Initialize(); err != nil {
			return err
		}
	}

	return nil
}
