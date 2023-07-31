package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const ProposalTypeMultiTokenStakingAddition = "MultiTokenStakingAddition"

var _ govtypes.Content = &AddNonNativeTokenToStakingProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeMultiTokenStakingAddition)
}

// GetDescription implements v1beta1.Content
func (cap *AddNonNativeTokenToStakingProposal) GetDescription() string {
	return cap.Description
}

// GetTitle implements v1beta1.Content
func (cap *AddNonNativeTokenToStakingProposal) GetTitle() string {
	return cap.Title
}

// ProposalRoute implements v1beta1.Content
func (*AddNonNativeTokenToStakingProposal) ProposalRoute() string {
	return RouterKey
}

// ProposalType implements v1beta1.Content
func (*AddNonNativeTokenToStakingProposal) ProposalType() string {
	return ProposalTypeMultiTokenStakingAddition
}

// ValidateBasic implements v1beta1.Content
func (*AddNonNativeTokenToStakingProposal) ValidateBasic() error {
	// TODO more validate
	return nil
}
