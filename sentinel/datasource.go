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
	"github.com/StanzaSystems/sdk-go/global"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/ext/datasource/file"
)

// type ApolloOptions struct {
// 	Conf        *config.AppConfig
// 	PropertyKey string
// }

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

// type NacosOptions {
// 	Client config_client.IConfigClient
// 	Group  string
// 	DataId string
// }

type DataSourceOptions struct {
	// apollo ApolloOptions
	Consul ConsulOptions
	// etcdv3 Etcdv3Options
	File FileOptions
	// k8s K8sOptions
	// nacos NacosOptions
}

// Initialize a sentinel datasource
func InitDataSource(options DataSourceOptions) error {
	// TODO: Put a case statement here for each of the supported datasources
	if options.File.ConfigFilePath != "" {
		ds, err := InitFileDataSource(options.File.ConfigFilePath)
		if err != nil {
			return err
		}
		if err := global.SetDataSource(ds); err != nil {
			return err
		}
	}

	// TODO: can disable system metrics polling if no system rules? where?
	//       (has to be evaluated everytime new rules are loaded)
	return nil
}

// Initialize a new refreshable file datasource
func InitFileDataSource(sourceFilePath string) (datasource.DataSource, error) {
	ds := file.NewFileDataSource(sourceFilePath,
		datasource.NewCircuitBreakerRulesHandler(datasource.CircuitBreakerRuleJsonArrayParser),
		datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser),
		datasource.NewIsolationRulesHandler(datasource.IsolationRuleJsonArrayParser),
		datasource.NewSystemRulesHandler(datasource.SystemRuleJsonArrayParser),
	)
	if err := ds.Initialize(); err != nil {
		// do we want to return a failed datasource connection?
		// or should we setup a background poller (with exponential backoff, etc)
		// so we keep trying?
		return nil, err
	}
	return ds, nil
}
