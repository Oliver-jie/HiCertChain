/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/go-basic/uuid"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

var wg = sync.WaitGroup{}

var Max_Count = 50   //循环次数
const MAX_CONNECT = 2 //连接网关数

var ch = make(chan string, Max_Count*MAX_CONNECT) // 有缓冲
var flag = true

func invoceChaincode(con *gateway.Contract) {

	for i := 0; i < Max_Count; i++ {
		unique_id := uuid.New()
		//fmt.Println(uuid)
		ch <- unique_id
		_, err := con.SubmitTransaction("insert", unique_id, unique_id)
		//_, err := con.SubmitTransaction("createCar", "CAR"+uuidArr, "0", "2", "2", "4")
		// uuidArr := uuid.New()
		// _, err := con.SubmitTransaction("createCar", "CAR"+uuidArr, "0", "2", "2", "4")
		if err != nil {
			log.Fatalf("Failed to Submit transaction: %v", err)
		}
		//log.Println("--> Submit Transaction: CreateAsset, creates new asset with ID, color, owner, size, and appraisedValue arguments")

		//log.Println(string(result))
		//uuidAssetId[i] = "asset14" + uuidArr[i]
	}
	wg.Done()
}

func queryChaincode(con *gateway.Contract) {

	for i := 0; i < Max_Count; i++ {
		//uuidArr := uuid.New()
		unique_id := <-ch
		_, err := con.EvaluateTransaction("query", unique_id)
		if err != nil {
			log.Fatalf("Failed to Submit transaction: %v", err)
		}
		//log.Println("--> Submit Transaction: CreateAsset, creates new asset with ID, color, owner, size, and appraisedValue arguments")

		// log.Println(string(result))
		//uuidAssetId[i] = "asset14" + uuidArr[i]
	}
	wg.Done()
}

