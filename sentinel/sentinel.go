package sentinel

import (
	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
)

func Init(name string, rulesFiles map[string]string) (func(), error) {
	shutdown := func() {
		logging.Debug("gracefully shutdown sentinel watcher")
	}

	conf := config.NewDefaultConfig()
	conf.Sentinel.App.Name = name                 // overload this with environment?
	conf.Sentinel.Log.Logger = &loggerAdapter{}   // log via the Stanza global logger
	conf.Sentinel.Log.Metric.FlushIntervalSec = 0 // disable default logging of metrics to on disk files
	if err := api.InitWithConfig(conf); err != nil {
		return shutdown, err
	}
	if err := InitFileDataSource(rulesFiles); err != nil {
		return shutdown, err
	}
	return shutdown, nil
}
