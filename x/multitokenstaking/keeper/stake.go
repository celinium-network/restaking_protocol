package keeper

import (
	"fmt"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

// MTStakingDelegate defines a method for performing a delegation of non native coins from a delegator to a validator
func (k Keeper) MTStakingDelegate(ctx sdk.Context, msg types.MsgMTStakingDelegate) error {
	defaultBondDenom := k.stakingKeeper.BondDenom(ctx)
	if strings.Compare(msg.Balance.Denom, defaultBondDenom) == 0 {
		return sdkerrors.Wrapf(types.ErrForbidStakingDenom, "denom: %s is native token", msg.Balance.Denom)
	}

	if !k.denomInWhiteList(ctx, msg.Balance.Denom) {
		return sdkerrors.Wrapf(types.ErrForbidStakingDenom, "denom: %s not in white list", msg.Balance.Denom)
	}

	agent := k.GetOrCreateMTStakingAgent(ctx, msg.Balance.Denom, msg.ValidatorAddress)
	delegatorAccAddr := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)

	if err := k.depositAndDelegate(ctx, delegatorAccAddr, agent.AgentAddress, agent.ValidatorAddress, msg.Balance); err != nil {
		return err
	}

	shares := agent.CalculateSharesFromTokenAmount(msg.Balance.Amount)
	agent.Shares = agent.Shares.Add(shares)
	agent.StakedAmount = agent.StakedAmount.Add(msg.Balance.Amount)

	k.SetMTStakingAgent(ctx, agent)
	k.SetMTStakingDenomAndValWithAgentAddress(ctx, agent.AgentAddress, agent.StakeDenom, agent.ValidatorAddress)

	return k.IncreaseDelegatorAgentShares(ctx, shares, agent.AgentAddress, msg.DelegatorAddress)
}

// depositAndDelegate defines a method deposit coin for delegator to agent and mint shares to delegator.
func (k Keeper) depositAndDelegate(ctx sdk.Context, delegator sdk.AccAddress, agentAddress, validatorAddress string, balance sdk.Coin) error {
	agentAccAddr := sdk.MustAccAddressFromBech32(agentAddress)
	validator, err := k.getValidator(ctx, validatorAddress)
	if err != nil {
		return err
	}

	if err := k.sendCoinsFromAccountToAccount(ctx, delegator, agentAccAddr, sdk.Coins{balance}); err != nil {
		return err
	}

	eqNativeCoin, err := k.CalculateEquivalentNativeCoin(ctx, balance)
	if err != nil {
		return err
	}

	// mint equivalent coin to agent account then agent delegate to validator.
	return k.mintAndDelegate(ctx, agentAccAddr, validator, eqNativeCoin)
}

// depositAndDelegate defines a method mint coin to agent account and delegate to a validator.
func (k Keeper) mintAndDelegate(ctx sdk.Context, agentAccAddr sdk.AccAddress, validator *stakingtypes.Validator, balance sdk.Coin) error {
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{balance}); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, agentAccAddr, sdk.Coins{balance}); err != nil {
		return err
	}

	if _, err := k.stakingKeeper.Delegate(ctx, agentAccAddr, balance.Amount, stakingtypes.Unbonded, *validator, true); err != nil {
		return err
	}
	return nil
}

// MTStakingUndelegate defines a method for performing an undelegation from a delegate and a validator.
// Delegator burn the shares of the agents. Then agent account begin undelegate.
func (k Keeper) MTStakingUndelegate(ctx sdk.Context, msg *types.MsgMTStakingUndelegate) error {
	agent, found := k.GetMTStakingAgent(ctx, msg.Balance.Denom, msg.ValidatorAddress)
	if !found {
		return types.ErrNotExistedAgent
	}

	delegatorAddr := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return err
	}

	if k.Unbond(ctx, delegatorAddr, valAddr, msg.Balance) != nil {
		return err
	}

	unbonding := k.GetOrCreateMTStakingUnbonding(ctx, agent.AgentAddress, msg.DelegatorAddress)
	unbondingTime := k.stakingKeeper.GetParams(ctx).UnbondingTime

	// TODO Whether the length of the entries should be limited ?
	undelegateCompleteTime := ctx.BlockTime().Add(unbondingTime)
	unbonding.Entries = append(unbonding.Entries, types.MTStakingUnbondingEntry{
		CompletionTime: undelegateCompleteTime,
		InitialBalance: msg.Balance,
		Balance:        msg.Balance,
	})

	k.SetMTStakingUnbonding(ctx, agent.AgentAddress, msg.DelegatorAddress, unbonding)
	k.InsertUBDQueue(ctx, unbonding, undelegateCompleteTime)

	return nil
}

