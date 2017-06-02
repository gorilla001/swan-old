package types

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/utils/dfs"
	"github.com/aanand/compose-file/types"
)

// save to -> keyInstance
type Instance struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Desc      string    `json:"desc"`
	Status    string    `json:"status"` // op status
	ErrMsg    string    `json:"errmsg"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// request settings
	ServiceGroup ServiceGroup          `json:"service_group"`
	YAMLRaw      string                `json:"yaml_raw"`
	YAMLEnv      map[string]string     `json:"yaml_env"`
	YAMLExtra    map[string]*YamlExtra `json:"yaml_extra"`
}

func (ins *Instance) RequireConvert() bool {
	return len(ins.ServiceGroup) == 0 && ins.YAMLRaw != ""
}

func (ins *Instance) Valid() error {
	reg := regexp.MustCompile(`^[a-zA-Z0-9]{1,32}$`)
	if !reg.MatchString(ins.Name) {
		return errors.New("instance name should be regexp matched by: " + reg.String())
	}
	if ins.Name == "default" {
		return errors.New("instance name reserved")
	}
	return ins.ServiceGroup.Valid()
}

type Resource struct {
	CPU   float64  `json:"cpu"`
	Mem   float64  `json:"mem"`
	Disk  float64  `json:"disk"`
	Ports []uint64 `json:"ports"`
}

type YamlExtra struct {
	Priority    uint              `json:"priority"`
	WaitDelay   uint              `json:"wait_delay"` // by second
	PullAlways  bool              `json:"pull_always"`
	Resource    *Resource         `json:"resource"`
	Constraints string            `json:"constraints"`
	RunAs       string            `json:"runas"`
	URIs        []string          `json:"uris"`
	IPs         []string          `json:"ips"`
	Labels      map[string]string `json:"labels"` // extra labels: uid, username, vcluster ...
}

type ServiceGroup map[string]*DockerService

func (sg ServiceGroup) Valid() error {
	if len(sg) == 0 {
		return errors.New("serviceGroup empty")
	}
	for name, svr := range sg {
		if name == "" {
			return errors.New("service name required")
		}
		if strings.ContainsRune(name, '-') {
			return errors.New(`char '-' not allowed for service name`)
		}
		if name != svr.Name {
			return errors.New("service name mismatched")
		}
		if err := svr.Valid(); err != nil {
			return fmt.Errorf("validate service %s error: %v", name, err)
		}
	}
	return sg.circled()
}

func (sg ServiceGroup) PrioritySort() ([]string, error) {
	m, err := sg.dependMap()
	if err != nil {
		return nil, err
	}
	o := dfs.NewDfsOrder(m)
	return o.PostOrder(), nil
}

func (sg ServiceGroup) circled() error {
	m, err := sg.dependMap()
	if err != nil {
		return err
	}
	c := dfs.NewDirectedCycle(m)
	if cs := c.Cycle(); len(cs) > 0 {
		return fmt.Errorf("dependency circled: %v", cs)
	}
	return nil
}

func (sg ServiceGroup) dependMap() (map[string][]string, error) {
	ret := make(map[string][]string)
	for name, svr := range sg {
		// ensure exists
		for _, d := range svr.Service.DependsOn {
			if _, ok := sg[d]; !ok {
				return nil, fmt.Errorf("missing dependency: %s -> %s", name, d)
			}
		}
		ret[name] = svr.Service.DependsOn
	}
	return ret, nil
}

type DockerService struct {
	Name    string               `json:"name"`
	Service *types.ServiceConfig `json:"service"`
	Network *types.NetworkConfig `json:"network"`
	Volume  *types.VolumeConfig  `json:"volume"`
	Extra   *YamlExtra           `json:"extra"`
}

func (s *DockerService) Valid() error {
	return nil
}
