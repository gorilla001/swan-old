package mesos

import (
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan/proto/sched"
	log "github.com/Sirupsen/logrus"
)

const (
	defaultOfferTimeout  = 30 * time.Second
	defaultRefuseTimeout = 5 * time.Second
)

var lock sync.Mutex

type eventHandler func(*sched.Event)

func (s *Scheduler) subscribedHandler(event *sched.Event) {
	var (
		id = event.GetSubscribed().FrameworkId
	)

	log.Infof("Subscription successful with frameworkId %s", id.GetValue())

	s.framework.Id = id

	if err := s.db.UpdateFrameworkId(id.GetValue()); err != nil {
		log.Errorf("update frameworkid got error:%s", err)
	}
}

func (s *Scheduler) offersHandler(event *sched.Event) {
	var (
		offers = event.Offers.Offers
	)

	log.Debugf("Receiving %d offer(s) from mesos", len(offers))

	for _, offer := range offers {
		agentId := offer.AgentId.GetValue()

		a := s.getAgent(agentId)
		if a == nil {
			a = newAgent(agentId)
			s.addAgent(a)
		}

		s.addOffer(offer)
	}
}

func (s *Scheduler) rescindedHandler(event *sched.Event) {
	var (
		offerId = event.GetRescind().OfferId.GetValue()
	)

	log.Debugln("Receiving rescind msg for offer ", offerId)

	for _, agent := range s.getAgents() {
		if offer := agent.getOffer(offerId); offer != nil {
			s.removeOffer(offer)
			break
		}
	}
}

func (s *Scheduler) updateHandler(event *sched.Event) {
	lock.Lock()
	defer lock.Unlock()

	var (
		status  = event.GetUpdate().GetStatus()
		taskId  = status.TaskId.GetValue()
		healthy = status.GetHealthy()
	)

	log.Printf("Received status update %s for task %s", status.GetState(), taskId)

	if err := s.AckUpdateEvent(status); err != nil {
		log.Errorf("send status update %s for task %s error: %v", status.GetState(), taskId, err)
	}

	if task, ok := s.tasks[taskId]; ok {
		task.SendStatus(status)
	}

	task, err := s.db.GetTask(taskId)
	if err != nil {
		log.Errorf("find app from zk got error: %v", err)
		return
	}

	task.Status = status.String()
	task.Healthy = healthy
	task.ErrMsg = status.GetReason().String() + ":" + status.GetMessage()

	if err := s.db.UpdateTask(task); err != nil {
		log.Errorf("update task status error: %v", err)
	}
}

func (s *Scheduler) heartbeatHandler(event *sched.Event) {
	log.Debug("Receive heartbeat msg from mesos")

	s.reconnectTimer.Reset(reconnectDuration)
}
