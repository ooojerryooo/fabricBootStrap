'use strict';
/*
* Copyright IBM Corp All Rights Reserved
*
* SPDX-License-Identifier: Apache-2.0
*/
/*
 * Chaincode Invoke
 */

var Fabric_Client = require('fabric-client');
var path = require('path');
var util = require('util');
var fs = require('fs');
// TLS配置pem路径
var basicnetwork_path = path.resolve('..', 'fabric-docker');
var org1tlscacert_path = path.resolve(basicnetwork_path, 'crypto-config', 'peerOrganizations', 'org1.example.com', 'tlsca', 'tlsca.org1.example.com-cert.pem');
var org1tlscacert = fs.readFileSync(org1tlscacert_path, 'utf8');
//两个组织都执行的情况吧
var org2tlscacert_path = path.resolve(basicnetwork_path, 'crypto-config', 'peerOrganizations', 'org2.example.com', 'tlsca', 'tlsca.org2.example.com-cert.pem');
var org2tlscacert = fs.readFileSync(org2tlscacert_path, 'utf8');

var ordertlscacert_path = path.resolve(basicnetwork_path, 'crypto-config', 'ordererOrganizations', 'example.com', 'orderers','orderer.example.com','msp','tlscacerts', 'tlsca.example.com-cert.pem');
var ordertlscacert = fs.readFileSync(ordertlscacert_path, 'utf8');

var fabric_client = new Fabric_Client();
// 创建一个名字为mychannel的通道
var channel = fabric_client.newChannel('mychannel');
// 创建一个peer,这里使用tls连接，不使用tls的话参数只传一个grpc://localhost:7051,默认不使用tls
var peer = fabric_client.newPeer('grpcs://localhost:7051',{
    'ssl-target-name-override': 'peer0.org1.example.com',
    pem: org1tlscacert
});

var peer2 = fabric_client.newPeer('grpcs://localhost:9051',{
    'ssl-target-name-override': 'peer0.org2.example.com',
    pem: org2tlscacert
});
console.log('peerorg1TLS:%s',org1tlscacert_path);
console.log('peerorg2TLS:%s',org2tlscacert_path);
channel.addPeer(peer);
channel.addPeer(peer2);
// 试试没有order有没有问题,没有order是不行的
// 创建一个order,这里使用tls连接，不使用tls的话参数只传一个grpc://localhost:7050,默认不使用tls
var order = fabric_client.newOrderer('grpcs://localhost:7050',{
    'ssl-target-name-override': 'orderer.example.com',
    pem: ordertlscacert
})
channel.addOrderer(order);

//
var member_user = null;
var store_path = path.join(__dirname, 'hfc-key-store');
console.log('Store path:'+store_path);
var tx_id = null;

