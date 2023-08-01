#!/bin/bash

set -o errexit -o nounset

rly transact link coordinator-consumer-path 

rly start coordinator-consumer-path 