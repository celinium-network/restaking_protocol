package keeper

import (
	"fmt"
	"strings"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	multistakingtypes "github.com/celinium-network/restaking_protocol/x/multistaking/types"
	"github.com/celinium-network/restaking_protocol/x/restaking/consumer/types"
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

	valsetUpdateID := k.GetValidatorSetUpdateID(ctx)

	var valUpdates []abci.ValidatorUpdate

	if valsetUpdateID == 0 {
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
	vsc := restaking.ValidatorSetChange{
		ValidatorUpdates: valUpdates,
		ValsetUpdateId:   valsetUpdateID,
	}

	k.AppendPendingVSCPackets(ctx, vsc)

	valsetUpdateID++
	k.SetValidatorSetUpdateID(ctx, valsetUpdateID)
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
				ctx.Logger().Debug("IBC client is expired, cannot send VSC, leaving packet data stored:")
				return
			}
			panic(fmt.Errorf("packet could not be sent over IBC: %w", err))
		}
	}

	k.DeletePendingVSCPackets(ctx)
}

func (k Keeper) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, restakingPacket *restaking.RestakingPacket) exported.Acknowledgement {
	var ack exported.Acknowledgement

	switch restakingPacket.Type {
	case 0:
		var restakingDelegatePacket restaking.DelegationPacket
		k.cdc.MustUnmarshal([]byte(restakingPacket.Data), &restakingDelegatePacket)
		k.HandleRestakingDelegationPacket(ctx, packet, &restakingDelegatePacket)
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
	channelID, err := k.GetCoordinatorChannelID(ctx)
	if err != nil {
		return err
	}

	if strings.Compare(packet.SourceChannel, channelID) != 0 {
		return types.ErrUnknownPacketChannel
	}

	validatorPkBz := k.cdc.MustMarshal(&delegation.ValidatorPk)
	operatorLocalAddress := k.GetOrCreateOperatorLocalAddress(ctx, packet.SourceChannel, packet.SourcePort, delegation.OperatorAddress, validatorPkBz)

	k.SetOperatorLocalAddress(ctx, delegation.OperatorAddress, validatorPkBz, operatorLocalAddress)

	sdkVaPk, err := cryptocodec.FromTmProtoPublicKey(delegation.ValidatorPk)
	if err != nil {
		return err
	}

	consAddress := sdk.ConsAddress(sdkVaPk.Address())
	validator, found := k.standaloneStakingKeeper.GetValidatorByConsAddr(ctx, consAddress)
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
