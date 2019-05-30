### 借用fabric-sample工程的First Network启动程序，其他自己微调

    * 升级链码使用upgrade_chaincode.sh
        docker exec cli scripts/upgrade_chaincode.sh
    * 升级完后删除之前的链码
        docker rm -f <container id>
        rm /var/hyperledger/production/chaincodes/fabcar.2.0