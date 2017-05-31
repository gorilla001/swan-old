package scheduler

// SubscribeUpdate waitting for the specified taskID until finished or error.
func (s *Scheduler) SubscribeTaskUpdate(taskID string) *sched.Event {
	tf := func(v interface{}) bool {
		ev, ok := v.(*sched.Event)
		if !ok {
			return false
		}

		if ev.GetType() != sched.Event_UPDATE {
			return false
		}

		status := ev.GetUpdate().GetStatus()
		return taskID == status.TaskId.GetValue()
	}

	sub := pub.subcribe(tf)
	defer pub.evict(sub)

	ev := <-sub
	return ev.(*sched.Event)
}
