package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"

	"github.com/bbklab/swan-ng/mesos/protobuf/mesos"
	"github.com/bbklab/swan-ng/mesos/protobuf/sched"
)

// Scheduler represents a client interacting with mesos master via x-protobuf
type Scheduler struct {
	http      *http.Client
	zkUrl     *url.URL
	framework *mesos.FrameworkInfo

	eventCh chan *sched.Event // mesos events
	errCh   chan error        // subscriber's error events

	//endPoint string // eg: http://master/api/v1/scheduler
	leader  string // mesos leader address
	cluster string // name of mesos cluster
}

// NewClient ...
func NewClient(url *url.URL) (*Client, error) {
	s := &Scheduler{
		zkUrl:     url,
		framework: defaultFramework(),
		eventCh:   make(chan *sched.Event, 1024),
		errCh:     make(chan error, 1),
	}

	if err := s.init(); err != nil {
		return nil, err
	}

	return c, nil
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

	s.http = NewHTTPClient()
	s.leader = state.Leader

	s.cluster = state.Cluster
	if c.cluster == "" {
		c.cluster = "none" // set default cluster name
	}

	return nil
}

// Cluster return current mesos cluster's name
func (s *Scheduler) Cluster() string {
	return s.cluster
}

func (s *Scheduler) FrameworkId() *mesos.FrameworkID {
	return s.framework.Id
}

// Send send mesos request against the mesos master's scheduler api endpoint.
// NOTE it's the caller's responsibility to deal with the Send() error
func (s *Scheduler) Send(call *sched.Call) error {
	payload, err := proto.Marshal(call)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.send(payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if code := resp.StatusCode; code != 202 {
		return fmt.Errorf("expect 202, got %d: [%s]", code, string(bs))
	}

	return nil
}

// ReSubscribe ...
// TODO
func (c *Client) ReSubscribe(call *sched.Call) error {
	// reinit
	// resub
	return nil
}

// Subscribe ...
func (s *Scheduler) Subscribe() error {
	log.Infof("subscribing to mesos leader: %s", s.leader)

	call := &sched.Call{
		Type: sched.Call_SUBSCRIBE.Enum(),
		Subscribe: &sched.Call_Subscribe{
			FrameworkInfo: c.framework,
		},
	}
	if c.framework.Id != nil {
		call.FrameworkId = &mesos.FrameworkID{
			Value: proto.String(c.framework.Id.GetValue()),
		}
	}

	resp, err := c.send(call)
	if err != nil {
		return fmt.Errorf("subscribe to mesos leader [%s] error [%v]", s.leader, err)
	}

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("subscribe with unexpected response [%d] - [%s]", code, string(bs))
	}
	log.Printf("subscribed to mesos leader: %s succeed.")

	go s.watchEvents(resp)
	return nil
}

func (s *Scheduler) watchEvents(resp *http.Response) {
	log.Println("mesos event subscriber starting")

	defer func() {
		log.Warnln("mesos event subscriber quited")
		resp.Body.Close()
	}()

	r := NewReader(resp.Body)
	dec := json.NewDecoder(r)

	for {
		ev := new(sched.Event)
		if err := dec.Decode(ev); err != nil {
			log.Errorln("mesos events subscriber decode events error:", err)
			s.errCh <- err
			return
		}

		s.eventCh <- ev
	}
}
