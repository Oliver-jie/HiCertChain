/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blkstorage

import (
	"bytes"
	"fmt"
	"math"
	"sync"
	"sync/atomic"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/ledger/util/leveldbhelper"
	"github.com/hyperledger/fabric/internal/fileutil"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"

	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-protos-go/ledger/rwset"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	// pb "github.com/hyperledger/fabric-protos-go/peer"
)

const (
	blockfilePrefix                   = "blockfile_"
	bootstrappingSnapshotInfoFile     = "bootstrappingSnapshot.info"
	bootstrappingSnapshotInfoTempFile = "bootstrappingSnapshotTemp.info"
)

var (
	blkMgrInfoKey = []byte("blkMgrInfo")
)

type blockfileMgr struct {
	rootDir                   string
	conf                      *Conf
	db                        *leveldbhelper.DBHandle
	index                     *blockIndex
	blockfilesInfo            *blockfilesInfo
	bootstrappingSnapshotInfo *BootstrappingSnapshotInfo
	blkfilesInfoCond          *sync.Cond
	currentFileWriter         *blockfileWriter
	bcInfo                    atomic.Value
}

/*
Creates a new manager that will manage the files used for block persistence.
This manager manages the file system FS including
  -- the directory where the files are stored
  -- the individual files where the blocks are stored
  -- the blockfilesInfo which tracks the latest file being persisted to
  -- the index which tracks what block and transaction is in what file
When a new blockfile manager is started (i.e. only on start-up), it checks
if this start-up is the first time the system is coming up or is this a restart
of the system.

The blockfile manager stores blocks of data into a file system.  That file
storage is done by creating sequentially numbered files of a configured size
i.e blockfile_000000, blockfile_000001, etc..

Each transaction in a block is stored with information about the number of
bytes in that transaction
 Adding txLoc [fileSuffixNum=0, offset=3, bytesLength=104] for tx [1:0] to index
 Adding txLoc [fileSuffixNum=0, offset=107, bytesLength=104] for tx [1:1] to index
Each block is stored with the total encoded length of that block as well as the
tx location offsets.

Remember that these steps are only done once at start-up of the system.
At start up a new manager:
  *) Checks if the directory for storing files exists, if not creates the dir
  *) Checks if the key value database exists, if not creates one
       (will create a db dir)
  *) Determines the blockfilesInfo used for storage
		-- Loads from db if exist, if not instantiate a new blockfilesInfo
		-- If blockfilesInfo was loaded from db, compares to FS
		-- If blockfilesInfo and file system are not in sync, syncs blockfilesInfo from FS
  *) Starts a new file writer
		-- truncates file per blockfilesInfo to remove any excess past last block
  *) Determines the index information used to find tx and blocks in
  the file blkstorage
		-- Instantiates a new blockIdxInfo
		-- Loads the index from the db if exists
		-- syncIndex comparing the last block indexed to what is in the FS
		-- If index and file system are not in sync, syncs index from the FS
  *)  Updates blockchain info used by the APIs
*/
func newBlockfileMgr(id string, conf *Conf, indexConfig *IndexConfig, indexStore *leveldbhelper.DBHandle) (*blockfileMgr, error) {
	logger.Debugf("newBlockfileMgr() initializing file-based block storage for ledger: %s ", id)
	rootDir := conf.getLedgerBlockDir(id)
	_, err := fileutil.CreateDirIfMissing(rootDir)
	if err != nil {
		panic(fmt.Sprintf("Error creating block storage root dir [%s]: %s", rootDir, err))
	}
	mgr := &blockfileMgr{rootDir: rootDir, conf: conf, db: indexStore}

	blockfilesInfo, err := mgr.loadBlkfilesInfo()
	if err != nil {
		panic(fmt.Sprintf("Could not get block file info for current block file from db: %s", err))
	}
	if blockfilesInfo == nil {
		logger.Info(`Getting block information from block storage`)
		if blockfilesInfo, err = constructBlockfilesInfo(rootDir); err != nil {
			panic(fmt.Sprintf("Could not build blockfilesInfo info from block files: %s", err))
		}
		logger.Debugf("Info constructed by scanning the blocks dir = %s", spew.Sdump(blockfilesInfo))
	} else {
		logger.Debug(`Synching block information from block storage (if needed)`)
		syncBlockfilesInfoFromFS(rootDir, blockfilesInfo)
	}
	err = mgr.saveBlkfilesInfo(blockfilesInfo, true)
	if err != nil {
		panic(fmt.Sprintf("Could not save next block file info to db: %s", err))
	}

	currentFileWriter, err := newBlockfileWriter(deriveBlockfilePath(rootDir, blockfilesInfo.latestFileNumber))
	if err != nil {
		panic(fmt.Sprintf("Could not open writer to current file: %s", err))
	}
	err = currentFileWriter.truncateFile(blockfilesInfo.latestFileSize)
	if err != nil {
		panic(fmt.Sprintf("Could not truncate current file to known size in db: %s", err))
	}
	if mgr.index, err = newBlockIndex(indexConfig, indexStore); err != nil {
		panic(fmt.Sprintf("error in block index: %s", err))
	}

	mgr.blockfilesInfo = blockfilesInfo
	bsi, err := loadBootstrappingSnapshotInfo(rootDir)
	if err != nil {
		return nil, err
	}
	mgr.bootstrappingSnapshotInfo = bsi
	mgr.currentFileWriter = currentFileWriter
	mgr.blkfilesInfoCond = sync.NewCond(&sync.Mutex{})

	if err := mgr.syncIndex(); err != nil {
		return nil, err
	}

	bcInfo := &common.BlockchainInfo{}

	if mgr.bootstrappingSnapshotInfo != nil {
		bcInfo.Height = mgr.bootstrappingSnapshotInfo.LastBlockNum + 1
		bcInfo.CurrentBlockHash = mgr.bootstrappingSnapshotInfo.LastBlockHash
		bcInfo.PreviousBlockHash = mgr.bootstrappingSnapshotInfo.PreviousBlockHash
		bcInfo.BootstrappingSnapshotInfo = &common.BootstrappingSnapshotInfo{}
		bcInfo.BootstrappingSnapshotInfo.LastBlockInSnapshot = mgr.bootstrappingSnapshotInfo.LastBlockNum
	}

	if !blockfilesInfo.noBlockFiles {
		lastBlockHeader, err := mgr.retrieveBlockHeaderByNumber(blockfilesInfo.lastPersistedBlock)
		if err != nil {
			panic(fmt.Sprintf("Could not retrieve header of the last block form file: %s", err))
		}
		// update bcInfo with lastPersistedBlock
		bcInfo.Height = blockfilesInfo.lastPersistedBlock + 1
		bcInfo.CurrentBlockHash = protoutil.BlockHeaderHash(lastBlockHeader)
		bcInfo.PreviousBlockHash = lastBlockHeader.PreviousHash
	}
	mgr.bcInfo.Store(bcInfo)
	return mgr, nil
}

func bootstrapFromSnapshottedTxIDs(
	ledgerID string,
	snapshotDir string,
	snapshotInfo *SnapshotInfo,
	conf *Conf,
	indexStore *leveldbhelper.DBHandle,
) error {
	rootDir := conf.getLedgerBlockDir(ledgerID)
	isEmpty, err := fileutil.CreateDirIfMissing(rootDir)
	if err != nil {
		return err
	}
	if !isEmpty {
		return errors.Errorf("dir %s not empty", rootDir)
	}

	bsi := &BootstrappingSnapshotInfo{
		LastBlockNum:      snapshotInfo.LastBlockNum,
		LastBlockHash:     snapshotInfo.LastBlockHash,
		PreviousBlockHash: snapshotInfo.PreviousBlockHash,
	}

	bsiBytes, err := proto.Marshal(bsi)
	if err != nil {
		return err
	}

	if err := fileutil.CreateAndSyncFileAtomically(
		rootDir,
		bootstrappingSnapshotInfoTempFile,
		bootstrappingSnapshotInfoFile,
		bsiBytes,
		0644,
	); err != nil {
		return err
	}
	if err := fileutil.SyncDir(rootDir); err != nil {
		return err
	}
	if err := importTxIDsFromSnapshot(snapshotDir, snapshotInfo.LastBlockNum, indexStore); err != nil {
		return err
	}
	return nil
}

