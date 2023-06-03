package sentinel

import (
	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
)

var SentinelTempDir string

func Init(name, ds string) {
	SentinelTempDir = ds
	conf := config.NewDefaultConfig()
	conf.Sentinel.App.Name = name                 // overload this with environment?
	conf.Sentinel.Log.Logger = &loggerAdapter{}   // log via the Stanza global logger
	conf.Sentinel.Log.Metric.FlushIntervalSec = 0 // disable default logging of metrics to on disk files
	if err := api.InitWithConfig(conf); err != nil {
		logging.Error(err)
	}
	if err := InitFileDataSource(ds); err != nil {
		logging.Error(err)
	}
}
