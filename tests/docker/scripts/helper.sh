#!/bin/bash

set -o errexit -o nounset

WALLET_KEY_NAME=$VALIDATOR_NAME 

_get_node_address() {
    wallet_address="$($CHAIN_NODE keys show $WALLET_KEY_NAME --keyring-backend=test --output json | jq -r .address)" 
    echo "$wallet_address"
}

_query_delegation() {
    $CHAIN_NODE query staking delegation $2 $3 --output json | jq .balance | jq .amount | sed 's/\"//g'
}

_get_wallet_balance() {
    wallet_address="$(_get_node_address)"  
    $CHAIN_NODE query bank balances "$wallet_address"
}

_transfer() {
  $CHAIN_NODE tx bank send \
    $2 $3 $4 \
    --chain-id=$CHAIN_ID \
    --gas="auto" \
    --gas-adjustment=1.5 \
    --fees="5000$DENOM" \
    --from=$WALLET_KEY_NAME \
    --keyring-backend=test
}

_ibc_transfer(){
    $CHAIN_NODE tx ibc-transfer transfer transfer channel-0 \
    $2 $3 \
    --gas="auto" \
    --gas-adjustment=1.5 \
    --fees="5000$DENOM" \
    --keyring-backend test \
    --chain-id=$CHAIN_ID \
    --from $WALLET_KEY_NAME
}

if [ "$1" = 'wallet:balance' ]; then
  _get_wallet_balance
elif [ "$1" = 'wallet:address' ]; then
  _get_node_address  
elif [ "$1" = 'validator:query_delegation' ]; then
   _query_delegation "$@"
elif [ "$1" = 'wallet:transfer' ]; then
   _transfer "$@"
elif [ "$1" = 'wallet:ibc_transfer' ]; then
   _ibc_transfer "$@"
else
  $CHAIN_NODE "$@"
fi    