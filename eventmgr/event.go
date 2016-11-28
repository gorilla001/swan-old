package eventmgr

import (
	"github.com/Dataman-Cloud/swan/types"
	fifo "github.com/foize/go.fifo"
	//"time"
)

type EventCenter struct {
	q *fifo.Queue
}

func NewEventCenter() *EventCenter {
	return &EventCenter{
		q: fifo.NewQueue(),
	}
}

func (ec *EventCenter) Push(ev *types.Event) {
	ec.q.Add(ev)
}

func (ec *EventCenter) Next() interface{} {
	return ec.q.Next()
}

//func (ec *EventCenter) Next() <-chan *types.Event {
//	evC := make(chan *types.Event, 64)
//	go func() {
//		for {
//			ev := ec.q.Next()
//			if ev != nil {
//				evC <- ev.(*types.Event)
//			}
//		}
//	}()
//
//	return evC
//}

func (ec *EventCenter) Pop() <-chan *types.Event {
	//evC := make(chan *types.Event, 64)
	//ticker := time.NewTicker(time.Duration(1) * time.Second)
	//go func() {
	//	for {
	//		select {
	//		case <-ticker.C:
	//			ev := ec.eventQ.Next()
	//			if ev != nil {
	//				evC <- ev.(*types.Event)
	//			}

	//		}
	//	}
	//}()

	//return evC
	evC := make(chan *types.Event)
	go func() {
		for {
			ev := ec.q.Next()
			if ev != nil {
				evC <- ev.(*types.Event)
			}
		}
	}()

	return evC
}
