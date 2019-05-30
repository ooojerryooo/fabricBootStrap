'use strict';
/*
* Copyright IBM Corp All Rights Reserved
*
* SPDX-License-Identifier: Apache-2.0
*/
/*
 * Chaincode query
 */

var Fabric_Client = require('fabric-client');
var path = require('path');
var fs = require('fs');
// TLS配置pem路径
var basicnetwork_path = path.resolve('..', 'fabric-docker');
var org1tlscacert_path = path.resolve(basicnetwork_path, 'crypto-config', 'peerOrganizations', 'org1.example.com', 'tlsca', 'tlsca.org1.example.com-cert.pem');
var org1tlscacert = fs.readFileSync(org1tlscacert_path, 'utf8');

//
var fabric_client = new Fabric_Client();

// setup the fabric network
var channel = fabric_client.newChannel('mychannel');
// 创建一个peer,这里使用tls连接，不使用tls的话参数只传一个grpc://localhost:7051,默认不使用tls
var peer = fabric_client.newPeer('grpcs://localhost:7051',{
    'ssl-target-name-override': 'peer0.org1.example.com',
    pem: org1tlscacert
});
channel.addPeer(peer);

//
var member_user = null;
var store_path = path.join(__dirname, 'hfc-key-store');
console.log('Store path:'+store_path);
var tx_id = null;

exports.query = function(ChaincodeID,Function,Args,callback) {
// create the key value store as defined in the fabric-client/config/default.json 'key-value-store' setting
    Fabric_Client.newDefaultKeyValueStore({
        path: store_path
    }).then((state_store) => {
        // assign the store to the fabric client
        fabric_client.setStateStore(state_store);
        var crypto_suite = Fabric_Client.newCryptoSuite();
        // use the same location for the state store (where the users' certificate are kept)
        // and the crypto store (where the users' keys are kept)
        var crypto_store = Fabric_Client.newCryptoKeyStore({path: store_path});
        crypto_suite.setCryptoKeyStore(crypto_store);
        fabric_client.setCryptoSuite(crypto_suite);

        // get the enrolled user from persistence, this user will sign all requests
        return fabric_client.getUserContext('user1', true);
    }).then((user_from_store) => {
        if (user_from_store && user_from_store.isEnrolled()) {
            console.log('Successfully loaded user1 from persistence');
            member_user = user_from_store;
        } else {
            throw new Error('Failed to get user1.... run registerUser.js');
        }

        const request = {
            chaincodeId: ChaincodeID,
            fcn: Function,
            args: Args
        };

        // send the query proposal to the peer
        return channel.queryByChaincode(request);
    }).then((query_responses) => {
    	callback(null,query_responses[0].toString())
    }).catch((err) => {
        console.error('Failed to query successfully :: ' + err);
    })
}
;