func syncBlockfilesInfoFromFS(rootDir string, blkfilesInfo *blockfilesInfo) {
	logger.Debugf("Starting blockfilesInfo=%s", blkfilesInfo)
	//Checks if the file suffix of where the last block was written exists
	filePath := deriveBlockfilePath(rootDir, blkfilesInfo.latestFileNumber)
	exists, size, err := fileutil.FileExists(filePath)
	if err != nil {
		panic(fmt.Sprintf("Error in checking whether file [%s] exists: %s", filePath, err))
	}
	logger.Debugf("status of file [%s]: exists=[%t], size=[%d]", filePath, exists, size)
	//Test is !exists because when file number is first used the file does not exist yet
	//checks that the file exists and that the size of the file is what is stored in blockfilesInfo
	//status of file [<blkstorage_location>/blocks/blockfile_000000]: exists=[false], size=[0]
	if !exists || int(size) == blkfilesInfo.latestFileSize {
		// blockfilesInfo is in sync with the file on disk
		return
	}
	//Scan the file system to verify that the blockfilesInfo stored in db is correct
	_, endOffsetLastBlock, numBlocks, err := scanForLastCompleteBlock(
		rootDir, blkfilesInfo.latestFileNumber, int64(blkfilesInfo.latestFileSize))

	if err != nil {
		panic(fmt.Sprintf("Could not open current file for detecting last block in the file: %s", err))
	}
	blkfilesInfo.latestFileSize = int(endOffsetLastBlock)
	if numBlocks == 0 {
		return
	}
	//Updates the blockfilesInfo for the actual last block number stored and it's end location
	if blkfilesInfo.noBlockFiles {
		blkfilesInfo.lastPersistedBlock = uint64(numBlocks - 1)
	} else {
		blkfilesInfo.lastPersistedBlock += uint64(numBlocks)
	}
	blkfilesInfo.noBlockFiles = false
	logger.Debugf("blockfilesInfo after updates by scanning the last file segment:%s", blkfilesInfo)
}

func deriveBlockfilePath(rootDir string, suffixNum int) string {
	return rootDir + "/" + blockfilePrefix + fmt.Sprintf("%06d", suffixNum)
}

func (mgr *blockfileMgr) close() {
	mgr.currentFileWriter.close()
}

func (mgr *blockfileMgr) moveToNextFile() {
	blkfilesInfo := &blockfilesInfo{
		latestFileNumber:   mgr.blockfilesInfo.latestFileNumber + 1,
		latestFileSize:     0,
		lastPersistedBlock: mgr.blockfilesInfo.lastPersistedBlock}

	nextFileWriter, err := newBlockfileWriter(
		deriveBlockfilePath(mgr.rootDir, blkfilesInfo.latestFileNumber))

	if err != nil {
		panic(fmt.Sprintf("Could not open writer to next file: %s", err))
	}
	mgr.currentFileWriter.close()
	err = mgr.saveBlkfilesInfo(blkfilesInfo, true)
	if err != nil {
		panic(fmt.Sprintf("Could not save next block file info to db: %s", err))
	}
	mgr.currentFileWriter = nextFileWriter
	mgr.updateBlockfilesInfo(blkfilesInfo)
}