// Unbond defines a method for removing shares from an agent by a delegator then agent undelegate funds from a validator.
func (k Keeper) Unbond(ctx sdk.Context, delegatorAccAddr sdk.AccAddress, valAddr sdk.ValAddress, balance sdk.Coin) error {
	var removeShares math.Int
	agent, found := k.GetMTStakingAgent(ctx, balance.Denom, valAddr.String())
	if !found {
		return types.ErrNotExistedAgent
	}

	delegatorAddr := delegatorAccAddr.String()
	removeShares = agent.CalculateSharesFromTokenAmount(balance.Amount)
	if err := k.DecreaseDelegatorAgentShares(ctx, removeShares, agent.AgentAddress, delegatorAddr); err != nil {
		return err
	}

	nativeCoinDenom := k.stakingKeeper.BondDenom(ctx)
	agentAccAddr := sdk.MustAccAddressFromBech32(agent.AgentAddress)
	rewards, err := k.distributionKeeper.WithdrawDelegationRewards(ctx, agentAccAddr, valAddr)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("withdraw delegation rewards failed %s", err))
		return err
	}

	// Currently only native coin rewards are considered
	agent.RewardAmount = agent.RewardAmount.Add(rewards.AmountOf(nativeCoinDenom))
	if !agent.RewardAmount.IsZero() {
		delegatorRewardAmt := agent.RewardAmount.Mul(removeShares).Quo(agent.Shares)
		if !delegatorRewardAmt.IsZero() {
			// delegator get staking rewards immediately.
			rewardCoins := sdk.Coins{sdk.NewCoin(nativeCoinDenom, delegatorRewardAmt)}
			if err := k.sendCoinsFromAccountToAccount(ctx, agentAccAddr, delegatorAccAddr, rewardCoins); err != nil {
				return err
			}

			agent.RewardAmount.Sub(delegatorRewardAmt)
		}
	}

	eqNativeBalance, err := k.CalculateEquivalentNativeCoin(ctx, balance)
	if err != nil {
		return err
	}

	if err := k.undelegateAndBurn(ctx, agentAccAddr, valAddr, eqNativeBalance); err != nil {
		return err
	}

	agent.Shares = agent.Shares.Sub(removeShares)
	agent.StakedAmount = agent.StakedAmount.Sub(balance.Amount)
	k.SetMTStakingAgent(ctx, agent)

	return nil
}

// undelegateAndBurn performs immediate undelegation from the staking module and burns the undelegated funds.
func (k Keeper) undelegateAndBurn(ctx sdk.Context, agentAccAddr sdk.AccAddress, valAddr sdk.ValAddress, balance sdk.Coin) error {
	stakedShares, err := k.stakingKeeper.ValidateUnbondAmount(ctx, agentAccAddr, valAddr, balance.Amount)
	if err != nil {
		return err
	}

	undelegationCoins, err := k.instantUndelegate(ctx, agentAccAddr, valAddr, stakedShares)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, agentAccAddr, types.ModuleName, undelegationCoins); err != nil {
		return err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, undelegationCoins); err != nil {
		return err
	}

	return nil
}

func (k Keeper) getValidator(ctx sdk.Context, validatorAddress string) (*stakingtypes.Validator, error) {
	valAddr, err := sdk.ValAddressFromBech32(validatorAddress)
	if err != nil {
		return nil, err
	}

	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrNotExistedValidator, "address %s", valAddr)
	}
	return &validator, nil
}

func (k Keeper) denomInWhiteList(ctx sdk.Context, denom string) bool {
	whiteList, found := k.GetMTStakingDenomWhiteList(ctx)
	if !found {
		return false
	}
	for _, wd := range whiteList.DenomList {
		if wd == denom {
			return true
		}
	}
	return false
}

func (k Keeper) GetOrCreateMTStakingAgent(ctx sdk.Context, denom, valAddr string) *types.MTStakingAgent {
	agent, found := k.GetMTStakingAgent(ctx, denom, valAddr)
	if found {
		return agent
	}

	newAccount := k.GenerateAccount(ctx, denom, valAddr)

	agent = &types.MTStakingAgent{
		StakeDenom:       denom,
		AgentAddress:     newAccount.Address,
		ValidatorAddress: valAddr,
		RewardAddress:    newAccount.Address,
		StakedAmount:     math.ZeroInt(),
		Shares:           math.ZeroInt(),
		RewardAmount:     math.ZeroInt(),
	}

	return agent
}

func (k Keeper) GenerateAccount(ctx sdk.Context, prefix, suffix string) *authtypes.ModuleAccount {
	header := ctx.BlockHeader()

	buf := []byte(types.ModuleName + prefix)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)

	addrBuf := string(buf) + suffix

	return authtypes.NewEmptyModuleAccount(addrBuf, authtypes.Staking)
}

