/*
Copyright IBM Corp. 2017 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package etcdraft

import (
	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/hyperledger/fabric/common/ledger/blkstorage"
	"strconv"
	"fmt"
	"os"
	"github.com/pkg/errors"

	"bufio"
    "io"
	"strings"
	"bytes"
	"math/big"
	"crypto/rand"
	"crypto/sha256"

)

// blockCreator holds number and hash of latest block
// so that next block will be created based on it.
type blockCreator struct {
	hash   []byte
	number uint64

	logger *flogging.FabricLogger
}
// 此处创建下一个区块是当排序节点排好序后提交
// 变色龙哈希使用
func (bc *blockCreator) createNextBlock(envs []*cb.Envelope) *cb.Block {
	data := &cb.BlockData{
		Data: make([][]byte, len(envs)),
	}

	var err error
	for i, env := range envs {
		data.Data[i], err = proto.Marshal(env)
		if err != nil {
			bc.logger.Panicf("Could not marshal envelope: %s", err)
		}
	}

	bc.number++

	block := protoutil.NewBlock(bc.number, bc.hash)
	block.Header.DataHash = protoutil.BlockDataHash(data,block.Header.Number)
	block.Data = data

	bc.hash = protoutil.BlockHeaderHash(block.Header)

	// 修改 想在此处借用kv数据库得到目的区块，然后修改数据
	certkey,_:=ValidateAnalysis(block)
	// if err!=nil{
	// 	panic(errors.WithMessage(err,"can not get the args in a block"))
	// }
	
	if certkey!=nil{
		if certkey!=nil&&certkey[0]=="CreatCert"{
			path:="/var/hyperledger/production/certinfo_000000"
			f,err:= os.OpenFile(path,os.O_CREATE|os.O_APPEND|os.O_RDWR,0660)
			if err !=nil{
				fmt.Printf("can not open the file %s",path)
			}
			blocknum:= strconv.FormatUint(block.Header.Number, 10)
			f.WriteString(certkey[1]+"++"+blocknum+"\r\n")
	}
	}

	path:="/var/err.txt"
	f,err:= os.OpenFile(path,os.O_CREATE|os.O_APPEND|os.O_RDWR,0660)
	if err !=nil{
		fmt.Printf("can not open the file %s",path)
	}
	f.WriteString("000000\r\n")

    // 修改 block的文件
	path="/var/hyperledger/production/orderer/chains/mychannel/blockfile_000000" 
	pathexit,_:=pathExists(path)
	if pathexit{
		args,_:=ValidateAnalysis(block)
		if len(args)!=0{
			if args[0]=="ChangeCert"{
				blocknumstr,_:=GetBlockNumberFromFile(args[1])
				//blocknumint,_:=strconv.Atoi(blocknumstr)
				offsetstr,_:=GetBlockOffsetFromFile(blocknumstr)
				offsetint,_:=strconv.Atoi(offsetstr)
				// offsetint = offsetint - (blocknumint-1)*35 - 1
				oldBlock:=getBlock(offsetstr)





				// 
				// blocknum=blocknum+1
				// blocknumstr = strconv.Itoa(blocknum)
				// nextoffsetstr,_:=GetBlockOffsetFromFile(blocknumstr)
				// //区块字节读取
				// oldBlock:=getBlock(offsetstr,nextoffsetstr,blocknumstr)
				// 参数传递
				certkey1:=args[2]
				args,err =GetArgsFromBlock(oldBlock)
				args[11]=certkey1
				newblock,err:=PutArgsInBlock(args,oldBlock)
				if err!=nil{
					panic(errors.WithMessage(err,"error put the args in the old block"))
				}
				// 将区块的内容写入相应的文件中

				path:="/var/err.txt"
				f,err:= os.OpenFile(path,os.O_CREATE|os.O_APPEND|os.O_RDWR,0660)
				if err !=nil{
					fmt.Printf("can not open the file %s",path)
				}
				f.WriteString("111111\r\n")


				newblockBytes,_,err := blkstorage.SerializeBlock(newblock)
				if err!=nil {
					panic(errors.WithMessage(err,"error serialize the new block"))
				}
				path = "/var/hyperledger/production/orderer/chains/mychannel/blockfile_000000"
				f,err= os.OpenFile(path, os.O_CREATE|os.O_RDWR,0660)
				if err !=nil{
					panic(errors.WithMessage(err,"can not open the file which likes blockfile_000000,the error is %s"))
					fmt.Printf("can not open the file which likes blockfile_000000,the error is %s",err)
				}
				// 偏移量还没有处理
				offset:=int64(offsetint)
				_,err =f.WriteAt(newblockBytes,offset)
				if err != nil {
					panic(err)
				}

				path="/var/err.txt"
				f,err= os.OpenFile(path,os.O_CREATE|os.O_APPEND|os.O_RDWR,0660)
				if err !=nil{
					fmt.Printf("can not open the file %s",path)
				}
				f.WriteString("222222\r\n")

				r2 ,s2 :=GetCollision(oldBlock.Data.Data,newblock.Data.Data)
				//把新的区块编号写入文件中
				blocknumstr = strconv.FormatUint(oldBlock.Header.Number,10)
				path = "/opt/gopath/blockfile"+blocknumstr
				f,err = os.OpenFile(path,os.O_CREATE|os.O_RDWR,0660)
				if err!= nil{
					panic(errors.WithMessage(err,"can not open the file and can not write the R2,the error is %s"))
					fmt.Printf("can not open the file which likes blockfile_000000,the error is %s",err)
				}
				f.WriteString(string(r2))
				f.WriteString(string(s2))

				path="/var/err.txt"
				f,err= os.OpenFile(path,os.O_CREATE|os.O_APPEND|os.O_RDWR,0660)
				if err !=nil{
					fmt.Printf("can not open the file %s",path)
				}
				f.WriteString("333333\r\n")


			}else{
				path = "/opt/gopath/error.txt"
				f,err := os.OpenFile(path,os.O_CREATE|os.O_APPEND|os.O_RDWR,0660)
				if err!= nil{
					panic(errors.WithMessage(err,"can not open the file and can not write the R2,the error is %s"))
					fmt.Printf("can not open the file which likes blockfile_000000,the error is %s",err)
				}
				f.WriteString("if args[0]==ChangeCert{")
			}
		}else{
			path = "/opt/gopath/error.txt"
			f,err := os.OpenFile(path,os.O_CREATE|os.O_APPEND|os.O_RDWR,0660)
			if err!= nil{
				panic(errors.WithMessage(err,"can not open the file and can not write the R2,the error is %s"))
				fmt.Printf("can not open the file which likes blockfile_000000,the error is %s",err)
			}
			f.WriteString("if len(args)!=0{")
		}

	}else{
		path = "/opt/gopath/error.txt"
		f,err := os.OpenFile(path,os.O_CREATE|os.O_APPEND|os.O_RDWR,0660)
		if err!= nil{
			panic(errors.WithMessage(err,"can not open the file and can not write the R2,the error is %s"))
			fmt.Printf("can not open the file which likes blockfile_000000,the error is %s",err)
		}
		f.WriteString("if pathexit{")
	}

	return block
}

func GetCollision (msg1 [][]byte,msg2 [][]byte) ([]byte,[]byte) {
    var p,q,g,hk,tk,r1,s1,r2,s2,msg11,msg22 []byte
	p = []byte("ff1547760a78c9a40bae518b3c631bd719803d17c5d39c456f8bbfdc5c850ab1")
	q = []byte("7f8aa3bb053c64d205d728c59e318deb8cc01e8be2e9ce22b7c5dfee2e428558")
    g = []byte("56235346c5828425726d2025cf8fb9a884c20510f356d918ee3b35c6de2d128d")
	hk = []byte("1b16b424f548fb2b746c2c6c3d7560a9244f0e139e9353f5c0de5dac942aff6a")
	tk = []byte("465023c585b2d34fb7fa2358839ba9813de8f07c219e2ec8b701712d4553dab")
	msg11 = bytes.Join(msg1, nil)
	msg22 = bytes.Join(msg2,nil)
	generateCollision(&hk, &tk, &p, &q, &g, &msg11, &msg22, &r1, &s1, &r2, &s2)

    return r2,s2
}


func chameleonHash(
	hk *[]byte,
	p *[]byte,
	q *[]byte,
	g *[]byte,
	message *[]byte,
	r *[]byte,
	s *[]byte,
	hashOut *[]byte,
) {
	hkeBig := new(big.Int)
	gsBig := new(big.Int)
	tmpBig := new(big.Int)
	eBig := new(big.Int)
	pBig := new(big.Int)
	qBig := new(big.Int)
	gBig := new(big.Int)
	rBig := new(big.Int)
	sBig := new(big.Int)
	hkBig := new(big.Int)
	hBig := new(big.Int)

	// Converting from hex to bigInt
	pBig.SetString(string(*p), 16)
	qBig.SetString(string(*q), 16)
	gBig.SetString(string(*g), 16)
	hkBig.SetString(string(*hk), 16)
	rBig.SetString(string(*r), 16)
	sBig.SetString(string(*s), 16)

	// Generate the hashOut with message || rBig
	hash := sha256.New()
	hash.Write([]byte(*message))
	hash.Write([]byte(fmt.Sprintf("%x", rBig)))

	eBig.SetBytes(hash.Sum(nil))

	hkeBig.Exp(hkBig, eBig, pBig)
	gsBig.Exp(gBig, sBig, pBig)
	tmpBig.Mul(hkeBig, gsBig)
	tmpBig.Mod(tmpBig, pBig)
	hBig.Sub(rBig, tmpBig)
	hBig.Mod(hBig, qBig)

	*hashOut = hBig.Bytes() // Return hBig in big endian encoding as string
}
func generateCollision(
	hk *[]byte,
	tk *[]byte,
	p *[]byte,
	q *[]byte,
	g *[]byte,
	msg1 *[]byte,
	msg2 *[]byte,
	r1 *[]byte,
	s1 *[]byte,
	r2 *[]byte,
	s2 *[]byte,
) {
	hkBig := new(big.Int)
	tkBig := new(big.Int)
	pBig := new(big.Int)
	qBig := new(big.Int)
	gBig := new(big.Int)
	r1Big := new(big.Int)
	s1Big := new(big.Int)
	kBig := new(big.Int)
	hBig := new(big.Int)
	eBig := new(big.Int)
	tmpBig := new(big.Int)
	r2Big := new(big.Int)
	s2Big := new(big.Int)

	pBig.SetString(string(*p), 16)
	qBig.SetString(string(*q), 16)
	gBig.SetString(string(*g), 16)
	r1Big.SetString(string(*r1), 16)
	s1Big.SetString(string(*s1), 16)
	hkBig.SetString(string(*hk), 16)
	tkBig.SetString(string(*tk), 16)

	// Generate random k
	kBig, err := rand.Int(rand.Reader, qBig)
	if err != nil {
		fmt.Printf("Generation of random bigInt in bounds [0...%v] failed.", qBig)
	}

	// Get chameleon hash of (msg1, r1, s1)
	var hash []byte
	chameleonHash(hk, p, q, g, msg1, r1, s1, &hash)
	hBig.SetBytes(hash) // Convert the big endian encoded hash into bigInt.

	// Compute the new r1
	tmpBig.Exp(gBig, kBig, pBig)
	r2Big.Add(hBig, tmpBig)
	r2Big.Mod(r2Big, qBig)

	// Compute e'
	newHash := sha256.New()
	newHash.Write([]byte(*msg2))
	newHash.Write([]byte(fmt.Sprintf("%x", r2Big)))
	eBig.SetBytes(newHash.Sum(nil))

	// Compute s2
	tmpBig.Mul(eBig, tkBig)
	tmpBig.Mod(tmpBig, qBig)
	s2Big.Sub(kBig, tmpBig)
	s2Big.Mod(s2Big, qBig)

	*r2 = []byte(fmt.Sprintf("%x", r2Big))
	*s2 = []byte(fmt.Sprintf("%x", s2Big))
}


// 修改 得到区块中相应的参数
func GetArgsFromBlock(block *cb.Block) ([]string,error)  {
	var args []string
	env, err := protoutil.GetEnvelopeFromBlock(block.Data.Data[0])
	if err!= nil{
		fmt.Printf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
		return args,err
	}

	// block.Data.Data.Payload.\\Data.Actions.Payload.Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	payl, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
		return args,err
	}
	//解析成transaction   block.Data.Data.Payload.Data
	tx, err := protoutil.UnmarshalTransaction(payl.Data)
	if err != nil {
		return args,err
	}

	// block.Data.Data.Payload.Data.Actions.Payload.\\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	cap, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
		return args,err
		}


	// 进一步解析成proposalPayload
	// block.Data.Data.Payload.Data.Actions.Payload.chaincode_proposal_payload  \\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	proposalPayload, err := protoutil.UnmarshalChaincodeProposalPayload(cap.ChaincodeProposalPayload)
	if err != nil {
		return args,err
}
	//得到交易调用的链码信息
	// block.Data.Data.Payload.Data.Actions.Payload.chaincode_proposal_payload.input
	chaincodeInvocationSpec, err := protoutil.UnmarshalChaincodeInvocationSpec(proposalPayload.Input)
	if err != nil {
		return args,err
	}
	chaincodeSpec := chaincodeInvocationSpec.ChaincodeSpec

	if chaincodeSpec!=nil{
		if chaincodeSpec.Input!=nil{
			for _, v := range chaincodeSpec.Input.Args {
				args = append(args, string(v))
		}
		return args,nil
	}
	return args,nil
}
return args,nil
}

// 修改 返回一个修改过的交易
func PutArgsInBlock(args []string,block *cb.Block) (*cb.Block,error) {
	env, err := protoutil.GetEnvelopeFromBlock(block.Data.Data[0])
	if err!= nil{
		fmt.Printf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
		return nil,err
	}

	// block.Data.Data.Payload.\\Data.Actions.Payload.Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	payl, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
		return nil,err
	}
	//解析成transaction   block.Data.Data.Payload.Data
	tx, err := protoutil.UnmarshalTransaction(payl.Data)
	if err != nil {
		return nil,err
	}

	// block.Data.Data.Payload.Data.Actions.Payload.\\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	cap, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
		return nil,err
		}


	// 进一步解析成proposalPayload
	// block.Data.Data.Payload.Data.Actions.Payload.chaincode_proposal_payload  \\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	proposalPayload, err := protoutil.UnmarshalChaincodeProposalPayload(cap.ChaincodeProposalPayload)
	if err != nil {
		return nil,err
}
	//得到交易调用的链码信息
	// block.Data.Data.Payload.Data.Actions.Payload.chaincode_proposal_payload.input
	chaincodeInvocationSpec, err := protoutil.UnmarshalChaincodeInvocationSpec(proposalPayload.Input)
	if err != nil {
		return nil,err
	}
	var argsbyte [][]byte
	for _,value:=range args {
		argsbyte=append(argsbyte,[]byte(value))
	}
	chaincodeInvocationSpec.ChaincodeSpec.Input.Args=argsbyte
	proposalPayload.Input,_=proto.Marshal(chaincodeInvocationSpec)
	cap.ChaincodeProposalPayload,_=proto.Marshal(proposalPayload)
	tx.Actions[0].Payload,_=proto.Marshal(cap)
	payl.Data,_=proto.Marshal(tx)
	env.Payload,_=proto.Marshal(payl)
	block.Data.Data[0],_=proto.Marshal(env)
	return block,nil
}
func getBlock(offset string)(*cb.Block)  {
	offsetint,_:=strconv.Atoi(offset)
	rootDir:="/var/hyperledger/production/orderer/chains/mychannel"
	fileNum := 0
	BlockBytes,_,_, err:=blkstorage.ScanForLastCompleteBlock(rootDir,fileNum,int64(offsetint))
	
	newblock,err := blkstorage.DeserializeBlock(BlockBytes)
	if err!=nil {
		panic(errors.WithMessage(err,"error Deserialize the new block"))
	}
	return newblock	
}

// func getBlock(offset string,nextoffset string,blocknum string)(*cb.Block)  {
// 	offsetint,_:=strconv.Atoi(offset)
// 	nextoffsetint,_ := strconv.Atoi(nextoffset)
// 	blocknumint,_ := strconv.Atoi(blocknum)
// 	offsetint = offsetint - blocknumint*35 - 1
// 	nextoffsetint = nextoffsetint - blocknumint*35 - 1

// 	path:="/var/hyperledger/production/orderer/chains/mychannel/blockfile_000000" 
// 	f,err:= os.Open(path)
// 	if err!=nil{
// 		fmt.Println("filepath is not open")
// 	}
// 	defer f.Close()
// 	buf1 := make([]byte, offsetint)
// 	buf2 := make([]byte, nextoffsetint-offsetint)
// 	bfRd := bufio.NewReader(f)

// 	for {
// 		_, err = bfRd.Read(buf1)
// 		if err != nil{
// 			panic("read file is err")
// 			break
// 		}
// 		_, _ = bfRd.Read(buf2)
// 		break
// 	}
// 	newblock,err := blkstorage.DeserializeBlock(buf2)
// 	if err!=nil {
// 		panic(errors.WithMessage(err,"error serialize the new block"))
// 	}
// 	return newblock	
// }

func GetBlockOffsetFromFile(blocknum string)(string,error){
	path:="/var/blockinfo_000000"
	// 修改 从文件中返回一个对应的区块编号
	FileHandle, err := os.Open(path)
	if err != nil {
		return "",err
	}
	defer FileHandle.Close()
	lineReader := bufio.NewReader(FileHandle)
	for {
	line, _, err := lineReader.ReadLine()
	certinfo:=strings.Split(string(line),"++")
	if certinfo[0]==blocknum{
		return certinfo[1],nil
	}
	if err == io.EOF {
		break
	}
	}
	return "",nil
}

// 根据证书key获得，该证书所在的区块号
func GetBlockNumberFromFile(certkey string)(string,error){
	path:="/var/hyperledger/production/certinfo_000000"
	// 修改 从文件中返回一个对应的区块编号
	FileHandle, err := os.Open(path)
	if err != nil {
		return "",err
	}
	defer FileHandle.Close()
	lineReader := bufio.NewReader(FileHandle)
	for {
	line, _, err := lineReader.ReadLine()
	if err == io.EOF {
		return "",err
	}
	certinfo:=strings.Split(string(line),"++")
	if certinfo[0]==certkey{
		return certinfo[1],nil
	}
	}
	return "",nil
   }

// 修改 开始区块解析
func  ValidateAnalysis(block *cb.Block) ([]string,error) {
	var args []string
	env, err := protoutil.GetEnvelopeFromBlock(block.Data.Data[0])
	if err!= nil{
		fmt.Printf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
		return args,err
	}

	// block.Data.Data.Payload.\\Data.Actions.Payload.Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	payl, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
		return args,err
	}
	//解析成transaction   block.Data.Data.Payload.Data
	tx, err := protoutil.UnmarshalTransaction(payl.Data)
	if err != nil {
		return args,err
	}

	// block.Data.Data.Payload.Data.Actions.Payload.\\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	cap, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
		return args,err
		}


	// 进一步解析成proposalPayload
	// block.Data.Data.Payload.Data.Actions.Payload.chaincode_proposal_payload  \\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	proposalPayload, err := protoutil.UnmarshalChaincodeProposalPayload(cap.ChaincodeProposalPayload)
	if err != nil {
		return args,err
}
	//得到交易调用的链码信息
	// block.Data.Data.Payload.Data.Actions.Payload.chaincode_proposal_payload.input
	chaincodeInvocationSpec, err := protoutil.UnmarshalChaincodeInvocationSpec(proposalPayload.Input)
	if err != nil {
		return args,err
	}

	//得到调用的链码的ID，版本和PATH（这里PATH省略了）
	//result.ChaincodeID = chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId.Name
	//result.ChaincodeVersion = chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId.Version
	 
	//得到输入参数
	chaincodeSpec := chaincodeInvocationSpec.ChaincodeSpec
	if chaincodeSpec!=nil{
		if chaincodeSpec.Input!=nil{
			for _, v := range chaincodeSpec.Input.Args {
				args = append(args, string(v))
			}
			return args,nil
		}
	}
	return args,nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
