package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func BuildRestakingProtocolPacket(cdc codec.Codec, packet interface{}) (*RestakingPacket, error) {
	var (
		bz         []byte
		packetType RestakingPacket_PacketType
		err        error
	)

	switch value := packet.(type) {
	case DelegationPacket:
		packetType = 0
		bz, err = cdc.Marshal(&value)
	default:
		err = ErrBuildRestakingPacketFailed
	}

	if err != nil {
		return nil, err
	}

	return &RestakingPacket{
		Type: packetType,
		Data: string(bz),
	}, nil
}
