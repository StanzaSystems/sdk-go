package global

import (
	"github.com/alibaba/sentinel-golang/ext/datasource"
)

type sentinel struct {
	cb        datasource.DataSource
	flow      datasource.DataSource
	isolation datasource.DataSource
	system    datasource.DataSource
	resources []string
}

func GetSentinelConfig() *sentinel {
	return gs.sentinel
}

func NewResource(resName string) error {
	gsLock.Lock()
	defer gsLock.Unlock()

	gs.sentinel.resources = append(gs.sentinel.resources, resName)
	return nil
}

func SetCircuitBreakerDataSource(ds datasource.DataSource) error {
	gsLock.Lock()
	defer gsLock.Unlock()

	gs.sentinel.cb = ds
	return nil
}

func SetFlowDataSource(ds datasource.DataSource) error {
	gsLock.Lock()
	defer gsLock.Unlock()

	gs.sentinel.flow = ds
	return nil
}

func SetIsolationDataSource(ds datasource.DataSource) error {
	gsLock.Lock()
	defer gsLock.Unlock()

	gs.sentinel.isolation = ds
	return nil
}

func SetSystemDataSource(ds datasource.DataSource) error {
	gsLock.Lock()
	defer gsLock.Unlock()

	gs.sentinel.system = ds
	return nil
}
