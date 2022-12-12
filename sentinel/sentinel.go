package sentinel

import (
	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
)

func Init(appName string, dsOptions DataSourceOptions) error {
	conf := config.NewDefaultConfig()
	conf.Sentinel.App.Name = appName // overload this with environment?
	conf.Sentinel.Log.Logger = &logging.SentinelAdaptor{}
	if err := api.InitWithConfig(conf); err != nil {
		return err
	}
	if err := InitDataSource(dsOptions); err != nil {
		return err
	}
	return nil
}
