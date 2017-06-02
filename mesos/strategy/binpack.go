package strategy

import (
	"github.com/Dataman-Cloud/swan/mesos"
)

type BinpackStrategy struct{}

func (b *BinpackStrategy) RankAndSort(agents []*mesos.Agent) []*mesos.Agent {
	return nil
}
