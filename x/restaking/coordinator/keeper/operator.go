package keeper

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/exp/slices"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func (k Keeper) RegisterOperator(ctx sdk.Context, msg types.MsgRegisterOperatorRequest) error {
	// TODO The number of consumers in a operator should be limited.
	if len(msg.ConsumerChainIDs) != len(msg.ConsumerValidatorAddresses) {
		return types.ErrMismatchParams
	}

	var operatedValidators []types.OperatedValidator
	for i, chainID := range msg.ConsumerChainIDs {
		clientID, found := k.GetConsumerClientID(ctx, chainID)
		if !found {
			return errorsmod.Wrap(types.ErrUnknownConsumer, fmt.Sprintf("consumer chainID %s", chainID))
		}

		restakingTokens := k.GetConsumerRestakingToken(ctx, string(clientID))
		if !slices.Contains[string](restakingTokens, msg.RestakingDenom) {
			return errorsmod.Wrap(types.ErrUnsupportedRestakingToken,
				fmt.Sprintf("chainID %s, unsupported token: %s", chainID, msg.RestakingDenom))
		}

		if _, found := k.GetConsumerValidator(ctx, string(clientID), msg.ConsumerValidatorAddresses[i]); !found {
			return errorsmod.Wrap(types.ErrNotExistedValidator,
				fmt.Sprintf("chainID %s, validator %s not existed", clientID, msg.ConsumerValidatorAddresses[i]))
		}

		operatedValidators = append(operatedValidators, types.OperatedValidator{
			ChainID:          chainID,
			ClientID:         string(clientID),
			ValidatorAddress: msg.ConsumerValidatorAddresses[i],
		})
	}

	operatorAccount := generateOperatorAddress(ctx)

	operator := types.Operator{
		RestakingDenom:     msg.RestakingDenom,
		OperatorAddress:    operatorAccount.Address,
		RestakedAmount:     math.ZeroInt(),
		Shares:             math.ZeroInt(),
		OperatedValidators: operatedValidators,
		Owner:              msg.Sender,
	}

	k.SetOperator(ctx, operatorAccount.GetAddress(), &operator)

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

func (k Keeper) Delegate(ctx sdk.Context, delegatorAccAddr sdk.AccAddress, operatorAccAddr sdk.AccAddress, amount math.Int) error {
	k.WithdrawDelegatorRewards(ctx, delegatorAccAddr, operatorAccAddr)

	operatorAddress := operatorAccAddr.String()
	operator, found := k.GetOperator(ctx, operatorAccAddr)
	if !found {
		return errorsmod.Wrap(types.ErrUnknownOperator, fmt.Sprintf("operator address %s", operatorAddress))
	}

	if err := k.sendCoinsFromAccountToAccount(
		ctx, delegatorAccAddr, operatorAccAddr, sdk.Coins{sdk.NewCoin(operator.RestakingDenom, amount)},
	); err != nil {
		return err
	}

	delegatorRecord, found := k.GetOperatorDelegateRecord(ctx, uint64(ctx.BlockHeight()), operatorAccAddr)
	if !found {
		delegatorRecord = &types.OperatorDelegationRecord{
			OperatorAddress:  operatorAddress,
			DelegationAmount: math.ZeroInt(),
			Status:           types.OpDelRecordPending,
			IbcCallbackIds:   []string{},
		}
	}

	var addedShares math.Int
	if operator.RestakedAmount.IsZero() {
		addedShares = amount
	} else {
		addedShares = operator.Shares.Mul(amount).Quo(operator.RestakedAmount.Add(amount))
	}

	operator.Shares = operator.Shares.Add(addedShares)

	delegatorRecord.DelegationAmount = delegatorRecord.DelegationAmount.Add(amount)
	// TODO maybe use epoch replace blockHight
	k.SetOperatorDelegateRecord(ctx, uint64(ctx.BlockHeight()), delegatorRecord)

	delegatorAddress := delegatorAccAddr.String()
	delegation, found := k.GetDelegation(ctx, delegatorAccAddr, operatorAccAddr)
	if !found {
		delegation = &types.Delegation{
			Delegator: delegatorAddress,
			Operator:  operatorAddress,
			Shares:    math.ZeroInt(),
		}
	}

	// TODO shares should be math.Dec?
	delegation.Shares = delegation.Shares.Add(addedShares)
	k.SetDelegation(ctx, delegatorAccAddr, operatorAccAddr, delegation)

	k.AfterDelegationSharesModified(ctx, delegatorAccAddr, operatorAccAddr)
	return nil
}

