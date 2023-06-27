package keeper

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	multistakingtypes "github.com/celinium-network/restaking_protocol/x/multistaking/types"
	"github.com/celinium-network/restaking_protocol/x/restaking/consumer/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) {
	k.SendConsumerPendingPacket(ctx)
}

func (k Keeper) QueueInitialVSC(ctx sdk.Context) error {
	var vsc restaking.ValidatorSetChange

	var lastPowers []stakingtypes.LastValidatorPower
	k.stakingKeeper.IterateLastValidatorPowers(ctx, func(addr sdk.ValAddress, power int64) (stop bool) {
		lastPowers = append(lastPowers, stakingtypes.LastValidatorPower{Address: addr.String(), Power: power})
		return false
	})

	var initialValidators []string
	for _, p := range lastPowers {
		initialValidators = append(initialValidators, p.Address)
	}

	vsc.ValidatorAddresses = initialValidators
	vsc.Type = restaking.ValidatorSetChange_ADD

	k.AppendPendingVSC(ctx, vsc)

	return nil
}

func (k Keeper) QueueVSC(ctx sdk.Context, vsc restaking.ValidatorSetChange) {
	k.AppendPendingVSC(ctx, vsc)
}

func (k Keeper) SendConsumerPendingPacket(ctx sdk.Context) {
	channelID, err := k.GetCoordinatorChannelID(ctx)
	if err != nil {
		ctx.Logger().Info("restaking protocol ibc channel not found")
		return
	}

	pendingVSCList := k.GetPendingVSCList(ctx)
	pendingSlashList := k.GetPendingConsumerSlashList(ctx)
	if len(pendingVSCList) == 0 && len(pendingSlashList) == 0 {
		return
	}

	pendingConsumerPacket := restaking.ConsumerPacket{
		ValidatorSetChanges: pendingVSCList,
		ConsumerSlashList:   pendingSlashList,
	}

	bz := k.cdc.MustMarshal(&pendingConsumerPacket)

	if _, err := restaking.SendIBCPacket(
		ctx,
		k.scopedKeeper,
		k.channelKeeper,
		channelID,
		restaking.ConsumerPortID,
		bz,
		time.Minute*10,
	); err != nil {
		if clienttypes.ErrClientNotActive.Is(err) {
			ctx.Logger().Debug("IBC client is expired, cannot send VSC, leaving packet data stored:")
			return
		}
		// TODO panic or return ?
		panic(fmt.Errorf("packet could not be sent over IBC: %w", err))
	}

	k.DeletePendingConsumerSlashList(ctx)
	k.DeletePendingVSCList(ctx)
}

func (k Keeper) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, coordinatorPacket *restaking.CoordinatorPacket) exported.Acknowledgement {
	var ack exported.Acknowledgement

	channelID, err := k.GetCoordinatorChannelID(ctx)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if strings.Compare(packet.SourceChannel, channelID) != 0 {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	switch coordinatorPacket.Type {
	case restaking.CoordinatorPacket_Delegation:
		var delegatePacket restaking.DelegationPacket
		k.cdc.MustUnmarshal([]byte(coordinatorPacket.Data), &delegatePacket)

		err := k.HandleRestakingDelegationPacket(ctx, packet, &delegatePacket)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		} else {
			return channeltypes.NewResultAcknowledgement([]byte{1})
		}
	case restaking.CoordinatorPacket_Undelegation:
		var undelegatePacket restaking.UndelegationPacket
		k.cdc.MustUnmarshal([]byte(coordinatorPacket.Data), &undelegatePacket)

		err := k.HandleRestakingUndelegationPacket(ctx, packet, &undelegatePacket)
		if err != nil {
			ack = channeltypes.NewErrorAcknowledgement(err)
		} else {
			unbondingTime := k.stakingKeeper.GetParams(ctx).UnbondingTime
			resp := restaking.ConsumerUndelegateResponse{
				CompletionTime: ctx.BlockTime().Add(unbondingTime).UnixNano(),
			}
			respBz := k.cdc.MustMarshal(&resp)
			ack = channeltypes.NewResultAcknowledgement(respBz)
		}
	case restaking.CoordinatorPacket_Slash:
		var slashPacket restaking.SlashPacket
		k.cdc.MustUnmarshal([]byte(coordinatorPacket.Data), &slashPacket)

		err := k.HandleRestakingSlashPacket(ctx, packet, &slashPacket)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		} else {
			return channeltypes.NewResultAcknowledgement([]byte{1})
		}
	default:
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("unknown restaking protocol packet type"))
	}

	return ack
}

func (k Keeper) HandleRestakingDelegationPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	delegation *restaking.DelegationPacket,
) error {
	operatorLocalAddress := k.GetOrCreateOperatorLocalAddress(ctx, packet.SourceChannel, packet.SourcePort, delegation.OperatorAddress, delegation.ValidatorAddress)

	k.SetOperatorLocalAddress(ctx, delegation.OperatorAddress, delegation.ValidatorAddress, operatorLocalAddress)

	validator, found := k.getValidator(ctx, delegation.ValidatorAddress)
	if !found {
		return types.ErrUnknownValidator
	}

	// TODO how to delegate to validator
	// (1) adjust delegation of staking module ?
	// (2) mint coins and delegate by multistaking module
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{delegation.Balance}); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, operatorLocalAddress, sdk.Coins{delegation.Balance}); err != nil {
		return err
	}

	return k.multiStakingKeeper.MultiStakingDelegate(ctx, multistakingtypes.MsgMultiStakingDelegate{
		DelegatorAddress: operatorLocalAddress.String(),
		ValidatorAddress: validator.OperatorAddress,
		Amount:           delegation.Balance,
	})
}

func (k Keeper) HandleRestakingUndelegationPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	delegation *restaking.UndelegationPacket,
) error {
	operatorLocalAddress := k.GetOrCreateOperatorLocalAddress(ctx, packet.SourceChannel, packet.SourcePort, delegation.OperatorAddress, delegation.ValidatorAddress)

	validator, found := k.getValidator(ctx, delegation.ValidatorAddress)
	if !found {
		return types.ErrUnknownValidator
	}

	valAddress, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	if err != nil {
		return err
	}

	_, err = k.multiStakingKeeper.Unbond(ctx, operatorLocalAddress, valAddress, delegation.Balance)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) getValidator(ctx sdk.Context, valAddr string) (stakingtypes.Validator, bool) {
	accAddr, err := sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return stakingtypes.Validator{}, false
	}
	return k.stakingKeeper.GetValidator(ctx, accAddr)
}

func (k Keeper) GenerateOperatorAccount(
	ctx sdk.Context,
	channel, portID, operatorAddress string,
	valAddr string,
) authtypes.AccountI {
	header := ctx.BlockHeader()

	buf := []byte(types.ModuleName)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)
	buf = append(buf, []byte(channel)...)
	buf = append(buf, []byte(portID)...)
	buf = append(buf, []byte(operatorAddress)...)
	buf = append(buf, valAddr...)

	return authtypes.NewEmptyModuleAccount(string(buf), authtypes.Staking)
}

func (k Keeper) HandleRestakingSlashPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	slash *restaking.SlashPacket,
) error {
	operatorLocalAddress := k.GetOrCreateOperatorLocalAddress(ctx, packet.SourceChannel, packet.SourcePort, slash.OperatorAddress, slash.ValidatorAddress)

	validator, found := k.getValidator(ctx, slash.ValidatorAddress)
	if !found {
		return types.ErrUnknownValidator
	}

	valAddress, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	if err != nil {
		return err
	}

	err = k.multiStakingKeeper.SlashDelegator(ctx, valAddress, operatorLocalAddress, slash.Balance)
	if err != nil {
		return err
	}

	return nil
}
