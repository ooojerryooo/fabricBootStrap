# 食品追溯改成医疗器械追溯

[![Build Status](https://travis-ci.org/zhazhalaila/hyperledger-simple-app.svg?branch=master)](https://travis-ci.org/zhazhalaila/hyperledger-simple-app) ![](https://img.shields.io/badge/language-javascript-blue.svg)

基于 Hyperledger Fabric 的一个极简 App

chaincode 由@[DevilExileSu](https://github.com/DevilExileSu)所编写,感谢他

在此之前请确保已安装 Hyperledger Fabric（本项目基于 Hyperledger Fabric v1.4.1)

库版本:

```
npm 5.6.0 (建议大于或等于此版本)
node.js v8.11.3 (建议大于或等于此版本)
angularjs 1.4.3
```

<h4>安装及配置</h4>

Fabric网络
```
如果网络没有启动，执行startFabric.sh
chmod +x startFabric.sh
如果格式不对，vi startFabric.sh
:set ff=unix
:wq

./startFabric.sh
```

Step 1:安装 node 依赖包

```js
npm install --unsafe-perm
//安装时速度慢，静等即可
```

Step 2:登记 admin 并注册 user2

```shell
node enrollAdmin.js
node registerUser.js
```

Step 3:安装链码并实例化

```
先给installCC.sh赋权再执行
chmod +x installCC.sh
./installCC.sh
```

Step 4:启动 app

```
先给app.sh赋权再执行
chmod +x app.sh
./app.sh
```

访问`http://localhost:3389`

由于没有初始化信息，因此需要先提交表单信息才可以查询出信息

在填写表单信息时，没有做过多的处理，因此每个选项都要尽量填写（配料的表单可以不填写完）

配料及食品信息对于同一个 Id 只能添加一次，中转信息可以添加多次

<h4>url & json 格式</h4>

获取食品信息

`http://localhost:8000/source/:id`

```json
{
  "FoodName": "Apple",
  "FoodSpec": "123456",
  "FoodMFGDate": "2018-8-24",
  "FoodEXPDate": "10day",
  "FoodLOT": "123",
  "FoodQSID": "456",
  "FoodMFRSName": "lalala",
  "FoodProPrice": "2",
  "FoodProPlace": "zhengzhou"
}
```

获取食品配料信息

`http://localhost:8000/part/:id`

```
[{"IngID":"1","IngName":"a"},{"IngID":"2","IngName":"b"},{"IngID":"3","IngName":"c"},{"IngID":"4","IngName":"d"},{"IngID":"5","IngName":"e"}]
```

获取交易（运输）信息

`http://localhost:8000/transit/:id`

```
[{"LogDepartureTm":"14:20","LogArrivalTm":"16:40","LogMission":"Store","LogDeparturePl":"zhengzhou","LogDest":"wuhan","LogToSeller":"lalala","LogStorageTm":"1day","LogMOT":"truck","LogCopName":"shunfeng","LogCost":"10"},{"LogDepartureTm":"16:50","LogArrivalTm":"18:50","LogMission":"Store","LogDeparturePl":"wuhan","LogDest":"guangzhou","LogToSeller":"lalala","LogStorageTm":"1day","LogMOT":"truck","LogCopName":"shunfeng","LogCost":"10"}]
```

 <h4>部署在服务器上</h4>
 
 仅需修改server.js文件最后几行代码
 
 ```javascript
var port = process.env.PORT || 3389; //端口修改为服务器的安全组端口

app.listen(port, '0.0.0.0', function(){
console.log("Live on port: " + port);
});

```

<h4>重要的文件说明:</h4>

```

server.js 启动
routes.js 定义路由
controller.js 路由
chaincode
medical-app.go chaincode

```

一般来说，路由会在routes.js文件中定义，但是这样不方便路由复用，因此分为两个脚本

<h4>参考链接</h4>

  [Education](https://github.com/hyperledger/education/tree/master/LFS171x)

  [启动网络参考](https://github.com/ooojerryooo/fabricBootStrap/tree/master/fabric-docker)


<h4>注释说明:</h4>

除了chaincode，其余大部分.js文件都是参照以上参考链接所写，因此注释为英文
```