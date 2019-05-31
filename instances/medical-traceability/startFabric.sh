#!/bin/bash

set -e

cp ./chaincode/medical-app.go ././../../fabric-docker/chaincode/medical-traceability/

cd ../../fabric-docker
./byfn.sh up -s couchdb -o etcdraft -a


printf "\nTotal setup execution time : $(($(date +%s) - starttime)) secs ...\n\n\n"