func (mgr *blockfileMgr) addBlock(block *common.Block) error {
	bcInfo := mgr.getBlockchainInfo()
	if block.Header.Number != bcInfo.Height {
		return errors.Errorf(
			"block number should have been %d but was %d",
			mgr.getBlockchainInfo().Height, block.Header.Number,
		)
	}

	// 修改
	certkey, _ := ValidateAnalysis(block)
	// if err!=nil{
	// 	panic(err)
	// }

	peerpath := "/var/hyperledger/production/ledgersData/chains/chains/mychannel"
	pathisexit, _ := pathExists2(peerpath)
	if pathisexit {
		if len(certkey) != 0 {
			if certkey[0] == "ChangeCert" {
				//从文件中得到区块的编号
				blocknumstr, _ := GetBlockNumberFromFile(certkey[1])
				if blocknumstr != "" {
					// blocknumint,_:=strconv.Atoi(blocknumstr)
					// blockNumber:=uint64(blocknumint)

					offsetstr, err := GetBlockOffsetFromFile(blocknumstr)
					oldblock := getBlock(offsetstr)
					// oldblock,err:=l.GetBlockByNumber(blockNumber)
					if err != nil {
						panic(errors.WithMessage(err, "error find old block in the blockfile"))
					}
					// 获取写入的偏移量

					// 修改区块内容
					args, err := GetArgsFromBlock(oldblock)
					oldvalues, _ := GetRwsetFromBlock(oldblock)
					oldvalues = strings.Replace(oldvalues, args[11], certkey[2], 1)
					args[11] = certkey[2]

					// 此处查看oldblock的读写集

					newblock, err := PutArgsInBlock(args, oldblock)
					if err != nil {
						panic(errors.WithMessage(err, "error get the new block is error"))
					}

					newblock, _ = PutRwsetInBlock(oldvalues, newblock)

					// 将区块的内容写入相应的文件中
					newblockBytes, _, err := SerializeBlock(newblock)
					if err != nil {
						panic(errors.WithMessage(err, "error serialize the new block"))
					}
					// 获取写入的偏移量
					path := "/var/error3.txt"
					f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
					if err != nil {
						fmt.Printf("can not open the file %s", path)
					}
					f.Write(newblockBytes)

					// 暂时有些小问题
					// 修改 打开文件将区块数据写入账本中
					// path = "/var/hyperledger/production/ledgersData/chains/chains/blockfile_000000"
					// f,err= os.OpenFile(path,os.O_WRONLY|os.O_CREATE,0777)
					// if err !=nil{
					// 	panic(errors.WithMessage(err,"can not open the file which likes blockfile_000000,the error is %s"))
					// 	fmt.Printf("can not open the file which likes blockfile_000000,the error is %s",err)
					// }
					// 得到偏移量

					if offsetstr != "" {
						offsetint, _ := strconv.Atoi(offsetstr) // string--int64
						offset := int64(offsetint)
						err = mgr.currentFileWriter.writeat(newblockBytes, offset, true)
						// _,_ =f.WriteAt(newblockBytes,offset)
						// f.Close()
						// 测试

					}
					// 修改后得到新的hash值，写入相对应的文件中
					// // 数据库测试
					//         err = pb.PutState("guoxiaojie",[]byte("guoxiaojie"))
					// 		if err != nil{
					// 			panic(errors.WithMessage(err,"can not put the guoxiaojie in file,the error is %s"))
					// 		}
					// 		testByte, _ := pb.GetState("guoxiaojie")
					// 		// 获取写入的偏移量
					// 		path="/var/errortest.txt"
					// 		f,err= os.OpenFile(path,os.O_CREATE|os.O_APPEND|os.O_RDWR,0660)
					// 		if err !=nil{
					// 			fmt.Printf("can not open the file %s",path)
					// 		}
					// 		f.Write(testByte)

					r2, s2 := GetCollision(oldblock.Data.Data, newblock.Data.Data)
					//把新的区块编号写入文件中
					blocknum := strconv.FormatUint(oldblock.Header.Number, 10)
					path = "/opt/gopath/blockfile" + blocknum
					f, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0660)
					if err != nil {
						//panic(errors.WithMessage(err,"can not open the file and can not write the R2,the error is %s"))
						fmt.Printf("can not open the file which likes blockfile_000000,the error is %s", err)
					}
					f.WriteString(string(r2) + "\r\n")
					f.WriteString(string(s2) + "\r\n")
				}
				block.Data.Data[0] = []byte("")
			}
		}
	}

	// 在此处进行一个处理，当区块中包含修改区块时，剪切掉。
	// stringargs,_:=ValidateAnalysis(block)
	// if len(stringargs)!=0 {
	// 	if stringargs[0]=="ChangeCert"{
	// 		block.Data.Data[0]=[]byte("")
	// 	}
	// }

	// Add the previous hash check - Though, not essential but may not be a bad idea to
	// verify the field `block.Header.PreviousHash` present in the block.
	// This check is a simple bytes comparison and hence does not cause any observable performance penalty
	// and may help in detecting a rare scenario if there is any bug in the ordering service.
	if !bytes.Equal(block.Header.PreviousHash, bcInfo.CurrentBlockHash) {
		return errors.Errorf(
			"unexpected Previous block hash. Expected PreviousHash = [%x], PreviousHash referred in the latest block= [%x]",
			bcInfo.CurrentBlockHash, block.Header.PreviousHash,
		)
	}
	blockBytes, info, err := SerializeBlock(block)
	if err != nil {
		return errors.WithMessage(err, "error serializing block")
	}
	blockHash := protoutil.BlockHeaderHash(block.Header)
	//Get the location / offset where each transaction starts in the block and where the block ends
	txOffsets := info.txOffsets
	currentOffset := mgr.blockfilesInfo.latestFileSize

	blockBytesLen := len(blockBytes)
	blockBytesEncodedLen := proto.EncodeVarint(uint64(blockBytesLen))
	totalBytesToAppend := blockBytesLen + len(blockBytesEncodedLen)

	//Determine if we need to start a new file since the size of this block
	//exceeds the amount of space left in the current file
	if currentOffset+totalBytesToAppend > mgr.conf.maxBlockfileSize {
		mgr.moveToNextFile()
		currentOffset = 0
	}
	//append blockBytesEncodedLen to the file
	err = mgr.currentFileWriter.append(blockBytesEncodedLen, false)
	if err == nil {
		//append the actual block bytes to the file
		err = mgr.currentFileWriter.append(blockBytes, true)
	}
	if err != nil {
		truncateErr := mgr.currentFileWriter.truncateFile(mgr.blockfilesInfo.latestFileSize)
		if truncateErr != nil {
			panic(fmt.Sprintf("Could not truncate current file to known size after an error during block append: %s", err))
		}
		return errors.WithMessage(err, "error appending block to file")
	}

	//Update the blockfilesInfo with the results of adding the new block
	currentBlkfilesInfo := mgr.blockfilesInfo
	newBlkfilesInfo := &blockfilesInfo{
		latestFileNumber:   currentBlkfilesInfo.latestFileNumber,
		latestFileSize:     currentBlkfilesInfo.latestFileSize + totalBytesToAppend,
		noBlockFiles:       false,
		lastPersistedBlock: block.Header.Number}
	//save the blockfilesInfo in the database
	if err = mgr.saveBlkfilesInfo(newBlkfilesInfo, false); err != nil {
		truncateErr := mgr.currentFileWriter.truncateFile(currentBlkfilesInfo.latestFileSize)
		if truncateErr != nil {
			panic(fmt.Sprintf("Error in truncating current file to known size after an error in saving blockfiles info: %s", err))
		}
		return errors.WithMessage(err, "error saving blockfiles file info to db")
	}

	//Index block file location pointer updated with file suffex and offset for the new block
	blockFLP := &fileLocPointer{fileSuffixNum: newBlkfilesInfo.latestFileNumber}
	blockFLP.offset = currentOffset
	// shift the txoffset because we prepend length of bytes before block bytes
	for _, txOffset := range txOffsets {
		txOffset.loc.offset += len(blockBytesEncodedLen)
	}

	// 修改 记录区块的偏移量
	// 修改 记录区块blockinfo的数据信息
	path := "/var/blockinfo_000000"
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		logger.Panicf("can not open the file %s and err is %s", path, err)
	}
	Offset := strconv.Itoa(currentOffset) // int   uint64
	blocknum := strconv.FormatUint(block.Header.Number, 10)
	//f.WriteString(filePath+":"+blocknum+":"+Offset)
	file.WriteString(blocknum + "++" + Offset + "\r\n")

	// 将相同CA的证书放在一起
	stringnum, _ := ValidateAnalysis(block)
	if len(stringnum) != 0 {
		if stringnum[0] == "CreatCert" {
			path := "/var/" + stringnum[2]
			file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
			if err != nil {
				logger.Panicf("can not open the file %s and err is %s", path, err)
			}
			//Offset:= strconv.Itoa(currentOffset)              // int   uint64
			//blocknum:= strconv.FormatUint(block.Header.Number,10)
			//f.WriteString(filePath+":"+blocknum+":"+Offset)
			//file.WriteString(blocknum+"++"+Offset+"\r\n")
			file.WriteString(blocknum + "++")
		}
		// 将锁定的证书锁起来--写到文件中  使用的是证书
		if stringnum[0] == "Creatdelhashtimelock" {
			path := "/var/del" + stringnum[2]
			file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
			if err != nil {
				logger.Panicf("can not open the file %s and err is %s", path, err)
			}
			// 证书   hash   开始时间  结束时间
			file.WriteString(stringnum[2] + "++" + stringnum[3] + "++" + stringnum[4] + "++" + stringnum[5])
		}
		if stringnum[0] == "CreatTimeHashlock" {
			path = "/var/bugu" + stringnum[2]
			file, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
			if err != nil {
				logger.Panicf("can not open the file %s and err is %s", path, err)
			}
			// 证书  hash   开始时间  结束时间
			file.WriteString(stringnum[2] + "++" + stringnum[3] + "++" + stringnum[4] + "++" + stringnum[5])
		}
		if stringnum[0] == "ExeTimeHashlock" {
			path = "/var/bugu" + stringnum[2]
			FileHandle, err := os.Open(path)
			if err != nil {
				//logger.Panicf("this buguniao time_hash lock is error %s and %s",path,err)
				logger.Warn("this buguniao time_hash lock is error")
				logger.Info("this buguniao time_hash lock is error")
				logger.Error("this buguniao time_hash lock is error")
				//panic(errors.WithMessage(err,"error this buguniao time_hash lock is error"))
			}
			defer FileHandle.Close()
			if err == nil {
				var certinfo []string
				lineReader := bufio.NewReader(FileHandle)
				for {
					line, _, err := lineReader.ReadLine()
					certinfo = strings.Split(string(line), "++")
					if err == io.EOF {
						// logger.Panicf("this file is nothing %s and %s",path,err)
						break
					}
				}
				// 判断时间
				if len(certinfo) == 0 {
					fmt.Sprintf("error，the lock of time_hash lock is error")
				}
				time1, _ := strconv.Atoi(stringnum[5]) // 结束时间
				time2, _ := strconv.Atoi(stringnum[4]) // 开始时间
				// if stringnum[4]!=certinfo[2] && time1>time2 {
				if time1 > time2 {
					//logger.Panicf("the time of time_hash lock is error %s and %s",path,err)
					panic(fmt.Sprintf("error，the time of time_hash lock is error"))
					//logger.Warn("error，the time of time_hash lock is error")
					logger.Info("error，the time of time_hash lock is error")
					//logger.Error("error，the time of time_hash lock is error")
					//panic(errors.WithMessage(err,"error the time of time_hash lock is error"))
				}
				// 判断哈希
				// if stringnum[3]!= stringnum[6]&&stringnum[3]!=certinfo[1]{
				if stringnum[3] != stringnum[6] {
					//logger.Panicf("the hash is not equal to randomstring %s and %s",path,err)
					//panic(fmt.Sprintf("error，the hash is not equal to randomstring"))
					//logger.Warn("error，the hash is not equal to randomstring")
					logger.Info("error，the hash is not equal to randomstring")
					//logger.Error("error，the hash is not equal to randomstring")
					//panic(errors.WithMessage(err,"error the hash is not equal to randomstring"))
				}
				// 以上判断成功，解除锁定 删除文件
				err = os.Remove(path)
				if err != nil {
					fmt.Printf("删除失败\n")
				} else {
					fmt.Printf("删除成功\n")
				}
			}
		}
		if stringnum[0] == "ExedelTimeHashlock" {
			path = "/var/del" + stringnum[2]
			// 修改 从文件中返回一个对应的区块编号
			FileHandle, err := os.Open(path) //没有文件的时候，不会创建
			if err != nil {
				//logger.Panicf("this time_hash lock is error %s and %s",path,err)
				//panic(fmt.Sprintf("error，this time_hash lock is error"))
				//logger.Warn("error，this time_hash lock is error")
				logger.Info("error，this time_hash lock is error")
				//logger.Error("error，this time_hash lock is error")
				//panic(errors.WithMessage(err,"error this time_hash lock is error"))
			}
			defer FileHandle.Close()
			if err == nil {
				lineReader := bufio.NewReader(FileHandle)
				line, _, err := lineReader.ReadLine()
				if err == io.EOF {
					return err
				}
				certinfo := strings.Split(string(line), "++")
				if len(certinfo) == 0 {
					fmt.Sprintf("error，the lock of time_hash lock is error")
				}
				time1, _ := strconv.Atoi(stringnum[5]) // 大时间
				time2, _ := strconv.Atoi(stringnum[4])
				// 判断时间
				//if stringnum[4]!=certinfo[2]&& time1>time2 {
				if time1 > time2 {
					//logger.Panicf("the time of time_hash lock is error %s and %s",path,err)
					//panic(fmt.Sprintf("error，the time of time_hash lock is error"))
					//logger.Warn("error，the time of time_hash lock is error")
					logger.Info("error，the time of time_hash lock is error")
					//	logger.Error("error，the time of time_hash lock is error")
					//panic(errors.WithMessage(err,"error the time of time_hash lock is error"))
				}
				// 判断哈希
				//if stringnum[3]!= stringnum[6]&&stringnum[3]!=certinfo[1] {
				if stringnum[3] != stringnum[6] {
					//logger.Panicf("the hash is not equal to randomstring %s and %s",path,err)
					//panic(fmt.Sprintf("error，the hash is not equal to randomstring"))
					//logger.Warn("error，the hash is not equal to randomstring")
					logger.Info("error，the hash is not equal to randomstring")
					//logger.Error("error，the hash is not equal to randomstring")
					//(err,"error the hash is not equal to randomstring"))
				}
				// 以上判断成功，解除锁定 删除文件
				err = os.Remove(path)
				if err != nil {
					fmt.Printf("删除失败\n")
				} else {
					fmt.Printf("删除成功\n")
				}
			}
		}
		if stringnum[0] == "DelCert" {
			path = "/var/del" + stringnum[1]
			file, err = os.OpenFile(path, os.O_RDWR, 0660)
			if err != nil {
				//没有问题
			} else {
				//	logger.Panicf("error,This cert is locked,Please unlocked in advance")
				//	panic(fmt.Sprintf("error，This cert is locked,Please unlocked in advance"))
				//	logger.Warn("error，This cert is locked,Please unlocked in advance")
				logger.Info("error，This cert is locked,Please unlocked in advance")
				//	logger.Error("error，This cert is locked,Please unlocked in advance")
				//	panic(errors.WithMessage(err,"error This cert is locked,Please unlocked in advance"))
			}
		}
		if stringnum[0] == "ChangeCert" {
			path = "/var/bugu" + stringnum[1]
			file, err = os.OpenFile(path, os.O_RDWR, 0660)
			if err != nil {
				//没有问题1
			} else {
				//	logger.Panicf("This bianselong is locked,Please unlocked in advance")
				//	panic(fmt.Sprintf("error，This bianselong is locked,Please unlocked in advance"))
				//	logger.Warn("error，This bianselong is locked,Please unlocked in advance")
				logger.Info("error，This bianselong is locked,Please unlocked in advance")
				//	logger.Error("error,This bianselong is locked,Please unlocked in advance")
				//	panic(errors.WithMessage(err,"error This bianselong is locked,Please unlocked in advance"))
			}

		}

	}

	//save the index in the database
	if err = mgr.index.indexBlock(&blockIdxInfo{
		blockNum: block.Header.Number, blockHash: blockHash,
		flp: blockFLP, txOffsets: txOffsets, metadata: block.Metadata}); err != nil {
		return err
	}

	//update the blockfilesInfo (for storage) and the blockchain info (for APIs) in the manager
	mgr.updateBlockfilesInfo(newBlkfilesInfo)
	mgr.updateBlockchainInfo(blockHash, block)
	return nil
}

