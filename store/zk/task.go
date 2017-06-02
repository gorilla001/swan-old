package zk

import (
	"strings"

	"github.com/Dataman-Cloud/swan/types"
)

// As Nested Field of AppHolder, UpdateCurrentTask Require Transaction Lock

func (zk *ZKStore) CreateTask(task *types.Task) error {
	bs, err := encode(task)
	if err != nil {
		return err
	}

	var (
		fields = strings.Split(task.Name, ".")
		appId  = fields[1]
	)

	path := keyApp + "/" + appId + "/tasks/" + task.ID

	return zk.createAll(path, bs)
}

func (zk *ZKStore) UpdateTask(task *types.Task) error {
	bs, err := encode(task)
	if err != nil {
		return err
	}

	var (
		fields = strings.Split(task.Name, ".")
		appId  = fields[1]
	)

	path := keyApp + "/" + appId + "/tasks/" + task.ID

	return zk.set(path, bs)
}

func (zk *ZKStore) ListTaskHistory(aid, sid string) []*types.Task {
	return nil
}

func (zk *ZKStore) DeleteTask(taskid string) error {
	return nil
}

func (zk *ZKStore) GetTask(id string) (*types.Task, error) {
	var (
		fields = strings.Split(id, ".")
		appId  = fields[1]
	)

	path := keyApp + "/" + appId + "/tasks/" + id

	data, _, err := zk.get(path)
	if err != nil {
		return nil, err
	}

	var task types.Task
	if err := decode(data, &task); err != nil {
		return nil, err
	}

	return &task, nil
}
