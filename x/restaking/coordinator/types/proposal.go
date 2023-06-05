package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const ProposalTypeConsumerAddition = "ConsumerAddition"

var _ govtypes.Content = &ConsumerAdditionProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeConsumerAddition)
}

// GetDescription implements v1beta1.Content
func (cap *ConsumerAdditionProposal) GetDescription() string {
	return cap.Description
}

// GetTitle implements v1beta1.Content
func (cap *ConsumerAdditionProposal) GetTitle() string {
	return cap.Title
}

// ProposalRoute implements v1beta1.Content
func (*ConsumerAdditionProposal) ProposalRoute() string {
	return RouterKey
}

// ProposalType implements v1beta1.Content
func (*ConsumerAdditionProposal) ProposalType() string {
	return ProposalTypeConsumerAddition
}

// ValidateBasic implements v1beta1.Content
func (*ConsumerAdditionProposal) ValidateBasic() error {
	// TODO more validate
	return nil
}
