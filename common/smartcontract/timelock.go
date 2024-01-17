/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a car
type SmartContract struct {
	contractapi.Contract
}

//
type Hashtime_lock struct{
	Id string `json:"id"`
	CertId string `json:"certId"`
	Hash string `json:"hash"`
	Nowtime string `json:"nowtime"`
	Endtime string `json:"endtime"`
	RandomString string `json:"randomString"`
	PublicKey string `json:"publicKey"`
	OptionalString string `json:"optionalString"`   // 更改时，此变量不为0，其他情况，可为0
}


//Cert describes details of what make up a cert
type Cert struct{
	Id string `json:"id"`
	CA string `json:"ca"`
	Version string    `json:"version"`
	SerialNumber string `json:"serialNumber"`
	Signature string `json:"signature"`
    SignatureAlgorithm string `json:"signatureAlgorithm"`
    Issure string `json:"issure"`
    CreateTime string `json:"createTime"`
    EndTime string `json:"endTime"`
    EntityName string `json:"entityName"`
    EntityIdentifier string `json:"entityIdentifier"`
    PublicKey string `json:"publicKey"`
    PublicKeyAlgorithm string `json:"publicKeyAlgorithm"`
    OptionalString string `json:"optionalString"`   // 更改时，此变量不为0，其他情况，可为0
}



