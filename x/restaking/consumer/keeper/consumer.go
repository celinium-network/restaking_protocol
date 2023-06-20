package keeper

import (
	"fmt"
	"strings"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/proto/tendermint/crypto"

	errorsmod "cosmossdk.io/errors"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
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

	initialUpdates := []abci.ValidatorUpdate{}
	for _, p := range lastPowers {
		addr, err := sdk.ValAddressFromBech32(p.Address)
		if err != nil {
			return err
		}

		val, found := k.stakingKeeper.GetValidator(ctx, addr)
		if !found {
			return errorsmod.Wrapf(stakingtypes.ErrNoValidatorFound, "error getting validator from LastValidatorPowers: %s", err)
		}

		tmProtoPk, err := val.TmConsPublicKey()
		if err != nil {
			return err
		}

		initialUpdates = append(initialUpdates, abci.ValidatorUpdate{
			PubKey: tmProtoPk,
			Power:  p.Power,
		})
	}

	vsc.ValidatorUpdates = initialUpdates
	vsc.Type = restaking.ValidatorSetChange_Add

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

	pendingConsumerPacketData := restaking.ConsumerPacketData{
		ValidatorSetChanges: pendingVSCList,
		SlashPacketData:     pendingSlashList,
	}

	bz := k.cdc.MustMarshal(&pendingConsumerPacketData)

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

func (k Keeper) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, restakingPacket *restaking.RestakingPacket) exported.Acknowledgement {
	var ack exported.Acknowledgement

	channelID, err := k.GetCoordinatorChannelID(ctx)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if strings.Compare(packet.SourceChannel, channelID) != 0 {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	switch restakingPacket.Type {
	case restaking.RestakingPacket_Delegation:
		var delegatePacket restaking.DelegationPacket
		k.cdc.MustUnmarshal([]byte(restakingPacket.Data), &delegatePacket)

		err := k.HandleRestakingDelegationPacket(ctx, packet, &delegatePacket)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		} else {
			return channeltypes.NewResultAcknowledgement([]byte{1})
		}
	case restaking.RestakingPacket_Undelegation:
		var undelegatePacket restaking.UndelegationPacket
		k.cdc.MustUnmarshal([]byte(restakingPacket.Data), &undelegatePacket)

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
	validatorPkBz := k.cdc.MustMarshal(&delegation.ValidatorPk)
	operatorLocalAddress := k.GetOrCreateOperatorLocalAddress(ctx, packet.SourceChannel, packet.SourcePort, delegation.OperatorAddress, validatorPkBz)

	k.SetOperatorLocalAddress(ctx, delegation.OperatorAddress, validatorPkBz, operatorLocalAddress)

	validator, found := k.getValidatorFromTmPublicKey(ctx, delegation.ValidatorPk)
	if !found {
		return types.ErrUnknownValidator
	}

	// TODO how to delegate to validator
	// (1) adjust delegation of staking module ?
	// (2) mint coins and delegate by multistaking module
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{delegation.Amount}); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, operatorLocalAddress, sdk.Coins{delegation.Amount}); err != nil {
		return err
	}

	return k.multiStakingKeeper.MultiStakingDelegate(ctx, multistakingtypes.MsgMultiStakingDelegate{
		DelegatorAddress: operatorLocalAddress.String(),
		ValidatorAddress: validator.OperatorAddress,
		Amount:           delegation.Amount,
	})
}

func (k Keeper) HandleRestakingUndelegationPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	delegation *restaking.UndelegationPacket,
) error {
	validatorPkBz := k.cdc.MustMarshal(&delegation.ValidatorPk)
	operatorLocalAddress := k.GetOrCreateOperatorLocalAddress(ctx, packet.SourceChannel, packet.SourcePort, delegation.OperatorAddress, validatorPkBz)

	validator, found := k.getValidatorFromTmPublicKey(ctx, delegation.ValidatorPk)
	if !found {
		return types.ErrUnknownValidator
	}

	valAddress, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	if err != nil {
		return err
	}

	_, err = k.multiStakingKeeper.Unbond(ctx, operatorLocalAddress, valAddress, delegation.Amount)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) getValidatorFromTmPublicKey(ctx sdk.Context, tmpk crypto.PublicKey) (stakingtypes.Validator, bool) {
	sdkVaPk, err := cryptocodec.FromTmProtoPublicKey(tmpk)
	if err != nil {
		return stakingtypes.Validator{}, false
	}

	consAddress := sdk.ConsAddress(sdkVaPk.Address())
	return k.stakingKeeper.GetValidatorByConsAddr(ctx, consAddress)
}

func (k Keeper) GenerateOperatorAccount(
	ctx sdk.Context,
	channel, portID, operatorAddress string,
	validatorPk []byte,
) authtypes.AccountI {
	header := ctx.BlockHeader()

	buf := []byte(types.ModuleName)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)
	buf = append(buf, []byte(channel)...)
	buf = append(buf, []byte(portID)...)
	buf = append(buf, []byte(operatorAddress)...)
	buf = append(buf, validatorPk...)

	return authtypes.NewEmptyModuleAccount(string(buf), authtypes.Staking)
}
