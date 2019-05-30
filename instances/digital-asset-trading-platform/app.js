var express = require('express');
var http = require('http');
var multipart= require('connect-multiparty');
var bodyParser = require('body-parser');
var ipfs = require('./js/ipfs-server')
var fs = require('fs')
var crypto = require('crypto')

var fabricQuery = require('./fabric/query')
var fabricInvoke = require('./fabric/invoke')

var multipartMiddeware = multipart();
var urlencodedParser = bodyParser.urlencoded({ extended: false })
var app = express();

var router = express.Router();
router.all('*',function (req,res,next) {
    res.header("Access-Control-Allow-Origin", "*");
    res.header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept");
    res.header("Access-Control-Allow-Methods","PUT,POST,GET,DELETE,OPTIONS");
    next();
});

var tempFilePath = './temp/'
fs.exists(tempFilePath,(exists)=>{
    if(!exists) {
        fs.mkdirSync(tempFilePath)
    }
})

var request = require('request');

app.post('/getchaincode',urlencodedParser,function(reqq,ress){
    ress.header("Access-Control-Allow-Origin", "*");
    ress.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    ress.header("Access-Control-Allow-Headers", "X-Requested-With");
    ress.header('Access-Control-Allow-Headers', 'Content-Type');
    var options = {
        hostname: reqq.body.hostname,
        port: reqq.body.port,
        path: reqq.body.url,
        method: 'GET'
    };
    var req = http.request(options, function (res) {
        console.log('STATUS: ' + res.statusCode);
        console.log('HEADERS: ' + JSON.stringify(res.headers));
        res.setEncoding('utf8');
        res.on('data', function (chunk) {
            console.log('BODY: ' + chunk);
            ress.end(JSON.stringify(chunk));
        });
    });
    req.on('error', function (e) {
        console.log('problem with request: ' + e.message);
    });
    req.end();
})

// 获取文件扩展名
function suffix(filename){
    //var result =/\.[^\.]+/.exec(file_name);
    var indexStart = filename.lastIndexOf(".") + 1;
    var indexEnd = filename.length;
    var result = filename.substring(indexStart,indexEnd);
    return result;
}

// 生成随机字符串
function randomString(len) {
    len = len || 64;
    var chars = '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
    var maxPos = chars.length;
    var str = '';
    for (i = 0; i < len; i++) {
        str += chars.charAt(Math.floor(Math.random() * maxPos));
    }
    return str;
}

// 文件加密
function fileEncrypt(buffer, key,algorithm) {
    var encrypted = "";
    var cip = crypto.createCipher(algorithm, key);
    encrypted += cip.update(buffer, 'binary', 'hex');
    encrypted += cip.final('hex');
    return Buffer.from( encrypted, 'utf8' );
}

function fileDecrypt(buffer, str, key,algorithm) {
    var decrypted = "";
    var decip = crypto.createDecipher(algorithm, key);
    //decrypted += decip.update(buffer.toString('utf8'), 'hex', 'binary');

    if(buffer){
        decrypted += decip.update(buffer.toString('utf8'), 'hex', 'binary');
    } else if(str){
        decrypted += decip.update(str, 'hex', 'binary');
    }

    decrypted += decip.final('binary');
    return decrypted;
}

app.use(multipart({uploadDir:tempFilePath}));