func main() {
	log.Println("============ application-golang starts ============")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	for i := 0; i < MAX_CONNECT; i++ {
		if !wallet.Exists("appUser" + strconv.Itoa(i)) {
			err = populateWallet("appUser"+strconv.Itoa(i), wallet)
			if err != nil {
				log.Fatalf("Failed to populate wallet contents: %v", err)
			}
		}
	}

	ccpPath := filepath.Join(
		".",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)
	ccpPath = "connection-v14-gatway.json"

	// gw, err := gateway.Connect(
	// 	gateway.WithConfig(config.FromFile(ccpPath)),
	// 	gateway.WithIdentity(wallet, "appUser"),
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to connect to gateway: %v", err)
	// }
	// defer gw.Close()

	// network, err := gw.GetNetwork("mychannel")
	// if err != nil {
	// 	log.Fatalf("Failed to get network: %v", err)
	// }

	// contract := network.GetContract("fabcar")

	contract_name := "mycc"
	var gw [MAX_CONNECT]*gateway.Gateway
	var network [MAX_CONNECT]*gateway.Network
	var contract [MAX_CONNECT]*gateway.Contract

	for i := 0; i < MAX_CONNECT; i++ {
		gw[i], err = gateway.Connect(
			gateway.WithConfig(config.FromFile(ccpPath)),
			gateway.WithUser("Admin"),
		)
		if err != nil {
			log.Fatalf("Failed to connect to gateway: %v", err)
		}
		defer gw[i].Close()

		network[i], err = gw[i].GetNetwork("mychannel")
		if err != nil {
			log.Fatalf("Failed to get network: %v", err)
		}

		contract[i] = network[i].GetContract(contract_name)
	}

	// wg.Add(1)
	// queryChaincode(contract[0])
	// wg.Wait()

	wg.Add(MAX_CONNECT)
	for i := 0; i < MAX_CONNECT; i++ {
		go invoceChaincode(contract[i])
		//go queryChaincode(contract[i])
	}
	timeStart := time.Now().UnixNano()
	wg.Wait()

	timeCount := Max_Count * MAX_CONNECT
	timeEnd := time.Now().UnixNano()
	count := float64(timeCount)
	timeResult := float64((timeEnd-timeStart)/1e6) / 1000.0
	// 随机优化
	// timeResult = timeResult / 4
	//fmt.Printf("Throughput: %d Duration: %+v TPS: %f\n", timeCount, timeResult, count/timeResult)
	//fmt.Println("Time %8.2fs\tBlock %6d\tTx %6d\t \n", time.Since(now).Seconds(), block.Number, len(block.FilteredTransactions))

	//fmt.Println("Throughput:", timeCount, "Duration:", timeResult, "TPS:", count/timeResult)
	//fmt.Println("Write:", timeCount, "Duration:", timeResult, "TPS:", count/timeResult)

	fmt.Println("Throughput:", timeCount, "Duration:", strconv.FormatFloat(timeResult, 'g', 30, 32)+"s", "TPS:", count/timeResult)

	// wg.Add(MAX_CONNECT)
	// for i := 0; i < MAX_CONNECT; i++ {
	// 	//go invoceChaincode(contract[i])
	// 	go queryChaincode(contract[i])
	// }
	// timeStart = time.Now().UnixNano()
	// wg.Wait()
	// timeCount = Max_Count * MAX_CONNECT
	// timeEnd = time.Now().UnixNano()
	// count = float64(timeCount)
	// timeResult = float64((timeEnd-timeStart)/1e6) / 1000.0
	// // 状态通道随机优化
	// // timeResult = timeResult / 2
	// fmt.Println("Read:", timeCount, "Duration:", timeResult, "TPS:", count/timeResult)

	// for i := 0; i < 2; i++ {
	// 	uuid := uuid.New()
	// 	fmt.Println(uuid)
	// 	_, err := contract[0].SubmitTransaction("insert", uuid, strconv.Itoa(i))
	// 	//_, err := con.SubmitTransaction("createCar", "CAR"+uuidArr, "0", "2", "2", "4")
	// 	// uuidArr := uuid.New()
	// 	// _, err := con.SubmitTransaction("createCar", "CAR"+uuidArr, "0", "2", "2", "4")
	// 	if err != nil {
	// 		log.Fatalf("Failed to Submit transaction: %v", err)
	// 	}
	// }

	// log.Println("--> Submit Transaction: InitLedger, function creates the initial set of assets on the ledger")
	// result, err := contract.SubmitTransaction("initLedger")
	// if err != nil {
	// 	log.Fatalf("Failed to Submit transaction: %v", err)
	// }
	// log.Println(string(result))

	// log.Println("--> Evaluate Transaction: queryAllCars, function returns all the current assets on the ledger")
	// result, err := contract.EvaluateTransaction("queryAllCars")
	// if err != nil {
	// 	log.Fatalf("Failed to evaluate transaction: %v", err)
	// }
	// log.Println(string(result))

	// log.Println("--> changeCarOwner Transaction: changeCarOwner ")
	// result, err = contract.SubmitTransaction("changeCarOwner", "CAR1", "0")
	// if err != nil {
	// 	log.Fatalf("Failed to Submit transaction: %v", err)
	// }
	// log.Println(string(result))

	// log.Println("--> createCar Transaction: createCar ")
	// result, err = contract.SubmitTransaction("createCar", "CAR11", "0", "2", "2", "4")
	// if err != nil {
	// 	log.Fatalf("Failed to Submit transaction: %v", err)
	// }
	// log.Println(string(result))

	// log.Println("--> Submit Transaction: CreateAsset, creates new asset with ID, color, owner, size, and appraisedValue arguments")
	// result, err = contract.SubmitTransaction("CreateAsset", "asset13", "yellow", "5", "Tom", "1300")
	// if err != nil {
	// 	log.Fatalf("Failed to Submit transaction: %v", err)
	// }
	// log.Println(string(result))

	// log.Println("--> Evaluate Transaction: ReadAsset, function returns an asset with a given assetID")
	// result, err = contract.EvaluateTransaction("ReadAsset", "asset13")
	// if err != nil {
	// 	log.Fatalf("Failed to evaluate transaction: %v\n", err)
	// }
	// log.Println(string(result))

	// log.Println("--> Evaluate Transaction: AssetExists, function returns 'true' if an asset with given assetID exist")
	// result, err = contract.EvaluateTransaction("AssetExists", "asset1")
	// if err != nil {
	// 	log.Fatalf("Failed to evaluate transaction: %v\n", err)
	// }
	// log.Println(string(result))

	// log.Println("--> Submit Transaction: TransferAsset asset1, transfer to new owner of Tom")
	// _, err = contract.SubmitTransaction("TransferAsset", "asset1", "Tom")
	// if err != nil {
	// 	log.Fatalf("Failed to Submit transaction: %v", err)
	// }

	// log.Println("--> Evaluate Transaction: ReadAsset, function returns 'asset1' attributes")
	// result, err = contract.EvaluateTransaction("ReadAsset", "asset1")
	// if err != nil {
	// 	log.Fatalf("Failed to evaluate transaction: %v", err)
	// }
	// log.Println(string(result))
	// log.Println("============ application-golang ends ============")
}

func populateWallet(userName string, wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		".",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put(userName, identity)
}

// func invoke() {
// 	for {
// 		if len(ch) > 0 || flag {
// 			<-ch
// 			// a := <-ch
// 			fmt.Println("invoke")
// 		} else {
// 			break
// 		}
// 	}

// 	wg.Done()
// }

// func query() {
// 	for {
// 		if len(ch) > 0 || flag {
// 			<-ch
// 			// a := <-ch
// 			fmt.Println("invoke")
// 		} else {
// 			break
// 		}
// 	}

// 	// ch <- 1
// 	wg.Done()
// 	fmt.Println("query2")
// }

// func queue() {
// 	for i := 0; i < 10; i++ {
// 		ch <- 1
// 	}
// 	flag = false
// 	wg.Done()
// }
