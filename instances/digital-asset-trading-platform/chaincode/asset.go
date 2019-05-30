package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	marand "math/rand"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"

	"io/ioutil"
	"net/http"
	"net/url"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// User
type User struct {
	ObjectType string `json:"docType"`
	UserID string `json:"UserID"`
	UserName string `json:"UserName"`
	LoginPwd string `json:"LoginPwd"`
	Mobile string `json:"Mobile"`
	PrivateKey string `json:"PrivateKey"`
	PublicKey string `json:"PublicKey"`
}

// UserAccount
type UserAccount struct {
	ObjectType string `json:"docType"`
	AccountID string `json:"AccountID"`
	UserID string `json:"UserID"`
	Balance float64 `json:"Balance,string"`
}

// Asset
type Asset struct {
	ObjectType string `json:"docType"`
	AssetID string `json:"AssetID"`
	OwnerID string `json:"OwnerID"`
	AssetName string `json:"AssetName"`
	FileType string `json:"FileType"`  // 扩展名
	AssetHash string `json:"AssetHash"`  // 将加密后的文件上传至IPFS，由IPFS生成的hash
	FileHash string `json:"FileHash"`    // 未加密文件的hash
}

// Contract
type Contract struct {
	ObjectType string `json:"docType"`
	ContractID string `json:"ContractID"`
	AssetID string `json:"AssetID"`
	AssetName string `json:"AssetName"`
	AssetHash string `json:"AssetHash"`
	FileHash string `json:"FileHash"`
	OwnerID string `json:"OwnerID"`
	AgnetID string `json:"AgentID"`
	ContractName string `json:"ContractName"`
	Authority string `json:"Authority"`  // sell-卖合约，view-卖资产
	UseType string `json:"UseType"`  // 1-拷贝，2-播放，3-翻录
	IsAgent string `json:"IsAgent"`
	ExpireDate string `json:"ExpireDate"`  // 当前合约失效日期,yyyy-mm-dd
	AgentContractExpireDate string `json:"AgentContractExpireDate"`  // 代理合约失效日期,yyyy-mm-dd
	TradePrice float64 `json:"TradePrice,string"`
	AgentPrice float64 `json:"AgentPrice,string"`
}

// Transaction
type Transaction struct {
	ObjectType string `json:"docType"`
	TxID string `json:"TxID"`    // 交易流水号
	AgreementID string `json:"AgreementID"`  // 版权转让协议号
	OwnerID string `json:"OwnerID"`
	AgentID string `json:"SellerID"`  // IsDistributed == 'true'时使用
	BuyerID string `json:"BuyerID"`
	AssetID string `json:"AssetID"`
	ContractID string `json:"ContractID"`
	AgentContractID string `json:"AgentContractID"` // 买合约时(Authority=='sell')使用,写新生成的代理商合约ID
	IsDistributed string `json:"IsDistributed"` //当前交易是否经由代理商。从代理商处买资产为"true",其他为"false"
	TxDate string `json:"TxDate"`
	Authority string `json:"Authority"`  // sell-买合约，view-买资产
	UseType string `json:"UseType"`  // 1-拷贝，2-播放，3-翻录
	BuyerKey string `json:"BuyerKey"`
	AgentKey string `json:"AgentKey"`
	SellerKey string `json:"SellerKey"`
	AssetHash string `json:"AssetHash"`
	FileHash string `json:"FileHash"`
	EncryptedAssetHash string `json:"EncryptedAssetHash"`  // authority!='sell'时使用
	OriginPrice float64 `json:"OriginPrice,string"`  // authority=='sell'时使用
	AgentPrice float64 `json:"AgentPrice,string"`  // authority=='sell'时使用
	TradePrice float64 `json:"TradePrice,string"`  // authority!='sell'时使用
	EffectiveTime string `json:"EffectiveTime"`  //有效期，单位：年
	EffectiveDate string `json:"EffectiveDate"`  // 生效日期
	ExpireDate string `json:"ExpireDate"`  // 失效日期
}

// prefix of key
const User_Prefix = "User_"
const Account_Prefix = "Acco_"
const Asset_Prefix = "Asset_"
const Contract_Prefix = "Contract_"
const Tx_Prefix = "Tx_"

const appUrlPrefix = "http://114.115.238.2:8002/"

func GenRsaKey(bits int) (string ,string) {
	// 生成私钥文件
	privateKey,_ := rsa.GenerateKey(rand.Reader, bits)
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	//err = pem.Encode(file, block)

	pemRsaPrvbyte:=pem.EncodeToMemory(block)
	pemRsaPrvstring:= string(pemRsaPrvbyte)
	fmt.Println("生成私钥成功！\n" + pemRsaPrvstring +"\n")

	// 生成公钥文件
	publicKey := &privateKey.PublicKey
	derPkix, _ := x509.MarshalPKIXPublicKey(publicKey)
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pemRsaPlcbyte:=pem.EncodeToMemory(block)
	pemRsaPlcstring:= string(pemRsaPlcbyte)
	fmt.Println("生成公钥成功！\n"+pemRsaPlcstring +"\n")

	return pemRsaPrvstring,pemRsaPlcstring
}



// 加密
func RsaEncrypt(origData []byte,publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// 解密
func RsaDecrypt(ciphertext []byte,privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}


// RandomStr 生成随机字符串
func RandomStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	byteData := []byte(str)
	result := []byte{}
	r := marand.New(marand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, byteData[r.Intn(len(byteData))])
	}
	return string(result)
}

// Md5
func  GetMd5(str string) string {
	x := md5.New();
	x.Write([]byte(str))
	y := x.Sum([]byte(""))
	return fmt.Sprintf("%x", y)
}

// 校验用户密钥
func (s *SmartContract) CheckPrivateKey(APIstub shim.ChaincodeStubInterface, userID string, privateKey string) bool {
	userAsBytes, err := APIstub.GetState(User_Prefix + userID)
	if err != nil {
		return  false
	}
	user := User{}
	json.Unmarshal(userAsBytes, &user)
	fmt.Println("111111111111111:"+user.PrivateKey)
	fmt.Println("222222222222222:"+privateKey)
	if user.PrivateKey == privateKey {
		return  true
	} else {
		return false
	}
}

//couchdb拼查询条件
func queryString(a string,b string) string{
	queryString :="{\""+a+"\":\""+b+"\"}}"
	return queryString
}

