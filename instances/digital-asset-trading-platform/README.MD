# 资产交易平台

     这个版本使用的network有两个org，每个org两个peer
     由于每个org的锚节点都安装并实例化了链码
     sdk调用的时候要在channel中将两个org的peer锚节点都加入通道invoke的策略才会通过
	 
	 已经把AND策略改为OR，现在只在peer0.org1.example.com安装链码

## 启动步骤

	启动fabric网络
    	./startFabric.sh
	安装链码并实例化
		./installCC.sh
    安装node依赖包
        npm install --unsafe-perm
    登记admin并注册user1
        cd fabric
        rm -rf hfc-key-store
        node enrollAdmin.js
        node registerUser.js
	启动app
		./appStart.sh
    访问地址
		http://ip:8001/WEB-INF/login.html
    默认账号密码
		ssj\123,agent\123,buyer1\123,buyer2\123

### 软件基础

     ipfs
        下载https://pan.baidu.com/s/1H9DRYZLKmGvdEzP0-DzjJA
        安装
            ./install.sh
        初始化
            ipfs init
        配置参数
            ipfs config --json API.HTTPHeaders.Access-Control-Allow-Origin '[""]'
            ipfs config --json Addresses.API '"/ip4/0.0.0.0/tcp/5001"'
        设置为系统服务并开机启动
            # 创建ipfs.service并写入内容
            cd /lib/systemd/system/
            vim ipfs.service

            [Unit]
            Description=IPFS
            [Service]
            ExecStart=/usr/local/bin/ipfs daemon
            Restart=always
            User=root
            Group=root
            [Install]
            WantedBy=multi-user.target

            # 保存退出
            systemctl start ipfs
            systemctl enable ipfs
        访问地址
            http://127.0.0.1:5001/webui
     其他基础软件参考ready.sh脚本