app.set('port',8002);
http.createServer(app).listen(app.get('port'),function () {
    console.log("Starting......")
});
//（1）查看当前用户资产列表
app.post('/getAssentForOwner',urlencodedParser,function(req,res){
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var userId =req.body.userId;
    console.log("前台参数："+userId);
    new Promise(function(resolve,reject){
        fabricQuery.query('fabcar','getAssentForOwner',[userId],
            function(err,data){
                if(err) {
                    reject(err);
                } else {
                    resolve(data);
                }
            })
    }).then((data)=>{
        console.log(data);
        res.end(JSON.stringify(data));
    })
})
//（2-1）我发布的代理合约
app.post('/getContractForOwner',urlencodedParser,function(req,res){
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var userId =req.body.userId;
    console.log("前台参数："+userId);
    new Promise(function(resolve,reject){
        fabricQuery.query('fabcar','getContractForOwner',[userId],
            function(err,data){
                if(err) {
                    reject(err);
                } else {
                    resolve(data);
                }
            })
    }).then((data)=>{
        console.log(data);
        res.end(JSON.stringify(data));
    })
})
//（2-2）我发布的直销合约
app.post('/getContractForOwner2',urlencodedParser,function(req,res){
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var userId =req.body.userId;
    console.log("前台参数："+userId);
    new Promise(function(resolve,reject){
        fabricQuery.query('fabcar','getContractForOwner2',[userId],
            function(err,data){
                if(err) {
                    reject(err);
                } else {
                    resolve(data);
                }
            })
    }).then((data)=>{
        console.log(data);
        res.end(JSON.stringify(data));
    })
})
//（2-3）我代理的合约
app.post('/getContractForOwner3',urlencodedParser,function(req,res){
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var userId =req.body.userId;
    console.log("前台参数："+userId);
    new Promise(function(resolve,reject){
        fabricQuery.query('fabcar','getContractForOwner3',[userId],
            function(err,data){
                if(err) {
                    reject(err);
                } else {
                    resolve(data);
                }
            })
    }).then((data)=>{
        console.log(data);
        res.end(JSON.stringify(data));
    })
})
//（3）查看当前用户资产使用权列表
app.post('/getAssentForBuyer',urlencodedParser,function(req,res){
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var userId =req.body.userId;
    console.log("前台参数："+userId);
    new Promise(function(resolve,reject){
        fabricQuery.query('fabcar','getAssentForBuyer',[userId],
            function(err,data){
                if(err) {
                    reject(err);
                } else {
                    resolve(data);
                }
            })
    }).then((data)=>{
        console.log(data);
        res.end(JSON.stringify(data));
    })
})
//（4）查看当前用户交易流水
app.post('/getTransactionForUser',urlencodedParser,function(req,res){
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var userId =req.body.userId;
    var txType =req.body.txType;
    console.log("前台参数："+userId);
    new Promise(function(resolve,reject){
        fabricQuery.query('fabcar','getTransactionForUser',[userId,txType],
            function(err,data){
                if(err) {
                    reject(err);
                } else {
                    resolve(data);
                }
            })
    }).then((data)=>{
        console.log(data);
        res.end(JSON.stringify(data));
    })
})
//（5）查看当前用户账户余额
app.post('/getAccountForUser',urlencodedParser,function(req,res){
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var userId =req.body.userId;
    console.log("前台参数："+userId);
    new Promise(function(resolve,reject){
        fabricQuery.query('fabcar','getAccountForUser',[userId],
            function(err,data){
                if(err) {
                    reject(err);
                } else {
                    resolve(data);
                }
            })
    }).then((data)=>{
        console.log(data);
        res.end(JSON.stringify(data));
    })
})
//（6）查看全部有效合约
app.post('/getContractList',urlencodedParser,function(req,res){
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var userId =req.body.userId;
    console.log("前台参数："+userId);
    new Promise(function(resolve,reject){
        fabricQuery.query('fabcar','getContractList',[userId],
            function(err,data){
                if(err) {
                    reject(err);
                } else {
                    resolve(data);
                }
        })
    }).then((data)=>{
        console.log(data);
        res.end(JSON.stringify(data));
    })
})
//（7）//根据ID取合约
app.post('/getContractByid',urlencodedParser,function(req,res){
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var contractID =req.body.contractID;
    console.log("前台参数："+contractID);
    new Promise(function(resolve,reject){
        fabricQuery.query('fabcar','getContractByid',[contractID],
            function(err,data){
                if(err) {
                    reject(err);
                } else {
                    resolve(data);
                }
            })
    }).then((data)=>{
        console.log(data);
        res.end(JSON.stringify(data));
    })
})
//（8）//根据ID取资产
app.post('/getAssetByid',urlencodedParser,function(req,res){
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var AssetID =req.body.AssetID;
    console.log("前台参数："+AssetID);
    new Promise(function(resolve,reject){
        fabricQuery.query('fabcar','getAssetByid',[AssetID],
            function(err,data){
                if(err) {
                    reject(err);
                } else {
                    resolve(data);
                }
            })
    }).then((data)=>{
        console.log(data);
        res.end(JSON.stringify(data));
    })
})

