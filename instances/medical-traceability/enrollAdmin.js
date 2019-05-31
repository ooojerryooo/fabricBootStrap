'use strict';

var Fabric_Client = require('fabric-client');
var Fabric_CA_Client = require('fabric-ca-client');

var path = require('path');
var os = require('os');
var fs = require('fs');

var firstnetwork_path = path.resolve('..','fabric-docker');
var org1tlscacert_path = path.resolve(firstnetwork_path, 'crypto-config', 'peerOrganizations', 'org1.example.com', 'tlsca', 'tlsca.org1.example.com-cert.pem');
var org1tlscacert = fs.readFileSync(org1tlscacert_path, 'utf8');
console.log("Org1的tlsca证书路径是:%s",org1tlscacert_path)
//
var fabric_client = new Fabric_Client();
var fabric_ca_client = null;
var admin_user = null;
var store_path = path.join(__dirname, 'hfc-key-store');
console.log(' Store path:'+store_path);

Fabric_Client.newDefaultKeyValueStore({ path: store_path
}).then((state_store) => {
    fabric_client.setStateStore(state_store);
    var crypto_suite = Fabric_Client.newCryptoSuite();
    var crypto_store = Fabric_Client.newCryptoKeyStore({path: store_path});
    crypto_suite.setCryptoKeyStore(crypto_store);
    fabric_client.setCryptoSuite(crypto_suite);
    var	tlsOptions = {
    	trustedRoots: [org1tlscacert],
    	verify: false
    };
    fabric_ca_client = new Fabric_CA_Client('https://localhost:7054', tlsOptions , 'ca-org1', crypto_suite);

    return fabric_client.getUserContext('admin', true);
}).then((user_from_store) => {
    if (user_from_store && user_from_store.isEnrolled()) {
        console.log('从持久层成功加载admin');
        admin_user = user_from_store;
        return null;
    } else {
        return fabric_ca_client.enroll({
          enrollmentID: 'admin',
          enrollmentSecret: 'adminpw'
        }).then((enrollment) => {
          console.log('成功注册管理用户“admin”');
          return fabric_client.createUser(
              {username: 'admin',
                  mspid: 'Org1MSP',
                  cryptoContent: { privateKeyPEM: enrollment.key.toBytes(), signedCertPEM: enrollment.certificate }
              });
        }).then((user) => {
          admin_user = user;
          return fabric_client.setUserContext(admin_user);
        }).catch((err) => {
          console.error('注册和持久化admin失败。错误原因: ' + err.stack ? err.stack : err);
          throw new Error('登记admin失败');
        });
    }
}).then(() => {
    console.log('将admin用户分配给fabric客户机 ::' + admin_user.toString());
}).catch((err) => {
    console.error('登记admin失败: ' + err);
});
