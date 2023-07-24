package keeper

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

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
	case restaking.CoordinatorPacket_WithdrawReward:
		var withdrawPacket restaking.WithdrawRewardPacket
		k.cdc.MustUnmarshal([]byte(coordinatorPacket.Data), &withdrawPacket)

		validatorAddr, err := sdk.ValAddressFromBech32(withdrawPacket.ValidatorAddress)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}
		coordinatorOperatorAccAddr, err := sdk.AccAddressFromBech32(withdrawPacket.OperatorAddress)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}
		operatorLocalAddress := k.GetOrCreateOperatorLocalAddress(ctx, packet.SourceChannel, packet.SourcePort, coordinatorOperatorAccAddr, validatorAddr)

		coin, err := k.multiStakingKeeper.WithdrawReward(ctx, validatorAddr, withdrawPacket.Denom, operatorLocalAddress)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}

		transferSeq, err := k.sendCoinToCoordinator(ctx, operatorLocalAddress, coordinatorOperatorAccAddr, coin, withdrawPacket.TransferChanel)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		} else {
			withdrawResp := restaking.ConsumerWithdrawRewardResponse{
				TransferDestChannel: withdrawPacket.TransferChanel,
				TransferDestPort:    ibctransfertypes.PortID,
				TransferDestSeq:     transferSeq,
				Balance:             coin,
			}
			withdrawRespBz := k.cdc.MustMarshal(&withdrawResp)
			return channeltypes.NewResultAcknowledgement(withdrawRespBz)
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
	delegationOperatorAccAddr, err := sdk.AccAddressFromBech32(delegation.OperatorAddress)
	if err != nil {
		return err
	}

	delegationValAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
	if err != nil {
		return err
	}

	operatorLocalAddress := k.GetOrCreateOperatorLocalAddress(ctx, packet.SourceChannel, packet.SourcePort, delegationOperatorAccAddr, delegationValAddr)
	k.SetOperatorLocalAddress(ctx, delegationOperatorAccAddr, delegationValAddr, operatorLocalAddress)

	// TODO how to delegate to validator
	// (1) adjust delegation of staking module ?
	// (2) mint coins and delegate by multistaking module
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{delegation.Balance}); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, operatorLocalAddress, sdk.Coins{delegation.Balance}); err != nil {
		return err
	}

	if _, err = k.multiStakingKeeper.MTStakingDelegate(ctx, operatorLocalAddress, delegationValAddr, delegation.Balance); err != nil {
		return err
	}
	return nil
}

func (k Keeper) HandleRestakingUndelegationPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	delegation *restaking.UndelegationPacket,
) error {
	delegationOperatorAccAddr, err := sdk.AccAddressFromBech32(delegation.OperatorAddress)
	if err != nil {
		return err
	}
	valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
	if err != nil {
		return err
	}

	operatorLocalAddress := k.GetOrCreateOperatorLocalAddress(ctx, packet.SourceChannel, packet.SourcePort, delegationOperatorAccAddr, valAddr)

	err = k.multiStakingKeeper.Unbond(ctx, operatorLocalAddress, valAddr, delegation.Balance)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) GenerateOperatorAccount(
	ctx sdk.Context, channel, portID string, operatorAccAddr sdk.AccAddress, valAddr sdk.ValAddress,
) authtypes.AccountI {
	header := ctx.BlockHeader()

	buf := []byte(types.ModuleName)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)
	buf = append(buf, []byte(channel)...)
	buf = append(buf, []byte(portID)...)
	buf = append(buf, []byte(operatorAccAddr)...)
	buf = append(buf, valAddr...)

	return authtypes.NewEmptyModuleAccount(string(buf), authtypes.Staking)
}

func (k Keeper) HandleRestakingSlashPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	slash *restaking.SlashPacket,
) error {
	slashOperatorAccAddr, err := sdk.AccAddressFromBech32(slash.OperatorAddress)
	if err != nil {
		return err
	}
	slashValAddr, err := sdk.ValAddressFromBech32(slash.ValidatorAddress)
	if err != nil {
		return err
	}

	operatorLocalAddress := k.GetOrCreateOperatorLocalAddress(ctx, packet.SourceChannel, packet.SourcePort, slashOperatorAccAddr, slashValAddr)

	err = k.multiStakingKeeper.InstantSlash(ctx, slashValAddr, operatorLocalAddress, slash.Balance)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) sendCoinToCoordinator(ctx sdk.Context, from, to sdk.AccAddress, balance sdk.Coin, channelID string) (uint64, error) {
	timeoutTimestamp := ctx.BlockTime().UnixNano() + 1800000000000
	msg := ibctransfertypes.MsgTransfer{
		SourcePort:       ibctransfertypes.PortID,
		SourceChannel:    channelID,
		Token:            balance,
		Sender:           from.String(),
		Receiver:         to.String(),
		TimeoutHeight:    clienttypes.Height{},
		TimeoutTimestamp: uint64(timeoutTimestamp),
		Memo:             "",
	}

	resp, err := k.ibcTransferKeeper.Transfer(ctx, &msg)
	if err != nil {
		return 0, err
	}

	return resp.Sequence, nil
}
