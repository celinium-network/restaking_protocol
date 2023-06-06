package integration

import (
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/cosmos/ibc-go/v7/testing/simapp"
)

func SendMsgs(chain *ibctesting.TestChain, msgs ...sdk.Msg) []abci.Event {
	chain.Coordinator.UpdateTimeForChain(chain)

	_, _, err := simapp.SignAndDeliver(
		chain.T,
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		[]uint64{chain.SenderAccount.GetAccountNumber()},
		[]uint64{chain.SenderAccount.GetSequence()},
		true, true, chain.SenderPrivKey,
	)
	if err != nil {
		panic(err)
	}

	// NextBlock calls app.Commit()
	events := NextBlockWithEvents(chain)

	// increment sequence for successful transaction execution
	err = chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		panic(err)
	}
	chain.Coordinator.IncrementTime()

	return events
}

func (s *IntegrationTestSuite) ChanOpenAck(endpoint *ibctesting.Endpoint) []abci.Event {
	err := endpoint.UpdateClient()
	require.NoError(endpoint.Chain.T, err)

	channelKey := host.ChannelKey(endpoint.Counterparty.ChannelConfig.PortID, endpoint.Counterparty.ChannelID)
	proof, height := endpoint.Counterparty.Chain.QueryProof(channelKey)

	msg := channeltypes.NewMsgChannelOpenAck(
		endpoint.ChannelConfig.PortID, endpoint.ChannelID,
		endpoint.Counterparty.ChannelID, endpoint.Counterparty.ChannelConfig.Version, // testing doesn't use flexible selection
		proof, height,
		endpoint.Chain.SenderAccount.GetAddress().String(),
	)

	events := SendMsgs(endpoint.Chain, msg)

	endpoint.ChannelConfig.Version = endpoint.GetChannel().Version

	return events
}

func NextBlockWithEvents(chain *ibctesting.TestChain) []abci.Event {
	res := chain.App.EndBlock(abci.RequestEndBlock{Height: chain.CurrentHeader.Height})

	chain.App.Commit()

	// set the last header to the current header
	// use nil trusted fields
	chain.LastHeader = chain.CurrentTMClientHeader()

	// val set changes returned from previous block get applied to the nweext validators
	// of this block. See tendermint spec for details.
	chain.Vals = chain.NextVals
	chain.NextVals = ibctesting.ApplyValSetChanges(chain.T, chain.Vals, res.ValidatorUpdates)

	// increment the current header
	chain.CurrentHeader = tmproto.Header{
		ChainID: chain.ChainID,
		Height:  chain.App.LastBlockHeight() + 1,
		AppHash: chain.App.LastCommitID().Hash,
		// NOTE: the time is increased by the coordinator to maintain time synchrony amongst
		// chains.
		Time:               chain.CurrentHeader.Time,
		ValidatorsHash:     chain.Vals.Hash(),
		NextValidatorsHash: chain.NextVals.Hash(),
		ProposerAddress:    chain.CurrentHeader.ProposerAddress,
	}

	chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})
	return res.Events
}

func parseMsgRecvPacketFromEvents(fromChain *ibctesting.TestChain, events []abci.Event, sender string) []channeltypes.MsgRecvPacket {
	var msgRecvPackets []channeltypes.MsgRecvPacket
	for _, ev := range events {
		sdkevents := sdk.Events{sdk.Event{
			Type:       ev.Type,
			Attributes: ev.Attributes,
		}}

		packet, err := ibctesting.ParsePacketFromEvents(sdkevents)
		if err != nil {
			continue
		}

		commitKey := host.PacketCommitmentKey(packet.SourcePort, packet.SourceChannel, packet.Sequence)
		proof, height := fromChain.QueryProof(commitKey)

		backProofType := ibccommitmenttypes.MerkleProof{}
		backProofType.Unmarshal(proof)

		msgRecvPacket := channeltypes.MsgRecvPacket{
			Packet:          packet,
			ProofCommitment: proof,
			ProofHeight:     height,
			Signer:          sender,
		}

		msgRecvPackets = append(msgRecvPackets, msgRecvPacket)
	}

	return msgRecvPackets
}

