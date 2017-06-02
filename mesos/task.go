package mesos

import (
	"errors"
	"strings"

	mesosproto "github.com/Dataman-Cloud/swan/proto/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/gogo/protobuf/proto"
)

type Task struct {
	mesosproto.TaskInfo

	updates chan *mesosproto.TaskStatus

	cfg *types.Version
}

func NewTask(cfg *types.Version, id, name string) *Task {
	task := &Task{
		cfg:     cfg,
		updates: make(chan *mesosproto.TaskStatus),
	}

	task.Name = &name
	task.TaskId = &mesosproto.TaskID{Value: &id}

	return task
}
func (t *Task) Build(offer *mesosproto.Offer) {
	t.AgentId = offer.GetAgentId()
	t.Command = t.command()
	t.Resources = t.resources(offer)
	t.Container = t.container(offer)
	//t.HealthCheck = t.cfg.HealthCheck
	//t.KillPolicy = t.cfg.KillPolicy
	t.Labels = t.labels()
}

func (t *Task) command() *mesosproto.CommandInfo {
	if cmd := t.cfg.Command; len(cmd) > 0 {
		return &mesosproto.CommandInfo{
			Shell:     proto.Bool(false),
			Value:     proto.String(string(cmd[0])),
			Arguments: strings.Split(cmd[1:], " "),
		}
	}
	return &mesosproto.CommandInfo{Shell: proto.Bool(false)}
}

func (t *Task) container(offer *mesosproto.Offer) *mesosproto.ContainerInfo {
	var (
		image      = t.cfg.Container.Docker.Image
		privileged = t.cfg.Container.Docker.Privileged
		force      = t.cfg.Container.Docker.ForcePullImage
	)

	return &mesosproto.ContainerInfo{
		Type:    mesosproto.ContainerInfo_DOCKER.Enum(),
		Volumes: t.volumes(),
		Docker: &mesosproto.ContainerInfo_DockerInfo{
			Image:          proto.String(image),
			Privileged:     proto.Bool(privileged),
			Network:        t.network(),
			PortMappings:   t.portMappings(offer),
			Parameters:     t.parameters(),
			ForcePullImage: proto.Bool(force),
		},
	}
}

func (t *Task) resources(offer *mesosproto.Offer) []*mesosproto.Resource {
	var (
		pmsLen = len(t.cfg.Container.Docker.PortMappings)
		rs     = make([]*mesosproto.Resource, 0, 0)
		cpus   = t.cfg.CPUs
		mem    = t.cfg.Mem
		disk   = t.cfg.Disk
	)

	if cpus > 0 {
		rs = append(rs, &mesosproto.Resource{
			Name: proto.String("cpus"),
			Type: mesosproto.Value_SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{
				Value: proto.Float64(cpus),
			},
		})
	}

	if mem > 0 {
		rs = append(rs, &mesosproto.Resource{
			Name: proto.String("mem"),
			Type: mesosproto.Value_SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{
				Value: proto.Float64(mem),
			},
		})
	}

	if disk > 0 {
		rs = append(rs, &mesosproto.Resource{
			Name: proto.String("disk"),
			Type: mesosproto.Value_SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{
				Value: proto.Float64(disk),
			},
		})
	}

	if pmsLen > 0 {
		ports := getPorts(offer)
		for i := 0; i < pmsLen; i++ {
			rs = append(rs, &mesosproto.Resource{
				Name: proto.String("ports"),
				Type: mesosproto.Value_RANGES.Enum(),
				Ranges: &mesosproto.Value_Ranges{
					Range: []*mesosproto.Value_Range{
						{
							Begin: proto.Uint64(uint64(ports[i])),
							End:   proto.Uint64(uint64(ports[i])),
						},
					},
				},
			})
		}
	}

	return rs
}

// SendStatus method writes the task status in the updates channel
func (t *Task) SendStatus(status *mesosproto.TaskStatus) {
	t.updates <- status
}

// GetStatus method reads the task status on the updates channel
func (t *Task) GetStatus() chan *mesosproto.TaskStatus {
	return t.updates
}

// IsDone check that if a task is done or not according by task status.
func (t *Task) IsDone(status *mesosproto.TaskStatus) bool {
	state := status.GetState()

	switch state {
	case mesosproto.TaskState_TASK_RUNNING,
		mesosproto.TaskState_TASK_FINISHED,
		mesosproto.TaskState_TASK_FAILED,
		mesosproto.TaskState_TASK_KILLED,
		mesosproto.TaskState_TASK_ERROR,
		mesosproto.TaskState_TASK_LOST,
		mesosproto.TaskState_TASK_DROPPED,
		mesosproto.TaskState_TASK_GONE:

		return true
	}

	return false
}

func (t *Task) DetectError(status *mesosproto.TaskStatus) error {
	var (
		state = status.GetState()
		//data  = status.GetData() // docker container inspect result
	)

	switch state {
	case mesosproto.TaskState_TASK_FAILED,
		mesosproto.TaskState_TASK_ERROR,
		mesosproto.TaskState_TASK_LOST,
		mesosproto.TaskState_TASK_DROPPED,
		mesosproto.TaskState_TASK_UNREACHABLE,
		mesosproto.TaskState_TASK_GONE,
		mesosproto.TaskState_TASK_GONE_BY_OPERATOR,
		mesosproto.TaskState_TASK_UNKNOWN:

		return errors.New(status.GetMessage())
	}

	return nil
}

