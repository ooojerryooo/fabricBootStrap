#!/bin/bash

echo "#############################################################"
echo "#############################################################"
echo "#############################################################"
echo "#############################################################"
echo "################# #################### ######################"
echo "################### ################ ########################"
echo "#################### ############## #########################"
echo "###############    ##################    ####################"
echo "#############################################################"
echo "##########################   ################################"
echo "#############################################################"
echo "#############################################################"
echo "######################           ############################"
echo "#############################################################"
echo "#############################################################"
echo "#######        医疗器械实例链码安装并实例化         #########"
echo "#############################################################"
echo "#############################################################"
echo "######################作者：孙世江#########################"
echo "#############################################################"

starttime=$(date +%s)

CORE_PEER_TLS_ENABLED=true
CC_SRC_PATH="github.com/chaincode/medical-traceability/"
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
PEER0_ORG1_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
PEER0_ORG2_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt


verifyResult() {
  if [ $1 -ne 0 ]; then
    echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
    echo
    exit 1
  fi
}

set -x
docker exec cli peer chaincode install -n medical-cc -v 1.0 -p ${CC_SRC_PATH}
res=$?
set +x
verifyResult $res "Chaincode installed has failed"
echo "===================== Chaincode is installed ===================== "


set -x
docker exec cli peer chaincode instantiate -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C mychannel -n medical-cc -v 1.0 -c '{"Args":[""]}' -P "OR ('Org1MSP.member','Org2MSP.member')"
res=$?
set +x
verifyResult $res "Chaincode instantiate has failed"
echo "===================== Chaincode is instantiate ===================== "

sleep 10

printf "\nTotal execution time : $(($(date +%s) - starttime)) secs ...\n\n"

