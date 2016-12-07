package types

type Application struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Instances         int      `json:"instances"`
	UpdatedInstances  int      `json:"instanceUpdated"`
	RunningInstances  int      `json:"runningInstances"`
	RollbackInstances int      `json:"rollbackInstances"`
	Tasks             []string `json:"tasks"`
	Versions          []string `json:"versions"`
	CurrentVersion    string   `json:"currentVersion"`
	UserId            string   `json:"userId"`
	ClusterId         string   `json:"clusterId"`
	Status            string   `json:"status"`
	Created           int64    `json:"created"`
	Updated           int64    `json:"updated"`
}
