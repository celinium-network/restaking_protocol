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
		delegatorAccAddr := sdk.AccAddress(dvPair.DelegatorAddress)
		agentAccAddr := sdk.AccAddress(dvPair.AgentAddress)
		_, err := k.CompleteUnbonding(ctx, delegatorAccAddr, agentAccAddr)
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

func (k Keeper) CompleteUnbonding(ctx sdk.Context, delegatorAccAddr, agentAccAddr sdk.AccAddress) (sdk.Coins, error) {
	ubd, found := k.GetMTStakingUnbonding(ctx, agentAccAddr, delegatorAccAddr)
	if !found {
		return nil, types.ErrNoUnbondingDelegation
	}

	balances := sdk.NewCoins()
	ctxTime := ctx.BlockHeader().Time

	for i := 0; i < len(ubd.Entries); i++ {
		entry := ubd.Entries[i]
		if entry.IsMature(ctxTime) {
			ubd.RemoveEntry(int64(i))
			i--

			if !entry.Balance.IsZero() {
				err := k.sendCoinsFromAccountToAccount(ctx, agentAccAddr, delegatorAccAddr, sdk.Coins{entry.Balance})
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("sendCoinsFromAccountToAccount has err %s", err))
				}
				balances = balances.Add(entry.Balance)
			}
		}
	}

	if len(ubd.Entries) == 0 {
		k.RemoveMTStakingUnbonding(ctx, agentAccAddr, delegatorAccAddr)
	} else {
		k.SetMTStakingUnbondingDelegation(ctx, ubd)
	}

	return balances, nil
}
