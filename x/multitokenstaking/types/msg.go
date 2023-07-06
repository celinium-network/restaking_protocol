package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgMTStakingDelegate{}
	_ sdk.Msg = &MsgMTStakingUndelegate{}
	_ sdk.Msg = &MsgMTStakingWithdrawReward{}
)

func NewMsgMTStakingDelegate(delegator sdk.AccAddress, validator sdk.ValAddress, balance sdk.Coin) (*MsgMTStakingDelegate, error) {
	msg := MsgMTStakingDelegate{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
		Balance:          balance,
	}
	return &msg, nil
}

// GetSigners implements sdk.Msg.
func (msg MsgMTStakingDelegate) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

// ValidateBasic implements sdk.Msg.
func (msg MsgMTStakingDelegate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.Wrapf(errors.ErrInvalidAddress, "invalid delegator address: %s", err)
	}

	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.Wrapf(errors.ErrInvalidAddress, "invalid validator address: %s", err)
	}

	if !msg.Balance.IsValid() || !msg.Balance.Amount.IsPositive() {
		return sdkerrors.Wrap(errors.ErrInvalidRequest, "invalid delegation amount")
	}

	return nil
}

// GetSigners implements sdk.Msg.
func (msg MsgMTStakingUndelegate) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

// ValidateBasic implements sdk.Msg.
func (msg MsgMTStakingUndelegate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.Wrapf(errors.ErrInvalidAddress, "invalid delegator address: %s", err)
	}

	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.Wrapf(errors.ErrInvalidAddress, "invalid validator address: %s", err)
	}

	if !msg.Balance.IsValid() || !msg.Balance.Amount.IsPositive() {
		return sdkerrors.Wrap(errors.ErrInvalidRequest, "invalid delegation amount")
	}

	return nil
}

// GetSigners implements sdk.Msg.
func (msg MsgMTStakingWithdrawReward) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

// ValidateBasic implements sdk.Msg.
func (msg MsgMTStakingWithdrawReward) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.Wrapf(errors.ErrInvalidAddress, "invalid delegator address: %s", err)
	}

	if _, err := sdk.AccAddressFromBech32(msg.AgentAddress); err != nil {
		return sdkerrors.Wrapf(errors.ErrInvalidAddress, "invalid validator address: %s", err)
	}

	return nil
}