// 用户注册
app.post('/register',urlencodedParser,function (req,res) {
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var userId = req.body.userCode;
    var userName = req.body.userName;
    var mobile = req.body.mobile;
    var pwd = req.body.pwd;
    console.log("userId:" + userId);
    console.log("pwd:" + pwd)

    new Promise((resolve,reject)=>{
        fabricInvoke.invoke('fabcar','register',[userId,userName,pwd,mobile],(err,data)=>{
            if(err) {
                reject(err);
            } else {
                resolve(data);
            }
        })
    }).then((data)=>{
        console.log(data);
        res.send(data);
    }).catch((err)=>{
        console.error(err);
    })
})

// 用户登录
app.post('/login',urlencodedParser,function (req,res) {
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var userId = req.body.userCode;
    var pwd = req.body.pwd;
    console.log("userId:" + userId);
    console.log("pwd:" + pwd)

    new Promise((resolve,reject)=>{
        fabricQuery.query('fabcar','login',[userId,pwd],(err,data)=>{
            if(err) {
                reject(err);
            } else {
                resolve(data);
            }
        })
    }).then((data)=>{
        console.log(data);
        //res.send(data);
        res.end(data);
    }).catch((err)=>{
        console.error(err);
        res.end(err);
    })
})

// 资产上传
app.post('/fileUpload',multipartMiddeware,function (req,res) {
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');

    var fileProperty = req.files;
    var filePath = fileProperty.assetFile.path
    var fileType = suffix(fileProperty.assetFile.originalFilename)
    var ownerId = req.body.userId;
    var privateKey = req.body.privateKey;
    var publicKey = req.body.publicKey;
    var assetName = req.body.assetName;
    var fileHash = '';
    var enRanKey = '';
    var rsaFileHash = '';
    console.log("ownerId:" + ownerId);
    console.log("assetName:" + assetName);
    console.log("fileType:" + fileType)
    console.log("测试条件：1公钥:" + publicKey);
    console.log("测试条件：2私钥:" + privateKey)

    new Promise((resolve,reject)=>{
        fs.readFile(filePath,(err,data)=>{
            if(err){
                reject(err);
            } else {
                //未加密文件数据流MD5哈希值
                fileHash = crypto.createHash('md5').update(data).digest('hex');
                resolve(data);
            }
        })
    }).then((data)=>{
        //使用hash加密上传文件
        var encryptedBuf = fileEncrypt(data,fileHash,'rc4')
        //使用RSA加密该密钥，存入数据库
        //enRanKey = crypto.publicEncrypt(publicKey, Buffer.from(ranKey))
        return new Promise((resolve,reject)=>{
            var encryptendFilePath = tempFilePath + randomString()
            console.log("encryptendFilePath" + encryptendFilePath)
            fs.writeFile(encryptendFilePath,encryptedBuf,(err)=>{
                if(err) {
                    reject(err);
                } else {
                    resolve(encryptendFilePath)
                }
            })
        })
    }).then((encryptendFilePath)=>{
        fabricInvoke.invoke('fabcar','fileUpload',[ownerId,assetName,fileType,encryptendFilePath,privateKey,fileHash],(err,data)=>{
            console.log(data);
            res.send(data);
        })
    }).catch((err)=>{
        console.log(err);
    })
})

// 合约发布
// args:[0-Authority,
// 		1-OwnerID,
// 		2-AssetID,
// 		3-ContractName,
// 		4-ExpireDate,
// 		5-AgentContractExpireDate,
// 		6-TradePrice,
// 		7-AgentPrice,
// 		8-privateKey]
app.post('/contractIssue',urlencodedParser,function (req,res) {
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');

    new Promise((resolve,reject)=>{

        fabricInvoke.invoke('fabcar','contractIssue',
            [req.body.authority,
                req.body.userId,
                req.body.assetId,
                req.body.contractName,
                req.body.expireDate,
                req.body.agentContractExpireDate || "",
                req.body.tradePrice,
                req.body.agentPrice || "",
                req.body.privateKey
                //req.body.useType
            ],(err,data)=>{
                if(err) {
                    reject(err);
                } else {
                    resolve(data);
                }
            })
    }).then((data)=>{
        console.log(data);
        res.send(data);
    }).catch((err)=>{
        console.error(err);
    })
})

