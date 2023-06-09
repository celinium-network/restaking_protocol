#!/usr/bin/env bash

# generate restaking expected_keeper mock
mockgen -source=./x/restaking/types/expected_keeper.go  -destination=./testutil/keeper/mocks.go -package=keeper