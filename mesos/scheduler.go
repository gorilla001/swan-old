package mesos

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"

	mesosproto "github.com/Dataman-Cloud/swan/proto/mesos"
	"github.com/Dataman-Cloud/swan/proto/sched"
	"github.com/Dataman-Cloud/swan/store"
	"github.com/Dataman-Cloud/swan/types"
)

const (
	reconnectDuration = time.Duration(20 * time.Second)
	resourceTimeout   = time.Duration(10 * time.Second)
	creationTimeout   = time.Duration(300 * time.Second)
)

var (
	errResourceNotEnough = errors.New("timeout for resource not enough")
	errCreationTimeout   = errors.New("task create timeout")
)

type ZKConfig struct {
	Host []string
	Path string
}

// Scheduler represents a client interacting with mesos master via x-protobuf
type Scheduler struct {
	http      *httpClient
	zkCfg     *ZKConfig
	framework *mesosproto.FrameworkInfo

	eventCh chan *sched.Event // mesos events
	errCh   chan error        // subscriber's error events
	quit    chan struct{}

	//endPoint string // eg: http://master/api/v1/scheduler
	leader  string // mesos leader address
	cluster string // name of mesos cluster

	db store.Store

	sync.RWMutex
	agents map[string]*Agent

	handlers map[sched.Event_Type]eventHandler
	tasks    map[string]*Task

	offerTimeout time.Duration

	reconnectTimer *time.Timer

	strategy Strategy
	filters  []Filter

	eventmgr *eventManager

	launch sync.Mutex
}

// NewScheduler...
func NewScheduler(cfg *ZKConfig, db store.Store, strategy Strategy, mgr *eventManager) (*Scheduler, error) {
	s := &Scheduler{
		zkCfg:        cfg,
		framework:    defaultFramework(),
		errCh:        make(chan error, 1),
		quit:         make(chan struct{}),
		agents:       make(map[string]*Agent),
		tasks:        make(map[string]*Task),
		offerTimeout: time.Duration(10 * time.Second),
		db:           db,
		strategy:     strategy,
		filters:      make([]Filter, 0),
		eventmgr:     mgr,
	}

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

// init setup mesos sched api endpoint & cluster name
func (s *Scheduler) init() error {
	state, err := s.MesosState()
	if err != nil {
		return err
	}

	l := state.Leader
	if l == "" {
		return fmt.Errorf("no mesos leader found.")
	}

	s.http = NewHTTPClient(l)
	s.leader = state.Leader

	s.cluster = state.Cluster
	if s.cluster == "" {
		s.cluster = "none" // set default cluster name
	}

	s.handlers = map[sched.Event_Type]eventHandler{
		sched.Event_SUBSCRIBED: s.subscribedHandler,
		sched.Event_OFFERS:     s.offersHandler,
		sched.Event_RESCIND:    s.rescindedHandler,
		sched.Event_UPDATE:     s.updateHandler,
		sched.Event_HEARTBEAT:  s.heartbeatHandler,
	}

	s.reconnectTimer = time.AfterFunc(reconnectDuration, s.resubscribe)

	if id := s.db.GetFrameworkId(); id != "" {
		s.framework.Id = &mesosproto.FrameworkID{
			Value: proto.String(id),
		}
	}

	return nil
}

func (s *Scheduler) InitFilters(filters []Filter) {
	s.filters = filters
}

// Cluster return current mesos cluster's name
func (s *Scheduler) ClusterName() string {
	return s.cluster
}

func (s *Scheduler) FrameworkId() *mesosproto.FrameworkID {
	return s.framework.Id
}

// Send send mesos request against the mesos master's scheduler api endpoint.
// NOTE it's the caller's responsibility to deal with the Send() error
func (s *Scheduler) Send(call *sched.Call) (*http.Response, error) {
	payload, err := proto.Marshal(call)
	if err != nil {
		return nil, err
	}

	return s.http.send(payload)
}

// Subscribe ...
func (s *Scheduler) Subscribe() error {
	log.Infof("Subscribing to mesos leader: %s", s.leader)

	call := &sched.Call{
		Type: sched.Call_SUBSCRIBE.Enum(),
		Subscribe: &sched.Call_Subscribe{
			FrameworkInfo: s.framework,
		},
	}
	if s.framework.Id != nil {
		call.FrameworkId = &mesosproto.FrameworkID{
			Value: proto.String(s.framework.Id.GetValue()),
		}
	}

	resp, err := s.Send(call)
	if err != nil {
		return fmt.Errorf("subscribe to mesos leader [%s] error [%v]", s.leader, err)
	}

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("subscribe with unexpected response [%d] - [%s]", code, string(bs))
	}

	go s.watchEvents(resp)
	return nil
}

func (s *Scheduler) Unsubscribe() error {
	log.Infoln("Unscribing from mesos leader: %s", s.leader)
	close(s.quit)
	return nil
}

