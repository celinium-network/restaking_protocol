#!/usr/bin/env bash

# generate restaking expected_keeper mock
mockgen -source=./x/restaking/types/expected_keeper.go  -destination=./testutil/keeper/restaking/mocks.go -package=keeper

mockgen -source=./x/multitokenstaking/types/expected_keeper.go  -destination=./testutil/keeper/multitokenstaking/mocks.go -package=keeper