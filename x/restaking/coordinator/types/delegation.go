package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (ubd *UnbondingDelegation) AddEntry(creationHeight uint64, coin sdk.Coin, unbondingID uint64) {
	// Check the entries exists with creation_height and complete_time
	entryIndex := -1
	for index, ubdEntry := range ubd.Entries {
		if ubdEntry.CreateHeight == creationHeight {
			entryIndex = index
			break
		}
	}
	// entryIndex exists
	if entryIndex != -1 {
		ubdEntry := ubd.Entries[entryIndex]
		ubdEntry.Amount = ubdEntry.Amount.Add(coin)

		// update the entry
		ubd.Entries[entryIndex] = ubdEntry
	} else {
		// append the new unbond delegation entry
		entry := NewUnbondingDelegationEntry(creationHeight, coin, unbondingID)
		ubd.Entries = append(ubd.Entries, entry)
	}
}

func NewUnbondingDelegationEntry(creationHeight uint64, coin sdk.Coin, unbondingID uint64) UnbondingEntry {
	return UnbondingEntry{
		Id:           unbondingID,
		Amount:       coin,
		CreateHeight: creationHeight,
		CompleteTime: -1,
	}
}

func NewUnbondingDelegation(
	delegatorAddr, operatorAddr sdk.AccAddress, creationHeight uint64, balance sdk.Coin, id uint64,
) UnbondingDelegation {
	return UnbondingDelegation{
		DelegatorAddress: delegatorAddr.String(),
		OperatorAddress:  operatorAddr.String(),
		Entries:          []UnbondingEntry{NewUnbondingDelegationEntry(creationHeight, balance, id)},
	}
}
