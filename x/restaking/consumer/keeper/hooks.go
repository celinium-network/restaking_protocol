package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/celinium-network/restaking_protocol/x/restaking/consumer/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
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
func (Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// AfterUnbondingInitiated implements types.StakingHooks.
func (Hooks) AfterUnbondingInitiated(ctx sdk.Context, id uint64) error {
	return nil
}

// AfterValidatorBeginUnbonding implements types.StakingHooks.
func (Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

// AfterValidatorBonded implements types.StakingHooks.
func (Hooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

// AfterValidatorCreated implements types.StakingHooks.
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error {
	// TODO make added validators in a block into a ValidatorSetChange
	_, err := h.k.GetCoordinatorChannelID(ctx)
	if err != nil {
		return nil
	}

	vsc := restaking.ValidatorSetChange{
		Type:               restaking.ValidatorSetChange_ADD,
		ValidatorAddresses: []string{valAddr.String()},
	}
	h.k.AppendPendingVSC(ctx, vsc)
	return nil
}

// AfterValidatorRemoved implements types.StakingHooks.
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	_, err := h.k.GetCoordinatorChannelID(ctx)
	if err != nil {
		return nil
	}

	vsc := restaking.ValidatorSetChange{
		Type:               restaking.ValidatorSetChange_REMOVE,
		ValidatorAddresses: []string{valAddr.String()},
	}
	h.k.AppendPendingVSC(ctx, vsc)
	return nil
}

// BeforeDelegationCreated implements types.StakingHooks.
func (Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeDelegationRemoved implements types.StakingHooks.
func (Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeDelegationSharesModified implements types.StakingHooks.
func (Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeValidatorModified implements types.StakingHooks.
func (Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeValidatorSlashed implements types.StakingHooks.
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction math.LegacyDec) error {
	iterator := h.k.ValidatorsOperatorStoreIterator(ctx, valAddr.String())
	defer iterator.Close()

	var slashes []restaking.ConsumerSlash
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		operatorAddr := types.ParseValidatorOperatorKey(key)
		slashes = append(slashes, restaking.ConsumerSlash{
			OperatorAddress: string(operatorAddr),
			SlashFactor:     fraction,
		})
	}

	h.k.AppendPendingConsumerSlash(ctx, slashes...)

	return nil
}
