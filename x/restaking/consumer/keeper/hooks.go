package keeper

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Hooks are utilized to monitor and capture the events of validator
// addition and removal on the consumer chain.
type Hooks struct {
	k *Keeper
}

var _ stakingtypes.StakingHooks = Hooks{}

// Returns new provider hooks
func (k *Keeper) Hooks() Hooks {
	return Hooks{k}
}

// AfterDelegationModified implements types.StakingHooks.
func (Hooks) AfterDelegationModified(ctx types.Context, delAddr types.AccAddress, valAddr types.ValAddress) error {
	return nil
}

// AfterUnbondingInitiated implements types.StakingHooks.
func (Hooks) AfterUnbondingInitiated(ctx types.Context, id uint64) error {
	return nil
}

// AfterValidatorBeginUnbonding implements types.StakingHooks.
func (Hooks) AfterValidatorBeginUnbonding(ctx types.Context, consAddr types.ConsAddress, valAddr types.ValAddress) error {
	return nil
}

// AfterValidatorBonded implements types.StakingHooks.
func (Hooks) AfterValidatorBonded(ctx types.Context, consAddr types.ConsAddress, valAddr types.ValAddress) error {
	return nil
}

// AfterValidatorCreated implements types.StakingHooks.
func (Hooks) AfterValidatorCreated(ctx types.Context, valAddr types.ValAddress) error {
	return nil
}

// AfterValidatorRemoved implements types.StakingHooks.
func (Hooks) AfterValidatorRemoved(ctx types.Context, consAddr types.ConsAddress, valAddr types.ValAddress) error {
	return nil
}

// BeforeDelegationCreated implements types.StakingHooks.
func (Hooks) BeforeDelegationCreated(ctx types.Context, delAddr types.AccAddress, valAddr types.ValAddress) error {
	return nil
}

// BeforeDelegationRemoved implements types.StakingHooks.
func (Hooks) BeforeDelegationRemoved(ctx types.Context, delAddr types.AccAddress, valAddr types.ValAddress) error {
	return nil
}

// BeforeDelegationSharesModified implements types.StakingHooks.
func (Hooks) BeforeDelegationSharesModified(ctx types.Context, delAddr types.AccAddress, valAddr types.ValAddress) error {
	return nil
}

// BeforeValidatorModified implements types.StakingHooks.
func (Hooks) BeforeValidatorModified(ctx types.Context, valAddr types.ValAddress) error {
	return nil
}

// BeforeValidatorSlashed implements types.StakingHooks.
func (Hooks) BeforeValidatorSlashed(ctx types.Context, valAddr types.ValAddress, fraction math.LegacyDec) error {
	return nil
}
