package filter

import (
	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
)

type ResourceFilter struct {
}

func NewResourceFilter() *ResourceFilter {
	return &ResourceFilter{}
}

func (f *ResourceFilter) Filter(config *types.Version, agents []*mesos.Agent) []*mesos.Agent {
	candidates := make([]*mesos.Agent, 0)

	for _, agent := range agents {
		cpus, mem, disk, ports := agent.Measure()
		if cpus >= config.CPUs &&
			mem >= config.Mem &&
			disk >= config.Disk &&
			len(ports) >= len(config.Container.Docker.PortMappings) {
			candidates = append(candidates, agent)
		}
	}

	return candidates
}
