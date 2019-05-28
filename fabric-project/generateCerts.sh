#!/bin/bash

function generateCerts() {
  which cryptogen
  if [ "$?" -ne 0 ];then
    echo "cryptogen工具没发现.退出！"
    exit 1
  else
    cryptogen version
  fi


  echo
  echo "##########################################################"
  echo "############## 使用cryptogen工具生成证书文件 #############"
  echo "##########################################################"

  if [ -d "crypto-config" ]; then
    rm -Rf crypto-config
  fi

  set -x
  cryptogen generate --config=crypto-config.yaml --output ./crypto-config
  res=$?
  set +x

  if [ $res -ne 0 ]; then
    echo "未能生成证书..."
    exit 1
  else
    echo "证书文件已生成！目录结构如下："
      which tree
      if [ "$?" -ne 0 ];then
        echo "tree工具没发现.退出！"
        yum install -y tree
        tree -L 3 crypto-config
      else
        tree -L 3 crypto-config
      fi
  fi
  echo
}

generateCerts