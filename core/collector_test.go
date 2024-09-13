package core

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"testing"

	"github.com/mavryk-network/mvgo/mavryk"
	"github.com/mavryk-network/mvgo/rpc"
	"github.com/mavryk-network/protocol-rewards/constants"
	"github.com/mavryk-network/protocol-rewards/test"
	"github.com/stretchr/testify/assert"
)

var (
	defaultCtx context.Context = context.Background()
)

func getTransport(path string) *test.TestTransport {
	transport, err := test.NewTestTransport(http.DefaultTransport, path, path+".gob.lz4")
	if err != nil {
		panic(err)
	}
	return transport
}

func TestGetActiveDelegates(t *testing.T) {
	assert := assert.New(t)

	cycle := 745
	lastBlockInTheCycle := rpc.BlockLevel(5799936)
	collector, err := newRpcCollector(defaultCtx, []string{"https://atlasnet.rpc.mavryk.network/"}, []string{"https://atlasnet.api.mavryk.network/"}, getTransport(fmt.Sprintf("../test/data/%d", cycle)))
	assert.Nil(err)

	delegates, err := collector.GetActiveDelegatesFromCycle(defaultCtx, lastBlockInTheCycle)
	assert.Nil(err)
	assert.Equal(354, len(delegates))
}

func TestGetDelegationStateNoStaking(t *testing.T) {
	assert := assert.New(t)
	debug.SetMaxThreads(1000000)

	// cycle 745
	cycle := int64(745)
	lastBlockInTheCycle := rpc.BlockLevel(5799936)
	collector, err := newRpcCollector(defaultCtx, []string{"https://atlasnet.rpc.mavryk.network/"}, []string{"https://atlasnet.api.mavryk.network/"}, getTransport(fmt.Sprintf("../test/data/%d", cycle)))
	assert.Nil(err)

	delegates, err := collector.GetActiveDelegatesFromCycle(defaultCtx, lastBlockInTheCycle)
	assert.Nil(err)

	err = runInParallel(defaultCtx, delegates, 100, func(ctx context.Context, addr mavryk.Address, mtx *sync.RWMutex) bool {
		delegate, err := collector.GetDelegateFromCycle(defaultCtx, lastBlockInTheCycle, addr)
		if err != nil {
			assert.Nil(err)
			return true
		}

		_, err = collector.GetDelegationState(defaultCtx, delegate, cycle, lastBlockInTheCycle)
		if err != nil && err != constants.ErrDelegateHasNoMinimumDelegatedBalance {
			assert.Nil(err)
			return true
		}
		return false
	})
	assert.Nil(err)

	// cycle 746
	cycle = int64(746)
	lastBlockInTheCycle = rpc.BlockLevel(5824512)
	collector, err = newRpcCollector(defaultCtx, []string{"https://atlasnet.rpc.mavryk.network/"}, []string{"https://atlasnet.api.mavryk.network/"}, getTransport(fmt.Sprintf("../test/data/%d", cycle)))
	assert.Nil(err)

	delegates, err = collector.GetActiveDelegatesFromCycle(defaultCtx, lastBlockInTheCycle)
	assert.Nil(err)

	err = runInParallel(defaultCtx, delegates, 100, func(ctx context.Context, addr mavryk.Address, mtx *sync.RWMutex) bool {
		delegate, err := collector.GetDelegateFromCycle(defaultCtx, lastBlockInTheCycle, addr)
		if err != nil {
			assert.Nil(err)
			return true
		}

		_, err = collector.GetDelegationState(defaultCtx, delegate, cycle, lastBlockInTheCycle)
		if err != nil && err != constants.ErrDelegateHasNoMinimumDelegatedBalance {
			assert.Nil(err)
			return true
		}
		return false
	})
	assert.Nil(err)
}

func TestGetDelegationState(t *testing.T) {
	assert := assert.New(t)
	debug.SetMaxThreads(1000000)

	// cycle 748
	cycle := int64(748)
	lastBlockInTheCycle := rpc.BlockLevel(5873664)
	collector, err := newRpcCollector(defaultCtx, []string{"https://atlasnet.rpc.mavryk.network/"}, []string{"https://atlasnet.api.mavryk.network/"}, getTransport(fmt.Sprintf("../test/data/%d", cycle)))
	assert.Nil(err)

	delegates, err := collector.GetActiveDelegatesFromCycle(defaultCtx, lastBlockInTheCycle)
	assert.Nil(err)

	err = runInParallel(defaultCtx, delegates, 100, func(ctx context.Context, addr mavryk.Address, mtx *sync.RWMutex) bool {
		delegate, err := collector.GetDelegateFromCycle(defaultCtx, lastBlockInTheCycle, addr)
		if err != nil {
			assert.Nil(err)
			return true
		}

		_, err = collector.GetDelegationState(defaultCtx, delegate, cycle, lastBlockInTheCycle)
		if err != nil && err != constants.ErrDelegateHasNoMinimumDelegatedBalance {
			assert.Nil(err)
			return true
		}
		return false
	})
	assert.Nil(err)
}

func TestCycle749RaceConditions(t *testing.T) {
	assert := assert.New(t)
	debug.SetMaxThreads(1000000)

	cycle := int64(749)
	lastBlockInTheCycle := rpc.BlockLevel(5898240)
	collector, err := newRpcCollector(defaultCtx, []string{"https://atlasnet.rpc.mavryk.network/"}, []string{"https://atlasnet.api.mavryk.network/"}, getTransport(fmt.Sprintf("../test/data/%d", cycle)))
	assert.Nil(err)

	delegates := []mavryk.Address{
		mavryk.MustParseAddress("mv18vxoSEtntT8WJnjrXKD8qxcepcJeTGmkA"),
		mavryk.MustParseAddress("mv1MVC17roTyHPTb3kDMiNzQmjacq6zCYXeM"),
		mavryk.MustParseAddress("mv3QxcvapxQuE784gCfGoUJScygDiiLiCsbK"),
		mavryk.MustParseAddress("mv1MfKc4giVD7GmqJnj82s6VQi6ufWF5JBtt"),
		mavryk.MustParseAddress("mv1B6CNnAbLdB7etMdXW7r4AmiNtJVggJios"),
		mavryk.MustParseAddress("mv1AMtXT4JpBZBtMpQ3KKcqLFCtecB3xzznj"),
		mavryk.MustParseAddress("mv3CXh2o75d43pBZMvkgXBQDYeUea1gMYG1Z"),
	}

	err = runInParallel(defaultCtx, delegates, 100, func(ctx context.Context, addr mavryk.Address, mtx *sync.RWMutex) bool {
		delegate, err := collector.GetDelegateFromCycle(defaultCtx, lastBlockInTheCycle, addr)
		if err != nil {
			assert.Nil(err)
			return true
		}

		_, err = collector.GetDelegationState(defaultCtx, delegate, cycle, lastBlockInTheCycle)
		if err != nil && err != constants.ErrDelegateHasNoMinimumDelegatedBalance {
			fmt.Println(delegate.Delegate.String())
			assert.Nil(err)
			return true
		}
		return false
	})
	assert.Nil(err)
}
