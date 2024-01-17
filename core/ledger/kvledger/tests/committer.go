/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tests

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/core/ledger"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/stretchr/testify/assert"
)

// committer helps in cutting a block and commits the block (with pvt data) to the ledger
type committer struct {
	lgr    ledger.PeerLedger
	blkgen *blkGenerator
	assert *assert.Assertions
}

func newCommitter(lgr ledger.PeerLedger, t *testing.T) *committer {
	return &committer{lgr, newBlockGenerator(lgr, t), assert.New(t)}
}

// cutBlockAndCommitLegacy cuts the next block from the given 'txAndPvtdata' and commits the block (with pvt data) to the ledger
// This function return a copy of 'ledger.BlockAndPvtData' that was submitted to the ledger to commit.
// A copy is returned instead of the actual one because, ledger makes some changes to the submitted block before commit
// (such as setting the metadata) and the test code would want to have the exact copy of the block that was submitted to the ledger
//cutBlockAndCommitLegacy从给定的“txAndPvtdata”中剪切下一个块，并将该块（带有pvt数据）提交到分类帐
//此函数返回已提交到要提交的分类帐的“ledger.BlockAndPvtData”副本。
//将返回一个副本而不是实际的副本，因为ledger在提交之前对提交的块进行了一些更改
//（例如设置元数据）并且测试代码希望获得提交到分类帐的块的精确副本
func (c *committer) cutBlockAndCommitLegacy(trans []*txAndPvtdata, missingPvtData ledger.TxMissingPvtDataMap) *ledger.BlockAndPvtData {
	blk := c.blkgen.nextBlockAndPvtdata(trans, missingPvtData)
	blkCopy := c.copyOfBlockAndPvtdata(blk)
	c.assert.NoError(
		c.lgr.CommitLegacy(blk, &ledger.CommitOptions{}),
	)
	return blkCopy
}

//切割区块，并且提交错误
func (c *committer) cutBlockAndCommitExpectError(trans []*txAndPvtdata, missingPvtData ledger.TxMissingPvtDataMap) (*ledger.BlockAndPvtData, error) {
	blk := c.blkgen.nextBlockAndPvtdata(trans, missingPvtData)
	blkCopy := c.copyOfBlockAndPvtdata(blk)
	err := c.lgr.CommitLegacy(blk, &ledger.CommitOptions{})
	c.assert.Error(err)
	return blkCopy, err
}

func (c *committer) copyOfBlockAndPvtdata(blk *ledger.BlockAndPvtData) *ledger.BlockAndPvtData {
	blkBytes, err := proto.Marshal(blk.Block)
	c.assert.NoError(err)
	blkCopy := &common.Block{}
	c.assert.NoError(proto.Unmarshal(blkBytes, blkCopy))
	return &ledger.BlockAndPvtData{Block: blkCopy, PvtData: blk.PvtData,
		MissingPvtData: blk.MissingPvtData}
}

/////////////////   block generation code  ///////////////////////////////////////////
// blkGenerator helps creating the next block for the ledger
type blkGenerator struct {
	lastNum  uint64
	lastHash []byte
	assert   *assert.Assertions
}

// newBlockGenerator constructs a 'blkGenerator' and initializes the 'blkGenerator' from the last block available in the ledger so that the next block can be populated with the correct block number and previous block hash
// newBlockGenerator构造一个“blkGenerator”，并从分类帐中可用的最后一个块初始化“blkGenerator”，以便用正确的块号和上一个块哈希填充下一个块
func newBlockGenerator(lgr ledger.PeerLedger, t *testing.T) *blkGenerator {
	assert := assert.New(t)
	info, err := lgr.GetBlockchainInfo()
	assert.NoError(err)
	return &blkGenerator{info.Height - 1, info.CurrentBlockHash, assert}
}

// nextBlockAndPvtdata cuts the next block
// 修改 此处是commiter节点切割区块，要看 
func (g *blkGenerator) nextBlockAndPvtdata(trans []*txAndPvtdata, missingPvtData ledger.TxMissingPvtDataMap) *ledger.BlockAndPvtData {
	block := protoutil.NewBlock(g.lastNum+1, g.lastHash)
	blockPvtdata := make(map[uint64]*ledger.TxPvtData)
	for i, tran := range trans {
		seq := uint64(i)
		envelopeBytes, _ := proto.Marshal(tran.Envelope)
		block.Data.Data = append(block.Data.Data, envelopeBytes)
		if tran.Pvtws != nil {
			blockPvtdata[seq] = &ledger.TxPvtData{SeqInBlock: seq, WriteSet: tran.Pvtws}
		}
	}
	// 此处的设计是为了test
	block.Block.Header.DataHash = protoutil.BlockDataHash(block.Block.Data,block.Header.Number)
	g.lastNum++
	g.lastHash = protoutil.BlockHeaderHash(block.Block.Header)
	setBlockFlagsToValid(block)
	return &ledger.BlockAndPvtData{Block: block, PvtData: blockPvtdata,
		MissingPvtData: missingPvtData}
}

