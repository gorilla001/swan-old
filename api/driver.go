package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/mesos"
)

type Driver interface {
	//CreateApp(*types.Version, string) error
	//GetApp(string) (*types.App, error)
	//DeleteApp(string) error
	//GetApps(types.AppFilterOptions) ([]*types.App, error)
	//ScaleUp(string, int, []string) error
	//ScaleDown(string, int) error
	//CancelScale(string) error
	//UpdateApp(string, *types.Version) error
	//CancelUpdate(string) error
	//ProceedUpdate(string, int, map[string]float64) error

	LaunchTask(*mesos.Task) error
	KillTask(string, string) error
	ClusterName() string
	SubscribeEvent(http.ResponseWriter, string) error
}
