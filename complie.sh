#!/bin/bash

echo "开始编译fabric"
set -x
cd  ~/go/src/github.com/hyperledger/fabric
make release
echo "export PATH=$PATH:$HOME/go/src/github.com/hyperledger/fabric/release/linux-amd64/bin" >> /etc/profile
source /etc/profile
echo "查看编译是否成功！"
peer version
orderer version
cryptogen version
configtxgen -version
configtxlator version
set +x


