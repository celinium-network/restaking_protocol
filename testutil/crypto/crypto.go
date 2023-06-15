package crypto

import (
	cryptocodec "github.com/cometbft/cometbft/crypto/encoding"
	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	sdkmock "github.com/cosmos/cosmos-sdk/testutil/mock"
)

func CreateTmProtoPublicKey() (tmprotocrypto.PublicKey, error) {
	pv := sdkmock.NewPV()
	cpv, err := pv.GetPubKey()
	if err != nil {
		return tmprotocrypto.PublicKey{}, err
	}

	return cryptocodec.PubKeyToProto(cpv)
}
