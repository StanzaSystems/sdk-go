package sentinel

import (
	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	sl "github.com/alibaba/sentinel-golang/logging"
)

func Init(appName string, dsOptions DataSourceOptions) error {
	conf := config.NewDefaultConfig()
	conf.Sentinel.App.Name = appName // overload this with environment?
	conf.Sentinel.Log.Logger = sl.NewConsoleLogger() // not logr; can we make a simple adapater?
	if err := api.InitWithConfig(conf); err != nil {
		return err
	}
	if err := InitDataSource(dsOptions); err != nil {
		return err
	}
	return nil
}
