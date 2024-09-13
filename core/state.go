package core

import (
	"slices"
	"sync"

	"github.com/mavryk-network/mvgo/mavryk"
	"github.com/samber/lo"
)

var (
	mtx sync.RWMutex
)

type state struct {
	lastFetchedCycle      int64
	delegatesBeingFetched map[int64][]mavryk.Address
}

func newState() *state {
	return &state{
		delegatesBeingFetched: make(map[int64][]mavryk.Address),
	}
}

func (s *state) AddDelegateBeingFetched(cycle int64, delegate ...mavryk.Address) {
	mtx.Lock()
	defer mtx.Unlock()

	s.delegatesBeingFetched[cycle] = append(s.delegatesBeingFetched[cycle], delegate...)
}

func (s *state) RemoveCycleBeingFetched(cycle int64, delegate ...mavryk.Address) {
	mtx.Lock()
	defer mtx.Unlock()

	if _, ok := s.delegatesBeingFetched[cycle]; !ok {
		return
	}

	s.delegatesBeingFetched[cycle] = lo.Filter(s.delegatesBeingFetched[cycle], func(d mavryk.Address, _ int) bool {
		for _, del := range delegate {
			if d.Equal(del) {
				return false
			}
		}
		return true
	})
}

func (s *state) IsDelegateBeingFetched(cycle int64, delegate mavryk.Address) bool {
	mtx.RLock()
	defer mtx.RUnlock()

	if _, ok := s.delegatesBeingFetched[cycle]; !ok {
		return false
	}

	return slices.Contains(s.delegatesBeingFetched[cycle], delegate)
}

func (s *state) SetLastFetchedCycle(cycle int64) {
	mtx.Lock()
	defer mtx.Unlock()

	s.lastFetchedCycle = cycle
}

func (s *state) GetLastFetchedCycle() int64 {
	mtx.RLock()
	defer mtx.RUnlock()

	return s.lastFetchedCycle
}
