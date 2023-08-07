package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterOperatorRequest{},
		&MsgDelegateRequest{},
		&MsgUndelegateRequest{},
		&MsgWithdrawRewardRequest{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&ConsumerAdditionProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
