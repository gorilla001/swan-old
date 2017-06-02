package strategy

import (
	"math/rand"
	"time"

	"github.com/Dataman-Cloud/swan/mesos"
)

type RandomStrategy struct {
	r *rand.Rand
}

func NewRandomStrategy() *RandomStrategy {
	return &RandomStrategy{
		r: rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
	}
}

func (m *RandomStrategy) RankAndSort(agents []*mesos.Agent) []*mesos.Agent {
	for i := 0; i < len(agents); i++ {
		j := m.r.Intn(i + 1)
		agents[i], agents[j] = agents[j], agents[i]
	}

	return agents
}
