var ipfsAPI = require('ipfs-api');
var ipfs = ipfsAPI({host: 'localhost', port: '5001', protocol: 'http'});

exports.add = function (buffer) {
    return new Promise(function (resolve,reject) {
        try {
            ipfs.add(buffer, function (err, files) {
                if (err || typeof files == "undefined") {
                    reject(err);
                } else {
                    resolve(files[0].hash);
                }
            })
        }catch(err) {
            reject(err);
        }
    })
}
exports.get = function (hash) {
    return new Promise(function (resolve,reject) {
        try{
            ipfs.get(hash,function (err,files) {
                if (err || typeof files == "undefined") {
                    reject(err);
                }else{
                    resolve(files[0].content);
                }
            })
        }catch (err){
            reject(err);
        }
    })
}
