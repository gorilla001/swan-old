package zk

import (
	"errors"

	"github.com/Dataman-Cloud/swan/types"

	"github.com/Sirupsen/logrus"
)

func (zk *ZKStore) CreateInstance(ins *types.Instance) error {
	bs, err := encode(ins)
	if err != nil {
		return err
	}

	path := keyInstance + "/" + ins.ID
	return zk.createAll(path, bs)
}

func (zk *ZKStore) UpdateInstance(ins *types.Instance) error {
	if i, _ := zk.GetInstance(ins.ID); i == nil {
		return errInstanceNotFound
	}

	bs, err := encode(ins)
	if err != nil {
		return err
	}

	path := keyInstance + "/" + ins.ID
	return zk.create(path, bs)
}

func (zk *ZKStore) GetInstance(idOrName string) (*types.Instance, error) {
	// by id
	bs, _, err := zk.get(keyInstance + "/" + idOrName)
	if err == nil {
		ins := new(types.Instance)
		if err := decode(bs, &ins); err != nil {
			logrus.Errorln("zk GetInstance.decode error:", err)
			return nil, err
		}
		return ins, nil
	}

	// by name
	inss, err := zk.ListInstances()
	if err != nil {
		return nil, err
	}
	for _, ins := range inss {
		if ins.Name == idOrName {
			return ins, nil
		}
	}

	return nil, errors.New("no such instance")
}

func (zk *ZKStore) ListInstances() ([]*types.Instance, error) {
	ret := make([]*types.Instance, 0, 0)

	nodes, err := zk.list(keyInstance)
	if err != nil {
		logrus.Errorln("zk ListInstances error:", err)
		return ret, err
	}

	for _, node := range nodes {
		bs, _, err := zk.get(keyInstance + "/" + node)
		if err != nil {
			logrus.Errorln("zk ListInstance.getnode error:", err)
			continue
		}

		ins := new(types.Instance)
		if err := decode(bs, &ins); err != nil {
			logrus.Errorln("zk ListInstance.decode error:", err)
			continue
		}

		ret = append(ret, ins)
	}

	return ret, nil
}

func (zk *ZKStore) DeleteInstance(idOrName string) error {
	ins, err := zk.GetInstance(idOrName)
	if err != nil {
		return err
	}

	return zk.del(keyInstance + "/" + ins.ID)
}
