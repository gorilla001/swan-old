package types

type Application struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Instances         int    `json:"instances"`
	UpdatedInstances  int    `json:"instance_updated"`
	RunningInstances  int    `json:"running_instances"`
	RollbackInstances int    `json:"rollback_instances"`
	RunAS             string `json:"run_as"`
	ClusterId         string `json:"cluster_id"`
	Status            string `json:"status"`
	Created           int64  `json:"created"`
	Updated           int64  `json:"updated"`
}
