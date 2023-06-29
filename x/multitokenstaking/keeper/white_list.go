package keeper

import (
	"strings"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/slices"
)

func (k Keeper) GetMTStakingDenomWhiteList(ctx sdk.Context) (*types.MTStakingDenomWhiteList, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DenomWhiteListKey)
	if bz == nil {
		return nil, false
	}

	whiteList := &types.MTStakingDenomWhiteList{}
	if err := k.cdc.Unmarshal(bz, whiteList); err != nil {
		return nil, false
	}

	return whiteList, true
}

func (k Keeper) AddMTStakingDenom(ctx sdk.Context, denom string) bool {
	whiteList, found := k.GetMTStakingDenomWhiteList(ctx)
	if !found || whiteList == nil {
		whiteList = &types.MTStakingDenomWhiteList{
			DenomList: []string{denom},
		}
	} else {
		for _, existedDenom := range whiteList.DenomList {
			if strings.Compare(existedDenom, denom) == 0 {
				return false
			}
		}

		whiteList.DenomList = append(whiteList.DenomList, denom)
	}

	bz, err := k.cdc.Marshal(whiteList)
	if err != nil {
		return false
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.DenomWhiteListKey, bz)

	return true
}

func (k Keeper) RemoveMTStakingDenom(ctx sdk.Context, denom string) bool {
	whiteList, found := k.GetMTStakingDenomWhiteList(ctx)
	if !found || whiteList == nil {
		return false
	}

	index := slices.Index(whiteList.DenomList, denom)
	if index == -1 {
		return false
	}

	whiteList.DenomList = append(whiteList.DenomList[:index], whiteList.DenomList[index+1:]...)
	bz, err := k.cdc.Marshal(whiteList)
	if err != nil {
		return false
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.DenomWhiteListKey, bz)

	return true
}