// instantUndelegate define a method for immediately undelegate from staking module
func (k Keeper) instantUndelegate(ctx sdk.Context, delegatorAccAddr sdk.AccAddress, validatorAddr sdk.ValAddress, sharesAmount sdk.Dec) (sdk.Coins, error) {
	validator, found := k.stakingKeeper.GetValidator(ctx, validatorAddr)
	if !found {
		return nil, stakingtypes.ErrNoValidatorFound
	}

	unbondAmount, err := k.stakingKeeper.Unbond(ctx, delegatorAccAddr, validatorAddr, sharesAmount)
	if err != nil {
		return nil, err
	}

	bondDenom := k.stakingKeeper.GetParams(ctx).BondDenom
	unbondCoin := sdk.NewCoin(bondDenom, unbondAmount)
	unbondCoins := sdk.NewCoins(unbondCoin)

	moduleName := stakingtypes.NotBondedPoolName
	if validator.IsBonded() {
		moduleName = stakingtypes.BondedPoolName
	}

	err = k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, moduleName, delegatorAccAddr, unbondCoins)
	if err != nil {
		return nil, err
	}

	return unbondCoins, nil
}

// UpdateEquivalentNativeCoinMultiplier defines a method for updating the equivalent
// native coin multiplier for all token in white list
func (k Keeper) UpdateEquivalentNativeCoinMultiplier(ctx sdk.Context, epoch int64) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.MTStakingDenomWhiteListKey)
	iterator := sdk.KVStorePrefixIterator(prefixStore, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		denom := iterator.Value()
		multiplier, err := k.EquivalentNativeCoinMultiplier(ctx, string(denom))
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("get equivalent native coin multiplier of %s failed", denom))
			continue
		}

		k.SetEquivalentNativeCoinMultiplier(ctx, epoch, string(denom), multiplier)
	}
}

// RefreshAllAgentDelegation defines a method for updating all agent delegation amount base on current multiplier.
func (k Keeper) RefreshAllAgentDelegation(ctx sdk.Context) {
	agents := k.GetAllAgent(ctx)

	for i := 0; i < len(agents); i++ {
		err := k.refreshAgentDelegation(ctx, &agents[i])
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("refreshAgentDelegation failed, agentAddress %s", err))
		}
	}
}

func (k Keeper) refreshAgentDelegation(ctx sdk.Context, agent *types.MTStakingAgent) error {
	valAddress, err := sdk.ValAddressFromBech32(agent.ValidatorAddress)
	if err != nil {
		return err
	}

	validator, found := k.stakingKeeper.GetValidator(ctx, valAddress)
	if !found {
		return stakingtypes.ErrNoValidatorFound
	}

	var currentAmount math.Int
	agentAccAddr := sdk.MustAccAddressFromBech32(agent.AgentAddress)
	agentDelegation, found := k.stakingKeeper.GetDelegation(ctx, agentAccAddr, valAddress)
	if !found {
		return stakingtypes.ErrNoDelegation
	} else {
		currentAmount = validator.TokensFromShares(agentDelegation.Shares).RoundInt()
	}

	refreshedAmount, err := k.CalculateEquivalentNativeCoin(ctx, sdk.NewCoin(agent.StakeDenom, agent.StakedAmount))
	if err != nil {
		return err
	}

	if refreshedAmount.Amount.GT(currentAmount) {
		adjustment := refreshedAmount.Amount.Sub(currentAmount)
		err = k.mintAndDelegate(ctx, agentAccAddr, &validator, sdk.NewCoin(refreshedAmount.Denom, adjustment))
	} else if refreshedAmount.Amount.LT(currentAmount) {
		adjustment := currentAmount.Sub(refreshedAmount.Amount)
		err = k.undelegateAndBurn(ctx, agentAccAddr, valAddress, sdk.NewCoin(refreshedAmount.Denom, adjustment))
	}

	return err
}

// CollectAgentsReward defines a method for withdraw staking reward for all agents.
func (k Keeper) CollectAgentsReward(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.MTStakingAgentPrefix)
	defer iterator.Close()
	nativeCoinDenom := k.stakingKeeper.BondDenom(ctx)

	for ; iterator.Valid(); iterator.Next() {
		var agent types.MTStakingAgent
		// TODO panic or continue ?
		err := k.cdc.Unmarshal(iterator.Value(), &agent)
		if err != nil {
			ctx.Logger().Error(err.Error())
			continue
		}

		delegator := sdk.MustAccAddressFromBech32(agent.AgentAddress)
		valAddr, err := sdk.ValAddressFromBech32(agent.ValidatorAddress)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("convert validator address from bech32:%s failed, err: %s", agent.ValidatorAddress, err))
			continue
		}
		rewards, err := k.distributionKeeper.WithdrawDelegationRewards(ctx, delegator, valAddr)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("Withdraw delegation reward failed. AgentID: %s", agent.AgentAddress))
			continue
		}

		agent.RewardAmount = agent.RewardAmount.Add(rewards.AmountOf(nativeCoinDenom))
		agentBz, err := k.cdc.Marshal(&agent)
		if err != nil {
			ctx.Logger().Error(err.Error())
			continue
		}
		store.Set(iterator.Key(), agentBz)
	}
}