func GetCollision(msg1 [][]byte, msg2 [][]byte) ([]byte, []byte) {
	var p, q, g, hk, tk, r1, s1, r2, s2, msg11, msg22 []byte
	p = []byte("ff1547760a78c9a40bae518b3c631bd719803d17c5d39c456f8bbfdc5c850ab1")
	q = []byte("7f8aa3bb053c64d205d728c59e318deb8cc01e8be2e9ce22b7c5dfee2e428558")
	g = []byte("56235346c5828425726d2025cf8fb9a884c20510f356d918ee3b35c6de2d128d")
	hk = []byte("1b16b424f548fb2b746c2c6c3d7560a9244f0e139e9353f5c0de5dac942aff6a")
	tk = []byte("465023c585b2d34fb7fa2358839ba9813de8f07c219e2ec8b701712d4553dab")
	msg11 = bytes.Join(msg1, nil)
	msg22 = bytes.Join(msg2, nil)
	generateCollision(&hk, &tk, &p, &q, &g, &msg11, &msg22, &r1, &s1, &r2, &s2)

	return r2, s2
}

//修改读写数据分析
func GetRwsetFromBlock(block *common.Block) (string, error) {
	var args string
	env, err := protoutil.GetEnvelopeFromBlock(block.Data.Data[0])
	if err != nil {
		fmt.Printf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
		logger.Errorf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
		return args, err
	}

	// block.Data.Data.Payload.\\Data.Actions.Payload.Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	payl, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
		logger.Errorf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
		return args, err
	}
	//解析成transaction   block.Data.Data.Payload.Data
	tx, err := protoutil.UnmarshalTransaction(payl.Data)
	if err != nil {
		return args, err
	}

	// block.Data.Data.Payload.Data.Actions.Payload.\\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	chaincodeActionPayload, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
		logger.Errorf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
		return args, err
	}
	// 此处开始修改
	propRespPayload, err := protoutil.UnmarshalProposalResponsePayload(chaincodeActionPayload.Action.ProposalResponsePayload)
	if err != nil {
		return args, errors.WithMessage(err, "error unmarshal proposal response payload for block event")
	}
	// block.Data.Data.Payload.Data.Actions.Payload.action.proposal_response_payload.extension
	caPayload, err := protoutil.UnmarshalChaincodeAction(propRespPayload.Extension)
	if err != nil {
		return args, errors.WithMessage(err, "error unmarshal chaincode action for block event")
	}
	// vendor\github.com\hyperledger\fabric-protos-go\ledger\rwset
	txReadWriteSet := &rwset.TxReadWriteSet{}
	err = proto.Unmarshal(caPayload.Results, txReadWriteSet)
	if err != nil {
		return args, errors.WithMessage(err, "error unmarshal chaincode action for block event")
	}

	kvrwSet := &kvrwset.KVRWSet{}
	err = proto.Unmarshal(txReadWriteSet.NsRwset[1].Rwset, kvrwSet)
	if err != nil {
		return args, errors.WithMessage(err, "error unmarshal chaincode action for block event")
	}

	if len(kvrwSet.Writes) != 0 {
		return string(kvrwSet.Writes[0].Value), nil
	}
	return args, nil

}

