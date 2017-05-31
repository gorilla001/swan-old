package scheduler

import (
	log "github.com/Sirupsen/logrus"
)

func (s *Scheduler) LaunchTask(offer *mesos.Offer, task *mesos.TaskInfo) error {
	log.Info("launching task:", *task.Name)

	call := &sched.Call{
		FrameworkId: c.FrameworkId(),
		Type:        sched.Call_ACCEPT.Enum(),
		Accept: &sched.Call_Accept{
			OfferIds: []*mesos.OfferID{
				offer.GetId(),
			},
			Operations: []*mesos.Offer_Operation{
				&mesos.Offer_Operation{
					// TODO replace with LAUNCH_GROUP
					Type: mesos.Offer_Operation_LAUNCH.Enum(),
					Launch: &mesos.Offer_Operation_Launch{
						TaskInfos: []*mesos.TaskInfo{task},
					},
				},
			},
			Filters: &mesos.Filters{RefuseSeconds: proto.Float64(1)},
		},
	}

	// send call
	err := s.Send(call)
	if err != nil {
		return err
	}

	// subcribe waitting for task's update events here until task finished or met error.
	for {
		update := s.SubscribeTaskUpdate(task.TaskId.GetValue())

		// for debug
		//json.NewEncoder(os.Stdout).Encode(update)

		status := update.GetUpdate().GetStatus()
		err = s.AckUpdateEvent(status)
		if err != nil {
			break
		}

		if s.IsTaskDone(status) {
			err = s.DetectError(status) // check if we met an error.
			break
		}
	}

	return err
}

func (s *Scheduler) KillTask(taskID, agentID string) error {
	log.Info("stopping task:", taskID)

	call := &sched.Call{
		FrameworkId: c.FrameworkId(),
		Type:        sched.Call_KILL.Enum(),
		Kill: &sched.Call_Kill{
			TaskId: &mesos.TaskID{
				Value: proto.String(taskID),
			},
			AgentId: &mesos.AgentID{
				Value: proto.String(agentID),
			},
		},
	}

	// send call
	err := c.Send(call)
	if err != nil {
		return err
	}

	// subcribe waitting for task's update events here until task finished or met error.
	for {
		update := s.SubscribeTaskUpdate(taskID)

		// for debug
		//json.NewEncoder(os.Stdout).Encode(update)

		status := update.GetUpdate().GetStatus()
		err = s.AckUpdateEvent(status)
		if err != nil {
			break
		}

		if s.IsTaskDone(status) {
			err = s.DetectError(status) // check if we met an error.
			break
		}
	}

	return nil
}

// IsTaskDone check that if a task is done or not according by task status.
func (s *Scheduler) IsTaskDone(status *mesos.TaskStatus) bool {
	state := status.GetState()

	switch state {
	case mesos.TaskState_TASK_RUNNING,
		mesos.TaskState_TASK_FINISHED,
		mesos.TaskState_TASK_FAILED,
		mesos.TaskState_TASK_KILLED,
		mesos.TaskState_TASK_ERROR,
		mesos.TaskState_TASK_LOST,
		mesos.TaskState_TASK_DROPPED,
		mesos.TaskState_TASK_GONE:
		return true
	}

	return false
}

func (s *Scheduler) DetectError(status *mesos.TaskStatus) error {
	var (
		state = status.GetState()
		//data  = status.GetData() // docker container inspect result
	)

	switch state {
	case mesos.TaskState_TASK_FAILED,
		mesos.TaskState_TASK_ERROR,
		mesos.TaskState_TASK_LOST,
		mesos.TaskState_TASK_DROPPED,
		mesos.TaskState_TASK_UNREACHABLE,
		mesos.TaskState_TASK_GONE,
		mesos.TaskState_TASK_GONE_BY_OPERATOR,
		mesos.TaskState_TASK_UNKNOWN:
		bs, _ := json.Marshal(map[string]interface{}{
			"state":   state.String(),
			"message": status.GetMessage(),
			"source":  status.GetSource().String(),
			"reason":  status.GetReason().String(),
			"healthy": status.GetHealthy(),
		})
		return errors.New(string(bs))
	}

	return nil
}
