package types

import (
	abci "github.com/cometbft/cometbft/abci/types"
)

func ValidatorUpdateToConsumerValidator(update abci.ValidatorUpdate) ConsumerValidator {
	return ConsumerValidator{
		ValidatorPk: update.PubKey,
		Power:       update.Power,
	}
}