func getBlock(offset string) *common.Block {
	offsetint, _ := strconv.Atoi(offset)
	rootDir := "/var/hyperledger/production/ledgersData/chains/chains/mychannel"
	fileNum := 0
	BlockBytes, _, _, err := ScanForLastCompleteBlock(rootDir, fileNum, int64(offsetint))

	path := "/var/error4.txt"
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		fmt.Printf("can not open the file %s", path)
	}
	f.Write(BlockBytes)

	newblock, err := DeserializeBlock(BlockBytes)
	if err != nil {
		panic(errors.WithMessage(err, "error Deserialize the new block"))
	}
	return newblock
}

func pathExists2(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
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

// 根据证书key获得，该证书所在的区块号
func GetBlockNumberFromFile(certkey string) (string, error) {
	path := "/var/hyperledger/production/certinfo_000000"
	// 修改 从文件中返回一个对应的区块编号
	FileHandle, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer FileHandle.Close()
	lineReader := bufio.NewReader(FileHandle)
	for {
		line, _, err := lineReader.ReadLine()
		if err == io.EOF {
			return "", err
		}
		certinfo := strings.Split(string(line), "++")
		if certinfo[0] == certkey {
			return certinfo[1], nil
		}
	}
	return "", nil
}

func GetBlockOffsetFromFile(blocknum string) (string, error) {
	path := "/var/blockinfo_000000"
	// 修改 从文件中返回一个对应的区块编号
	FileHandle, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer FileHandle.Close()
	lineReader := bufio.NewReader(FileHandle)
	for {
		line, _, err := lineReader.ReadLine()
		certinfo := strings.Split(string(line), "++")
		if certinfo[0] == blocknum {
			return certinfo[1], nil
		}
		if err == io.EOF {
			break
		}
	}
	return "", nil
}

// 修改 开始区块解析
func ValidateAnalysis(block *common.Block) ([]string, error) {
	var args []string
	if len(block.Data.Data) != 0 {
		env, err := protoutil.GetEnvelopeFromBlock(block.Data.Data[0])
		if err != nil {
			// fmt.Printf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
			// logger.Errorf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
			return args, err
		}

		// block.Data.Data.Payload.\\Data.Actions.Payload.Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
		payl, err := protoutil.UnmarshalPayload(env.Payload)
		if err != nil {
			fmt.Printf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
			logger.Errorf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
			return args, err
		}
		//解析成transaction   block.Data.Data.Payload.Data
		tx, err := protoutil.UnmarshalTransaction(payl.Data)
		if err != nil {
			return args, err
		}

		// block.Data.Data.Payload.Data.Actions.Payload.\\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
		cap, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload)
		if err != nil {
			// fmt.Printf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
			// logger.Errorf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
			return args, err
		}

		// 进一步解析成proposalPayload
		// block.Data.Data.Payload.Data.Actions.Payload.chaincode_proposal_payload  \\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
		proposalPayload, err := protoutil.UnmarshalChaincodeProposalPayload(cap.ChaincodeProposalPayload)
		if err != nil {
			return args, err
		}
		//得到交易调用的链码信息
		// block.Data.Data.Payload.Data.Actions.Payload.chaincode_proposal_payload.input
		chaincodeInvocationSpec, err := protoutil.UnmarshalChaincodeInvocationSpec(proposalPayload.Input)
		if err != nil {
			return args, err
		}

		//得到调用的链码的ID，版本和PATH（这里PATH省略了）
		//result.ChaincodeID = chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId.Name
		//result.ChaincodeVersion = chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId.Version

		//得到输入参数
		chaincodeSpec := chaincodeInvocationSpec.ChaincodeSpec
		if chaincodeSpec != nil {
			if chaincodeSpec.Input != nil {
				for _, v := range chaincodeSpec.Input.Args {
					args = append(args, string(v))
				}
				return args, nil
			}
		}
	}
	return args, nil
}

// 修改 得到区块中相应的参数
func GetArgsFromBlock(block *common.Block) ([]string, error) {
	var args []string
	env, err := protoutil.GetEnvelopeFromBlock(block.Data.Data[0])
	if err != nil {
		fmt.Printf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
		logger.Errorf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
		return args, err
	}

	// block.Data.Data.Payload.\\Data.Actions.Payload.Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	payl, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
		logger.Errorf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
		return args, err
	}
	//解析成transaction   block.Data.Data.Payload.Data
	tx, err := protoutil.UnmarshalTransaction(payl.Data)
	if err != nil {
		return args, err
	}

	// block.Data.Data.Payload.Data.Actions.Payload.\\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	cap, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
		logger.Errorf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
		return args, err
	}

	// 进一步解析成proposalPayload
	// block.Data.Data.Payload.Data.Actions.Payload.chaincode_proposal_payload  \\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	proposalPayload, err := protoutil.UnmarshalChaincodeProposalPayload(cap.ChaincodeProposalPayload)
	if err != nil {
		return args, err
	}
	//得到交易调用的链码信息
	// block.Data.Data.Payload.Data.Actions.Payload.chaincode_proposal_payload.input
	chaincodeInvocationSpec, err := protoutil.UnmarshalChaincodeInvocationSpec(proposalPayload.Input)
	if err != nil {
		return args, err
	}
	chaincodeSpec := chaincodeInvocationSpec.ChaincodeSpec

	if chaincodeSpec != nil {
		if chaincodeSpec.Input != nil {
			for _, v := range chaincodeSpec.Input.Args {
				args = append(args, string(v))
			}
			return args, nil
		}
		return args, nil
	}
	return args, nil
}

