package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (k Keeper) OnRecvConsumerPacketData(
	ctx sdk.Context,
	packet channeltypes.Packet,
	consumerPacket restaking.ConsumerPacket,
) ibcexported.Acknowledgement {
	consumerClientID, err := k.GetConsumerClientIDByChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
	if err != nil {
		ctx.Logger().Error("Coordinator can't get consumer clientID at receive VSC of consumer")
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// TODO more clearer ack
	if err := k.OnRecvConsumerSlash(ctx, consumerClientID, consumerPacket.ConsumerSlashList); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}
	return k.OnRecvConsumerVSC(ctx, consumerClientID, consumerPacket.ValidatorSetChanges)
}

func (k Keeper) OnRecvConsumerVSC(
	ctx sdk.Context,
	consumerClientID string,
	changeList []restaking.ValidatorSetChange,
) ibcexported.Acknowledgement {
	for _, change := range changeList {
		if change.Type == restaking.ValidatorSetChange_ADD {
			for _, addr := range change.ValidatorAddresses {
				k.SetConsumerValidator(ctx, consumerClientID, types.ConsumerValidator{
					Address: addr,
				})
			}
		} else if change.Type == restaking.ValidatorSetChange_REMOVE {
			for _, addr := range change.ValidatorAddresses {
				k.DeleteConsumerValidator(ctx, consumerClientID, addr)
			}
		}
	}

	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})
	return ack
}

func (k Keeper) OnRecvConsumerSlash(
	ctx sdk.Context,
	consumerClientID string,
	slashList []restaking.ConsumerSlash,
) error {
	// TODO how process error in this loop? panic or not?
	for _, slash := range slashList {
		operatorAccAddr := sdk.MustAccAddressFromBech32(slash.OperatorAddress)
		operator, found := k.GetOperator(ctx, operatorAccAddr)
		if !found {
			ctx.Logger().Error("slashing operator no exist", slash.OperatorAddress)
			continue
		}

		slashAmt := slash.SlashFactor.MulInt(operator.RestakedAmount).TruncateInt()
		operator.RestakedAmount = operator.RestakedAmount.Sub(slashAmt)

		k.SetOperator(ctx, operatorAccAddr, operator)

		slashCoin := sdk.NewCoin(operator.RestakingDenom, slashAmt)
		err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, operatorAccAddr, types.ModuleName, sdk.NewCoins(slashCoin))
		if err != nil {
			ctx.Logger().Error(
				fmt.Sprintf("prepare slash failed, send coin from operator to coordinator module failed, operator %s, coins: %s",
					slash.OperatorAddress, slashCoin))
			continue
		}
		err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(slashCoin))
		if err != nil {
			ctx.Logger().Error(
				fmt.Sprintf("slash failed, burn coin from operator to coordinator module failed, operator %s, coins: %s",
					slash.OperatorAddress, slashCoin))
			continue
		}

		// notify other consumer chain to slash
		var opValidators []types.OperatedValidator
		for _, ov := range operator.OperatedValidators {
			if ov.ClientID == consumerClientID {
				continue
			}
			opValidators = append(opValidators, ov)
		}
		k.NotifyConsumerSlashOperator(ctx, slashCoin, operator.OperatorAddress, opValidators)
	}

	return nil
}

func (k Keeper) NotifyConsumerSlashOperator(
	ctx sdk.Context,
	slashCoin sdk.Coin,
	operatorAddr string,
	slashValidator []types.OperatedValidator,
) {
	// TODO queue it and send packet at endblock?
	for _, va := range slashValidator {

		channel, found := k.GetConsumerClientIDToChannel(ctx, va.ClientID)
		if !found {
			ctx.Logger().Error(fmt.Sprintf(
				"the consumer chain of operator has't IBC Channel, chainID: %s, operator address: %s",
				va.ChainID, operatorAddr))
			continue
		}

		// TODO correct TIMEOUT
		timeout := time.Minute * 10

		slashPacket := restaking.SlashPacket{
			OperatorAddress:  operatorAddr,
			ValidatorAddress: va.ValidatorAddress,
			Balance:          slashCoin,
		}

		bz := k.cdc.MustMarshal(&slashPacket)
		restakingPacket := restaking.CoordinatorPacket{
			Type: restaking.CoordinatorPacket_Slash,
			Data: string(bz),
		}

		restakingProtocolPacketBz, err := k.cdc.Marshal(&restakingPacket)
		if err != nil {
			ctx.Logger().Error("marshal restaking.Delegation has err: ", err)
			// TODO continue ?
			continue
		}
		_, err = restaking.SendIBCPacket(
			ctx,
			k.scopedKeeper,
			k.channelKeeper,
			channel,
			restaking.CoordinatorPortID,
			restakingProtocolPacketBz,
			timeout,
		)
		if err != nil {
			ctx.Logger().Error("send ibc packet has error:", err)
		}
	}
}