// trade
// args:[ContractID,BuyerID,EffectiveTime,ContractName,PrivateKey]
app.post('/trade',urlencodedParser,function (req,res) {
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');

    new Promise((resolve,reject)=>{

        fabricInvoke.invoke('fabcar','trade',
            [req.body.contractID,
                req.body.userId,
                req.body.effectiveTime || "",
                req.body.contractName || "",
                req.body.privateKey],(err,data)=>{
                if(err) {
                    reject(err);
                } else {
                    resolve(data);
                }
            })
    }).then((data)=>{
        console.log(data);
        res.send(data);
    }).catch((err)=>{
        console.error(err);
    })
})

// 用户下载文件
app.post('/fileDownload',urlencodedParser,function (req,res) {
    res.header("Access-Control-Allow-Origin", "*");
    res.header('Access-Control-Allow-Methods', 'PUT, GET, POST, DELETE, OPTIONS');
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    var userId = req.body.userId;
    var privateKey = req.body.privateKey;
    var txId = req.body.txId;
    console.log("userId:" + userId);
    console.log("privateKey:" + privateKey)
    console.log("txId:" + txId)

    new Promise((resolve,reject)=>{
        fabricInvoke.invoke('fabcar','assetGet',[userId,txId,privateKey],(err,data)=>{
            if(err) {
                reject(err);
            } else {
                resolve(data);
            }
        })
    }).then((data)=>{
        console.log(data);
        console.log(data["res_message"][0])
        if(data["event_status"] == "INVALID") {
            res.redirect('http://localhost:8001/WEB-INF/downloadError.html');
            //res.send(data)
        } else if(data["event_status"] == "VALID") {
            console.log(data["res_message"]["filePath"]);
            console.log(data["res_message"]["fileName"]);
            res.download(data["res_message"]["filePath"],data["res_message"]["fileName"])
        }
    }).catch((err)=>{
        console.error(err);
    })
})

// from fabric
// 将用户文件上传至IPFS
app.post('/HFassetUpload',urlencodedParser,function (req,res) {

    var filePath = req.body.filePath;
    console.log(filePath);

    new Promise((resolve,reject) => {
        fs.readFile(filePath,(err,data) => {
            if(err) {
                reject(err)
            } else {
                resolve(data)
            }
        })
    }).then((buffer) => {
        return ipfs.add(buffer)
    }).then((hash) => {
        console.log(hash);
        res.send(hash);
    }).catch((err)=>{
        console.error(err);
    })
});

// from fabric
// 用户购买资产时，对原始文件加密后上传至IPFS，返回加密文件hash
app.post('/HFencryptFile',urlencodedParser,function(req,res) {
    var assetHash = req.body.assetHash;
    var key = req.body.key;

    ipfs.get(assetHash).then((buffer) => {
        var encryptedBuf = fileEncrypt(buffer, key,'rc4');
        return ipfs.add(encryptedBuf)
    }).then((hash) => {
        res.send(hash);
    }).catch((err)=>{
        console.error(err);
    })
});


// from fabric
// 用户获取已购买资产时，fabric调用接口
app.post('/HFassetGet',urlencodedParser,function(req,res) {
    var encryptedFileHash = req.body.encryptedFileHash;
    var fileHash = req.body.fileHash;//取的是随机数加密后的字符串
    var key = req.body.key;
    var privateKey = req.body.privateKey;

    ipfs.get(encryptedFileHash).then((buffer) => {
        return new Promise((resolve,reject) => {
            //三段式解密
            var decrypted = fileDecrypt(buffer,'',key,'rc4');
            //使用hash解密下载的文件
            var originFile = fileDecrypt(decrypted,'',fileHash,'rc4')
            //设置文件路径
            var filePath = tempFilePath + randomString();
            //写文件流到设置的路径文件中
            fs.writeFile(filePath,originFile,(err) => {
                if(err) {
                    reject(err)
                } else {
                    resolve(filePath)
                }
            })
        })
    }).then((filePath) => {
        console.log(filePath);
        res.send(filePath);
    }).catch((err)=>{
        console.error(err);
        res.send("FAIL")
    })
});

app.post('/download',urlencodedParser,function(req,res) {
    ipfs.get(req.body.hash).then((buffer) => {
        fs.writeFileSync('./temp/filename',buffer)
        res.download('./temp/filename','filename.file')
    })
});