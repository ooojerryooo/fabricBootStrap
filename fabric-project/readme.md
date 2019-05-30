## v1.1版本
* 此版本采用了etcd实现的orderer服务
## 基于软件版本
* HyperLedger Fabric 1.4.1
* go1.12.5.linux-amd64
* docker 18.06.3-ce
* nodejs8.11.4
* docker-compose version 1.18.0

### `一、服务器准备工作`
执行ready.sh

###` 二、下载Fabric源码`
mkdir -p /root/go/src/github.com/hyperledger
cd /root/go/src/github.com/hyperledger
git clone https://github.com/hyperledger/fabric.git

### `三、编译源码，生成可执行文件`
执行complie.sh


### `四、可执行文件赋权、拷贝链码到GOPATHA下`

cd /root/fabric-project
cp chaincode/ $GOPATH/src/github.com/hyperledger/fabric/examples/chaincode/firstcc
chmod -R +x *


### `五、容器外运行一个简单的fabric网络`

    1、生成证书文件
    ./generateCerts.sh
        ①域名映射（一台服务器只做一次），执行下列操作
        echo "进行域名映射"
        echo "127.0.0.1 orderer.hwsj.com" >> /etc/hosts
        echo "127.0.0.1 peer0.org1.hwsj.com" >> /etc/hosts
        echo "127.0.0.1 peer1.org1.hwsj.com" >> /etc/hosts
        echo "127.0.0.1 peer0.org2.hwsj.com" >> /etc/hosts
        echo "127.0.0.1 peer1.org2.hwsj.com" >> /etc/hosts
        echo "测试是否能ping通"
        ping orderer.hwsj.com -c 3
    
    2、生成创世块、通道和锚点文件
    cd order
    ./generateblock.sh
    
    3、启动order节点
    orderer start >> log_orderer.log 2>&1 &
    
    4、启动peer节点
    cd ../peer0-org1
    ./peer0-org1-up.sh
    
    5、创建channel->加入channel->更新锚节点
    ./crea-join-upda-channel.sh
    
    6、安装链码，实例化，调用，查询
    ./chaincode.sh