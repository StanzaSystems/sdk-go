package sentinel

import (
	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
)

func Init(name, ds string) error {
	conf := config.NewDefaultConfig()
	conf.Sentinel.App.Name = name                 // overload this with environment?
	conf.Sentinel.Log.Logger = &loggerAdapter{}   // log via the Stanza global logger
	conf.Sentinel.Log.Metric.FlushIntervalSec = 0 // disable default logging of metrics to on disk files
	if err := api.InitWithConfig(conf); err != nil {
		return err
	}
	if err := InitDataSource(ds); err != nil {
		return err
	}
	return nil
}