func (s *Scheduler) resubscribe() {
}

func (s *Scheduler) watchEvents(resp *http.Response) {
	defer func() {
		resp.Body.Close()
	}()

	r := NewReader(resp.Body)
	dec := json.NewDecoder(r)

	for {
		select {
		case <-s.quit:
			return
		default:
			ev := new(sched.Event)
			if err := dec.Decode(ev); err != nil {
				log.Error("mesos events subscriber decode events error:", err)
				s.errCh <- err
				continue
			}

			s.handlers[ev.GetType()](ev)
		}
	}
}

func (s *Scheduler) addOffer(offer *mesosproto.Offer) {
	a, ok := s.agents[offer.AgentId.GetValue()]
	if !ok {
		return
	}

	a.addOffer(offer)
	//go func(offer *mesosproto.Offer) {
	//	time.Sleep(s.offerTimeout)
	//	// declining Mesos offers to make them available to other Mesos services
	//	if s.removeOffer(offer) {
	//		if err := s.declineOffer(offer); err != nil {
	//			log.Errorf("Error while declining offer %q: %v", offer.Id.GetValue(), err)
	//		} else {
	//			log.Debugf("Offer %q declined successfully", offer.Id.GetValue())
	//		}
	//	}
	//}(offer)
}

func (s *Scheduler) removeOffer(offer *mesosproto.Offer) bool {
	log.Debugf("Removing offer %s", offer.Id.GetValue())

	a := s.getAgent(offer.AgentId.GetValue())
	if a == nil {
		return false
	}

	found := a.removeOffer(offer.Id.GetValue())
	if a.empty() {
		s.removeAgent(offer.AgentId.GetValue())
	}

	return found
}

func (s *Scheduler) declineOffer(offer *mesosproto.Offer) error {
	call := &sched.Call{
		FrameworkId: s.FrameworkId(),
		Type:        sched.Call_DECLINE.Enum(),
		Decline: &sched.Call_Decline{
			OfferIds: []*mesosproto.OfferID{
				{
					Value: offer.GetId().Value,
				},
			},
			Filters: &mesosproto.Filters{
				RefuseSeconds: proto.Float64(1),
			},
		},
	}

	resp, err := s.Send(call)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("%d", resp.StatusCode)
	}

	return nil

}

func (s *Scheduler) rescindOffer(offerId string) {
}

func (s *Scheduler) addAgent(agent *Agent) {
	s.Lock()
	defer s.Unlock()

	s.agents[agent.id] = agent
}

func (s *Scheduler) getAgent(agentId string) *Agent {
	s.RLock()
	defer s.RUnlock()

	a, ok := s.agents[agentId]
	if !ok {
		return nil
	}

	return a
}

func (s *Scheduler) removeAgent(agentId string) {
	s.Lock()
	defer s.Unlock()

	delete(s.agents, agentId)
}

func (s *Scheduler) getAgents() map[string]*Agent {
	s.RLock()
	defer s.RUnlock()

	return s.agents
}

func (s *Scheduler) addTask(task *Task) {
	s.Lock()
	defer s.Unlock()

	s.tasks[task.TaskInfo.TaskId.GetValue()] = task
}

func (s *Scheduler) removeTask(taskID string) bool {
	s.Lock()
	defer s.Unlock()
	found := false
	_, found = s.tasks[taskID]
	if found {
		delete(s.tasks, taskID)
	}
	return found
}

