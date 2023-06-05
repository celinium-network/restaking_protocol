package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

var KeyTemplateClient = []byte("TemplateClient")

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyTemplateClient, p.TemplateClient, validateTemplateClient),
	}
}

func validateTemplateClient(i interface{}) error {
	cs, ok := i.(ibctm.ClientState)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T, expected: %T", i, ibctm.ClientState{})
	}

	copiedClient := cs

	if err := copiedClient.Validate(); err != nil {
		return err
	}

	return nil
}

// NewParams creates new provider parameters with provided arguments
func NewParams(
	cs *ibctm.ClientState,
) Params {
	return Params{
		TemplateClient: cs,
	}
}
