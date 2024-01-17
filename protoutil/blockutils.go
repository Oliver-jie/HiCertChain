/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoutil
// 创建一个新的区块，区块头和数据sha256.Sum256加密。搜索区块的元数据
import (
	"bytes"
	"crypto/sha256"
	"encoding/asn1"
	"math/big"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"

	"os"
	"fmt"
	"bufio"
	"strconv"
)

// NewBlock constructs a block with no data and no metadata.
func NewBlock(seqNum uint64, previousHash []byte) *cb.Block {
	block := &cb.Block{}
	block.Header = &cb.BlockHeader{}
	block.Header.Number = seqNum
	block.Header.PreviousHash = previousHash
	block.Header.DataHash = []byte{}
	block.Data = &cb.BlockData{}

	var metadataContents [][]byte
	for i := 0; i < len(cb.BlockMetadataIndex_name); i++ {
		metadataContents = append(metadataContents, []byte{})
	}
	block.Metadata = &cb.BlockMetadata{Metadata: metadataContents}

	return block
}

type asn1Header struct {
	Number       *big.Int
	PreviousHash []byte
	DataHash     []byte
}

func BlockHeaderBytes(b *cb.BlockHeader) []byte {
	asn1Header := asn1Header{
		PreviousHash: b.PreviousHash,
		DataHash:     b.DataHash,
		Number:       new(big.Int).SetUint64(b.Number),
	}
	result, err := asn1.Marshal(asn1Header)
	if err != nil {
		// Errors should only arise for types which cannot be encoded, since the
		// BlockHeader type is known a-priori to contain only encodable types, an
		// error here is fatal and should not be propagated
		panic(err)
	}
	return result
}

func BlockHeaderHash(b *cb.BlockHeader) []byte {
	sum := sha256.Sum256(BlockHeaderBytes(b))
	return sum[:]
}



// 修改 辅助函数
func PathExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
return false, err
}

// 修改 区块链的data哈希值

func BlockDataHash(b *cb.BlockData,num uint64) []byte {
	// sum := sha256.Sum256(bytes.Join(b.Data, nil))
	// return sum[:]

	// 如果文件不存在，创建新的文件，并且写进去随机数
	// 存在的话读取数据
	blocknum:= strconv.FormatUint(num,10)
	var p, q, g, hk, hash1,  r1, s1,msg []byte
	var message []string

	path := "/opt/gopath/blockfile"+blocknum
	pathisEmpty,_:=PathExists(path)
	if pathisEmpty {
		FileHandle, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer FileHandle.Close()
		lineReader := bufio.NewReader(FileHandle)
		for {
		line, _, err := lineReader.ReadLine()
		if err != nil {
			break
		}
		message = append(message,string(line))
		}
		r1=[]byte(message[0])
		s1=[]byte(message[1])
	}else{
		f,err:= os.OpenFile(path,os.O_CREATE|os.O_APPEND|os.O_RDWR,0660)
		if err !=nil{
			fmt.Printf("can not open the file %s",path)
		}
		f.WriteString("ff6a035ee82ebb3a88a59e8de2a3caa4722156ac54041a3e542c8adba1a6be8\r\n")
		f.WriteString("7c0ad2936ff5254a1b4ac1a016671bf18a7ad8bd1e7c8ce3f7241664db500704\r\n")
		r1 = []byte("ff6a035ee82ebb3a88a59e8de2a3caa4722156ac54041a3e542c8adba1a6be8")
		s1 = []byte("7c0ad2936ff5254a1b4ac1a016671bf18a7ad8bd1e7c8ce3f7241664db500704")
	}

	p = []byte("ff1547760a78c9a40bae518b3c631bd719803d17c5d39c456f8bbfdc5c850ab1")
	q = []byte("7f8aa3bb053c64d205d728c59e318deb8cc01e8be2e9ce22b7c5dfee2e428558")
    g = []byte("56235346c5828425726d2025cf8fb9a884c20510f356d918ee3b35c6de2d128d")
	hk = []byte("1b16b424f548fb2b746c2c6c3d7560a9244f0e139e9353f5c0de5dac942aff6a")


	msg = bytes.Join(b.Data, nil)

	chameleonHash(&hk, &p, &q, &g, &msg, &r1, &s1, &hash1)
	return hash1
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



// GetChannelIDFromBlockBytes returns channel ID given byte array which represents
// the block
func GetChannelIDFromBlockBytes(bytes []byte) (string, error) {
	block, err := UnmarshalBlock(bytes)
	if err != nil {
		return "", err
	}

	return GetChannelIDFromBlock(block)
}

// GetChannelIDFromBlock returns channel ID in the block
func GetChannelIDFromBlock(block *cb.Block) (string, error) {
	if block == nil || block.Data == nil || block.Data.Data == nil || len(block.Data.Data) == 0 {
		return "", errors.New("failed to retrieve channel id - block is empty")
	}
	var err error
	envelope, err := GetEnvelopeFromBlock(block.Data.Data[0])
	if err != nil {
		return "", err
	}
	payload, err := UnmarshalPayload(envelope.Payload)
	if err != nil {
		return "", err
	}

	if payload.Header == nil {
		return "", errors.New("failed to retrieve channel id - payload header is empty")
	}
	chdr, err := UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return "", err
	}

	return chdr.ChannelId, nil
}

