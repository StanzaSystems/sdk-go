package sentinel

// A sentinel DataSource watches for sentinel config changes and converts
// config JSON into structured sentinel Rules.
//
// Sentinel has the following datasource options builtin:
// https://github.com/alibaba/sentinel-golang/tree/master/ext/datasource/
// -- file (refreshable local)
// https://github.com/alibaba/sentinel-golang/tree/master/pkg/datasource/
// -- apollo
// -- consul
// -- etcdv3
// -- k8s
// -- nacos
//
// should we add a grpc datasource?
// or would that be done "outside" of the sentinel datasource model?

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/ext/datasource/file"
)

const (
	cb_rules        string = "circuitbreaker_rules.json"
	flow_rules      string = "flow_rules.json"
	isolation_rules string = "isolation_rules.json"
	system_rules    string = "system_rules.json"
)

// Initialize a sentinel datasource
func InitDataSource(dataSource string) error {
	ds := strings.Split(dataSource, ":")
	if len(ds) < 2 {
		return fmt.Errorf("invalid datasource: %v", ds)
	}
	switch ds[0] {
	case "consul":
		return fmt.Errorf("consul datasource support has not been implemented yet")

	case "grpc":
		return fmt.Errorf("grpc datasource support has not been implemented yet")

	case "local":
		return InitFileDataSource(ds[1])

	default:
		return fmt.Errorf("invalid datasource: %v", ds[0])
	}
}

// Initialize new refreshable file datasources
func InitFileDataSource(ConfigPath string) error {
	// if the file doesn't exist (yet), should we create a background poller which
	// keeps trying to add this datasource? (with expoential backoff, etc)

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