exports.invoke = function(ChaincodeID,Function,Args,callback) {
// 根据fabric-client/config/default.json 'key-value-store'配置创建键值对仓库
    Fabric_Client.newDefaultKeyValueStore({
        path: store_path
    }).then((state_store) => {
        // 分配这个仓库给fabric client
        fabric_client.setStateStore(state_store);
        var crypto_suite = Fabric_Client.newCryptoSuite();
        // 对状态存储(保存用户证书的地方)和加密存储(保存用户密钥的地方)使用相同的位置
        var crypto_store = Fabric_Client.newCryptoKeyStore({path: store_path});
        crypto_suite.setCryptoKeyStore(crypto_store);
        fabric_client.setCryptoSuite(crypto_suite);

        // 从持久性获取已注册用户，该用户将签署所有请求
        return fabric_client.getUserContext('user1', true);
    }).then((user_from_store) => {
        if (user_from_store && user_from_store.isEnrolled()) {
            console.log('Successfully loaded user1 from persistence');
            member_user = user_from_store;
        } else {
            throw new Error('Failed to get user1.... run registerUser.js');
        }

        // 根据分配给fabric client的当前用户获取事务id对象
        tx_id = fabric_client.newTransactionID();
        console.log("Assigning transaction_id: ", tx_id._transaction_id);

        // 必须把提案发送给peers背书
        var request = {
            chaincodeId: ChaincodeID,
            fcn: Function,
            args: Args,
            chainId: 'mychannel',
            txId: tx_id
        };

        // send the transaction proposal to the peers
        return channel.sendTransactionProposal(request);
    }).then((results) => {
        var proposalResponses = results[0];
        var proposal = results[1];
        let isProposalGood = false;
        console.log(proposalResponses)
        if (proposalResponses && proposalResponses[0].response &&
            proposalResponses[0].response.status === 200) {
            isProposalGood = true;
            console.log('Transaction proposal was good');
        } else {
            console.error('Transaction proposal was bad');
        }
        if (isProposalGood) {
            console.log(util.format(
                'Successfully sent Proposal and received ProposalResponse: Status - %s, message - "%s"',
                proposalResponses[0].response.status, proposalResponses[0].response.message));

            // 构建orderer提交事务的请求
            var request = {
                proposalResponses: proposalResponses,
                proposal: proposal
            };

            // 设置事务侦听器，并设置超时30秒,如果在超时期间没有提交事务，报告超时状态
            // 获取事件处理要使用的事务ID字符串
            var transaction_id_string = tx_id.getTransactionID();

            var promises = [];

            var sendPromise = channel.sendTransaction(request);
            promises.push(sendPromise); //we want the send transaction first, so that we know where to check status

            // 一旦fabric client分配了用户，就获得一个eventhub。之所以需要用户，是因为必须对事件注册进行签名
            // ssj修改：2019-05-16解决TypeError: fabric_client.newEventHub is not a function错误
            //let event_hub = fabric_client.newEventHub();
            //event_hub.setPeerAddr('grpc://localhost:7053');
            let event_hub = channel.newChannelEventHub(peer);

            // 使用resolve承诺，以便结果状态可以在then子句下处理，而不是让catch子句处理状态
            let txPromise = new Promise((resolve, reject) => {
                let handle = setTimeout(() => {
                    event_hub.disconnect();
                    resolve({event_status: 'TIMEOUT'}); //we could use reject(new Error('Trnasaction did not complete within 30 seconds'));
                }, 3000);
                event_hub.connect();
                event_hub.registerTxEvent(transaction_id_string, (tx, code) => {
                    // this is the callback for transaction event status
                    // first some clean up of event listener
                    clearTimeout(handle);
                    event_hub.unregisterTxEvent(transaction_id_string);
                    event_hub.disconnect();

                    // now let the application know what happened
                    var payloadString = proposalResponses[0].response.payload.toString();
                    var payloadJson = eval('(' + payloadString + ')')
                    var return_status = {event_status: code, tx_id: transaction_id_string, res_message: payloadJson};
                    if (code !== 'VALID') {
                        console.error('The transaction was invalid, code = ' + code);
                        resolve(return_status); // we could use reject(new Error('Problem with the tranaction, event status ::'+code));
                    } else {
                        // ssj修改：2019-05-16解决TypeError: fabric_client.newEventHub is not a function错误
                        //console.log('The transaction has been committed on peer ' + event_hub._ep._endpoint.addr);
                        console.log('The transaction has been committed on peer ' + event_hub.getPeerAddr());
                        resolve(return_status);
                    }
                }, (err) => {
                    //this is the callback if something goes wrong with the event registration or processing
                    reject(new Error('There was a problem with the eventhub ::' + err));
                });
            });
            promises.push(txPromise);

            return Promise.all(promises);
        } else if (proposalResponses) {
            console.error('Receive valid response. Response status is not 200. exiting...')
            /* for fabirc 1.0
            var errorStrWhole = proposalResponses[0].toString();
            var errorStr = errorStrWhole.substring(46,errorStrWhole.indexOf(')'));
            */
            var errorStr = proposalResponses[0].response.message;
            var errorJson = [{"status":"UNSUCCESS"},{"event_status":"INVALID",tx_id: transaction_id_string,"res_message":{"message":errorStr}}]
            return Promise.resolve(errorJson)
        } else {
            console.error('Failed to send Proposal or receive valid response. Response null or status is not 200. exiting...');
            throw new Error('Failed to send Proposal or receive valid response. Response null or status is not 200. exiting...');
        }
    }).then((results) => {

        console.log('Send transaction promise and event listener promise have completed');
        // check the results in the order the promises were added to the promise all list
        if (results && results[0] && results[0].status === 'SUCCESS') {
            console.log('Successfully sent transaction to the orderer.');
        } else {
            console.error('Failed to order the transaction.');
        }

        if (results && results[1] && results[1].event_status === 'VALID') {
            console.log('Successfully committed the change to the ledger by the peer');
        } else {
            console.log('Transaction failed to be committed to the ledger due to ::' + results[1].event_status);
        }

        if (results && results[0] && results[1]) {
            callback(null,results[1])
        } else {
            callback(null,null)
        }

    }).catch((err) => {
        console.error('Failed to invoke successfully :: ' + err);
    });
}