// 修改 返回一个修改过的交易
func PutArgsInBlock(args []string, block *common.Block) (*common.Block, error) {
	env, err := protoutil.GetEnvelopeFromBlock(block.Data.Data[0])
	if err != nil {
		fmt.Printf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
		logger.Errorf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
		return nil, err
	}

	// block.Data.Data.Payload.\\Data.Actions.Payload.Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	payl, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
		logger.Errorf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
		return nil, err
	}
	//解析成transaction   block.Data.Data.Payload.Data
	tx, err := protoutil.UnmarshalTransaction(payl.Data)
	if err != nil {
		return nil, err
	}

	// block.Data.Data.Payload.Data.Actions.Payload.\\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	cap, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
		logger.Errorf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
		return nil, err
	}

	// 进一步解析成proposalPayload
	// block.Data.Data.Payload.Data.Actions.Payload.chaincode_proposal_payload  \\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	proposalPayload, err := protoutil.UnmarshalChaincodeProposalPayload(cap.ChaincodeProposalPayload)
	if err != nil {
		return nil, err
	}
	//得到交易调用的链码信息
	// block.Data.Data.Payload.Data.Actions.Payload.chaincode_proposal_payload.input
	chaincodeInvocationSpec, err := protoutil.UnmarshalChaincodeInvocationSpec(proposalPayload.Input)
	if err != nil {
		return nil, err
	}
	var argsbyte [][]byte
	for _, value := range args {
		argsbyte = append(argsbyte, []byte(value))
	}
	chaincodeInvocationSpec.ChaincodeSpec.Input.Args = argsbyte
	proposalPayload.Input, _ = proto.Marshal(chaincodeInvocationSpec)
	cap.ChaincodeProposalPayload, _ = proto.Marshal(proposalPayload)
	tx.Actions[0].Payload, _ = proto.Marshal(cap)
	payl.Data, _ = proto.Marshal(tx)
	env.Payload, _ = proto.Marshal(payl)
	block.Data.Data[0], _ = proto.Marshal(env)
	return block, nil
}

func PutRwsetInBlock(oldvalues string, block *common.Block) (*common.Block, error) {
	env, err := protoutil.GetEnvelopeFromBlock(block.Data.Data[0])
	if err != nil {
		fmt.Printf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
		logger.Errorf("protoutil.GetEnvelopeFromBlock 此方法失败 in BlockAnaly")
		return nil, err
	}

	// block.Data.Data.Payload.\\Data.Actions.Payload.Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	payl, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
		logger.Errorf("protoutil.UnmarshalPayload failed,此方法失败 in BlockAnaly")
		return nil, err
	}
	//解析成transaction   block.Data.Data.Payload.Data
	tx, err := protoutil.UnmarshalTransaction(payl.Data)
	if err != nil {
		return nil, err
	}

	// block.Data.Data.Payload.Data.Actions.Payload.\\Action.Proposal_response_payload.Extension.Results.Ns_rwset.Rwset.Writes.Value.OptionalString
	chaincodeActionPayload, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload)
	if err != nil {
		fmt.Printf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
		logger.Errorf("protoutil.UnmarshalChaincodeActionPayload failed,此方法失败 in BlockAnaly")
		return nil, err
	}
	// 此处开始修改
	propRespPayload, err := protoutil.UnmarshalProposalResponsePayload(chaincodeActionPayload.Action.ProposalResponsePayload)
	if err != nil {
		return nil, errors.WithMessage(err, "error unmarshal proposal response payload for block event")
	}
	// block.Data.Data.Payload.Data.Actions.Payload.action.proposal_response_payload.extension
	caPayload, err := protoutil.UnmarshalChaincodeAction(propRespPayload.Extension)
	if err != nil {
		return nil, errors.WithMessage(err, "error unmarshal chaincode action for block event")
	}
	// vendor\github.com\hyperledger\fabric-protos-go\ledger\rwset
	txReadWriteSet := &rwset.TxReadWriteSet{}
	err = proto.Unmarshal(caPayload.Results, txReadWriteSet)
	if err != nil {
		return nil, errors.WithMessage(err, "error unmarshal chaincode action for block event")
	}

	kvrwSet := &kvrwset.KVRWSet{}
	err = proto.Unmarshal(txReadWriteSet.NsRwset[1].Rwset, kvrwSet)
	if err != nil {
		return nil, errors.WithMessage(err, "error unmarshal chaincode action for block event")
	}

	kvrwSet.Writes[0].Value = []byte(oldvalues)
	txReadWriteSet.NsRwset[1].Rwset, _ = proto.Marshal(kvrwSet)
	caPayload.Results, _ = proto.Marshal(txReadWriteSet)
	propRespPayload.Extension, _ = proto.Marshal(caPayload)
	chaincodeActionPayload.Action.ProposalResponsePayload, _ = proto.Marshal(propRespPayload)
	tx.Actions[0].Payload, _ = proto.Marshal(chaincodeActionPayload)
	payl.Data, _ = proto.Marshal(tx)
	env.Payload, _ = proto.Marshal(payl)
	block.Data.Data[0], _ = proto.Marshal(env)
	return block, nil
}

func (mgr *blockfileMgr) syncIndex() error {
	nextIndexableBlock := uint64(0)
	lastBlockIndexed, err := mgr.index.getLastBlockIndexed()
	if err != nil {
		if err != errIndexSavePointKeyNotPresent {
			return err
		}
	} else {
		nextIndexableBlock = lastBlockIndexed + 1
	}

	if nextIndexableBlock == 0 && mgr.bootstrappedFromSnapshot() {
		// This condition can happen only if there was a peer crash or failure during
		// bootstrapping the ledger from a snapshot or the index is dropped/corrupted afterward
		return errors.Errorf(
			"cannot sync index with block files. blockstore is bootstrapped from a snapshot and first available block=[%d]",
			mgr.firstPossibleBlockNumberInBlockFiles(),
		)
	}

	if mgr.blockfilesInfo.noBlockFiles {
		logger.Debug("No block files present. This happens when there has not been any blocks added to the ledger yet")
		return nil
	}

	if nextIndexableBlock == mgr.blockfilesInfo.lastPersistedBlock+1 {
		logger.Debug("Both the block files and indices are in sync.")
		return nil
	}

	startFileNum := 0
	startOffset := 0
	skipFirstBlock := false
	endFileNum := mgr.blockfilesInfo.latestFileNumber

	firstAvailableBlkNum, err := retrieveFirstBlockNumFromFile(mgr.rootDir, 0)
	if err != nil {
		return err
	}

	if nextIndexableBlock > firstAvailableBlkNum {
		logger.Debugf("Last block indexed [%d], Last block present in block files [%d]", lastBlockIndexed, mgr.blockfilesInfo.lastPersistedBlock)
		var flp *fileLocPointer
		if flp, err = mgr.index.getBlockLocByBlockNum(lastBlockIndexed); err != nil {
			return err
		}
		startFileNum = flp.fileSuffixNum
		startOffset = flp.locPointer.offset
		skipFirstBlock = true
	}

	logger.Infof("Start building index from block [%d] to last block [%d]", nextIndexableBlock, mgr.blockfilesInfo.lastPersistedBlock)

	//open a blockstream to the file location that was stored in the index
	var stream *blockStream
	if stream, err = newBlockStream(mgr.rootDir, startFileNum, int64(startOffset), endFileNum); err != nil {
		return err
	}
	var blockBytes []byte
	var blockPlacementInfo *blockPlacementInfo

	if skipFirstBlock {
		if blockBytes, _, err = stream.nextBlockBytesAndPlacementInfo(); err != nil {
			return err
		}
		if blockBytes == nil {
			return errors.Errorf("block bytes for block num = [%d] should not be nil here. The indexes for the block are already present",
				lastBlockIndexed)
		}
	}

	//Should be at the last block already, but go ahead and loop looking for next blockBytes.
	//If there is another block, add it to the index.
	//This will ensure block indexes are correct, for example if peer had crashed before indexes got updated.
	blockIdxInfo := &blockIdxInfo{}
	for {
		if blockBytes, blockPlacementInfo, err = stream.nextBlockBytesAndPlacementInfo(); err != nil {
			return err
		}
		if blockBytes == nil {
			break
		}
		info, err := extractSerializedBlockInfo(blockBytes)
		if err != nil {
			return err
		}

		//The blockStartOffset will get applied to the txOffsets prior to indexing within indexBlock(),
		//therefore just shift by the difference between blockBytesOffset and blockStartOffset
		numBytesToShift := int(blockPlacementInfo.blockBytesOffset - blockPlacementInfo.blockStartOffset)
		for _, offset := range info.txOffsets {
			offset.loc.offset += numBytesToShift
		}

		//Update the blockIndexInfo with what was actually stored in file system
		blockIdxInfo.blockHash = protoutil.BlockHeaderHash(info.blockHeader)
		blockIdxInfo.blockNum = info.blockHeader.Number
		blockIdxInfo.flp = &fileLocPointer{fileSuffixNum: blockPlacementInfo.fileNum,
			locPointer: locPointer{offset: int(blockPlacementInfo.blockStartOffset)}}
		blockIdxInfo.txOffsets = info.txOffsets
		blockIdxInfo.metadata = info.metadata

		logger.Debugf("syncIndex() indexing block [%d]", blockIdxInfo.blockNum)
		if err = mgr.index.indexBlock(blockIdxInfo); err != nil {
			return err
		}
		if blockIdxInfo.blockNum%10000 == 0 {
			logger.Infof("Indexed block number [%d]", blockIdxInfo.blockNum)
		}
	}
	logger.Infof("Finished building index. Last block indexed [%d]", blockIdxInfo.blockNum)
	return nil
}

