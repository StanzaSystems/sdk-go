package sentinel

import (
	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/logging"
)

func Init(appName string) error {
	conf := config.NewDefaultConfig()
	conf.Sentinel.App.Name = appName
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger() // not logr; can we make a simple adapater?
	if err := api.InitWithConfig(conf); err != nil {
		return err
	}
	return nil
}
