package keeper

import (
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"

	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (k Keeper) EndBlockValidatorSetUpdate(ctx sdk.Context) {
	k.QueueValidatorSetChangePackets(ctx)

	k.SendValidatorSetChangePackets(ctx)
}

func (k Keeper) QueueValidatorSetChangePackets(ctx sdk.Context) {
	_, err := k.GetCoordinatorChannelID(ctx)
	if err != nil {
		ctx.Logger().Info("restaking protocol ibc channel not found")
		return
	}

	valUpdateID := k.GetValidatorSetUpdateID(ctx)

	var valUpdates []abci.ValidatorUpdate

	// The first packet should contains all validators
	if valUpdateID == 0 {
		// TODO maybe loop too mach ？
		vals := k.standaloneStakingKeeper.GetLastValidators(ctx)
		for _, v := range vals {
			validatorUpdate := v.ABCIValidatorUpdateZero()
			validatorUpdate.Power = k.standaloneStakingKeeper.GetLastValidatorPower(ctx, v.GetOperator())
			valUpdates = append(valUpdates, validatorUpdate)
		}
	} else {
		valUpdates = k.standaloneStakingKeeper.GetValidatorUpdates(ctx)
	}

	// TODO apply delegation/undelegate operation for valUpdates ?
	vsc := restaking.ValidatorSetChangePacket{
		ValidatorUpdates: valUpdates,
		ValsetUpdateId:   valUpdateID,
	}

	k.AppendPendingVSCPackets(ctx, vsc)

	valUpdateID++
	k.SetValidatorSetUpdateID(ctx, valUpdateID)
}

func (k Keeper) SendValidatorSetChangePackets(ctx sdk.Context) {
	channelID, err := k.GetCoordinatorChannelID(ctx)
	if err != nil {
		ctx.Logger().Info("restaking protocol ibc channel not found")
		return
	}

	pendingPackets := k.GetPendingVSCPackets(ctx)

	for _, packet := range pendingPackets {
		p := packet
		bz := k.cdc.MustMarshal(&p)
		// TODO Timeout should get from params of module.
		_, err := restaking.SendIBCPacket(ctx, k.scopedKeeper, k.channelKeeper, channelID, restaking.ConsumerPortID, bz, time.Minute*10)
		if err != nil {
			if clienttypes.ErrClientNotActive.Is(err) {
				// IBC client is expired!
				// leave the packet data stored to be sent once the client is upgraded
				// the client cannot expire during iteration (in the middle of a block)
				ctx.Logger().Debug("IBC client is expired, cannot send VSC, leaving packet data stored:")
				return
			}
			panic(fmt.Errorf("packet could not be sent over IBC: %w", err))
		}
	}

	k.DeletePendingVSCPackets(ctx)
}
