package core

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/samber/lo"
	"github.com/tez-capital/ogun/store"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
)

type DefaultRpcCollector struct {
	rpcUrl string
	rpc    *rpc.Client
}

var (
	defaultCtx context.Context = context.Background()
)

func InitDefaultRpcCollector(rpcUrl string) (*DefaultRpcCollector, error) {

	client := http.Client{
		Timeout: 10 * time.Second,
	}
	rpcClient, err := rpc.NewClient(rpcUrl, &client)
	if err != nil {
		return nil, err
	}
	rpcClient.Init(defaultCtx)

	result := &DefaultRpcCollector{
		rpcUrl: rpcUrl,
		rpc:    rpcClient,
	}

	return result, result.RefreshParams()
}

func (engine *DefaultRpcCollector) GetId() string {
	return "DefaultRpcAndTzktCollector"
}

func (engine *DefaultRpcCollector) RefreshParams() error {
	return engine.rpc.Init(context.Background())
}

func (engine *DefaultRpcCollector) GetCurrentProtocol() (tezos.ProtocolHash, error) {
	params, err := engine.rpc.GetParams(context.Background(), rpc.Head)

	if err != nil {
		return tezos.ZeroProtocolHash, err
	}
	return params.Protocol, nil
}

func (engine *DefaultRpcCollector) GetCurrentCycleNumber(ctx context.Context) (int64, error) {
	head, err := engine.rpc.GetHeadBlock(ctx)
	if err != nil {
		return 0, err
	}

	return head.GetLevelInfo().Cycle, err
}

func (engine *DefaultRpcCollector) GetLastCompletedCycle(ctx context.Context) (int64, error) {
	cycle, err := engine.GetCurrentCycleNumber(ctx)
	return cycle - 1, err
}

func (engine *DefaultRpcCollector) determineLastBlockOfCycle(cycle int64) rpc.BlockID {
	height := engine.rpc.Params.CycleEndHeight(cycle)
	return rpc.BlockLevel(height)
}

func (engine *DefaultRpcCollector) GetActiveDelegatesFromCycle(ctx context.Context, cycle int64) (rpc.DelegateList, error) {
	id := engine.determineLastBlockOfCycle(cycle)
	dl, err := engine.rpc.ListActiveDelegates(ctx, id)
	if err != nil {
		return nil, err
	}

	return dl, nil
}

func (engine *DefaultRpcCollector) GetDelegateFromCycle(ctx context.Context, cycle int64, delegateAddress tezos.Address) (*rpc.Delegate, error) {
	blockId := engine.determineLastBlockOfCycle(cycle)

	return engine.rpc.GetDelegate(ctx, delegateAddress, blockId)
}

func (engine *DefaultRpcCollector) fetchDelegationState(ctx context.Context, delegate *rpc.Delegate, blockId rpc.BlockID) (*store.DelegationState, error) {
	state := &store.DelegationState{
		Baker:        delegate.Delegate,
		Balances:     make(map[tezos.Address]tezos.Z, len(delegate.DelegatedContracts)+1),
		TotalBalance: tezos.Z{},
	}

	state.Balances[delegate.Delegate] = tezos.NewZ(delegate.FullBalance - delegate.CurrentFrozenDeposits)

	for _, address := range delegate.DelegatedContracts {
		balance, err := engine.rpc.GetContractBalance(ctx, address, blockId)
		if err != nil {
			return nil, err
		}
		state.Balances[address] = balance
	}

	state.TotalBalance = lo.Reduce(lo.Values(state.Balances), func(acc tezos.Z, balance tezos.Z, _ int) tezos.Z {
		return acc.Add(balance)
	}, state.TotalBalance)

	return state, nil
}

func (engine *DefaultRpcCollector) GetDelegationState(ctx context.Context, delegate *rpc.Delegate) (*store.DelegationState, error) {
	if delegate.MinDelegated.Level.Level == 0 {
		return nil, errors.New("delegate has no minimum delegated balance")
	}

	blockWithMinimumBalance, err := engine.rpc.GetBlock(ctx, rpc.BlockLevel(delegate.MinDelegated.Level.Level))
	if err != nil {
		return nil, err
	}

	state, err := engine.fetchDelegationState(ctx, delegate, rpc.BlockLevel(delegate.MinDelegated.Level.Level-1))
	if err != nil {
		return nil, err
	}

	state.Cycle = delegate.MinDelegated.Level.Cycle
	state.Level = delegate.MinDelegated.Level.Level

	found := false

	allBalanceUpdates := make(ExtendedBalanceUpdates, 0, len(blockWithMinimumBalance.Operations)*2 /* thats minimum of balance updates we expect*/)
	// block balance updates
	allBalanceUpdates = allBalanceUpdates.AddBalanceUpdates(tezos.ZeroOpHash, -1, BlockBalanceUpdateSource, blockWithMinimumBalance.Metadata.BalanceUpdates...)

	for _, batch := range blockWithMinimumBalance.Operations {
		for _, operation := range batch {
			// first op fees
			for transactionIndex, content := range operation.Contents {
				allBalanceUpdates = allBalanceUpdates.AddBalanceUpdates(operation.Hash,
					int64(transactionIndex),
					TransactionMetadataBalanceUpdateSource,
					content.Meta().BalanceUpdates...,
				)
			}
			// then transfers
			for transactionIndex, content := range operation.Contents {
				allBalanceUpdates = allBalanceUpdates.AddBalanceUpdates(operation.Hash,
					int64(transactionIndex),
					TransactionContentsBalanceUpdateSource,
					content.Result().BalanceUpdates...,
				)

				for internalResultIndex, internalResult := range content.Meta().InternalResults {
					// TODO: check this
					slices.Reverse(internalResult.Result.BalanceUpdates)
					allBalanceUpdates = allBalanceUpdates.AddInternalResultBalanceUpdates(operation.Hash,
						state.Index,
						int64(internalResultIndex),
						internalResult.Result.BalanceUpdates...,
					)
				}
			}

		}
	}

	targetAmount := delegate.MinDelegated.Amount

	for _, balanceUpdate := range allBalanceUpdates {
		if _, found := state.Balances[balanceUpdate.Address()]; !found {
			continue
		}

		state.Balances[balanceUpdate.Address()] = state.Balances[balanceUpdate.Address()].Add64(balanceUpdate.Amount())
		state.TotalBalance = state.TotalBalance.Add64(balanceUpdate.Amount())
		// TODO: check this == sign here
		if state.TotalBalance.Int64() >= targetAmount {
			found = true
			state.Operation = balanceUpdate.Operation
			state.Index = balanceUpdate.Index
			state.InternalIndex = balanceUpdate.InternalIndex
			state.Source = balanceUpdate.Source
			break
		}
	}

	if !found {
		return nil, errors.New("delegate has not reached minimum delegated balance")
	}
	return state, nil
}
