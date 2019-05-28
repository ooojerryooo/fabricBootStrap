version1.0
* 基于fabric1.4配置的容器外启动orderer、peer
* 全是单节点，orderer使用solo模式
* 没有couchdb
* 没有kafka，zookeeper
* 今后的版本会逐步加入其它模块，并会持续更新使用最新版本的fabric
* 今天的版本还会加入一些调试好的实例
* fabric-ws文件夹是我参考的启动项目，下面两个文件夹，分别适用fabric1.0和1.4

## 基于软件版本
* 操作系统：Centos7.2
* HyperLedger Fabric 1.4
* go1.12.5.linux-amd64
* docker 18.06.3-ce
* nodejs8.11.4
* docker-compose-18

## 准备工作
* 执行ready.sh 预备系统环境
* 详细步骤查看fabric-project文件夹下的readme.md