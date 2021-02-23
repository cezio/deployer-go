#!/bin/bash

echo 'rebuilding..'

go build .

BUILD_EX=$?
if [[ ! $BUILD_EX -eq 0 ]]; then
    echo 'error when compiling'
    exit 1
fi;
echo 'serving..'
DEPLOYER_CONFIG=. ./deployer-go -port 8080