// InitLedger adds some Cert to the Ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error{
	certs := []Cert{
		Cert{
			Id:"000001",
			CA:"000001",
			Version:"1.0",
			SerialNumber:"1",
			Signature:"58:a9:98:e7:16:52:4c:40:e7:e1:47:92:19:1b:3a:8f:97:6c:7b:b7:b0:cb:20:6d:ad:b5:d3:47:58:d8:e4:f2:3e:32:e9:ef:87:77:e5:54:36:f4:8d:50:8d:07:b4:77:45:ea:9d:a4:33:36:9b:0b:e0:74:58:11:c5:01:7b:4d",
			SignatureAlgorithm:"md5WithRSAEncryption",
			Issure:"huawei",
			CreateTime:"2021-06-27 08:00:00",
			EndTime:"2023-06-27 08:00:00",
			EntityName:"rongyao",
			EntityIdentifier:"DCCAb4CAQEwDQYJKoZIhvcNAQEEBQAwgZ4xCzAJBgNVBAYTAlVTMRAwDgYDVQQIEwdNb250YW5hMRAwD",
			PublicKey:"00:c6:7b:c0:68:81:2f:de:82:3f:f9:ac:c3:86:4a:66:b7:ec:d4:f1:f6:64:21:ff:f5:a2:34:42:d0:38:9f:c6:dd:3b:6e:26:65:6a:54:96:dd:d2:7b:eb:36:a2:ae:7e:2a:9e:7e:56:a5:b6:87:9f:15:c7:18:66:7e:16:77:e2:a7",
			PublicKeyAlgorithm:"rsaEncryption",
			OptionalString:"0"},
		Cert{
			Id:"000002",
			CA:"000002",
			Version:"1.0",
			SerialNumber:"2",
			Signature:"58:a9:98:e7:16:52:4c:40:e7:e1:47:92:19:1b:3a:8f:97:6c:7b:b7:b0:cb:20:6d:ad:b5:d3:47:58:d8:e4:f2:3e:32:e9:ef:87:77:e5:54:36:f4:8d:50:8d:07:b4:77:45:ea:9d:a4:33:36:9b:0b:e0:74:58:11:c5:01:7b:4d",
			SignatureAlgorithm:"md5WithRSAEncryption",
			Issure:"huawei",
			CreateTime:"2021-06-27 08:00:00",
			EndTime:"2023-06-27 08:00:00",
			EntityName:"sike",
			EntityIdentifier:"DCCAb4CAQEwDQYJKoZIhvcNAQEEBQAwgZ4xCzAJBgNVBAYTAlVTMRAwDgYDVQQIEwdNb250YW5hMRAwD",
			PublicKey:"00:11:22:33:44:81:2f:de:82:3f:f9:ac:c3:86:4a:66:b7:ec:d4:f1:f6:64:21:ff:f5:a2:34:42:d0:38:9f:c6:dd:3b:6e:26:65:6a:54:96:dd:d2:7b:eb:36:a2:ae:7e:2a:9e:7e:56:a5:b6:87:9f:15:c7:18:66:7e:16:77:e2:a7",
			PublicKeyAlgorithm:"rsaEncryption",
			OptionalString:"0"},
		Cert{
			Id:"000003",
			CA:"000003",
			Version:"1.0",
			SerialNumber:"3",
			Signature:"58:a9:98:e7:16:52:4c:40:e7:e1:47:92:19:1b:3a:8f:97:6c:7b:b7:b0:cb:20:6d:ad:b5:d3:47:58:d8:e4:f2:3e:32:e9:ef:87:77:e5:54:36:f4:8d:50:8d:07:b4:77:45:ea:9d:a4:33:36:9b:0b:e0:74:58:11:c5:01:7b:4d",
			SignatureAlgorithm:"md5WithRSAEncryption",
			Issure:"huawei",
			CreateTime:"2021-06-27 08:00:00",
			EndTime:"2023-06-27 08:00:00",
			EntityName:"sike",
			EntityIdentifier:"DCCAb4CAQEwDQYJKoZIhvcNAQEEBQAwgZ4xCzAJBgNVBAYTAlVTMRAwDgYDVQQIEwdNb250YW5hMRAwD",
			PublicKey:"00:11:22:33:44:81:2f:de:82:3f:f9:ac:c3:86:4a:66:b7:ec:d4:f1:f6:64:21:ff:f5:a2:34:42:d0:38:9f:c6:dd:3b:6e:26:65:6a:54:96:dd:d2:7b:eb:36:a2:ae:7e:2a:9e:7e:56:a5:b6:87:9f:15:c7:18:66:7e:16:77:e2:a7",
			PublicKeyAlgorithm:"rsaEncryption",
			OptionalString:"0"},
		Cert{
			Id:"000004",
			CA:"000004",
			Version:"1.0",
			SerialNumber:"4",
			Signature:"58:a9:98:e7:16:52:4c:40:e7:e1:47:92:19:1b:3a:8f:97:6c:7b:b7:b0:cb:20:6d:ad:b5:d3:47:58:d8:e4:f2:3e:32:e9:ef:87:77:e5:54:36:f4:8d:50:8d:07:b4:77:45:ea:9d:a4:33:36:9b:0b:e0:74:58:11:c5:01:7b:4d",
			SignatureAlgorithm:"md5WithRSAEncryption",
			Issure:"huawei",
			CreateTime:"2021-06-27 08:00:00",
			EndTime:"2023-06-27 08:00:00",
			EntityName:"wangzherongyao",
			EntityIdentifier:"DCCAb4CAQEwDQYJKoZIhvcNAQEEBQAwgZ4xCzAJBgNVBAYTAlVTMRAwDgYDVQQIEwdNb250YW5hMRAwD",
			PublicKey:"00:11:22:33:44:81:2f:de:82:3f:f9:ac:c3:86:4a:66:b7:ec:d4:f1:f6:64:21:ff:f5:a2:34:42:d0:38:9f:c6:dd:3b:6e:26:65:6a:54:96:dd:d2:7b:eb:36:a2:ae:7e:2a:9e:7e:56:a5:b6:87:9f:15:c7:18:66:7e:16:77:e2:a7",
			PublicKeyAlgorithm:"rsaEncryption",
			OptionalString:"0"},
		Cert{
			Id:"000005",
			CA:"000005",
			Version:"1.0",
			SerialNumber:"5",
			Signature:"58:a9:98:e7:16:52:4c:40:e7:e1:47:92:19:1b:3a:8f:97:6c:7b:b7:b0:cb:20:6d:ad:b5:d3:47:58:d8:e4:f2:3e:32:e9:ef:87:77:e5:54:36:f4:8d:50:8d:07:b4:77:45:ea:9d:a4:33:36:9b:0b:e0:74:58:11:c5:01:7b:4d",
			SignatureAlgorithm:"md5WithRSAEncryption",
			Issure:"apple",
			CreateTime:"2021-06-27 08:00:00",
			EndTime:"2023-06-27 08:00:00",
			EntityName:"sike",
			EntityIdentifier:"DCCAb4CAQEwDQYJKoZIhvcNAQEEBQAwgZ4xCzAJBgNVBAYTAlVTMRAwDgYDVQQIEwdNb250YW5hMRAwD",
			PublicKey:"00:11:22:33:44:81:2f:de:82:3f:f9:ac:c3:86:4a:66:b7:ec:d4:f1:f6:64:21:ff:f5:a2:34:42:d0:38:9f:c6:dd:3b:6e:26:65:6a:54:96:dd:d2:7b:eb:36:a2:ae:7e:2a:9e:7e:56:a5:b6:87:9f:15:c7:18:66:7e:16:77:e2:a7",
			PublicKeyAlgorithm:"rsaEncryption",
			OptionalString:"0"},
	}
    for _, cert := range certs{
		certBytes ,err := json.Marshal(cert)
		if err != nil{
			return fmt.Errorf(" Failed to Marshal. %s", err.Error())
		}
		err = ctx.GetStub().PutState(cert.Id,certBytes)
		if err != nil{
			return fmt.Errorf("Failed to put in word state. %s",err.Error())
		}
	}
return nil
}