// TODO Too complex, need to simplify the logic
func (s *IntegrationTestSuite) RelayIBCPacket(path *ibctesting.Path, events []abci.Event, sender string) {
	sendChain := path.EndpointA.Chain
	recvChain := path.EndpointB.Chain

	msgRecvPackets := parseMsgRecvPacketFromEvents(sendChain, events, sender)

	channelPackets := make(map[string][]channeltypes.MsgRecvPacket)
	channelProcessed := make(map[string]int)

	for _, packet := range msgRecvPackets {
		ps, ok := channelPackets[packet.Packet.SourceChannel]
		if !ok {
			ps = make([]channeltypes.MsgRecvPacket, 0)
		}
		ps = append(ps, packet)

		channelPackets[packet.Packet.SourceChannel] = ps
		channelProcessed[packet.Packet.SourceChannel] = 0
	}

	for {
		for channelID, packets := range channelPackets {
			offset := channelProcessed[channelID]
			for ; offset < len(packets); offset++ {

				midAckMsg, _ := chainRecvPacket(recvChain, path.EndpointB, &packets[offset])
				if midAckMsg == nil {
					continue
				}

				midRecvMsg, _ := chainRecvAck(sendChain, path.EndpointA, midAckMsg)
				if midRecvMsg == nil {
					continue
				}
				if midRecvMsg.Packet.SourceChannel == channelID {
					packets = append(packets, *midRecvMsg)
				} else {
					ps, ok := channelPackets[midRecvMsg.Packet.SourceChannel]
					if !ok {
						ps = make([]channeltypes.MsgRecvPacket, 0)
					}
					ps = append(ps, *midRecvMsg)
					channelPackets[midRecvMsg.Packet.SourceChannel] = ps
				}
			}
			channelProcessed[channelID] = offset
		}

		done := true
		for channelID, packets := range channelPackets {
			if channelProcessed[channelID] != len(packets) {
				done = false
			}
		}
		if done {
			break
		}
	}
}

func chainRecvPacket(chain *ibctesting.TestChain, endpoint *ibctesting.Endpoint, msgRecvPacket *channeltypes.MsgRecvPacket) (*channeltypes.MsgAcknowledgement, error) {
	ctx := chain.GetContext()
	if _, err := chain.App.GetIBCKeeper().RecvPacket(ctx, msgRecvPacket); err != nil {
		return nil, err
	}

	// chain.NextBlock()
	// endpoint.UpdateClient()

	return assembleAckPacketFromEvents(chain, msgRecvPacket.Packet, ctx.EventManager().Events())
}

func chainRecvAck(chain *ibctesting.TestChain, endpoint *ibctesting.Endpoint, ack *channeltypes.MsgAcknowledgement) (*channeltypes.MsgRecvPacket, error) {
	ctx := chain.GetContext()
	if _, err := chain.App.GetIBCKeeper().Acknowledgement(ctx, ack); err != nil {
		return nil, err
	}

	chain.NextBlock()
	endpoint.UpdateClient()

	return assembleRecvPacketByEvents(chain, ctx.EventManager().Events())
}

func assembleRecvPacketByEvents(chain *ibctesting.TestChain, events sdk.Events) (*channeltypes.MsgRecvPacket, error) {
	packet, err := ibctesting.ParsePacketFromEvents(events)
	if err != nil {
		return nil, err
	}

	commitKey := host.PacketCommitmentKey(packet.SourcePort, packet.SourceChannel, packet.Sequence)
	proof, height := chain.QueryProof(commitKey)

	backProofType := ibccommitmenttypes.MerkleProof{}
	backProofType.Unmarshal(proof)

	msgRecvPacket := channeltypes.MsgRecvPacket{
		Packet:          packet,
		ProofCommitment: proof,
		ProofHeight:     height,
		Signer:          chain.SenderAccount.GetAddress().String(),
	}

	return &msgRecvPacket, nil
}

func assembleAckPacketFromEvents(chain *ibctesting.TestChain, packet channeltypes.Packet, events sdk.Events) (*channeltypes.MsgAcknowledgement, error) {
	ack, err := ibctesting.ParseAckFromEvents(events)
	if err != nil {
		return nil, err
	}
	key := host.PacketAcknowledgementKey(packet.GetDestPort(),
		packet.GetDestChannel(),
		packet.GetSequence())

	proof, height := chain.QueryProof(key)

	backProofType := ibccommitmenttypes.MerkleProof{}
	backProofType.Unmarshal(proof)

	ackMsg := channeltypes.MsgAcknowledgement{
		Packet:          packet,
		Acknowledgement: ack,
		ProofAcked:      proof,
		ProofHeight:     height,
		Signer:          chain.SenderAccount.GetAddress().String(),
	}

	return &ackMsg, nil
}
