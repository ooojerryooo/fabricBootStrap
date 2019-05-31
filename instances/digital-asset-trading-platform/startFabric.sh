#!/bin/bash

set -e

cp ./chaincode/asset.go ././../../fabric-docker/chaincode/digital-asset-trading-platform/

cd ../../fabric-docker

echo "启动Fabric网络，状态数据库使用couchdb，排序服务使用etcd，打开CAserver"
./byfn.sh up -s couchdb -o etcdraft -a


printf "\nTotal setup execution time : $(($(date +%s) - starttime)) secs ...\n\n\n"
