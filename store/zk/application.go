package zk

import (
	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

func (zk *ZKStore) CreateApp(app *types.Application) error {
	if zk.GetApp(app.ID) != nil {
		return errAppAlreadyExists
	}

	bs, err := encode(app)
	if err != nil {
		return err
	}

	path := keyApp + "/" + app.ID

	return zk.createAll(path, bs)
}

// All of AppHolder Write Ops Requires Transaction Lock
func (zk *ZKStore) UpdateApp(app *types.Application) error {
	bs, err := encode(app)
	if err != nil {
		return err
	}

	path := keyApp + "/" + app.ID
	return zk.set(path, bs)
}

func (zk *ZKStore) GetApp(id string) *types.Application {
	return nil
}

func (zk *ZKStore) ListApps() []*types.Application {
	ret := make([]*types.Application, 0, 0)

	nodes, err := zk.list(keyApp)
	if err != nil {
		log.Errorln("zk ListApps error:", err)
		return ret
	}

	for _, node := range nodes {
		if app := zk.GetApp(node); app != nil {
			ret = append(ret, app)
		}
	}

	return ret
}

// All of AppHolder Write Ops Requires Transaction Lock
func (zk *ZKStore) DeleteApp(id string) error {
	if zk.GetApp(id) == nil {
		return errAppNotFound
	}

	return zk.del(keyApp + "/" + id)
}
