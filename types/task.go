package types

import (
	"time"
)

type Task struct {
	ID      string    `json:"id,omitempty"`
	Name    string    `json:"name,omitempty"`
	CPU     float64   `json:"cpu"`
	Mem     float64   `json:"mem"`
	Disk    float64   `json:"disk"`
	IP      string    `json:"ip,omitempty"`
	Ports   []uint64  `json:"ports,omitempty"`
	Image   string    `json:"image"`
	Healthy bool      `json:"healthy"`
	Weight  float64   `json:"weight"`
	Status  string    `json:"status,omitempty"`
	ErrMsg  string    `json:"errmsg, omitempty"`
	Created time.Time `json:"created,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}

type TaskHistory struct {
	ID         string `json:"id"`
	AppID      string `json:"appID"`
	VersionID  string `json:"versionID"`
	AppVersion string `json:"appVersion"`

	OfferID       string `json:"offerID"`
	AgentID       string `json:"agentID"`
	AgentHostname string `json:"agentHostname"`

	CPU  float64 `json:"cpu"`
	Mem  float64 `json:"mem"`
	Disk float64 `json:"disk"`

	State   string `json:"state,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
	Stdout  string `json:"stdout,omitempty"`
	Stderr  string `json:"stderr,omitempty"`

	ArchivedAt    time.Time `json:"archivedAt, omitempty"`
	ContainerId   string    `json:"containerId"`
	ContainerName string    `json:"containerName"`
	Weight        float64   `json:"weight,omitempty"`
}