func (mgr *blockfileMgr) getBlockchainInfo() *common.BlockchainInfo {
	return mgr.bcInfo.Load().(*common.BlockchainInfo)
}

func (mgr *blockfileMgr) updateBlockfilesInfo(blkfilesInfo *blockfilesInfo) {
	mgr.blkfilesInfoCond.L.Lock()
	defer mgr.blkfilesInfoCond.L.Unlock()
	mgr.blockfilesInfo = blkfilesInfo
	logger.Debugf("Broadcasting about update blockfilesInfo: %s", blkfilesInfo)
	mgr.blkfilesInfoCond.Broadcast()
}

func (mgr *blockfileMgr) updateBlockchainInfo(latestBlockHash []byte, latestBlock *common.Block) {
	currentBCInfo := mgr.getBlockchainInfo()
	newBCInfo := &common.BlockchainInfo{
		Height:                    currentBCInfo.Height + 1,
		CurrentBlockHash:          latestBlockHash,
		PreviousBlockHash:         latestBlock.Header.PreviousHash,
		BootstrappingSnapshotInfo: currentBCInfo.BootstrappingSnapshotInfo,
	}

	mgr.bcInfo.Store(newBCInfo)
}

func (mgr *blockfileMgr) retrieveBlockByHash(blockHash []byte) (*common.Block, error) {
	logger.Debugf("retrieveBlockByHash() - blockHash = [%#v]", blockHash)
	loc, err := mgr.index.getBlockLocByHash(blockHash)
	if err != nil {
		return nil, err
	}
	return mgr.fetchBlock(loc)
}

func (mgr *blockfileMgr) retrieveBlockByNumber(blockNum uint64) (*common.Block, error) {
	logger.Debugf("retrieveBlockByNumber() - blockNum = [%d]", blockNum)

	// interpret math.MaxUint64 as a request for last block
	if blockNum == math.MaxUint64 {
		blockNum = mgr.getBlockchainInfo().Height - 1
	}
	if blockNum < mgr.firstPossibleBlockNumberInBlockFiles() {
		return nil, errors.Errorf(
			"cannot serve block [%d]. The ledger is bootstrapped from a snapshot. First available block = [%d]",
			blockNum, mgr.firstPossibleBlockNumberInBlockFiles(),
		)
	}
	loc, err := mgr.index.getBlockLocByBlockNum(blockNum)
	if err != nil {
		return nil, err
	}
	return mgr.fetchBlock(loc)
}

func (mgr *blockfileMgr) retrieveBlockByTxID(txID string) (*common.Block, error) {
	logger.Debugf("retrieveBlockByTxID() - txID = [%s]", txID)
	loc, err := mgr.index.getBlockLocByTxID(txID)
	if err == errNilValue {
		return nil, errors.Errorf(
			"details for the TXID [%s] not available. Ledger bootstrapped from a snapshot. First available block = [%d]",
			txID, mgr.firstPossibleBlockNumberInBlockFiles())
	}
	if err != nil {
		return nil, err
	}
	return mgr.fetchBlock(loc)
}

func (mgr *blockfileMgr) retrieveTxValidationCodeByTxID(txID string) (peer.TxValidationCode, error) {
	logger.Debugf("retrieveTxValidationCodeByTxID() - txID = [%s]", txID)
	validationCode, err := mgr.index.getTxValidationCodeByTxID(txID)
	if err == errNilValue {
		return peer.TxValidationCode(-1), errors.Errorf(
			"details for the TXID [%s] not available. Ledger bootstrapped from a snapshot. First available block = [%d]",
			txID, mgr.firstPossibleBlockNumberInBlockFiles())
	}
	return validationCode, err
}

func (mgr *blockfileMgr) retrieveBlockHeaderByNumber(blockNum uint64) (*common.BlockHeader, error) {
	logger.Debugf("retrieveBlockHeaderByNumber() - blockNum = [%d]", blockNum)
	if blockNum < mgr.firstPossibleBlockNumberInBlockFiles() {
		return nil, errors.Errorf(
			"cannot serve block [%d]. The ledger is bootstrapped from a snapshot. First available block = [%d]",
			blockNum, mgr.firstPossibleBlockNumberInBlockFiles(),
		)
	}
	loc, err := mgr.index.getBlockLocByBlockNum(blockNum)
	if err != nil {
		return nil, err
	}
	blockBytes, err := mgr.fetchBlockBytes(loc)
	if err != nil {
		return nil, err
	}
	info, err := extractSerializedBlockInfo(blockBytes)
	if err != nil {
		return nil, err
	}
	return info.blockHeader, nil
}

func (mgr *blockfileMgr) retrieveBlocks(startNum uint64) (*blocksItr, error) {
	if startNum < mgr.firstPossibleBlockNumberInBlockFiles() {
		return nil, errors.Errorf(
			"cannot serve block [%d]. The ledger is bootstrapped from a snapshot. First available block = [%d]",
			startNum, mgr.firstPossibleBlockNumberInBlockFiles(),
		)
	}
	return newBlockItr(mgr, startNum), nil
}

func (mgr *blockfileMgr) txIDExists(txID string) (bool, error) {
	return mgr.index.txIDExists(txID)
}

func (mgr *blockfileMgr) retrieveTransactionByID(txID string) (*common.Envelope, error) {
	logger.Debugf("retrieveTransactionByID() - txId = [%s]", txID)
	loc, err := mgr.index.getTxLoc(txID)
	if err == errNilValue {
		return nil, errors.Errorf(
			"details for the TXID [%s] not available. Ledger bootstrapped from a snapshot. First available block = [%d]",
			txID, mgr.firstPossibleBlockNumberInBlockFiles())
	}
	if err != nil {
		return nil, err
	}
	return mgr.fetchTransactionEnvelope(loc)
}

func (mgr *blockfileMgr) retrieveTransactionByBlockNumTranNum(blockNum uint64, tranNum uint64) (*common.Envelope, error) {
	logger.Debugf("retrieveTransactionByBlockNumTranNum() - blockNum = [%d], tranNum = [%d]", blockNum, tranNum)
	if blockNum < mgr.firstPossibleBlockNumberInBlockFiles() {
		return nil, errors.Errorf(
			"cannot serve block [%d]. The ledger is bootstrapped from a snapshot. First available block = [%d]",
			blockNum, mgr.firstPossibleBlockNumberInBlockFiles(),
		)
	}
	loc, err := mgr.index.getTXLocByBlockNumTranNum(blockNum, tranNum)
	if err != nil {
		return nil, err
	}
	return mgr.fetchTransactionEnvelope(loc)
}

