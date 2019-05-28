#!/bin/bash

echo "当前目录${PWD}"

echo "===========生成创世块=========="
configtxgen -profile TwoOrgsOrdererGenesis -channelID system-mychannel -outputBlock ./genesis.block

echo "===========创建channel文件=========="
configtxgen -profile TwoOrgsChannel  -outputCreateChannelTx ./mychannel.tx -channelID mychannel


echo "===========创建锚点文件=========="
configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./Org1MSPanchors.tx -channelID mychannel -asOrg Org1MSP
configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./Org2MSPanchors.tx -channelID mychannel -asOrg Org2MSP
