package scheduler

import (
	"errors"
	//"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"net/http"
	"time"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/types"
)

func (s *Scheduler) RequestOffers(resources []*mesos.Resource) ([]*mesos.Offer, error) {
	logrus.Info("Requesting offers")
	//s.EventManager().Push(&types.Event{
	//	Type:    "PROCESS",
	//	Message: "Requesting offers",
	//})

	var event *sched.Event

	select {
	case event = <-s.GetEvent(sched.Event_OFFERS):
	case <-time.After(time.Second * time.Duration(5)):
		s.EventManager().Push(&types.Event{
			Type:    "FINISHED",
			Message: "Offer timeout",
		})
		return nil, errors.New("Offer timeout")
	}

	if event == nil {
		call := &sched.Call{
			FrameworkId: s.framework.GetId(),
			Type:        sched.Call_REQUEST.Enum(),
			Request: &sched.Call_Request{
				Requests: []*mesos.Request{
					{
						Resources: resources,
					},
				},
			},
		}

		if _, err := s.send(call); err != nil {
			//s.EventManager().Push(&types.Event{
			//	Type:    "FINISHED",
			//	Message: fmt.Sprintf("Request offer failed: %s", err.Error()),
			//})
			logrus.Errorf("Request offer failed: %s", err.Error())
			return nil, err
		}
		event = <-s.GetEvent(sched.Event_OFFERS)

	}
	logrus.Infof("Received %d offer(s).", len(event.Offers.Offers))
	//s.EventManager().Push(&types.Event{
	//	Type:    "PROCESS",
	//	Message: fmt.Sprintf("Received %d offer(s).", len(event.Offers.Offers)),
	//})

	return event.Offers.Offers, nil
}

// DeclineResource is used to send DECLINE request to mesos to release offer. This
// is very important, otherwise resource will be taked until framework exited.
func (s *Scheduler) DeclineResource(offerId *string) (*http.Response, error) {
	call := &sched.Call{
		FrameworkId: s.framework.GetId(),
		Type:        sched.Call_DECLINE.Enum(),
		Decline: &sched.Call_Decline{
			OfferIds: []*mesos.OfferID{
				{
					Value: offerId,
				},
			},
			Filters: &mesos.Filters{
				RefuseSeconds: proto.Float64(1),
			},
		},
	}

	return s.send(call)
}

func (s *Scheduler) OfferedResources(offer *mesos.Offer) (cpus, mem, disk float64) {
	for _, res := range offer.GetResources() {
		if res.GetName() == "cpus" {
			cpus += *res.GetScalar().Value
		}
		if res.GetName() == "mem" {
			mem += *res.GetScalar().Value
		}
		if res.GetName() == "disk" {
			disk += *res.GetScalar().Value
		}
	}

	return
}