func (mgr *blockfileMgr) fetchBlock(lp *fileLocPointer) (*common.Block, error) {
	blockBytes, err := mgr.fetchBlockBytes(lp)
	if err != nil {
		return nil, err
	}
	block, err := DeserializeBlock(blockBytes)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (mgr *blockfileMgr) fetchTransactionEnvelope(lp *fileLocPointer) (*common.Envelope, error) {
	logger.Debugf("Entering fetchTransactionEnvelope() %v\n", lp)
	var err error
	var txEnvelopeBytes []byte
	if txEnvelopeBytes, err = mgr.fetchRawBytes(lp); err != nil {
		return nil, err
	}
	_, n := proto.DecodeVarint(txEnvelopeBytes)
	return protoutil.GetEnvelopeFromBlock(txEnvelopeBytes[n:])
}

func (mgr *blockfileMgr) fetchBlockBytes(lp *fileLocPointer) ([]byte, error) {
	stream, err := newBlockfileStream(mgr.rootDir, lp.fileSuffixNum, int64(lp.offset))
	if err != nil {
		return nil, err
	}
	defer stream.close()
	b, err := stream.nextBlockBytes()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (mgr *blockfileMgr) fetchRawBytes(lp *fileLocPointer) ([]byte, error) {
	filePath := deriveBlockfilePath(mgr.rootDir, lp.fileSuffixNum)
	reader, err := newBlockfileReader(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.close()
	b, err := reader.read(lp.offset, lp.bytesLength)
	if err != nil {
		return nil, err
	}
	return b, nil
}

//Get the current blockfilesInfo information that is stored in the database
func (mgr *blockfileMgr) loadBlkfilesInfo() (*blockfilesInfo, error) {
	var b []byte
	var err error
	if b, err = mgr.db.Get(blkMgrInfoKey); b == nil || err != nil {
		return nil, err
	}
	i := &blockfilesInfo{}
	if err = i.unmarshal(b); err != nil {
		return nil, err
	}
	logger.Debugf("loaded blockfilesInfo:%s", i)
	return i, nil
}

func (mgr *blockfileMgr) saveBlkfilesInfo(i *blockfilesInfo, sync bool) error {
	b, err := i.marshal()
	if err != nil {
		return err
	}
	if err = mgr.db.Put(blkMgrInfoKey, b, sync); err != nil {
		return err
	}
	return nil
}

func (mgr *blockfileMgr) firstPossibleBlockNumberInBlockFiles() uint64 {
	if mgr.bootstrappingSnapshotInfo == nil {
		return 0
	}
	return mgr.bootstrappingSnapshotInfo.LastBlockNum + 1
}

func (mgr *blockfileMgr) bootstrappedFromSnapshot() bool {
	return mgr.firstPossibleBlockNumberInBlockFiles() > 0
}

func ScanForLastCompleteBlock(rootDir string, fileNum int, startingOffset int64) ([]byte, int64, int, error) {
	numBlocks := 0
	var lastBlockBytes []byte
	blockStream, errOpen := newBlockfileStream(rootDir, fileNum, startingOffset)
	if errOpen != nil {
		return nil, 0, 0, errOpen
	}
	defer blockStream.close()
	var errRead error
	var blockBytes []byte
	for {
		blockBytes, errRead = blockStream.nextBlockBytes()
		if blockBytes == nil || errRead != nil {
			break
		}
		lastBlockBytes = blockBytes
		numBlocks++
		break
	}
	if errRead == ErrUnexpectedEndOfBlockfile {
		logger.Debugf(`Error:%s
		The error may happen if a crash has happened during block appending.
		Resetting error to nil and returning current offset as a last complete block's end offset`, errRead)
		errRead = nil
	}
	logger.Debugf("scanForLastCompleteBlock(): last complete block ends at offset=[%d]", blockStream.currentOffset)
	return lastBlockBytes, blockStream.currentOffset, numBlocks, errRead
}

// scanForLastCompleteBlock scan a given block file and detects the last offset in the file
// after which there may lie a block partially written (towards the end of the file in a crash scenario).
func scanForLastCompleteBlock(rootDir string, fileNum int, startingOffset int64) ([]byte, int64, int, error) {
	//scan the passed file number suffix starting from the passed offset to find the last completed block
	numBlocks := 0
	var lastBlockBytes []byte
	blockStream, errOpen := newBlockfileStream(rootDir, fileNum, startingOffset)
	if errOpen != nil {
		return nil, 0, 0, errOpen
	}
	defer blockStream.close()
	var errRead error
	var blockBytes []byte
	for {
		blockBytes, errRead = blockStream.nextBlockBytes()
		if blockBytes == nil || errRead != nil {
			break
		}
		lastBlockBytes = blockBytes
		numBlocks++
	}
	if errRead == ErrUnexpectedEndOfBlockfile {
		logger.Debugf(`Error:%s
		The error may happen if a crash has happened during block appending.
		Resetting error to nil and returning current offset as a last complete block's end offset`, errRead)
		errRead = nil
	}
	logger.Debugf("scanForLastCompleteBlock(): last complete block ends at offset=[%d]", blockStream.currentOffset)
	return lastBlockBytes, blockStream.currentOffset, numBlocks, errRead
}

// blockfilesInfo maintains the summary about the blockfiles
type blockfilesInfo struct {
	latestFileNumber   int
	latestFileSize     int
	noBlockFiles       bool
	lastPersistedBlock uint64
}

func (i *blockfilesInfo) marshal() ([]byte, error) {
	buffer := proto.NewBuffer([]byte{})
	var err error
	if err = buffer.EncodeVarint(uint64(i.latestFileNumber)); err != nil {
		return nil, errors.Wrapf(err, "error encoding the latestFileNumber [%d]", i.latestFileNumber)
	}
	if err = buffer.EncodeVarint(uint64(i.latestFileSize)); err != nil {
		return nil, errors.Wrapf(err, "error encoding the latestFileSize [%d]", i.latestFileSize)
	}
	if err = buffer.EncodeVarint(i.lastPersistedBlock); err != nil {
		return nil, errors.Wrapf(err, "error encoding the lastPersistedBlock [%d]", i.lastPersistedBlock)
	}
	var noBlockFilesMarker uint64
	if i.noBlockFiles {
		noBlockFilesMarker = 1
	}
	if err = buffer.EncodeVarint(noBlockFilesMarker); err != nil {
		return nil, errors.Wrapf(err, "error encoding noBlockFiles [%d]", noBlockFilesMarker)
	}
	return buffer.Bytes(), nil
}

func (i *blockfilesInfo) unmarshal(b []byte) error {
	buffer := proto.NewBuffer(b)
	var val uint64
	var noBlockFilesMarker uint64
	var err error

	if val, err = buffer.DecodeVarint(); err != nil {
		return err
	}
	i.latestFileNumber = int(val)

	if val, err = buffer.DecodeVarint(); err != nil {
		return err
	}
	i.latestFileSize = int(val)

	if val, err = buffer.DecodeVarint(); err != nil {
		return err
	}
	i.lastPersistedBlock = val
	if noBlockFilesMarker, err = buffer.DecodeVarint(); err != nil {
		return err
	}
	i.noBlockFiles = noBlockFilesMarker == 1
	return nil
}

func (i *blockfilesInfo) String() string {
	return fmt.Sprintf("latestFileNumber=[%d], latestFileSize=[%d], noBlockFiles=[%t], lastPersistedBlock=[%d]",
		i.latestFileNumber, i.latestFileSize, i.noBlockFiles, i.lastPersistedBlock)
}