// CreatCert adds a new cert   input 12 and output 1
func (s *SmartContract) CreatCert (ctx contractapi.TransactionContextInterface,id string,ca string,version string,serialNumber string,signature string,signatureAlgorithm string,issure string,createTime string,endTime string,entityName string,entityIdentifier string,publicKey string,publicKeyAlgorithm string,optionalString string) error {
	cert :=Cert{
		Id:id,
		CA:ca,
		Version: version,
		SerialNumber : serialNumber,
		Signature : signature,
		SignatureAlgorithm : signatureAlgorithm,
		Issure : issure,
		CreateTime : createTime,
		EndTime : endTime,
		EntityName : entityName,
		EntityIdentifier : entityIdentifier,
		PublicKey : publicKey,
		PublicKeyAlgorithm : publicKeyAlgorithm,
        OptionalString : optionalString,
	}
	certBytes,_:=json.Marshal(cert)
	return ctx.GetStub().PutState(id,certBytes)
}



// QueryCert return the cert    input 1 and out put 1
func (s *SmartContract) QueryCert (ctx contractapi.TransactionContextInterface,id string) (*Cert, error) {
	certBytes ,err := ctx.GetStub().GetState(id)
    if err != nil || certBytes == nil {
		return nil,fmt.Errorf("Failed get Cert from the world state. %s", err.Error())
	}
	cert := new(Cert)
	err = json.Unmarshal(certBytes,cert)
	if err != nil{
		return nil,fmt.Errorf("Failed Unmarshal certBytes. %s", err.Error())
	}
	return cert,nil
}





// change some information to some cert  插入三个数据  input 3 and output 1
func (s *SmartContract) ChangeCert (ctx contractapi.TransactionContextInterface,id string,publicKey string,optionalString string) error {
	certBytes ,err := ctx.GetStub().GetState(id)
    if err != nil || certBytes == nil {
		return fmt.Errorf("Failed get Cert from the world state. %s", err.Error())
	}
	cert := new(Cert)
	err = json.Unmarshal(certBytes,cert)
	if err != nil{
		return fmt.Errorf("Failed Unmarshal certBytes. %s", err.Error())
	}
	cert.PublicKey = publicKey
	cert.OptionalString = optionalString
	certBytes ,_ = json.Marshal(cert)
	return ctx.GetStub().PutState(id,certBytes)
}

// DeleteCert return the error, input 1 and return 1
func (s *SmartContract) DelCert (ctx contractapi.TransactionContextInterface,id string) error {
	certBytes ,err := ctx.GetStub().GetState(id)
	if err != nil || certBytes == nil {
		return fmt.Errorf("Failed get Cert from the world state. %s", err.Error())
	}
	return ctx.GetStub().DelState(id)
}

// CreatTimeHashlock adds a new lock   input 12 and output 1
func (s *SmartContract) CreatTimeHashlock (ctx contractapi.TransactionContextInterface,id string,certId string,hash string,nowtime string,endtime string,randomString string,publicKey string,optionalString string) error {
	hashtime_lock :=Hashtime_lock{
		Id : id,
		CertId : certId,
		Hash : hash,
		Nowtime : nowtime,
		Endtime : endtime,
		RandomString: randomString,   //生成时间哈希锁
		PublicKey : publicKey,
		OptionalString : optionalString,  
	}
	hashtime_lockBytes,_:=json.Marshal(hashtime_lock)
	return ctx.GetStub().PutState(id,hashtime_lockBytes)
}

func (s *SmartContract) ExeTimeHashlock (ctx contractapi.TransactionContextInterface,id string,certId string,hash string,nowtime string,endtime string,randomString string,publicKey string,optionalString string) error {
	hashtime_lock :=Hashtime_lock{
		Id : id,
		CertId : certId,
		Hash : hash,
		Nowtime : nowtime,
		Endtime : endtime,
		RandomString: randomString,   //生成时间哈希锁
		PublicKey : publicKey,
		OptionalString : optionalString,  
	}
	hashtime_lockBytes,_:=json.Marshal(hashtime_lock)
	return ctx.GetStub().PutState(id,hashtime_lockBytes)
}

// QueryCert return the timehashlock    input 1 and out put 1
func (s *SmartContract) QueryTimeHashlock (ctx contractapi.TransactionContextInterface,id string) (*Hashtime_lock, error) {
	timeBytes ,err := ctx.GetStub().GetState(id)
    if err != nil || timeBytes == nil {
		return nil,fmt.Errorf("Failed get time from the world state. %s", err.Error())
	}
	time := new(Hashtime_lock)
	err = json.Unmarshal(timeBytes,time)
	if err != nil{
		return nil,fmt.Errorf("Failed Unmarshal timeBytes. %s", err.Error())
	}
	return time,nil
}



func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}



