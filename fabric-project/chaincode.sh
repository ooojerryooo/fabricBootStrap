#!/bin/bash

export FABRIC_CFG_PATH=${PWD}/peer0-org1
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_MSPCONFIGPATH=../crypto-config/peerOrganizations/org1.hwsj.com/users/Admin@org1.hwsj.com/msp
export CORE_PEER_ADDRESS=peer0.org1.hwsj.com:7051

peer chaincode install -n mycc -v 1.0 -p github.com/hyperledger/fabric/examples/chaincode/firstcc

peer chaincode instantiate -o orderer.hwsj.com:7050 -C mychannel -n mycc -v 1.0 -c '{"Args":["init", "a", "100", "b", "200"]}'  -P "OR ('Org1MSP.member', 'Org2MSP.member')"

peer chaincode invoke -o orderer.hwsj.com:7050 -C mychannel -n mycc  -c '{"Args":["invoke", "a", "b", "1"]}'

peer chaincode query -o orderer.hwsj.com:7050 -C mychannel -n mycc  -c '{"Args":["query", "a"]}'
