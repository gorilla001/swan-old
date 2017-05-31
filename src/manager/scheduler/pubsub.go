package scheduler

var (
	pub *publisher // global publisher
)

func init() {
	pub = &publisher{
		m:         make(map[subcriber]topicFunc),
		timeout:   time.Second * 5,
		bufferLen: 1024,
	}
}

type publisher struct {
	sync.RWMutex                         // protect m
	m            map[subcriber]topicFunc // hold all of subcribers

	timeout   time.Duration // send topic timeout
	bufferLen int           // buffer length of each subcriber channel
}
type subcriber chan interface{}
type topicFunc func(v interface{}) bool

// numHitTopic return nb of waitting subscribers who cares about the specified value.
func (p *publisher) numHitTopic(v interface{}) (n int) {
	p.RLock()
	defer p.RUnlock()
	for _, tf := range p.m {
		if tf == nil || tf(v) {
			n++
		}
	}
	return
}

// subcribeAll adds a new subscriber that receive all messages.
func (p *publisher) subcribeAll() subcriber {
	return p.subcribe(nil)
}

// subcribe adds a new subscriber that filters messages sent by a topic.
func (p *publisher) subcribe(tf topicFunc) subcriber {
	ch := make(subcriber, p.bufferLen)
	p.Lock()
	p.m[ch] = tf
	p.Unlock()
	return ch
}

// evict removes the specified subscriber from receiving any more messages.
func (p *publisher) evict(sub subcriber) {
	p.Lock()
	delete(p.m, sub)
	close(sub)
	p.Unlock()
}

func (p *publisher) publish(v interface{}) {
	p.RLock()
	defer p.RUnlock()

	var wg sync.WaitGroup
	wg.Add(len(p.m))
	// broadcasting with concurrency
	for sub, tf := range p.m {
		go func(sub subcriber, v interface{}, tf topicFunc) {
			defer wg.Done()
			p.send(sub, v, tf)
		}(sub, v, tf)
	}
	wg.Wait()
}

func (p *publisher) send(sub subcriber, v interface{}, tf topicFunc) {
	// if a subcriber setup topic filter func and not matched by the topic filter
	// skip send message to this subcriber
	if tf != nil && !tf(v) {
		return
	}

	// send with timeout
	if p.timeout > 0 {
		select {
		case sub <- v:
		case <-time.After(p.timeout):
			log.Println("send to subcriber timeout after", p.timeout.String())
		}
		return
	}

	// directely send
	sub <- v
}
