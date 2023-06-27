package integration

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibchost "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	rsconsumer "github.com/celinium-network/restaking_protocol/app/consumer"
	rscoordinator "github.com/celinium-network/restaking_protocol/app/coordinator"
	"github.com/celinium-network/restaking_protocol/app/params"
)

func (s *IntegrationTestSuite) TestIBCTransfer() {
	s.coordinator.Setup(s.transferPath)

	consumerAccAddr := s.rsConsumerChain.SenderAccount.GetAddress()
	coordAccAddr := s.rsCoordinatorChain.SenderAccount.GetAddress()

	coin := sdk.NewCoin(params.DefaultBondDenom, sdk.NewIntFromUint64(1000000))
	err := mintCoin(s.rsConsumerChain, consumerAccAddr, coin)
	s.Require().NoError(err)

	s.IBCTransfer(consumerAccAddr.String(), coordAccAddr.String(), coin, s.transferPath, true)

	coordApp := getCoordinatorApp(s.rsCoordinatorChain)
	ibcDenom := calculateIBCDenom(ibctesting.TransferPort, s.transferPath.EndpointA.ChannelID, coin.Denom)

	balance := coordApp.BankKeeper.GetBalance(s.rsCoordinatorChain.GetContext(), coordAccAddr, ibcDenom)

	s.Require().True(balance.Amount.Equal(coin.Amount))
}

func mintCoin(
	chain *ibctesting.TestChain,
	to sdk.AccAddress,
	coin sdk.Coin,
) error {
	var bankkeeper bankkeeper.Keeper

	if app, ok := chain.App.(*rscoordinator.App); ok {
		bankkeeper = app.BankKeeper
	} else if app, ok := chain.App.(*rsconsumer.App); ok {
		bankkeeper = app.BankKeeper
	} else {
		return errors.New("known chain")
	}

	ctx := chain.GetContext()
	if err := bankkeeper.MintCoins(ctx, ibctransfertypes.ModuleName, sdk.Coins{coin}); err != nil {
		return err
	}

	return bankkeeper.SendCoinsFromModuleToAccount(ctx, ibctransfertypes.ModuleName, to, sdk.Coins{coin})
}

func (s *IntegrationTestSuite) IBCTransfer(
	from string,
	to string,
	coin sdk.Coin,
	transferPath *ibctesting.Path,
	transferForward bool,
) {
	srcEndpoint := transferPath.EndpointA
	destEndpoint := transferPath.EndpointB
	if !transferForward {
		srcEndpoint, destEndpoint = destEndpoint, srcEndpoint
	}

	destChainApp := destEndpoint.Chain.App

	timeout := srcEndpoint.Chain.CurrentHeader.Time.Add(time.Hour * 5).UnixNano()
	msg := ibctransfertypes.NewMsgTransfer(
		srcEndpoint.ChannelConfig.PortID,
		srcEndpoint.ChannelID,
		coin,
		from,
		to,
		ibcclienttypes.Height{},
		uint64(timeout),
		"",
	)

	res, err := srcEndpoint.Chain.SendMsgs(msg)
	s.NoError(err)

	s.Require().NoError(err)

	err = s.transferPath.EndpointB.UpdateClient()
	s.Require().NoError(err)

	for _, ev := range res.Events {
		events := sdk.Events{sdk.Event{
			Type:       ev.Type,
			Attributes: ev.Attributes,
		}}

		packet, err := ibctesting.ParsePacketFromEvents(events)
		if err != nil {
			continue
		}

		commitKey := ibchost.PacketCommitmentKey(packet.SourcePort, packet.SourceChannel, packet.Sequence)
		proof, height := srcEndpoint.Chain.QueryProof(commitKey)

		msgRecvPacket := channeltypes.MsgRecvPacket{
			Packet:          packet,
			ProofCommitment: proof,
			ProofHeight:     height,
			Signer:          from,
		}

		_, err = destChainApp.GetIBCKeeper().RecvPacket(destEndpoint.Chain.GetContext(), &msgRecvPacket)
		s.NoError(err)
	}
}

func calculateIBCDenom(portID, channelID string, denom string) string {
	sourcePrefix := ibctransfertypes.GetDenomPrefix(portID, channelID)
	denomTrace := ibctransfertypes.ParseDenomTrace(sourcePrefix + denom)

	return denomTrace.IBCDenom()
}
