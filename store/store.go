package store

import (
	"github.com/Dataman-Cloud/swan/types"
)

type Store interface {
	CreateApp(app *types.Application) error
	UpdateApp(app *types.Application) error
	GetApp(appId string) *types.Application
	ListApps() []*types.Application
	DeleteApp(appId string) error

	CreateTask(task *types.Task) error
	GetTask(string) (*types.Task, error)
	UpdateTask(*types.Task) error
	ListTaskHistory(appId, slotId string) []*types.Task

	CreateVersion(appId string, version *types.Version) error
	GetVersion(appId, versionId string) *types.Version
	ListVersions(appId string) []*types.Version

	UpdateFrameworkId(frameworkId string) error
	GetFrameworkId() string

	CreateInstance(ins *types.Instance) error
	DeleteInstance(idOrName string) error
	UpdateInstance(ins *types.Instance) error // status, errmsg, updateAt
	GetInstance(idOrName string) (*types.Instance, error)
	ListInstances() ([]*types.Instance, error)
}
