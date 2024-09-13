package common

import (
	"testing"

	"github.com/mavryk-network/mvgo/mavryk"
	"github.com/mavryk-network/mvgo/rpc"
	"github.com/stretchr/testify/assert"
)

func TestOverstake(t *testing.T) {
	assert := assert.New(t)

	baker := mavryk.MustParseAddress("mv1ELYevTeKz1tb8J8cqtYnz2vRdv9tamNmr")

	s := NewDelegationState(&rpc.Delegate{
		Delegate: baker,
	}, 745, rpc.BlockLevel(5799936))

	s.Parameters = &StakingParameters{
		LimitOfStakingOverBakingMillionth: 0,
	}

	s.AddBalance(mavryk.MustParseAddress("mv1ELYevTeKz1tb8J8cqtYnz2vRdv9tamNmr"), DelegationStateBalanceInfo{
		Balance:         1000000000,
		StakedBalance:   1000,
		UnstakedBalance: 0,
		Baker:           baker,
		StakeBaker:      baker,
	})

	delegator := mavryk.MustParseAddress("mv1VNRtHZdLzSJfyvvz2cxAoR1kWoNDWMisL")

	s.AddBalance(delegator, DelegationStateBalanceInfo{
		Balance:         1000000000,
		StakedBalance:   1000,
		UnstakedBalance: 0,
		Baker:           baker,
		StakeBaker:      baker,
	})

	assert.Equal(int64(1), s.overstakeFactor().Div64(OVERSTAKE_PRECISION).Int64())

	s.Parameters = &StakingParameters{
		LimitOfStakingOverBakingMillionth: 1000000,
	}

	assert.Equal(int64(0), s.overstakeFactor().Div64(OVERSTAKE_PRECISION).Int64())

	s.Parameters = &StakingParameters{
		LimitOfStakingOverBakingMillionth: 500000,
	}

	assert.Equal(int64(500000), s.overstakeFactor().Int64())
	assert.Equal(int64(500), s.GetDelegatorAndBakerBalances()[delegator].OverstakedBalance)

	delegator2 := mavryk.MustParseAddress("mv18vxoSEtntT8WJnjrXKD8qxcepcJeTGmkA")

	s.AddBalance(delegator2, DelegationStateBalanceInfo{
		Balance:         1000000000,
		StakedBalance:   1000,
		UnstakedBalance: 0,
		Baker:           baker,
		StakeBaker:      baker,
	})

	assert.Equal(int64(750000), s.overstakeFactor().Int64())
	assert.Equal(int64(750), s.GetDelegatorAndBakerBalances()[delegator].OverstakedBalance)
	assert.Equal(int64(750), s.GetDelegatorAndBakerBalances()[delegator2].OverstakedBalance)
	assert.Equal(int64(1000000000), s.GetDelegatorAndBakerBalances()[delegator].DelegatedBalance)
	assert.Equal(int64(1000000000), s.GetDelegatorAndBakerBalances()[delegator2].DelegatedBalance)
	assert.Equal(int64(1000), s.GetDelegatorAndBakerBalances()[delegator].StakedBalance)
	assert.Equal(int64(1000), s.GetDelegatorAndBakerBalances()[delegator2].StakedBalance)
}
