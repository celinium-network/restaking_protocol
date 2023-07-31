#!/bin/sh

WALLET_KEY_NAME=$VALIDATOR_NAME   
CHAINFLAG="--chain-id ${CHAIN_ID}"
TOKEN_AMOUNT="10000000000000000000000000NCT"
STAKING_AMOUNT="1000000000NCT"
NODEIP="--node http://127.0.0.1:26657"

./consumerd tendermint unsafe-reset-all
./consumerd init $VALIDATOR_NAME --chain-id $CHAIN_ID

./consumerd keys add $WALLET_KEY_NAME --keyring-backend test
./consumerd add-genesis-account $WALLET_KEY_NAME $TOKEN_AMOUNT --keyring-backend test

./consumerd keys add $WALLET_KEY_NAME --keyring-backend test
./consumerd add-genesis-account $WALLET_KEY_NAME $TOKEN_AMOUNT --keyring-backend test

./consumerd gentx $WALLET_KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test

./consumerd collect-gentxs

./consumerd start --rpc.laddr tcp://0.0.0.0:26657 --grpc.address 0.0.0.0:9090