func (s *Scheduler) LaunchTask(t *Task) error {
	log.Info("launching task ", *t.Name)

	//filter := filter.NewResourceFilter(
	//	t.cfg.CPUs,
	//	t.cfg.Mem,
	//	t.cfg.Disk,
	//	len(t.cfg.Container.Docker.PortMappings),
	//)
	s.launch.Lock()

	agents := make([]*Agent, 0)
	for _, agent := range s.getAgents() {
		agents = append(agents, agent)
	}

	filtered := make([]*Agent, 0)

	timeout := time.After(resourceTimeout)
	for {
		select {
		case <-timeout:
			s.launch.Unlock()

			var (
				task *types.Task
				err  error
			)

			id := t.TaskId.String()

			task, err = s.db.GetTask(id)
			if err != nil {
				log.Errorf("find task from zk got error: %v", err)
				return err
			}

			task.Status = "timeout"
			task.ErrMsg = "resource not enough"

			if err = s.db.UpdateTask(task); err != nil {
				log.Errorf("update task %s status got error: %v", id, err)
			}

			return err
		default:
			filtered = ApplyFilters(s.filters, t.cfg, agents)
			if len(filtered) > 0 {
				break
			}
		}
	}
	//res := resource{
	//	mem:      t.cfg.Mem,
	//	cpus:     t.cfg.CPUs,
	//	disk:     t.cfg.Disk,
	//	portsNum: len(t.cfg.Container.Docker.PortMappings),
	//}

	candidates := s.strategy.RankAndSort(filtered)

	chosen := candidates[0]

	var offer *mesosproto.Offer
	for _, ofr := range chosen.getOffers() {
		offer = ofr
	}

	t.Build(offer)
	s.removeOffer(offer)

	s.launch.Unlock()

	call := &sched.Call{
		FrameworkId: s.FrameworkId(),
		Type:        sched.Call_ACCEPT.Enum(),
		Accept: &sched.Call_Accept{
			OfferIds: []*mesosproto.OfferID{
				offer.GetId(),
			},
			Operations: []*mesosproto.Offer_Operation{
				&mesosproto.Offer_Operation{
					// TODO replace with LAUNCH_GROUP
					Type: mesosproto.Offer_Operation_LAUNCH.Enum(),
					Launch: &mesosproto.Offer_Operation_Launch{
						TaskInfos: []*mesosproto.TaskInfo{&t.TaskInfo},
					},
				},
			},
			Filters: &mesosproto.Filters{RefuseSeconds: proto.Float64(1)},
		},
	}

	// send call
	_, err := s.Send(call)
	if err != nil {
		return err
	}

	s.addTask(t)

	timeout := time.After(creationTimeout)
	// waitting for task's update events here until task finished or met error.
	for {
		select {
		case status := <-t.GetStatus():
			//if err := s.AckUpdateEvent(status); err != nil {
			//	log.Errorf("send status update %s for task %s error: %v", status.GetState(), t.TaskId.String(), err)
			//	break
			//}

			if t.IsDone(status) {
				//s.removeTask(task.Id)
				return t.DetectError(status)
			}
		case <-timeout:
			m := make(map[*mesosproto.TaskID]*mesosproto.AgentID)
			m[t.TaskId] = t.AgentId

			if err := s.reconcileTasks(m); err != nil {
				log.Errorf("reconcile tasks got error: %v", err)
				return errCreationTimeout
			}

			timeout.Reset()
		}
	}

	return nil
}

func (s *Scheduler) KillTask(taskID, agentID string) error {
	log.Info("stopping task ", taskID)

	call := &sched.Call{
		FrameworkId: s.FrameworkId(),
		Type:        sched.Call_KILL.Enum(),
		Kill: &sched.Call_Kill{
			TaskId: &mesosproto.TaskID{
				Value: proto.String(taskID),
			},
			AgentId: &mesosproto.AgentID{
				Value: proto.String(agentID),
			},
		},
	}

	// send call
	_, err := s.Send(call)
	if err != nil {
		return err
	}

	// subcribe waitting for task's update events here until task finished or met error.
	for {
		//update := s.SubscribeTaskUpdate(taskID)

		//// for debug
		////json.NewEncoder(os.Stdout).Encode(update)

		//status := update.GetUpdate().GetStatus()
		//_, err = s.AckUpdateEvent(status)
		//if err != nil {
		//	break
		//}

		//if s.IsTaskDone(status) {
		//	break
		//}
	}

	return nil
}

func (s *Scheduler) reconcileTasks(tasks map[interface{}]interface{}) error {
	call := &sched.Call{
		FrameworkId: s.FrameworkId(),
		Type:        sched.Call_Reconcile.Enum(),
		Reconcile: &sched.Call_Reconcile{
			Tasks: []*sched.Call_Reconcile_Task{},
		},
	}

	for t, a := range tasks {
		call.Reconcile.Tasks = append(call.Reconcile.Tasks, &sched.Call_Reconcile_Task{
			TaskId:  t,
			AgentId: a,
		})
	}

	resp, err := s.Send(call)
	if err != nil {
		return err
	}

	if code := resp.StatusCode; code != http.StatusAccepted {
		return fmt.Errorf("send reconcile call got %d not 202")
	}

	return nil
}

func (s *Scheduler) DetectError(status *mesosproto.TaskStatus) error {
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

func (s *Scheduler) AckUpdateEvent(status *mesosproto.TaskStatus) error {
	if status.GetUuid() != nil {
		call := &sched.Call{
			FrameworkId: s.FrameworkId(),
			Type:        sched.Call_ACKNOWLEDGE.Enum(),
			Acknowledge: &sched.Call_Acknowledge{
				AgentId: status.GetAgentId(),
				TaskId:  status.GetTaskId(),
				Uuid:    status.GetUuid(),
			},
		}
		resp, err := s.Send(call)
		if err != nil {
			return err
		}

		if code := resp.StatusCode; code != http.StatusAccepted {
			return fmt.Errorf("send ack got %d not 202")
		}
	}

	return nil
}

func (s *Scheduler) SubscribeEvent(w http.ResponseWriter, remote string) error {
	if s.eventmgr.Full() {
		return fmt.Errorf("%s", "too many event clients")
	}

	s.eventmgr.subscribe(remote, w)
	s.eventmgr.wait(remote)

	return nil
}
