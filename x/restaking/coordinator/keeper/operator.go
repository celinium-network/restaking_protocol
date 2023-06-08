package keeper

import (
	"fmt"

	"golang.org/x/exp/slices"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func (k Keeper) RegisterOperator(ctx sdk.Context, msg types.MsgRegisterOperator) error {
	// TODO The number of consumers in a operator should be limited.
	if len(msg.ConsumerChainIDs) != len(msg.ConsumerValidatorPks) {
		return types.ErrMismatchParams
	}

	var operatedValidators []types.OperatedValidator
	for i, chainID := range msg.ConsumerChainIDs {
		clientID, found := k.GetConsumerClientID(ctx, chainID)
		if !found {
			return errorsmod.Wrap(types.ErrUnknownConsumer, fmt.Sprintf("consumer chainID %s", chainID))
		}

		lastValidatorUpdates, found := k.GetConsumerValidator(ctx, string(clientID))
		if !found || len(lastValidatorUpdates) == 0 {
			return errorsmod.Wrap(types.ErrUnknownConsumer, fmt.Sprintf("unknown validator set of consumer %s", chainID))
		}

		restakingTokens := k.GetConsumerRestakingToken(ctx, string(clientID))
		if !slices.Contains[string](restakingTokens, msg.RestakingDenom) {
			return errorsmod.Wrap(types.ErrUnsupportedRestakingToken,
				fmt.Sprintf("chainID %s, unsupported token: %s", chainID, msg.RestakingDenom))
		}

		if !slices.ContainsFunc[abci.ValidatorUpdate](lastValidatorUpdates, func(vu abci.ValidatorUpdate) bool {
			return vu.PubKey.Equal(msg.ConsumerValidatorPks[i])
		}) {
			return errorsmod.Wrap(types.ErrNotExistedValidator,
				fmt.Sprintf("chainID %s, validator %s not existed", chainID, msg.ConsumerValidatorPks[i]))
		}

		operatedValidators = append(operatedValidators, types.OperatedValidator{
			ChainID:     chainID,
			ValidatorPk: msg.ConsumerValidatorPks[i],
		})
	}

	operatorAccount := generateOperatorAddress(ctx)

	operator := types.Operator{
		RestakingDenom:     msg.RestakingDenom,
		OperatorAddress:    operatorAccount.String(),
		RestakedAmount:     math.ZeroInt(),
		Shares:             math.ZeroInt(),
		OperatedValidators: operatedValidators,
		Owner:              msg.Sender,
	}

	k.SetOperator(ctx, &operator)

	return nil
}

// TODO more salt
func generateOperatorAddress(ctx sdk.Context) *authtypes.ModuleAccount {
	header := ctx.BlockHeader()

	buf := []byte(types.ModuleName)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)
	buf = append(buf, ctx.TxBytes()...)

	return authtypes.NewEmptyModuleAccount(string(buf), authtypes.Staking)
}

func (k Keeper) Delegate(ctx sdk.Context, delegator sdk.AccAddress, operatorAccAddr sdk.AccAddress, amount math.Int) error {
	operatorAddress := operatorAccAddr.String()
	operator, found := k.GetOperator(ctx, operatorAddress)
	if !found {
		return errorsmod.Wrap(types.ErrUnknownOperator, fmt.Sprintf("operator address %s", operatorAddress))
	}

	if err := k.sendCoinsFromAccountToAccount(
		ctx, delegator, operatorAccAddr, sdk.Coins{sdk.NewCoin(operator.RestakingDenom, amount)},
	); err != nil {
		return err
	}

	delegatorRecord, found := k.GetOperatorDelegateRecord(ctx, uint64(ctx.BlockHeight()))
	if !found {
		delegatorRecord = &types.OperatorDelegationRecord{
			OperatorAddress:  operatorAddress,
			DelegationAmount: math.ZeroInt(),
			Status:           types.OpDelRecordPending,
			IbcCallbackIds:   []string{},
		}
	}

	addedShares := operator.Shares.Mul(amount).Quo(operator.RestakedAmount.Add(amount))
	operator.Shares = operator.Shares.Add(addedShares)

	delegatorRecord.DelegationAmount = delegatorRecord.DelegationAmount.Add(amount)
	k.SetOperatorDelegateRecord(ctx, uint64(ctx.BlockHeight()), delegatorRecord)

	delegatorAddress := delegator.String()
	delegation, found := k.GetDelegation(ctx, delegatorAddress, operatorAddress)
	if !found {
		delegation = &types.Delegation{
			Delegator: delegatorAddress,
			Operator:  operatorAddress,
			Shares:    math.ZeroInt(),
		}
	}

	// TODO shares should be math.Dec?
	delegation.Shares = delegation.Shares.Add(addedShares)

	k.SetDelegation(ctx, delegatorAddress, operatorAddress, delegation)
	return nil
}
