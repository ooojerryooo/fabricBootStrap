#!/bin/bash

# 设定为peer路径后，其他路径都跟着改变
export FABRIC_CFG_PATH=${PWD}/peer0-org1
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_MSPCONFIGPATH=../crypto-config/peerOrganizations/org1.hwsj.com/users/Admin@org1.hwsj.com/msp

peer channel create -o orderer.hwsj.com:7050 -c mychannel -f order/mychannel.tx

export CORE_PEER_ADDRESS=peer0.org1.hwsj.com:7051
peer  channel join -b mychannel.block

peer channel update -o orderer.hwsj.com:7050 -c mychannel -f order/Org1MSPanchors.tx