func (k Keeper) Undelegate(ctx sdk.Context, delegatorAccAddr, operatorAccAddr sdk.AccAddress, amount math.Int) error {
	operatorAddress := operatorAccAddr.String()
	operator, found := k.GetOperator(ctx, operatorAccAddr)
	if !found {
		return errorsmod.Wrapf(types.ErrUnknownOperator, "operator address %s", operatorAddress)
	}

	// TODO check entry UnbondingDelegationEntry length? Set by module params?
	delegation, found := k.GetDelegation(ctx, delegatorAccAddr, operatorAccAddr)
	if !found {
		return errorsmod.Wrapf(types.ErrInsufficientDelegation, "delegation is't existed")
	}

	removedShared := operator.RestakedAmount.Mul(operator.Shares).Quo(amount)
	if delegation.Shares.LT(removedShared) {
		return errorsmod.Wrapf(types.ErrInsufficientDelegation, "remove shares too much")
	}

	operator.Shares = operator.Shares.Sub(removedShared)
	delegation.Shares = delegation.Shares.Sub(removedShared)

	k.SetOperator(ctx, operatorAccAddr, operator)
	k.SetDelegation(ctx, delegatorAccAddr, operatorAccAddr, delegation)

	blockHeight := uint64(ctx.BlockHeight())
	undelegationRecord, found := k.GetOperatorUndelegationRecord(ctx, blockHeight, operatorAccAddr)
	if !found {
		undelegationRecord = &types.OperatorUndelegationRecord{
			OperatorAddress:    operatorAddress,
			UndelegationAmount: math.ZeroInt(),
			Status:             types.OpUndelegationRecordPending,
			IbcCallbackIds:     []string{},
			LatestCompleteTime: -1,
		}
	}

	entryID := k.SetUnbondingDelegationEntry(ctx, blockHeight, delegatorAccAddr, operatorAccAddr, sdk.NewCoin(operator.RestakingDenom, amount))
	undelegationRecord.UndelegationAmount = undelegationRecord.UndelegationAmount.Add(amount)
	undelegationRecord.UnbondingEntryIds = append(undelegationRecord.UnbondingEntryIds, entryID)

	// TODO maybe use epoch replace blockHight
	k.SetOperatorUndelegationRecord(ctx, blockHeight, undelegationRecord)

	return nil
}

func (k Keeper) SetUnbondingDelegationEntry(
	ctx sdk.Context, creationHeight uint64, delAddr sdk.AccAddress, opAddr sdk.AccAddress, balance sdk.Coin,
) (entryID uint64) {
	ubd, found := k.GetUnbondingDelegation(ctx, delAddr, opAddr)
	id := k.IncrementUnbondingID(ctx)

	if found {
		ubd.AddEntry(creationHeight, balance, id)
	} else {
		ubd = types.NewUnbondingDelegation(delAddr, opAddr, creationHeight, balance, id)
	}

	k.SetUnbondingDelegation(ctx, ubd)

	k.SetUnbondingDelegationByUnbondingID(ctx, ubd, id)

	return id
}

func (k Keeper) GetUnbondingDelegation(ctx sdk.Context, delAddr sdk.AccAddress, opAddr sdk.AccAddress) (ubd types.UnbondingDelegation, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetUBDKey(delAddr, opAddr)
	value := store.Get(key)

	if value == nil {
		return ubd, false
	}

	k.cdc.MustUnmarshal(value, &ubd)

	return ubd, true
}

func (k Keeper) SetUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
	delAddr := sdk.MustAccAddressFromBech32(ubd.DelegatorAddress)

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&ubd)
	valAddr := sdk.MustAccAddressFromBech32(ubd.OperatorAddress)
	key := types.GetUBDKey(delAddr, valAddr)

	store.Set(key, bz)
}

func (k Keeper) IncrementUnbondingID(ctx sdk.Context) (unbondingID uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte{types.UnbondingIDKey})
	if bz != nil {
		unbondingID = binary.BigEndian.Uint64(bz)
	}

	unbondingID++

	// Convert back into bytes for storage
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, unbondingID)

	store.Set([]byte{types.UnbondingIDKey}, bz)

	return unbondingID
}

func (k Keeper) SetUnbondingDelegationByUnbondingID(ctx sdk.Context, ubd types.UnbondingDelegation, id uint64) {
	store := ctx.KVStore(k.storeKey)
	delAddr := sdk.MustAccAddressFromBech32(ubd.DelegatorAddress)
	valAddr := sdk.MustAccAddressFromBech32(ubd.OperatorAddress)

	ubdKey := types.GetUBDKey(delAddr, valAddr)
	store.Set(types.GetUnbondingIndexKey(id), ubdKey)
}

func (k Keeper) GetUnbondingDelegationByUnbondingID(ctx sdk.Context, id uint64) (ubd types.UnbondingDelegation, found bool) {
	store := ctx.KVStore(k.storeKey)

	ubdKey := store.Get(types.GetUnbondingIndexKey(id))
	if ubdKey == nil {
		return types.UnbondingDelegation{}, false
	}

	value := store.Get(ubdKey)
	if value == nil {
		return types.UnbondingDelegation{}, false
	}

	err := k.cdc.Unmarshal(value, &ubd)
	// An error here means that what we got wasn't the right type
	if err != nil {
		return types.UnbondingDelegation{}, false
	}

	return ubd, true
}