func (t *Task) healtcheck() *mesosproto.HealthCheck {
	var (
		hlc         *mesosproto.HealthCheck
		healthcheck = t.cfg.HealthCheck
		protocol    = strings.ToLower(healthcheck.Protocol)
		//delay         = t.HealthCheck.DelaySeconds
		interval      = healthcheck.IntervalSeconds
		timeout       = healthcheck.TimeoutSeconds
		consecutive   = healthcheck.ConsecutiveFailures
		gracep        = healthcheck.GracePeriodSeconds
		path          = healthcheck.Path
		namespacePort int32
	)

	switch protocol {
	case "cmd":
		hlc = &mesosproto.HealthCheck{
			Type: mesosproto.HealthCheck_COMMAND.Enum(),
			Command: &mesosproto.CommandInfo{
				Value: proto.String(""),
			},
		}
	case "http":
		hlc = &mesosproto.HealthCheck{
			Type: mesosproto.HealthCheck_HTTP.Enum(),
			Http: &mesosproto.HealthCheck_HTTPCheckInfo{
				Scheme:   proto.String(protocol),
				Port:     proto.Uint32(uint32(namespacePort)),
				Path:     proto.String(path),
				Statuses: []uint32{uint32(200), uint32(201), uint32(301), uint32(302)},
			},
		}
	case "tcp":
		hlc = &mesosproto.HealthCheck{
			Type: mesosproto.HealthCheck_TCP.Enum(),
			Tcp: &mesosproto.HealthCheck_TCPCheckInfo{
				Port: proto.Uint32(uint32(namespacePort)),
			},
		}
	}

	hlc.DelaySeconds = proto.Float64(t.cfg.HealthCheck.DelaySeconds)
	hlc.IntervalSeconds = proto.Float64(interval)
	hlc.TimeoutSeconds = proto.Float64(timeout)
	hlc.ConsecutiveFailures = proto.Uint32(consecutive)
	hlc.GracePeriodSeconds = proto.Float64(gracep)

	return hlc
}

func (t *Task) parameters() []*mesosproto.Parameter {
	var (
		parameters = t.cfg.Container.Docker.Parameters
		mps        = make([]*mesosproto.Parameter, 0, 0)
	)

	for _, para := range parameters {
		mps = append(mps, &mesosproto.Parameter{
			Key:   proto.String(para.Key),
			Value: proto.String(para.Value),
		})
	}

	return mps
}

func (t *Task) volumes() []*mesosproto.Volume {
	var (
		vlms = t.cfg.Container.Volumes
		mvs  = make([]*mesosproto.Volume, 0, 0)
	)

	for _, vlm := range vlms {
		mode := mesosproto.Volume_RO
		if vlm.Mode == "RW" {
			mode = mesosproto.Volume_RW
		}

		mvs = append(mvs, &mesosproto.Volume{
			ContainerPath: proto.String(vlm.ContainerPath),
			HostPath:      proto.String(vlm.HostPath),
			Mode:          &mode,
		})
	}

	return mvs
}

func (t *Task) network() *mesosproto.ContainerInfo_DockerInfo_Network {
	var (
		network = t.cfg.Container.Docker.Network
	)

	switch network {
	case "none":
		return mesosproto.ContainerInfo_DockerInfo_NONE.Enum()
	case "host":
		return mesosproto.ContainerInfo_DockerInfo_HOST.Enum()
	case "bridge":
		return mesosproto.ContainerInfo_DockerInfo_BRIDGE.Enum()
	case "user":
		return mesosproto.ContainerInfo_DockerInfo_USER.Enum()
	}

	return mesosproto.ContainerInfo_DockerInfo_NONE.Enum()
}

func (t *Task) portMappings(offer *mesosproto.Offer) []*mesosproto.ContainerInfo_DockerInfo_PortMapping {
	var (
		pms  = t.cfg.Container.Docker.PortMappings
		dpms = make([]*mesosproto.ContainerInfo_DockerInfo_PortMapping, 0, 0)
	)

	ports := getPorts(offer)
	for i, m := range pms {
		dpms = append(dpms,
			&mesosproto.ContainerInfo_DockerInfo_PortMapping{
				HostPort:      proto.Uint32(uint32(ports[i])),
				ContainerPort: proto.Uint32(uint32(m.ContainerPort)),
				Protocol:      proto.String(m.Protocol),
			})
	}

	return dpms
}

func (t *Task) labels() *mesosproto.Labels {
	return nil
}

func getPorts(offer *mesosproto.Offer) (ports []uint64) {
	for _, res := range offer.Resources {
		if res.GetName() == "ports" {
			for _, rang := range res.GetRanges().GetRange() {
				for i := rang.GetBegin(); i <= rang.GetEnd(); i++ {
					ports = append(ports, i)
				}
			}
		}
	}
	return ports
}
