#!/bin/bash
mkdir -p ./logs/
LOG_PATH="logs/app-$(date +%Y-%m-%d).log"
LOG2_PATH="logs/webapp-$(date +%Y-%m-%d).log"
echo "启动app.js和webApp.js，netstat -nlpt 查看启动情况 端口号8001 8002"
node app.js >>$LOG_PATH 2>&1 &
node webApp.js >>$LOG2_PATH 2>&1 &