// GetMetadataFromBlock retrieves metadata at the specified index.
func GetMetadataFromBlock(block *cb.Block, index cb.BlockMetadataIndex) (*cb.Metadata, error) {
	if block.Metadata == nil {
		return nil, errors.New("no metadata in block")
	}

	if len(block.Metadata.Metadata) <= int(index) {
		return nil, errors.Errorf("no metadata at index [%s]", index)
	}

	md := &cb.Metadata{}
	err := proto.Unmarshal(block.Metadata.Metadata[index], md)
	if err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling metadata at index [%s]", index)
	}
	return md, nil
}

// GetMetadataFromBlockOrPanic retrieves metadata at the specified index, or
// panics on error
func GetMetadataFromBlockOrPanic(block *cb.Block, index cb.BlockMetadataIndex) *cb.Metadata {
	md, err := GetMetadataFromBlock(block, index)
	if err != nil {
		panic(err)
	}
	return md
}

// GetConsenterMetadataFromBlock attempts to retrieve consenter metadata from the value
// stored in block metadata at index SIGNATURES (first field). If no consenter metadata
// is found there, it falls back to index ORDERER (third field).
func GetConsenterMetadataFromBlock(block *cb.Block) (*cb.Metadata, error) {
	m, err := GetMetadataFromBlock(block, cb.BlockMetadataIndex_SIGNATURES)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to retrieve metadata")
	}

	// TODO FAB-15864 Remove this fallback when we can stop supporting upgrade from pre-1.4.1 orderer
	if len(m.Value) == 0 {
		return GetMetadataFromBlock(block, cb.BlockMetadataIndex_ORDERER)
	}

	obm := &cb.OrdererBlockMetadata{}
	err = proto.Unmarshal(m.Value, obm)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal orderer block metadata")
	}

	res := &cb.Metadata{}
	err = proto.Unmarshal(obm.ConsenterMetadata, res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal consenter metadata")
	}

	return res, nil
}

// GetLastConfigIndexFromBlock retrieves the index of the last config block as
// encoded in the block metadata
func GetLastConfigIndexFromBlock(block *cb.Block) (uint64, error) {
	m, err := GetMetadataFromBlock(block, cb.BlockMetadataIndex_SIGNATURES)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to retrieve metadata")
	}
	// TODO FAB-15864 Remove this fallback when we can stop supporting upgrade from pre-1.4.1 orderer
	if len(m.Value) == 0 {
		m, err := GetMetadataFromBlock(block, cb.BlockMetadataIndex_LAST_CONFIG)
		if err != nil {
			return 0, errors.WithMessage(err, "failed to retrieve metadata")
		}
		lc := &cb.LastConfig{}
		err = proto.Unmarshal(m.Value, lc)
		if err != nil {
			return 0, errors.Wrap(err, "error unmarshaling LastConfig")
		}
		return lc.Index, nil
	}

	obm := &cb.OrdererBlockMetadata{}
	err = proto.Unmarshal(m.Value, obm)
	if err != nil {
		return 0, errors.Wrap(err, "failed to unmarshal orderer block metadata")
	}
	return obm.LastConfig.Index, nil
}

// GetLastConfigIndexFromBlockOrPanic retrieves the index of the last config
// block as encoded in the block metadata, or panics on error
func GetLastConfigIndexFromBlockOrPanic(block *cb.Block) uint64 {
	index, err := GetLastConfigIndexFromBlock(block)
	if err != nil {
		panic(err)
	}
	return index
}

// CopyBlockMetadata copies metadata from one block into another
func CopyBlockMetadata(src *cb.Block, dst *cb.Block) {
	dst.Metadata = src.Metadata
	// Once copied initialize with rest of the
	// required metadata positions.
	InitBlockMetadata(dst)
}

// InitBlockMetadata initializes metadata structure
func InitBlockMetadata(block *cb.Block) {
	if block.Metadata == nil {
		block.Metadata = &cb.BlockMetadata{Metadata: [][]byte{{}, {}, {}, {}, {}}}
	} else if len(block.Metadata.Metadata) < int(cb.BlockMetadataIndex_COMMIT_HASH+1) {
		for i := int(len(block.Metadata.Metadata)); i <= int(cb.BlockMetadataIndex_COMMIT_HASH); i++ {
			block.Metadata.Metadata = append(block.Metadata.Metadata, []byte{})
		}
	}
}

