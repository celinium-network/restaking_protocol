package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

var pk1 = ed25519.GenPrivKey().PubKey()

func TestMsgMTStakingDelegateValidation(t *testing.T) {
	delegatorAccAddr := sdk.AccAddress(pk1.Address())
	validatorAddr := sdk.ValAddress(pk1.Address())
	coin := sdk.NewCoin("test", sdk.NewIntFromUint64(10000000))

	cases := []struct {
		name      string
		delegator string
		validator string
		balance   sdk.Coin
		valid     bool
	}{
		{
			"invalid delegator",
			"bob",
			validatorAddr.String(),
			coin,
			false,
		},
		{
			"invalid validator",
			delegatorAccAddr.String(),
			"validator",
			coin,
			false,
		},
		{
			"invalid balance",
			delegatorAccAddr.String(),
			validatorAddr.String(),
			sdk.Coin{},
			false,
		},
		{
			"good",
			delegatorAccAddr.String(),
			validatorAddr.String(),
			coin,
			true,
		},
	}

	for _, c := range cases {
		msg := types.MsgMTStakingDelegate{
			DelegatorAddress: c.delegator,
			ValidatorAddress: c.validator,
			Balance:          c.balance,
		}
		err := msg.ValidateBasic()
		if c.valid {
			require.NoError(t, err, c.name)
		} else {
			require.Error(t, err, c.name)
		}
	}
}

func TestMsgMTStakingUndelegate(t *testing.T) {
	delegatorAccAddr := sdk.AccAddress(pk1.Address())
	validatorAddr := sdk.ValAddress(pk1.Address())
	coin := sdk.NewCoin("test", sdk.NewIntFromUint64(10000000))

	cases := []struct {
		name      string
		delegator string
		validator string
		balance   sdk.Coin
		valid     bool
	}{
		{
			"invalid delegator",
			"bob",
			validatorAddr.String(),
			coin,
			false,
		},
		{
			"invalid validator",
			delegatorAccAddr.String(),
			"validator",
			coin,
			false,
		},
		{
			"invalid balance",
			delegatorAccAddr.String(),
			validatorAddr.String(),
			sdk.Coin{},
			false,
		},
		{
			"good",
			delegatorAccAddr.String(),
			validatorAddr.String(),
			coin,
			true,
		},
	}

	for _, c := range cases {
		msg := types.MsgMTStakingUndelegate{
			DelegatorAddress: c.delegator,
			ValidatorAddress: c.validator,
			Balance:          c.balance,
		}
		err := msg.ValidateBasic()
		if c.valid {
			require.NoError(t, err, c.name)
		} else {
			require.Error(t, err, c.name)
		}
	}
}

func TestMsgMTStakingWithdrawReward(t *testing.T) {
	addr := sdk.AccAddress(pk1.Address())

	cases := []struct {
		name      string
		delegator string
		validator string
		valid     bool
	}{
		{
			"invalid delegator",
			"bob",
			addr.String(),
			false,
		},
		{
			"invalid validator",
			addr.String(),
			"validator",
			false,
		},
		{
			"good",
			addr.String(),
			addr.String(),
			true,
		},
	}

	for _, c := range cases {
		msg := types.MsgMTStakingWithdrawReward{
			DelegatorAddress: c.delegator,
			AgentAddress:     c.validator,
		}
		err := msg.ValidateBasic()
		if c.valid {
			require.NoError(t, err, c.name)
		} else {
			require.Error(t, err, c.name)
		}
	}
}
