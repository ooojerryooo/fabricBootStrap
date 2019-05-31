#!/bin/bash

mkdir -p ./logs/
LOG_PATH="logs/app-$(date +%Y-%m-%d).log"
LOG2_PATH="logs/webapp-$(date +%Y-%m-%d).log"
echo "启动server.js，netstat -nlpt 查看启动情况 端口号3389"
echo "查看Pid ps aux | grep node server.js"

set -x
node server.js >>$LOG_PATH 2>&1 &
netstat -nlpt|grep 3389
set +x