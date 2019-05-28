#!/bin/bash


echo "安装Git及依赖工具"
set -x
yum install -y git
yum install -y wget
yum install -y gcc  gcc-c++ kernel-devel
yum install -y libtool libtool-ltdl-devel
yum install -y epel-release
set +x


echo "安装Go"
set -x
cd ~
wget https://studygolang.com/dl/golang/go1.12.5.linux-amd64.tar.gz
tar -xzf go1.12.5.linux-amd64.tar.gz
mv go /usr/local

echo "export GOROOT=/usr/local/go" >> /etc/profile
echo "export GOPATH=$HOME/go" >> /etc/profile
echo "export PATH=$PATH:$HOME/go/bin:/usr/local/go/bin" >> /etc/profile
source /etc/profile

echo "创建go目录"
cd ~
mkdir  -p  go/src/github.com/hyperledger/fabric
chmod -R 777 go
set +x


echo "安装docker 18.06.3-ce,配置docker的yum仓库,安装依赖包"
set -x
yum install yum-utils lvm2 device-mapper-persistent-data -y
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum-config-manager --disable docker-ce-edge docker-ce-test

echo "查看可安装的docker-ce列表"
yum list docker-ce --showduplicates | sort -r
echo "要安装docker-ce-18.06.3.ce-3.el7"
yum install docker-ce-18.06.3.ce-3.el7 -y
echo "启动docker"
systemctl start docker
echo "设置为开机启动"
systemctl enable docker
set +x

echo "设置阿里云镜像加速"
set -x
mkdir -p /etc/docker
chmod -R 777 /var/run/docker.sock
tee /etc/docker/daemon.json <<-'EOF'
{
  "registry-mirrors": ["https://mzax3ftr.mirror.aliyuncs.com"]
}
EOF

systemctl daemon-reload
systemctl restart docker
set +x

echo "安装nodejs8.11.4"
set -x
cd /usr/local
wget https://nodejs.org/dist/v8.11.4/node-v8.11.4-linux-x64.tar.gz
tar -zxf node-v8.11.4-linux-x64.tar.gz
mv node-v8.11.4-linux-x64/ node
echo "export NODE_HOME=/usr/local/node" >> /etc/profile
echo "export PATH=$PATH:/usr/local/node/bin" >> /etc/profile
source /etc/profile
set +x

echo "安装docker-compose"
set -x
yum install -y python-pip
yum list docker-compose --showduplicates
yum install -y docker-compose
docker-compose -version
set +x