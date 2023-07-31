#!/bin/sh

WALLET_KEY_NAME=$VALIDATOR_NAME   
CHAINFLAG="--chain-id ${CHAIN_ID}"
TOKEN_AMOUNT="10000000000000000000000000CELI"
STAKING_AMOUNT="1000000000CNTR"
NODEIP="--node http://127.0.0.1:26657"

./coordinatord tendermint unsafe-reset-all
./coordinatord init $VALIDATOR_NAME --chain-id $CHAIN_ID

./coordinatord keys add $WALLET_KEY_NAME --keyring-backend test
./coordinatord add-genesis-account $WALLET_KEY_NAME $TOKEN_AMOUNT --keyring-backend test

./coordinatord keys add $WALLET_KEY_NAME --keyring-backend test
./coordinatord add-genesis-account $WALLET_KEY_NAME $TOKEN_AMOUNT --keyring-backend test

./coordinatord gentx $WALLET_KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test

./coordinatord collect-gentxs

./coordinatord start --rpc.laddr tcp://0.0.0.0:26657 --grpc.address 0.0.0.0:9090