//结果集拼json
func obj2json(resultsIterator shim.StateQueryIteratorInterface) (bytes.Buffer,error){
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	var err1 error
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		err1 = err
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{Key:\""+queryResponse.Key+"\", Record:"+string(queryResponse.Value)+"}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return buffer,err1
}
// 富查询，返回JSON数组:[{"key1":value1},{"key2":value2},...].其中,value是JSON对象.
func (s *SmartContract) RichQuery(APIstub shim.ChaincodeStubInterface, queryString string) (result []byte,err error) {
	resultsiterator, err := APIstub.GetQueryResult(queryString)
	if err != nil {
		return nil,err
	}
	defer resultsiterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsiterator.HasNext() {
		queryResponse, err := resultsiterator.Next()
		if err != nil {
			return nil, err
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString ( "," )
		}
		buffer.WriteString ( "{\"Key\":" )
		buffer.WriteString ("\"")
		buffer.WriteString( queryResponse.Key)
		buffer.WriteString ("\"")
		buffer.WriteString (",\"Record\":")
		buffer. WriteString (string(queryResponse.Value)) // Value是JSON 对象
		buffer.WriteString ("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return buffer.Bytes(),nil
}
/*
 * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "getAssentForOwner" {
		return s.getAssentForOwner(APIstub, args)
	} else if function == "getContractForOwner" {
		return s.getContractForOwner(APIstub, args)
	} else if function == "getContractForOwner2" {
		return s.getContractForOwner2(APIstub, args)
	} else if function == "getContractForOwner3" {
		return s.getContractForOwner3(APIstub, args)
	} else if function == "getAssentForBuyer" {
		return s.getAssentForBuyer(APIstub, args)
	} else if function == "getTransactionForUser" {
		return s.getTransactionForUser(APIstub, args)
	} else if function == "getAccountForUser" {
		return s.getAccountForUser(APIstub, args)
	} else if function == "getContractList" {
		return s.getContractList(APIstub, args)
	} else if function == "fileUpload" {
		return s.fileUpload(APIstub, args)
	}else if function == "register" {
		return s.register(APIstub, args)
	} else if function == "login" {
		return s.login(APIstub, args)
	} else if function == "trade" {
		return s.trade(APIstub, args)
	} else if function == "contractIssue" {
		return s.contractIssue(APIstub, args)
	} else if function == "assetGet" {
		return s.assetGet(APIstub, args)
	}else if function == "getContractByid" {
		return s.getContractByid(APIstub, args)
	}else if function == "getAssetByid" {
		return s.getAssetByid(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	//初始化用户
	user := User{ObjectType:"User",UserID:"ssj",UserName:"孙世江",
	LoginPwd:"202cb962ac59075b964b07152d234b70",Mobile:"15101046074",PrivateKey:"-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQDjF9ZT1yl4ph8Q3iqjRkpTJJOA5SXASG4LmPOvmOTl2rbAgAsD\nOgKXM5GshA/JbRG6DV+0m3YYsAdR1NKtxHxNwnQZY/7y2Sh9WowAZ1is7BIOeKGt\naHcm9P3Bln5B/zwAAMnJtWaWBB3459CHEsWx/2woNwHurz+g1Jq5ZsncbQIDAQAB\nAoGAdumLfDllaNyr1bxP3VE4KBM1/b9/phEHNqFvsNpBua5uFZy2p4jfxGbnd8tr\nLNElACRyofLpwwGhw5LKgt0pLN4KNgosUExTY0UCBC0NuKeeTbVDZvYIWdAEyf9Z\nqnfxcnMDBGBuWvKzKHnwZDKGze1eD9UNfvih03sdg3LyuyECQQDvbPbLnfsakXTR\nJl89dsVh8LiNT8tBWIa0VZzo9evP7XfOheAahI/vv/N2jrHcDBVUfZNOZRrZb4gM\nA0PbG/4lAkEA8tBRyngW/TPycfLSKNECklJK0JFnBks+m5K2AoKlaVKd+ciATFzH\nxj8wN23nmln5I4sFWaoRjLcZFRewn8/eqQJAQ7EaAa2Tzgor7eocaUNIQQ2PEBzs\nSXhe9WNzLoZA+pOYGOLO3pB2JYGusulbLeyYpz+twy9grPMUktjleXBrpQJBAIgi\n1Di4a+s6nMvWYI4R4Wc4TEnzu9eDiG6hKvrtVOBgOcI57/TgcAFunBH3xMp9d3m+\nEBndiElkvuNcjOKNIUECQQCk21ovKWGxyMUI2H9fLXOHxBLrhKzqn5Vwx8VkQJNr\n8mTo/+E7bbk7KV8zlZ6/MqgcsgXbH6nXAXrLL3LhPVit\n-----END RSA PRIVATE KEY-----\n",PublicKey:"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDjF9ZT1yl4ph8Q3iqjRkpTJJOA\n5SXASG4LmPOvmOTl2rbAgAsDOgKXM5GshA/JbRG6DV+0m3YYsAdR1NKtxHxNwnQZ\nY/7y2Sh9WowAZ1is7BIOeKGtaHcm9P3Bln5B/zwAAMnJtWaWBB3459CHEsWx/2wo\nNwHurz+g1Jq5ZsncbQIDAQAB\n-----END PUBLIC KEY-----\n"}
	userAsBytes, _ := json.Marshal(user)
	APIstub.PutState(User_Prefix+"ssj", userAsBytes)
	user2 := User{ObjectType:"User",UserID:"agent",UserName:"代理商",
		LoginPwd:"202cb962ac59075b964b07152d234b70",Mobile:"11312321312",PrivateKey:"-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQDPzDgZBl7kuZd32OQOKtLGXHT23O6azZQvQNhpT7Hh22povhBa\ncpeZb/zCkItKgGTzWnn/RTAvjxg5giINsXlmbl9+gwKSNclVDb/2fWwaL0Xpldai\n2o4owpuXh1f7vjE17Osc0sfMFAZFknzH8G04mAuZdU3iA/LeYV3hFVjnoQIDAQAB\nAoGAcYoTaNELTox9z7VJvtf1PP9NxYkoMpt7iWo0FS6/cjoyZTLcup78aJFBGYpj\nIX0QK8vW9vz1/DaCtQF+296KBsq6E4H+hqfnjm/M5ga7Keg75PoE+AVXGn2or+qk\nu695kNCin81GvyP0TptesIT8UG8OiDf7OJoTmfBnDuWAj5UCQQDu4wRJI4NiQYV7\nxml9jsMbdE6kW2sh7Uh5kZTmct5bnvbECazTNgI21j+IFPBSQJIJlvzHsArKknz6\n0wgNCMArAkEA3q8OI6M3Pm+NO1GipkABs2MD5iLcxNaNZNVTKs9D1MLYBN/52MoG\n7RV30eeW2ZHgxo3lbiiVy0Ij0ZXBTQFFYwJAY6rFq/osI79wgL68Xo1Eq9yvvvUB\nUqtcNJRfynIcgZ9mF05uE8UR0W08HnuV9MtJ4sRi/LFfHztU95U6Y63F9wJAIEMu\ngj1IaLRSuvBl5z5IwMusqfANGjuXeq9pgD9NLYsZLwOgOCd0/25n0LKD6Xu4HCw8\nvEwG/87ST7AptFVlzwJBAM+EHFzXUuW6ZKc5Vx1vqYbdiskUb4wmrueV6mrGT2Jc\nKGRa6SNhezkZziQSqbe1OCdJHYFY72F2JgopY9W0Ik0=\n-----END RSA PRIVATE KEY-----\n",PublicKey:"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDPzDgZBl7kuZd32OQOKtLGXHT2\n3O6azZQvQNhpT7Hh22povhBacpeZb/zCkItKgGTzWnn/RTAvjxg5giINsXlmbl9+\ngwKSNclVDb/2fWwaL0Xpldai2o4owpuXh1f7vjE17Osc0sfMFAZFknzH8G04mAuZ\ndU3iA/LeYV3hFVjnoQIDAQAB\n-----END PUBLIC KEY-----\n"}
	user2AsBytes, _ := json.Marshal(user2)
	APIstub.PutState(User_Prefix+"agent", user2AsBytes)
	buyer1 := User{ObjectType:"User",UserID:"buyer1",UserName:"买家1",
		LoginPwd:"202cb962ac59075b964b07152d234b70",Mobile:"15101034332",PrivateKey:"-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQDBSi7xUx1v2QveiuBrjGBy6PqGGyjU9RZLeS8xOpJilqyx/Z9M\n8u8UmiWIB8JdEaYWtGw+2qK8wsHcULKDfEbgKYL+zxH2MNYlbUAU9Zt324UhBYEl\nMJjqiGmsw/OeE+hsuGVmKadTnfLOGazujVpMPaw7DJUvookQZKgQTw2sCwIDAQAB\nAoGBAKJ0YSP/QVyCRhGTE7AQ2fd6jSUtqpHiNAzXG6z6l5I5JYirq7oW7I6aSDUe\noGNss/WdSuVBCUoxPpKXWJJeX4oMOLq74W8IiVKVPN+xdeg4kGD6xKk5lrHS7biU\nIRy+8eMqwRbV45UF3v89hScag8Ub24bTW3DwYokmelcagt+RAkEAyNXytofwrOP5\n0BftBqYFAZMYeN/krODN0gL4UKEwdHAig/5vPk1ltxqo708ZxD6D88xFCxY9+R1E\nFPv0MRpvnQJBAPZhoW4jXrsvlossUOCb2naO2vqQ93/TKfzjmRFCO1qiv/2rsgS+\ncZb/iIBRVeN7/WJ586/t0O8g49UunelivccCQGJcLjfpUh1KthGNdj+YVcFUqlqg\nxN1KaGMfoz/SAeo09SKSHSd1Poiz3OL/aY4sU/G2LGZmqUl1ZN+mGg1mdh0CQCh8\nOHuRolJd6n3qBUwzL/3FUaRUx+agO0kL2S4l1Pz6u1Oir/jpll66lKKJOvTLfgJ9\niZaCHI/+tpFkPJyKFicCQFnoj+jYnFY6EXmZ4lk0TLH3VoGmv7b12SG8TIITEGHQ\nPY5BLCg8On/Brizevbs29Ff82hFW8CcmdkANlH3G+Qg=\n-----END RSA PRIVATE KEY-----\n",PublicKey:"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDBSi7xUx1v2QveiuBrjGBy6PqG\nGyjU9RZLeS8xOpJilqyx/Z9M8u8UmiWIB8JdEaYWtGw+2qK8wsHcULKDfEbgKYL+\nzxH2MNYlbUAU9Zt324UhBYElMJjqiGmsw/OeE+hsuGVmKadTnfLOGazujVpMPaw7\nDJUvookQZKgQTw2sCwIDAQAB\n-----END PUBLIC KEY-----\n"}
	buyer1AsBytes, _ := json.Marshal(buyer1)
	APIstub.PutState(User_Prefix+"buyer1", buyer1AsBytes)
	buyer2 := User{ObjectType:"User",UserID:"buyer2",UserName:"买家2",
		LoginPwd:"202cb962ac59075b964b07152d234b70",Mobile:"15101034332",PrivateKey:"-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQDt1kVnPCSocuQ++9vUcmxgpPMvx/mJreV7jRpl8PdMsRTEAP79\neoCCB1ew4ItWvJRnS29j/jKhPBx8Wg9B6v4i/Uv+UZ3iHvMvdj7kuo7uEhUKflPz\nehXcKvwNo66q9NAlotbapDkBiIBu7o7cemurJ+dPyGRwDcftlsKV5aeyGQIDAQAB\nAoGBAL+DHne2cj5B5ZDh9W3ohWR87mW3PTSkFnmacXtMwZW7seDyfGFX10rc5rvC\n0/CQAV/72kJGWjaC1N09F7QYqXGtrp80fifXxgWD8Plq8ReIyaAQjEoUdlgwv14J\nnIoU+fAL67drdeXd6SRa3Dz2niyX1Wr8PilPJ3wOIrm/hJVRAkEA7nTiuGr5YGE3\nMdjHxjw4ccincU364da4QWhfH3E6LHuDnrKP+hha6PpqufeDu1Wc4lpSpIPVc1eM\n411xrKh5NwJBAP9Vt1JY8JgBRCXJTQnZf6dlC6m/48A8mFyRiKJ/g3EqrQxU3RfG\nQnClzxiWiFTGSe4RFEMN7LcxM48EoBuWly8CQHJBm2jWCQt/SV0fDAoWTM1oYaLO\nxIl1wu/EPN/p9v/dZuGhmY8yIE1Fv+G/kWUvzm4+7R5a9OnBZ4aB/bfHOd0CQHcV\nHfN84XCzHnpVAOX4Fy4V1TOs9+Y/HHwHr+bBe6b61UwsBBVDdNcerZB1HE4VUIOE\nWaPQSbdCbh5kdNuJBycCQQCpv2NCDvb09NC/yGibP3LNtMZBrnMWr/ONwOy+qHNj\ntpDazJrJYQnDyD1mMibqnm6YE8QC/7vwHdRIg3rnTwMO\n-----END RSA PRIVATE KEY-----\n",PublicKey:"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDt1kVnPCSocuQ++9vUcmxgpPMv\nx/mJreV7jRpl8PdMsRTEAP79eoCCB1ew4ItWvJRnS29j/jKhPBx8Wg9B6v4i/Uv+\nUZ3iHvMvdj7kuo7uEhUKflPzehXcKvwNo66q9NAlotbapDkBiIBu7o7cemurJ+dP\nyGRwDcftlsKV5aeyGQIDAQAB\n-----END PUBLIC KEY-----\n"}
	buyer2AsBytes, _ := json.Marshal(buyer2)
	APIstub.PutState(User_Prefix+"buyer2", buyer2AsBytes)
	//初始化账户
	ua := UserAccount{ObjectType:"UserAccount",AccountID:"Acco_ssj",UserID:"ssj",Balance:10000}
	uaAsBytes, _ := json.Marshal(ua)
	APIstub.PutState(Account_Prefix+"ssj", uaAsBytes)
	ua2 := UserAccount{ObjectType:"UserAccount",AccountID:"Acco_agent",UserID:"agent",Balance:10000}
	ua2AsBytes, _ := json.Marshal(ua2)
	APIstub.PutState(Account_Prefix+"agent", ua2AsBytes)
	ua3 := UserAccount{ObjectType:"UserAccount",AccountID:"Acco_buyer1",UserID:"buyer1",Balance:10000}
	ua3AsBytes, _ := json.Marshal(ua3)
	APIstub.PutState(Account_Prefix+"buyer1", ua3AsBytes)
	ua4 := UserAccount{ObjectType:"UserAccount",AccountID:"Acco_buyer2",UserID:"buyer2",Balance:10000}
	ua4AsBytes, _ := json.Marshal(ua4)
	APIstub.PutState(Account_Prefix+"buyer2", ua4AsBytes)

	//初始化文件
	assets := []Asset{
		Asset{ObjectType:"Asset",AssetID: "Asset_1", OwnerID: "ssj", AssetName: "音乐", FileType: "mp3",AssetHash:"567891fghjk"},
		Asset{ObjectType:"Asset",AssetID: "Asset_2", OwnerID: "ssj", AssetName: "影视", FileType: "mp4",AssetHash:"567892fghjk"},
		Asset{ObjectType:"Asset",AssetID: "Asset_3", OwnerID: "ssj", AssetName: "文件", FileType: "txt",AssetHash:"567893fghjk"},
		Asset{ObjectType:"Asset",AssetID: "Asset_4", OwnerID: "ssj", AssetName: "word", FileType: "doc",AssetHash:"567894fghjk"},
		Asset{ObjectType:"Asset",AssetID: "Asset_5", OwnerID: "ssj", AssetName: "excel", FileType: "xls",AssetHash:"567895fghjk"},
	}
	i := 0
	for i < len(assets) {
		fmt.Println("i is ", i)
		assentAsBytes, _ := json.Marshal(assets[i])
		APIstub.PutState(Asset_Prefix+strconv.Itoa(i), assentAsBytes)
		fmt.Println("Added", assets[i])
		i = i + 1
	}
	//初始化合约
	contracts := []Contract{
		Contract{ObjectType:"Contract",ContractID:"Contract_1",AssetID:"Asset_1",AssetName:"成都.mp3",AssetHash:"567891fghjk",OwnerID:"ssj",
			AgnetID:"",ContractName:"直销合约01",Authority:"view",IsAgent:"true",ExpireDate:"2019-08-19",
			AgentContractExpireDate:"",TradePrice:90,AgentPrice:0.0},
		Contract{ObjectType:"Contract",ContractID:"Contract_2",AssetID:"Asset_1",AssetName:"地球的起源.avi",AssetHash:"567892fghjk",OwnerID:"ssj",
			AgnetID:"",ContractName:"直销合约02",Authority:"view",IsAgent:"true",ExpireDate:"2019-08-19",
			AgentContractExpireDate:"",TradePrice:91,AgentPrice:0.0},
		Contract{ObjectType:"Contract",ContractID:"Contract_3",AssetID:"Asset_2",AssetName:"大闹天宫.avi",AssetHash:"567893fghjk",OwnerID:"ssj",
			AgnetID:"",ContractName:"直销合约03",Authority:"view",IsAgent:"true",ExpireDate:"2019-08-19",
			AgentContractExpireDate:"",TradePrice:92,AgentPrice:0.0},
		Contract{ObjectType:"Contract",ContractID:"Contract_4",AssetID:"Asset_2",AssetName:"我们的世界.mp4",AssetHash:"567894fghjk",OwnerID:"ssj",
			AgnetID:"",ContractName:"直销合约04",Authority:"view",IsAgent:"false",ExpireDate:"2019-08-19",
			AgentContractExpireDate:"",TradePrice:93,AgentPrice:0.0},
		Contract{ObjectType:"Contract",ContractID:"Contract_5",AssetID:"Asset_3",AssetName:"谁知道呢.mp3",AssetHash:"567895fghjk",OwnerID:"ssj",
			AgnetID:"",ContractName:"直销代理合约01",Authority:"sell",IsAgent:"false",ExpireDate:"2019-08-19",
			AgentContractExpireDate:"",TradePrice:94,AgentPrice:98},
		Contract{ObjectType:"Contract",ContractID:"Contract_6",AssetID:"Asset_4",AssetName:"走南闯北.mp3",AssetHash:"567896fghjk",OwnerID:"ssj",
			AgnetID:"",ContractName:"直销代理合约02",Authority:"sell",IsAgent:"false",ExpireDate:"2019-08-19",
			AgentContractExpireDate:"",TradePrice:94,AgentPrice:98},
		Contract{ObjectType:"Contract",ContractID:"Contract_7",AssetID:"Asset_5",AssetName:"谁知道呢.mp3",AssetHash:"567896fghjk",OwnerID:"ssj",
			AgnetID:"ssj",ContractName:"代理生成合约02",Authority:"view",IsAgent:"true",ExpireDate:"2019-08-19",
			AgentContractExpireDate:"",TradePrice:94,AgentPrice:98},
	}
	j := 0
	for j < len(contracts) {
		fmt.Println("i is ", j)
		contractAsBytes, _ := json.Marshal(contracts[j])
		APIstub.PutState(Contract_Prefix+strconv.Itoa(j), contractAsBytes)
		fmt.Println("Added", contracts[j])
		j = j + 1
	}

	//初始化交易
	transactions := []Transaction{
		Transaction{ObjectType:"Transaction",TxID:"Tx_1",AgreementID:"1",OwnerID:"ssj",BuyerID:"zq",AgentID:"",
			AssetID:"Asset_1",ContractID:"Contract_1",AgentContractID:"",
			IsDistributed:"false",TxDate:"2019-03-27",Authority:"view",BuyerKey:"123",AgentKey:"",SellerKey:"789",
			AssetHash:"567892fghjk",EncryptedAssetHash:"667891fghjk",OriginPrice:91,AgentPrice:0.0,
			TradePrice:91,EffectiveTime:"2",EffectiveDate:"2017-03-27",ExpireDate:"2019-03-26"},
		Transaction{ObjectType:"Transaction",TxID:"Tx_2",AgreementID:"2",OwnerID:"ssj",BuyerID:"pc",AgentID:"",
			AssetID:"Asset_1",ContractID:"Contract_2",AgentContractID:"",
			IsDistributed:"false",TxDate:"2019-03-27",Authority:"view",BuyerKey:"123",AgentKey:"",SellerKey:"789",
			AssetHash:"567892fghjk",EncryptedAssetHash:"667891fghjk",OriginPrice:91,AgentPrice:0.0,
			TradePrice:91,EffectiveTime:"2",EffectiveDate:"2019-03-27",ExpireDate:"2021-03-26"},
		Transaction{ObjectType:"Transaction",TxID:"Tx_3",AgreementID:"3",OwnerID:"ssj",BuyerID:"dl",AgentID:"ssj",
			AssetID:"Asset_2",ContractID:"Contract_3",AgentContractID:"5",
			IsDistributed:"true",TxDate:"2019-03-27",Authority:"view",BuyerKey:"123",AgentKey:"456",SellerKey:"789",
			AssetHash:"567892fghjk",EncryptedAssetHash:"667891fghjk",OriginPrice:94,AgentPrice:98,
			TradePrice:98,EffectiveTime:"2",EffectiveDate:"2019-03-27",ExpireDate:"2021-03-26"},
		Transaction{ObjectType:"Transaction",TxID:"Tx_4",AgreementID:"4",OwnerID:"ssj",BuyerID:"zq",AgentID:"ssj",
			AssetID:"Asset_2",ContractID:"Contract_4",AgentContractID:"6",
			IsDistributed:"true",TxDate:"2019-03-27",Authority:"view",BuyerKey:"123",AgentKey:"456",SellerKey:"789",
			AssetHash:"567893fghjk",EncryptedAssetHash:"667891fghjk",OriginPrice:94,AgentPrice:98,
			TradePrice:98,EffectiveTime:"2",EffectiveDate:"2019-03-27",ExpireDate:"2021-03-26"},

		Transaction{ObjectType:"Transaction",TxID:"Tx_5",AgreementID:"5",OwnerID:"dl",BuyerID:"ssj",AgentID:"",
			AssetID:"Asset_3",ContractID:"Contract_5",AgentContractID:"",
			IsDistributed:"false",TxDate:"2019-03-27",Authority:"view",BuyerKey:"123",AgentKey:"456",SellerKey:"789",
			AssetHash:"567893fghjk",EncryptedAssetHash:"667891fghjk",OriginPrice:94,AgentPrice:0.0,
			TradePrice:98,EffectiveTime:"2",EffectiveDate:"2019-03-27",ExpireDate:"2021-03-26"},
		Transaction{ObjectType:"Transaction",TxID:"Tx_6",AgreementID:"6",OwnerID:"dl",BuyerID:"ssj",AgentID:"pc",
			AssetID:"Asset_3",ContractID:"Contract_5",AgentContractID:"",
			IsDistributed:"true",TxDate:"2019-03-27",Authority:"view",BuyerKey:"123",AgentKey:"456",SellerKey:"789",
			AssetHash:"567893fghjk",EncryptedAssetHash:"667891fghjk",OriginPrice:94,AgentPrice:99,
			TradePrice:98,EffectiveTime:"2",EffectiveDate:"2019-03-27",ExpireDate:"2021-03-26"},

		Transaction{ObjectType:"Transaction",TxID:"Tx_7",AgreementID:"7",OwnerID:"dl",BuyerID:"ssj",AgentID:"",
			AssetID:"Asset_4",ContractID:"Contract_6",AgentContractID:"7",
			IsDistributed:"false",TxDate:"2019-03-27",Authority:"sell",BuyerKey:"123",AgentKey:"456",SellerKey:"789",
			AssetHash:"567893fghjk",EncryptedAssetHash:"667891fghjk",OriginPrice:94,AgentPrice:98,
			TradePrice:98,EffectiveTime:"2",EffectiveDate:"2019-03-27",ExpireDate:"2021-03-26"},
	}
	m := 0
	for m < len(transactions) {
		fmt.Println("i is ", m)
		transactionAsBytes,_ :=json.Marshal(transactions[m])
		APIstub.PutState(Tx_Prefix+strconv.Itoa(m), transactionAsBytes)
		fmt.Println("Added", transactions[m])
		m = m + 1
	}

	return shim.Success(nil)
}

//（1）查看当前用户资产列表
//request:UserID
//response:[{Key:"Assent_1",Record:{AssetID: "1", OwnerID: "ssj", AssetName: "音乐", FileType: "mp3",AssetHash:"567891fghjk"}}
// ,{Key:"Assent_2",Record:{AssetID: "2", OwnerID: "ssj", AssetName: "影视", FileType: "mp4",AssetHash:"567892fghjk"}]
func (s *SmartContract) getAssentForOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	queryString := "{\"selector\":{\"docType\":\"Asset\",\"OwnerID\":\""+args[0]+"\"}}"
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	//defer resultsIterator.Close()
	buffer,err1 := obj2json(resultsIterator)
	if err1 != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(buffer.Bytes())
}
//（2-1）我发布的代理合约
//request:UserID
func (s *SmartContract) getContractForOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	queryString := "{\"selector\":{\"docType\":\"Contract\",\"Authority\":\"sell\",\"OwnerID\":\""+args[0]+"\"}}"
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	//defer resultsIterator.Close()
	buffer,err1 := obj2json(resultsIterator)
	if err1 != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(buffer.Bytes())
}
//（2-2）我发布的直销合约
//request:UserID
func (s *SmartContract) getContractForOwner2(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	queryString := "{\"selector\":{\"docType\":\"Contract\"," +
		"\"Authority\":\"view\",\"AgentID\":\"\",\"OwnerID\":\""+args[0]+"\"}}"
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	//defer resultsIterator.Close()
	buffer,err1 := obj2json(resultsIterator)
	if err1 != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(buffer.Bytes())
}
//（2-3）我代理的合约
//request:UserID
func (s *SmartContract) getContractForOwner3(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	queryString := "{\"selector\":{\"docType\":\"Contract\"," +
		"\"Authority\":\"view\",\"AgentID\":\""+args[0]+"\"}}"
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	//defer resultsIterator.Close()
	buffer,err1 := obj2json(resultsIterator)
	if err1 != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(buffer.Bytes())
}
//（3）查看当前用户资产使用权列表
//request:UserID
func (s *SmartContract) getAssentForBuyer(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	currentTime:=time.Now().Format("2006-01-02")
	fmt.Println("当前日期："+currentTime)
	queryString := "{\"selector\":" +
		"{\"docType\":\"Transaction\",\"Authority\":\"view\",\"BuyerID\":\""+args[0]+"\",\"ExpireDate\":{\"$gte\":\""+currentTime+"\"}}}"
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	fmt.Println("当前查询语句："+queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		tx := Transaction{}
		json.Unmarshal(queryResponse.Value, &tx)	//转换为资产对象
		assetAsBytes, _ := APIstub.GetState(tx.AssetID)
		if assetAsBytes == nil {
			return  shim.Error("获取资产失败！")
		}
		fmt.Println("当前资产："+string(assetAsBytes))
		contractAsBytes,_:=APIstub.GetState(tx.ContractID)
		if contractAsBytes == nil {
			return  shim.Error("获取合约失败！")
		}

		buffer.WriteString("{Key:\""+queryResponse.Key+"\",Contract:"+string(contractAsBytes)+
			",Asset:"+string(assetAsBytes)+",Record:"+string(queryResponse.Value)+"}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return shim.Success(buffer.Bytes())
}
//（4）查看当前用户交易流水
//request:UserID
func (s *SmartContract) getTransactionForUser(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	var queryString = ""
	if(args[1]=="buyassent"){
		//买资产
		queryString = "{\"selector\":{\"docType\":\"Transaction\",\"Authority\":\"view\",\"BuyerID\":\""+
			args[0]+"\"}}"
	}else if(args[1]=="sellassent"){
		//卖资产（包含了直销和代理抽成的流水）
		queryString = "{\"selector\":{\"docType\":\"Transaction\",\"Authority\":\"view\",\"OwnerID\":\""+
			args[0]+"\"}}"
	}else if(args[1]=="agentsellassent"){
		//卖代理资产
		queryString = "{\"selector\":{\"docType\":\"Transaction\",\"Authority\":\"view\",\"SellerID\":\""+
			args[0]+"\",\"IsDistributed\":\"true\"}}"
	}else if(args[1]=="buycontract"){
		//买合约
		queryString = "{\"selector\":{\"docType\":\"Transaction\",\"Authority\":\"sell\",\"BuyerID\":\""+
			args[0]+"\"}}"
	}else {
		//卖合约
		queryString = "{\"selector\":{\"docType\":\"Transaction\",\"Authority\":\"sell\",\"OwnerID\":\""+
			args[0]+"\"}}"
	}
	//还有一个代理抽成IsDistributed=true，OwnerID=自己
	resultsIterator1, err1 := APIstub.GetQueryResult(queryString)
	if (err1 != nil ) {return shim.Error(err1.Error())}
	defer resultsIterator1.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator1.HasNext() {
		queryResponse, err := resultsIterator1.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		tx := Transaction{}
		json.Unmarshal(queryResponse.Value,&tx)
		assetAsBytes,_:=APIstub.GetState(tx.AssetID)
		if assetAsBytes == nil {
			return  shim.Error("获取资产失败！")
		}
		contractAsBytes,_:=APIstub.GetState(tx.ContractID)
		if contractAsBytes == nil {
			return  shim.Error("获取合约失败！")
		}
		buffer.WriteString("{Key:\""+queryResponse.Key+
			"\",Asset:"+string(assetAsBytes)+",Contract:"+string(contractAsBytes)+", Record:"+string(queryResponse.Value)+"}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return shim.Success(buffer.Bytes())
}
//（5）查看当前用户账户余额
//request:UserID
func (s *SmartContract) getAccountForUser(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	queryString := "{\"selector\":{\"docType\":\"UserAccount\",\"UserID\":\""+args[0]+"\"}}"
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	//defer resultsIterator.Close()
	buffer,err1 := obj2json(resultsIterator)
	if err1 != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(buffer.Bytes())
}
//（6）查看全部有效合约
func (s *SmartContract) getContractList(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	userId := args[0]
	fmt.Println("当前用户："+userId)
	currentTime:=time.Now().Format("2006-01-02")
	fmt.Println("当前日期："+currentTime)
	queryString := "{\"selector\":{\"docType\":\"Contract\",\"ExpireDate\":{\"$gte\":\""+currentTime+"\"}}}"
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		ct := Contract{}
		json.Unmarshal(queryResponse.Value, &ct)	//转换为合约对象
		//取该用户在交易流水中是否有该合约交易
		queryString2 := "{\"selector\":{\"docType\":\"Transaction\",\"ContractID\":\""+
			ct.ContractID+"\",\"BuyerID\":\""+userId+"\",\"ExpireDate\":{\"$gte\":\""+currentTime+"\"}}}"
		resultsIterator2, err := APIstub.GetQueryResult(queryString2)

		if(resultsIterator2.HasNext()){
			buffer.WriteString("{Key:\""+queryResponse.Key+"\",isBuy:\"true\",Record:"+string(queryResponse.Value)+"}")
		}else {
			buffer.WriteString("{Key:\""+queryResponse.Key+"\",isBuy:\"false\",Record:"+string(queryResponse.Value)+"}")
		}
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return shim.Success(buffer.Bytes())
}
//根据ID取合约
func (s *SmartContract) getContractByid(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	var buffer bytes.Buffer
	contractAsBytes,_:=APIstub.GetState(args[0])
	if contractAsBytes == nil {
		return  shim.Error("获取合约失败！")}
	buffer.WriteString("[{Key:\""+args[0]+ "\",Contract:"+string(contractAsBytes)+"}]")
	return shim.Success(buffer.Bytes())
}
//根据ID取资产
func (s *SmartContract) getAssetByid(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	var buffer bytes.Buffer
	contractAsBytes,_:=APIstub.GetState(args[0])
	if contractAsBytes == nil {
		return  shim.Error("获取合约失败！")}
	buffer.WriteString("[{Key:\""+args[0]+ "\",Asset:"+string(contractAsBytes)+"}]")
	return shim.Success(buffer.Bytes())
}
//注册
func (s *SmartContract) register(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
// args:[UserID,UserName,LoginPwd,Mobile]
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}
	existedUser, _ := APIstub.GetState(User_Prefix + args[0])
	if existedUser != nil {
		return  shim.Error("用户已存在！")
	}

	privateKey,publicKey := GenRsaKey(1024)

	var user = User{ObjectType: "User", UserID: args[0], UserName: args[1], LoginPwd: GetMd5(args[2]), Mobile: args[3], PrivateKey: privateKey,PublicKey:publicKey}
	userAsBytes, _ := json.Marshal(user)
	err1 := APIstub.PutState(User_Prefix + args[0], userAsBytes)
	if err1 != nil {
		return shim.Error("fail to put user to state")
	}

	var account = UserAccount{ObjectType: "UserAccount", AccountID:Account_Prefix + args[0],UserID:args[0],Balance:10000}
	accountAsBytes, _ := json.Marshal(account)
	err2 := APIstub.PutState(Account_Prefix + args[0], accountAsBytes)
	if err2 != nil {
		return shim.Error("fail to put account to state")
	}

	return shim.Success(userAsBytes)
}

func (s *SmartContract) login(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// arst:[UserID,LoginPwd]
	// res:User struct
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	userAsBytes, _ := APIstub.GetState(User_Prefix + args[0])
	if userAsBytes == nil {
		return  shim.Error("用户名或密码错误！")
	}

	user := User{}
	json.Unmarshal(userAsBytes, &user)
	if user.LoginPwd != GetMd5(args[1]) {
		return  shim.Error("用户名或密码错误！")
	} else {
		return shim.Success(userAsBytes)
	}
}

func (s *SmartContract) fileUpload(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
// args:[OwnerID,AssetName,FileType,filePath,PrivateKey,FileHash]
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	/*
	if !s.CheckPrivateKey(APIstub, args[0], args[4]) {
		return shim.Error("用户身份验证失败！")
	}
	*/

	//var enhash = hex.EncodeToString([]byte(args[5]))

	queryString := "{\"selector\":{\"docType\":\"Asset\",\"FileHash\":\""+args[5]+"\"}}"
	existedAsset, err := s.RichQuery(APIstub, queryString)

	if err != nil {
		return shim.Error("文件操作出错！")
	}

	if string(existedAsset) != "[]" {
		return shim.Error("文件已存在！")
	}

	queryString = "{\"selector\":{\"docType\":\"Asset\",\"OwnerID\":\""+args[0]+"\",\"AssetName\":\""+args[1]+"\"}}"
	existedAsset, err = s.RichQuery(APIstub, queryString)

	if err != nil {
		return shim.Error("资产操作出错！")
	}

	if string(existedAsset) != "[]" {
		return shim.Error("资产名称已存在！")
	}

	data := url.Values{"filePath":{args[3]}}
	urlString := appUrlPrefix + "HFassetUpload"
	res,err := http.PostForm(urlString,data)
	if err != nil{
		return shim.Error("文件上传失败！错误码："+err.Error())
	}
	defer res.Body.Close()
	hash, err := ioutil.ReadAll(res.Body)
	if err != nil{
		return shim.Error("文件处理失败！")
	}

	var asset = Asset{ObjectType:"Asset",
						AssetID:Asset_Prefix + APIstub.GetTxID(),
						OwnerID:args[0],
						AssetName:args[1],
						FileType:args[2],
						AssetHash:string(hash),
						FileHash:args[5]}

	assetAsBytes, _ := json.Marshal(asset)
	err1 := APIstub.PutState(Asset_Prefix + APIstub.GetTxID(), assetAsBytes)
	if err1 != nil {
		return shim.Error("fail to put asset to state")
	}
	return shim.Success(assetAsBytes)
}

func (s *SmartContract) contractIssue (APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
// args:[0-Authority,
// 		1-OwnerID,
// 		2-AssetID,
// 		3-ContractName,
// 		4-ExpireDate,
// 		5-AgentContractExpireDate,
// 		6-TradePrice,
// 		7-AgentPrice,
// 		8-privateKey,
// 		9-UseType]
// Authority:view/sell
	if len(args) != 9 {
		return shim.Error("Incorrect number of arguments. Expecting 9")
	}
/*
	if !s.CheckPrivateKey(APIstub, args[1], args[8]) {
		return shim.Error("用户身份验证失败！")
	}
*/
	queryString := "{\"selector\":{\"docType\":\"Contract\",\"OwnerID\":\""+args[1]+"\",\"ContractName\":\""+args[3]+"\"}}"
	existedContract, err := s.RichQuery(APIstub, queryString)

	if err != nil {
		return shim.Error("合约操作出错！")
	}

	if string(existedContract) != "[]" {
		return shim.Error("合约名称已存在！")
	}

	asset := Asset{}
	assetAsBytes, _ := APIstub.GetState(args[2])
	json.Unmarshal(assetAsBytes, &asset)
	tradePrice, _ := strconv.ParseFloat(args[6], 64)
	agentPrice, _ := strconv.ParseFloat(args[7], 64)

	contract := Contract{}

	if args[0] == "view" {
		contract = Contract{
			ObjectType:"Contract",
			ContractID:Contract_Prefix + APIstub.GetTxID(),
			AssetID:args[2],
			AssetName:asset.AssetName,
			AssetHash:asset.AssetHash,
			FileHash:asset.FileHash,
			OwnerID:args[1],
			AgnetID:"",
			ContractName:args[3],
			Authority:"view",
			//UseType:args[9],
			IsAgent:"false",
			ExpireDate:args[4],  // 当前合约失效日期,yyyy-mm-dd
			AgentContractExpireDate:"",  // 代理合约失效日期,yyyy-mm-dd
			TradePrice:tradePrice,
			AgentPrice:0}
	} else if args[0] == "sell" {
		contract = Contract{
			ObjectType:"Contract",
			ContractID:Contract_Prefix + APIstub.GetTxID(),
			AssetID:args[2],
			AssetName:asset.AssetName,
			AssetHash:asset.AssetHash,
			FileHash:asset.FileHash,
			OwnerID:args[1],
			AgnetID:"",
			ContractName:args[3],
			Authority:"sell",
			//UseType:args[9],
			IsAgent:"false",
			ExpireDate:args[4],  // 当前合约失效日期,yyyy-mm-dd
			AgentContractExpireDate:args[5],  // 代理合约失效日期,yyyy-mm-dd
			TradePrice:tradePrice,
			AgentPrice:agentPrice}
	}

	contractAsBytes, _ := json.Marshal(contract)
	err = APIstub.PutState(Contract_Prefix + APIstub.GetTxID(), contractAsBytes)
	if err != nil {
		return shim.Error("fail to put contract to state")
	}

	return shim.Success(contractAsBytes)
}

func (s *SmartContract) trade(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// args:[ContractID,BuyerID,EffectiveTime,ContractName,PrivateKey]
	// ContractID:购买标的合约ID，为平台中已存在的合约
	// ContractName:为代理商新生成的合约名称，由代理商在客户端填写

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
/*
	if !s.CheckPrivateKey(APIstub, args[1], args[4]) {
		return shim.Error("用户身份验证失败！")
	}
*/
	contractAsBytes, _ := APIstub.GetState(args[0])
	contract := Contract{}
	err := json.Unmarshal(contractAsBytes,&contract)
	if err != nil {
		return shim.Error("合约提取失败！")
	}

	curTime := time.Now().Format("2006-01-02 15:04:05")
	curDate := time.Now().Format("2006-01-02")
	i, _ := strconv.Atoi(args[2])
	expireDate := time.Now().AddDate(i,0,1).Format("2006-01-02")

	transaction := Transaction{}
	agentContract := Contract{} // 代理商合约
	sellerAcco := UserAccount{}
	buyerAcco := UserAccount{}
	agentAcco := UserAccount{}
	seller :=User{}
	agent := User{}
	buyer :=User{}


	sellerAccoAsBytes, _ := APIstub.GetState(Account_Prefix + contract.OwnerID)
	json.Unmarshal(sellerAccoAsBytes, &sellerAcco)
	sellerAsButes,_:= APIstub.GetState(User_Prefix + contract.OwnerID)
	json.Unmarshal(sellerAsButes, &seller)

	buyerAccoAsBytes, _ := APIstub.GetState(Account_Prefix + args[1])
	json.Unmarshal(buyerAccoAsBytes, &buyerAcco)
	buyerAsBytes,_:= APIstub.GetState(User_Prefix + args[1])
	json.Unmarshal(buyerAsBytes, &buyer)

	if contract.IsAgent == "true" {
		agnetAccoAsBytes, _ := APIstub.GetState(Account_Prefix + contract.AgnetID)
		json.Unmarshal(agnetAccoAsBytes, &agentAcco)
		agnetAsBytes,_:= APIstub.GetState(User_Prefix + contract.AgnetID)
		json.Unmarshal(agnetAsBytes, &agent)
	}

	if contract.Authority == "view" { // 买资产

		if buyerAcco.Balance < contract.TradePrice {
			return shim.Error("账户余额不足！")
		}

		key := RandomStr(66)
		fmt.Println("分段前的key："+key)
		assetHash := contract.AssetHash
		data := url.Values{"assetHash":{assetHash},"key":{key}}
		urlString := appUrlPrefix + "HFencryptFile"
		res,err := http.PostForm(urlString,data)
		if err != nil{
			return shim.Error("文件提取失败！")
		}
		defer res.Body.Close()
		encryptedFileHash, err := ioutil.ReadAll(res.Body)
		if err != nil{
			return shim.Error("文件处理失败！")
		}

		if contract.IsAgent == "false" { // 从资产所有者处买资产
			keyLength := len(key)/2
			sellerKey := string([]byte(key)[:keyLength])
			buyerKey := string([]byte(key)[keyLength:])
			fmt.Println("分段的sellerKey："+sellerKey)
			fmt.Println("分段的buyerKey："+buyerKey)
			ensellerKey,_ := RsaEncrypt([]byte(sellerKey),[]byte(seller.PublicKey))
			enbuyerKey,_ := RsaEncrypt([]byte(buyerKey),[]byte(buyer.PublicKey))
			enensellerKey := base64.StdEncoding.EncodeToString(ensellerKey)
			enenbuyerKey := base64.StdEncoding.EncodeToString(enbuyerKey)
			fmt.Println("加密后分段的sellerKey："+string(enensellerKey))
			fmt.Println("加密后分段的buyerKey："+string(enenbuyerKey))
			transaction = Transaction{
				ObjectType:"Transaction",
				TxID:Tx_Prefix + APIstub.GetTxID(),
				AgreementID:Tx_Prefix + APIstub.GetTxID(),
				OwnerID:contract.OwnerID,
				AgentID:"",
				BuyerID:args[1],
				AssetID:contract.AssetID,
				ContractID:contract.ContractID,
				AgentContractID:"",
				IsDistributed:"false",
				TxDate:curTime,
				Authority:"view",
				//UseType:contract.UseType,
				BuyerKey:enenbuyerKey,
				AgentKey:"",
				SellerKey:enensellerKey,
				AssetHash:contract.AssetHash,
				FileHash:contract.FileHash,
				EncryptedAssetHash:string(encryptedFileHash),  // authority!='sell'时使用
				AgentPrice:0,  // authority=='sell'时使用
				TradePrice:contract.TradePrice,
				EffectiveTime:args[2],
				EffectiveDate:curDate,
				ExpireDate:expireDate}

			sellerAcco.Balance = sellerAcco.Balance + contract.TradePrice
			buyerAcco.Balance = buyerAcco.Balance - contract.TradePrice

		} else if contract.IsAgent == "true" { // 从代理商处买资产
			keyLength := len(key)/3
			sellerKey := string([]byte(key)[:keyLength])
			agentKey := string([]byte(key)[keyLength:2*keyLength])
			buyerKey := string([]byte(key)[2*keyLength:])
			fmt.Println("分段的sellerKey："+sellerKey)
			fmt.Println("分段的agentKey："+agentKey)
			fmt.Println("分段的buyerKey："+buyerKey)
			ensellerKey,_ := RsaEncrypt([]byte(sellerKey),[]byte(seller.PublicKey))
			enagentKey,_:= RsaEncrypt([]byte(agentKey),[]byte(agent.PublicKey))
			enbuyerKey,_ := RsaEncrypt([]byte(buyerKey),[]byte(buyer.PublicKey))

			enensellerKey := base64.StdEncoding.EncodeToString(ensellerKey)
			enenagentKey := base64.StdEncoding.EncodeToString(enagentKey)
			enenbuyerKey := base64.StdEncoding.EncodeToString(enbuyerKey)
			fmt.Println("加密后分段的sellerKey："+string(enensellerKey))
			fmt.Println("加密后分段的enagentKey："+string(enenagentKey))
			fmt.Println("加密后分段的buyerKey："+string(enenbuyerKey))
			transaction = Transaction{
				ObjectType:"Transaction",
				TxID:Tx_Prefix + APIstub.GetTxID(),
				AgreementID:Tx_Prefix + APIstub.GetTxID(),
				OwnerID:contract.OwnerID,
				AgentID:contract.AgnetID,
				BuyerID:args[1],
				AssetID:contract.AssetID,
				ContractID:contract.ContractID,
				IsDistributed:"true",
				TxDate:curTime,
				Authority:"view",
				//UseType:contract.UseType,
				BuyerKey:enenbuyerKey,
				AgentKey:enenagentKey,
				SellerKey:enensellerKey,
				AssetHash:contract.AssetHash,
				FileHash:contract.FileHash,
				EncryptedAssetHash:string(encryptedFileHash),  // authority!='sell'时使用
				AgentPrice:contract.AgentPrice,
				TradePrice:contract.TradePrice,
				EffectiveTime:args[2],
				EffectiveDate:curDate,
				ExpireDate:expireDate}

			sellerAcco.Balance = sellerAcco.Balance + contract.AgentPrice
			buyerAcco.Balance = buyerAcco.Balance - contract.TradePrice
			agentAcco.Balance = agentAcco.Balance + (contract.TradePrice - contract.AgentPrice)

			angetAccoAsBytes, _ := json.Marshal(agentAcco)
			APIstub.PutState(Account_Prefix + contract.AgnetID, angetAccoAsBytes)

		}

		sellerAccoAsBytes, _ := json.Marshal(sellerAcco)
		APIstub.PutState(Account_Prefix + contract.OwnerID, sellerAccoAsBytes)

		buyerAccoAsBytes, _ := json.Marshal(buyerAcco)
		APIstub.PutState(Account_Prefix + args[1], buyerAccoAsBytes)

	} else if contract.Authority == "sell" { // 买合约

		queryString := "{\"selector\":{\"docType\":\"Contract\",\"OwnerID\":\""+args[1]+"\",\"ContractName\":\""+args[3]+"\"}}"
		existedContract, err := s.RichQuery(APIstub, queryString)

		if err != nil {
			return shim.Error("合约操作出错！")
		}

		if string(existedContract) != "[]" {
			return shim.Error("合约名称已存在！")
		}

		transaction = Transaction{
			ObjectType:"Transaction",
			TxID:Tx_Prefix + APIstub.GetTxID(),
			AgreementID:Tx_Prefix + APIstub.GetTxID(),
			OwnerID:contract.OwnerID,
			AgentID:args[1],
			BuyerID:args[1],
			AssetID:contract.AssetID,
			ContractID:contract.ContractID,
			AgentContractID:Contract_Prefix + APIstub.GetTxID(),
			IsDistributed:"false",
			TxDate:curTime,
			Authority:"sell",
			//UseType:contract.UseType,
			BuyerKey:"",
			AgentKey:"",
			SellerKey:"",
			AssetHash:contract.AssetHash,
			FileHash:contract.FileHash,
			EncryptedAssetHash:"",  // authority!='sell'时使用
			AgentPrice:contract.AgentPrice,  // authority=='sell'时使用
			TradePrice:contract.TradePrice,
			EffectiveTime:args[2],
			EffectiveDate:curDate,
			ExpireDate:expireDate}

		agentContract = Contract{
			ObjectType:"Contract",
			ContractID:Contract_Prefix + APIstub.GetTxID(),
			AssetID:contract.AssetID,
			AssetName:contract.AssetName,
			AssetHash:contract.AssetHash,
			FileHash:contract.FileHash,
			OwnerID:contract.OwnerID,
			AgnetID:args[1],
			ContractName:args[3],
			Authority:"view",
			//UseType:contract.UseType,
			IsAgent:"true",
			ExpireDate:contract.AgentContractExpireDate,  // 当前合约失效日期
			AgentContractExpireDate:"",  // 代理合约失效日期
			TradePrice:contract.TradePrice,
			AgentPrice:contract.AgentPrice}

		agentContractAsBytes, _ := json.Marshal(agentContract)
		err = APIstub.PutState(Contract_Prefix + APIstub.GetTxID(), agentContractAsBytes)
		if err != nil {
			return shim.Error("fail to put contract to state")
		}
	}

	txAsBytes, _ := json.Marshal(transaction)
	err = APIstub.PutState(Tx_Prefix + APIstub.GetTxID(), txAsBytes)
	if err != nil {
		return shim.Error("fail to put tx to state")
	}

	return shim.Success(txAsBytes)
}

func (s *SmartContract) assetGet(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// arst:[UserID,TxID,PrivateKey]

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
/*
	if !s.CheckPrivateKey(APIstub, args[0], args[2]) {
		return shim.Error("用户身份验证失败！")
	}
*/
	txAsBytes, _ := APIstub.GetState(args[1])
	if txAsBytes == nil {
		return  shim.Error("交易提取失败！")
	}

	tx := Transaction{}
	json.Unmarshal(txAsBytes, &tx)

	if tx.Authority != "view" {
		return shim.Error("无可获取资产！")
	}

	expireDate, _ := time.Parse("2006-01-02", tx.ExpireDate)
	if expireDate.Before(time.Now()) {
		return shim.Error("资产使用权已过期！")
	}

	//资产所有者seller
	user := User{}
	userAsBytes, _ := APIstub.GetState(User_Prefix+tx.OwnerID)
	if userAsBytes == nil {
		return  shim.Error("获取资产所有者失败！")
	}
	json.Unmarshal(userAsBytes,&user)
	//解密分段密钥
	debase64SellerKey, _ := base64.StdEncoding.DecodeString(tx.SellerKey)
	desellerKey,_:=RsaDecrypt(debase64SellerKey,[]byte(user.PrivateKey))
	fmt.Println("解密前卖家密钥："+tx.SellerKey)
	fmt.Println("解密前卖家私钥："+user.PrivateKey)
	fmt.Println("解密后卖家密钥："+string(desellerKey))
	//买家buyer
	buyer := User{}
	buyerAsBytes,_:= APIstub.GetState(User_Prefix + tx.BuyerID)
	json.Unmarshal(buyerAsBytes, &buyer)
	//解密分段密钥
	debase64BuyerKey, _ := base64.StdEncoding.DecodeString(tx.BuyerKey)
	debuyerKey,_:=RsaDecrypt(debase64BuyerKey,[]byte(buyer.PrivateKey))
	fmt.Println("解密前买家密钥："+tx.BuyerKey)
	fmt.Println("解密前买家私钥："+buyer.PrivateKey)
	fmt.Println("解密后买家密钥："+string(debuyerKey))

	var key string
	if tx.IsDistributed == "true" {
		//代理商agenter
		agent := User{}
		agentAsBytes,_:= APIstub.GetState(User_Prefix + tx.AgentID)
		json.Unmarshal(agentAsBytes, &agent)
		//解密分段密钥
		debase64agentKey, _ := base64.StdEncoding.DecodeString(tx.AgentKey)
		deagentKey,_:=RsaDecrypt(debase64agentKey,[]byte(agent.PrivateKey))
		fmt.Println("解密后代理商密钥："+string(deagentKey))

		key = string(desellerKey) + string(deagentKey) + string(debuyerKey)
		fmt.Println("解密后完整密钥："+key)
	} else if tx.IsDistributed == "false" {
		key = string(desellerKey) + string(debuyerKey)
		fmt.Println("解密后完整密钥："+key)
	}


	data := url.Values{"encryptedFileHash":{tx.EncryptedAssetHash},"fileHash":{tx.FileHash},"key":{key},"privateKey":{user.PrivateKey}}
	urlString := appUrlPrefix + "HFassetGet"
	res,err := http.PostForm(urlString,data)
	if err != nil{
		return shim.Error("文件请求失败！")
	}
	defer res.Body.Close()
	result, err := ioutil.ReadAll(res.Body)
	if err != nil{
		return shim.Error("请求处理失败！")
	}

	if string(result) == "FAIL" {
		return shim.Error("文件获取失败！")
	} else {
		asset := Asset{}
		assetAsBytes, _ := APIstub.GetState(tx.AssetID)
		json.Unmarshal(assetAsBytes,&asset)
		var buffer bytes.Buffer
		buffer.WriteString("{\"filePath\":\"" + string(result) + "\",\"fileName\":\""+ asset.AssetName + "." + asset.FileType + "\"}")
		return shim.Success(buffer.Bytes())
	}

	return shim.Error("请求失败！")
}


// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
