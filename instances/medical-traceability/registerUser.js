'use strict';

var Fabric_Client = require('fabric-client');
var Fabric_CA_Client = require('fabric-ca-client');

var path = require('path');
var os = require('os');
var fs = require('fs');
var firstnetwork_path = path.resolve('..', 'fabric-docker');
var org1tlscacert_path = path.resolve(firstnetwork_path, 'crypto-config', 'peerOrganizations', 'org1.example.com', 'tlsca', 'tlsca.org1.example.com-cert.pem');
var org1tlscacert = fs.readFileSync(org1tlscacert_path, 'utf8');
console.log("Org1的tlsca证书路径是:%s", org1tlscacert_path)

//
var fabric_client = new Fabric_Client();
var fabric_ca_client = null;
var admin_user = null;
var member_user = null;
var store_path = path.join(__dirname, 'hfc-key-store');
console.log(' Store path:' + store_path);

Fabric_Client.newDefaultKeyValueStore({
    path: store_path
}).then((state_store) => {
    fabric_client.setStateStore(state_store);
    var crypto_suite = Fabric_Client.newCryptoSuite();
    var crypto_store = Fabric_Client.newCryptoKeyStore({path: store_path});
    crypto_suite.setCryptoKeyStore(crypto_store);
    fabric_client.setCryptoSuite(crypto_suite);
    var tlsOptions = {
        trustedRoots: [org1tlscacert],
        verify: false
    };
    fabric_ca_client = new Fabric_CA_Client('https://localhost:7054', tlsOptions, 'ca-org1', crypto_suite);
    return fabric_client.getUserContext('admin', true);
}).then((user_from_store) => {
    if (user_from_store && user_from_store.isEnrolled()) {
        console.log('从持久层成功加载admin');
        admin_user = user_from_store;
    } else {
        throw new Error('运行registerAdmin.js无法获得admin');
    }
    return fabric_ca_client.register({enrollmentID: 'user2', affiliation: 'org1.department1'}, admin_user);
}).then((secret) => {
    console.log('成功注册user2 - 秘钥:' + secret);
    return fabric_ca_client.enroll({enrollmentID: 'user2', enrollmentSecret: secret});
}).then((enrollment) => {
    console.log('成功登记会员用户 "user2" ');
    return fabric_client.createUser(
        {
            username: 'user2',
            mspid: 'Org1MSP',
            cryptoContent: {privateKeyPEM: enrollment.key.toBytes(), signedCertPEM: enrollment.certificate}
        });
}).then((user) => {
    member_user = user;
    return fabric_client.setUserContext(member_user);
}).then(() => {
    console.log('user2已成功注册并注册，并准备对fabric网络进行研究');
}).catch((err) => {
    console.error('注册失败: ' + err);
    if (err.toString().indexOf('Authorization') > -1) {
        console.error('授权失败可能是由于具有来自以前CA实例的管理凭据造成的。删除存储目录的内容后再试一次 ' + store_path);
    }
});
