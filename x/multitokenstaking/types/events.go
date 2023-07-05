package types

// staking module event types
const (
	EventTypeCompleteUnbonding         = "complete_unbonding"
	EventTypeDelegate                  = "delegate"
	EventTypeUnbond                    = "unbond"
	EventTypeCancelUnbondingDelegation = "cancel_unbonding_delegation"

	AttributeKeyValidator         = "validator"
	AttributeKeyCommissionRate    = "commission_rate"
	AttributeKeyMinSelfDelegation = "min_self_delegation"
	AttributeKeySrcValidator      = "source_validator"
	AttributeKeyDstValidator      = "destination_validator"
	AttributeKeyDelegator         = "delegator"
	AttributeKeyCreationHeight    = "creation_height"
	AttributeKeyCompletionTime    = "completion_time"
	AttributeKeyNewShares         = "new_shares"
)
