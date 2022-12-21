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
	"os"
	"path/filepath"

	"github.com/StanzaSystems/sdk-go/global"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/ext/datasource/file"
)

const (
	cb_rules string = "circuitbreaker_rules.json"
	flow_rules string = "flow_rules.json"
	isolation_rules string = "isolation_rules.json"
	system_rules string = "system_rules.json"
)

type ConsulOptions struct {
	PropertyKey string
}

// type Etcdv3Options struct {
// 	Client *clientv3.Client
// 	Key    string
// }

type FileOptions struct {
	ConfigFilePath string
}

// type K8sOptions struct {
// 	Namespace string
// }

type DataSourceOptions struct {
	Consul ConsulOptions
	// etcdv3 Etcdv3Options
	File FileOptions
	// k8s K8sOptions
}

// Initialize a sentinel datasource
func InitDataSource(options DataSourceOptions) error {
	// TODO: Put a case statement here for each of the supported datasources

	// Refreshable File
	if options.File.ConfigFilePath != "" {
		if err := InitFileDataSource(options.File.ConfigFilePath); err != nil {
			return err
		}
	}

	// TODO: can disable system metrics polling if no system rules? where?
	//       (has to be evaluated everytime new rules are loaded)
	return nil
}

// Initialize new refreshable file datasources
func InitFileDataSource(ConfigFilePath string) error {
	// if the file doesn't exist (yet), should we create a background poller which
	// keeps trying to add this datasource? (with expoential backoff, etc)

	// circuitbreaker rules
	if _, err := os.Stat(filepath.Join(ConfigFilePath, cb_rules)); err == nil {
		cbDataSource := file.NewFileDataSource(
			filepath.Join(ConfigFilePath, cb_rules),
			datasource.NewCircuitBreakerRulesHandler(datasource.CircuitBreakerRuleJsonArrayParser))
		if err := cbDataSource.Initialize(); err == nil {
			global.SetCircuitBreakerDataSource(cbDataSource)
		}
	}

	// flow control rules
	if _, err := os.Stat(filepath.Join(ConfigFilePath, flow_rules)); err == nil {
		flowDataSource := file.NewFileDataSource(
			filepath.Join(ConfigFilePath, flow_rules),
			datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser))
		if err := flowDataSource.Initialize(); err == nil {
			global.SetFlowDataSource(flowDataSource)
		}
	}

	// isolation rules
	if _, err := os.Stat(filepath.Join(ConfigFilePath, isolation_rules)); err == nil {
		isolationDataSource := file.NewFileDataSource(
			filepath.Join(ConfigFilePath, isolation_rules),
			datasource.NewIsolationRulesHandler(datasource.IsolationRuleJsonArrayParser))
		if err := isolationDataSource.Initialize(); err == nil {
			global.SetIsolationDataSource(isolationDataSource)
		}
	}

	// system rules
	if _, err := os.Stat(filepath.Join(ConfigFilePath, system_rules)); err == nil {
		systemDataSource := file.NewFileDataSource(
			filepath.Join(ConfigFilePath, system_rules),
			datasource.NewSystemRulesHandler(datasource.SystemRuleJsonArrayParser))
		if err := systemDataSource.Initialize(); err == nil {
			global.SetSystemDataSource(systemDataSource)
		}
	}

	return nil
}
