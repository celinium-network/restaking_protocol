package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/utils"
)

var (
	_ sdk.Msg = &MsgRegisterOperatorRequest{}
	_ sdk.Msg = &MsgDelegateRequest{}
	_ sdk.Msg = &MsgUndelegateRequest{}
	_ sdk.Msg = &MsgWithdrawRewardRequest{}
)

// GetSigners implements types.Msg.
func (msg *MsgRegisterOperatorRequest) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements types.Msg.
func (msg *MsgRegisterOperatorRequest) ValidateBasic() error {
	if len(msg.ConsumerChainIDs) != len(msg.ConsumerValidatorAddresses) {
		return errorsmod.Wrapf(ErrParams, "mismatch length %s, %s", msg.ConsumerChainIDs, msg.ConsumerValidatorAddresses)
	}

	if utils.SliceHasRepeatedElement(msg.ConsumerChainIDs) {
		return errorsmod.Wrapf(ErrParams, "slice has repeated element %s", msg.ConsumerChainIDs)
	}

	if utils.SliceHasRepeatedElement(msg.ConsumerValidatorAddresses) {
		return errorsmod.Wrapf(ErrParams, "slice has repeated element %s", msg.ConsumerValidatorAddresses)
	}

	return nil
}

// GetSigners implements types.Msg.
func (msg *MsgDelegateRequest) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements types.Msg.
func (msg *MsgDelegateRequest) ValidateBasic() error {
	if msg.Amount.IsZero() {
		return errorsmod.Wrapf(ErrParams, "amount should't be zero")
	}

	return nil
}

// GetSigners implements types.Msg.
func (msg *MsgUndelegateRequest) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements types.Msg.
func (msg *MsgUndelegateRequest) ValidateBasic() error {
	if msg.Amount.IsZero() {
		return errorsmod.Wrapf(ErrParams, "amount should't be zero")
	}

	return nil
}

// GetSigners implements types.Msg.
func (msg *MsgWithdrawRewardRequest) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements types.Msg.
func (*MsgWithdrawRewardRequest) ValidateBasic() error {
	return nil
}
