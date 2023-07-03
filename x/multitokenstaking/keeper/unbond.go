package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

func (k Keeper) ProcessCompletedUnbonding(ctx sdk.Context) {
	matureUnbonds := k.DequeueAllMatureUBDQueue(ctx, ctx.BlockHeader().Time)
	for _, dvPair := range matureUnbonds {
		_, err := k.CompleteUnbonding(ctx, dvPair.DelegatorAddress, dvPair.AgentAddress)
		if err != nil {
			continue
		}
	}
}

func (k Keeper) DequeueAllMatureUBDQueue(ctx sdk.Context, curTime time.Time) (matureUnbonds []types.DAPair) {
	store := ctx.KVStore(k.storeKey)

	unbondingTimesliceIterator := k.UBDQueueIterator(ctx, curTime)
	defer unbondingTimesliceIterator.Close()

	for ; unbondingTimesliceIterator.Valid(); unbondingTimesliceIterator.Next() {
		timeslice := types.DAPairs{}
		value := unbondingTimesliceIterator.Value()
		k.cdc.MustUnmarshal(value, &timeslice)

		matureUnbonds = append(matureUnbonds, timeslice.Pairs...)

		store.Delete(unbondingTimesliceIterator.Key())
	}

	return matureUnbonds
}

func (k Keeper) CompleteUnbonding(ctx sdk.Context, delegator string, agentAddress string) (sdk.Coins, error) {
	ubd, found := k.GetMTStakingUnbonding(ctx, agentAddress, delegator)
	if !found {
		return nil, types.ErrNoUnbondingDelegation
	}

	agent, found := k.GetMTStakingAgentByAddress(ctx, agentAddress)
	if !found {
		return nil, types.ErrNoUnbondingDelegation
	}

	agentDelegateAddress := sdk.MustAccAddressFromBech32(agent.AgentAddress)

	balances := sdk.NewCoins()
	ctxTime := ctx.BlockHeader().Time

	delegatorAddress, err := sdk.AccAddressFromBech32(ubd.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(ubd.Entries); i++ {
		entry := ubd.Entries[i]
		if entry.IsMature(ctxTime) {
			ubd.RemoveEntry(int64(i))
			i--

			if !entry.Balance.IsZero() {
				err = k.sendCoinsFromAccountToAccount(ctx, agentDelegateAddress, delegatorAddress, sdk.Coins{entry.Balance})
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("sendCoinsFromAccountToAccount has err %s", err))
				}
				balances = balances.Add(entry.Balance)
			}
		}
	}

	if len(ubd.Entries) == 0 {
		k.RemoveMTStakingUnbonding(ctx, agentAddress, delegator)
	} else {
		k.SetMTStakingUnbondingDelegation(ctx, ubd)
	}

	return balances, nil
}
