package scheduler

// SubscribeOffer waitting for the first proper mesos offer
// which satisified the resources request.
func (c *Client) SubscribeOffer(rs []*mesos.Resource) *mesos.Offer {
	var (
		sets = make([]int, 0, 0) // subset idxes of matched offers
		idx  int                 // the final choosen idx
	)

	tf := func(v interface{}) bool {
		ev, ok := v.(*sched.Event)
		if !ok {
			return false
		}

		if ev.GetType() != sched.Event_OFFERS {
			return false
		}

		if n := len(ev.Offers.Offers); n == 0 {
			return false
		}

		// TODO
		// pick up offers those satisified requested resources.
		for i, of := range ev.Offers.Offers {
			if match(of, rs) {
				sets = append(sets, i)
			}
		}

		return len(sets) > 0
	}

	sub := pub.subcribe(tf)
	defer pub.evict(sub)

	ev := <-sub
	mesosOffers := ev.(*sched.Event).Offers.Offers

	// by random
	rand.Seed(time.Now().Unix())
	idx = sets[rand.Intn(len(sets))]

	return mesosOffers[idx]
}

// TODO(nmg)
func match(offer *mesos.Offer, res []*mesos.Resource) bool {
	for _, r := range res {
		for _, m := range offer.Resources {
			if r.GetName() == m.GetName() {
			}
		}
	}

	return true
}
