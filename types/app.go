package types

import (
	"time"

	"github.com/Dataman-Cloud/swan/utils/fields"
	"github.com/Dataman-Cloud/swan/utils/labels"
)

type Application struct {
	ID               string    `json:"id,omitempty"`
	Name             string    `json:"name"`
	Instances        int       `json:"instances"`
	UpdatedInstances int       `json:"updatedInstances"`
	RunningInstances int       `json:"runningInstances"`
	RunAs            string    `json:"runAs"`
	Priority         int       `json:"priority"`
	ClusterID        string    `json:"clusterID,omitempty"`
	Status           string    `json:"status,omitempty"`
	CreatedAt        time.Time `json:"created,omitempty"`
	UpdatedAt        time.Time `json:"updated,omitempty"`
	Mode             string    `json:"mode"`
	State            string    `json:"state,omitempty"`

	// use task for compatability now, should be slot here
	Tasks   []*Task  `json:"tasks,omitempty"`
	Version *Version `json:"currentVersion"`
	// use when app updated, ProposedVersion can either be commit or revert
	ProposedVersion *Version `json:"proposedVersion,omitempty"`
	Versions        []string `json:"versions,omitempty"`
	IP              []string `json:"ip,omitempty"`

	// current version related info
	Labels      map[string]string `json:"labels,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Constraints string            `json:"constraints,omitempty"`
	URIs        []string          `json:"uris,omitempty"`
}

type AppFilterOptions struct {
	LabelsSelector labels.Selector
	FieldsSelector fields.Selector
}

type ProceedUpdateParam struct {
	Instances  int                `json:"instances"`
	NewWeights map[string]float64 `json:"weights"`
}

type ScaleUpParam struct {
	Instances int      `json:"instances"`
	IPs       []string `json:"ips"`
}

type ScaleDownParam struct {
	Instances int `json:"instances"`
}

type UpdateWeightParam struct {
	Weight float64 `json:"weight"`
}

type UpdateWeightsParam struct {
	Weights map[string]float64 `json:"weights"`
}
