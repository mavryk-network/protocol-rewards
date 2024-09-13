package core

import (
	"github.com/mavryk-network/mvgo/mavryk"
	"github.com/mavryk-network/protocol-rewards/common"
)

var (
	defaultFetchOptions = FetchOptions{}
	DebugFetchOptions   = FetchOptions{Force: true, Debug: true}
	ForceFetchOptions   = FetchOptions{Force: true}
)

type PRBalanceUpdate struct {
	// rpc.BalanceUpdate
	Address mavryk.Address `json:"address"`
	Amount  int64          `json:"amount,string"`

	Kind     string `json:"kind"`
	Category string `json:"category"`

	Operation     mavryk.OpHash           `json:"operation"`
	Index         int                     `json:"index"`
	InternalIndex int                     `json:"internal_index"`
	Source        common.CreationInfoKind `json:"source"`

	Delegate mavryk.Address `json:"delegate"`
}

type PRBalanceUpdates []PRBalanceUpdate

func (e PRBalanceUpdates) Len() int {
	return len(e)
}

func (e PRBalanceUpdates) Add(updates ...PRBalanceUpdate) PRBalanceUpdates {
	return append(e, updates...)
}

type FetchOptions struct {
	Force bool
	Debug bool
}
