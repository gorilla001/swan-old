package backend

import (
	"fmt"
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	"net/http"
)

func (b *Backend) LaunchApplication(version *types.Version) error {
	b.sched.TaskLaunched = 0

	// Set scheduler's status to busy for accepting resource.
	b.sched.Status = "busy"
	// Set scheduler's status back to idle after launch applicaiton.
	defer func() {
		b.sched.Status = "idle"
	}()

	b.sched.EventManager().Push(&types.Event{
		ID:      version.ID,
		Type:    "PROCESS",
		Message: fmt.Sprintf("Building resources cpus(%.f) mem(%.f) disk(%.f)", version.Cpus, version.Mem, version.Disk),
	})
	resources := b.sched.BuildResources(version.Cpus, version.Mem, version.Disk)

	b.sched.EventManager().Push(&types.Event{
		ID:      version.ID,
		Type:    "PROCESS",
		Message: "Requesting offers",
	})

	offers, err := b.sched.RequestOffers(resources)
	if err != nil {
		logrus.Errorf("Request offers failed: %s", err.Error())
		b.sched.EventManager().Push(&types.Event{
			ID:      version.ID,
			Type:    "FINISHED",
			Message: fmt.Sprintf("Request offers failed: %s", err.Error()),
		})
		return err
	}
	b.sched.EventManager().Push(&types.Event{
		ID:      version.ID,
		Type:    "PROCESS",
		Message: fmt.Sprintf("Received %d offer(s).", len(offers)),
	})

	for _, offer := range offers {
		cpus, mem, disk := b.sched.OfferedResources(offer)
		var tasks []*mesos.TaskInfo
		for b.sched.TaskLaunched < version.Instances &&
			cpus >= version.Cpus &&
			mem >= version.Mem &&
			disk >= version.Disk {
			task, err := b.sched.BuildTask(offer, version, "")
			if err != nil {
				return fmt.Errorf("Build task failed: %s", err.Error())
			}

			taskInfo := b.sched.BuildTaskInfo(offer, resources, task)
			tasks = append(tasks, taskInfo)

			task.Port = int(*taskInfo.Container.Docker.PortMappings[0].HostPort)
			if err := b.store.SaveTask(task); err != nil {
				return fmt.Errorf("Save task failed: %s", err.Error())
			}

			if len(task.HealthChecks) != 0 {
				if err := b.store.SaveCheck(task,
					*taskInfo.Container.Docker.PortMappings[0].HostPort,
					version.ID); err != nil {
				}
				for _, healthCheck := range task.HealthChecks {
					check := types.Check{
						ID:       task.Name,
						Address:  *task.AgentHostname,
						Port:     int(*taskInfo.Container.Docker.PortMappings[0].HostPort),
						TaskID:   task.Name,
						AppID:    version.ID,
						Protocol: healthCheck.Protocol,
						Interval: int(healthCheck.IntervalSeconds),
						Timeout:  int(healthCheck.TimeoutSeconds),
					}
					if healthCheck.Command != nil {
						check.Command = healthCheck.Command
					}

					if healthCheck.Path != nil {
						check.Path = *healthCheck.Path
					}

					if healthCheck.ConsecutiveFailures != 0 {
						check.MaxFailures = int(healthCheck.ConsecutiveFailures)
					}

					b.sched.HealthCheckManager.Add(&check)
				}
			}

			b.sched.TaskLaunched++
			cpus -= version.Cpus
			mem -= version.Mem
			disk -= version.Disk
		}

		if len(tasks) == 0 {
			return fmt.Errorf("Not enough resource")
		}

		b.sched.EventManager().Push(&types.Event{
			ID:      version.ID,
			Type:    "PROCESS",
			Message: fmt.Sprintf("Launch %d tasks with offer %s", len(tasks), *offer.GetId().Value),
		})
		resp, err := b.sched.LaunchTasks(offer, tasks)
		if err != nil {
			return fmt.Errorf("Launchs task failed: %s", err.Error())
		}

		if resp != nil && resp.StatusCode != http.StatusAccepted {
			return fmt.Errorf("status code %d received", resp.StatusCode)
		}
	}

	return nil
}
