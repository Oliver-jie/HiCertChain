package main

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"math"
	//"bufio"
	"fmt"
	//"os"
	//"testing"
	"strconv"
)

const bucketHeight = 8

type fingerprint uint16

type target struct {
	bucketIndex uint
	fingerprint fingerprint
}

type bucket struct {
	entries [bucketHeight]fingerprint
	count   uint8
}

type table []bucket

/*
Dlcbf is a struct representing a d-left Counting Bloom Filter
*/
type Dlcbf struct {
	tables     []table
	numTables  uint
	numBuckets uint
}

/*
NewDlcbf returns a newly created Dlcbf
*/
func NewDlcbf(numTables uint, numBuckets uint) (*Dlcbf, error) {

	if numBuckets < numTables {
		return nil, errors.New("numBuckets has to be greater than numTables")
	}
// 第2个参数是切片的长度，第三个是预留的长度
	dlcbf := &Dlcbf{
		numTables:  numTables,
		numBuckets: numBuckets,
		tables:     make([]table, numTables, numTables),
	}

// 每个table放入bucket
	for i := range dlcbf.tables {
		dlcbf.tables[i] = make(table, numBuckets, numBuckets)
	}

	return dlcbf, nil
}

/*
NewDlcbfForCapacity returns a newly created Dlcbf for a given max Capacity
NewDlcbfForCapacity返回为给定的最大容量而新建的Dlcbf。
*/
func NewDlcbfForCapacity(capacity uint) (*Dlcbf, error) {
	t := capacity / (4096 * bucketHeight)
	return NewDlcbf(t, 4096)
}

func (dlcbf *Dlcbf) getTargets(data []byte) []target {
	hasher := fnv.New64a()
	hasher.Write(data)
	fp := hasher.Sum(nil)
	hsum := hasher.Sum64()

	h1 := uint32(hsum & 0xffffffff)
	h2 := uint32((hsum >> 32) & 0xffffffff)

	indices := make([]uint, dlcbf.numTables, dlcbf.numTables)
	for i := uint(0); i < dlcbf.numTables; i++ {
		saltedHash := uint((h1 + uint32(i)*h2))
		indices[i] = (saltedHash % dlcbf.numBuckets)
	}

	targets := make([]target, dlcbf.numTables, dlcbf.numTables)
	for i := uint(0); i < dlcbf.numTables; i++ {
		targets[i] = target{
			bucketIndex: uint(indices[i]),
			fingerprint: fingerprint(binary.LittleEndian.Uint16(fp)),
		}
	}
	return targets
}

/*
Add data to filter return true if insertion was successful,
returns false if data already in filter or size limit was exceeeded
*/
func (dlcbf *Dlcbf) Add(data []byte) bool {
	targets := dlcbf.getTargets(data)

	_, _, target := dlcbf.lookup(targets)
	if target != nil {
		return false
	}

	minCount := uint8(math.MaxUint8)
	tableI := uint(0)

	for i, target := range targets {
		tmpCount := dlcbf.tables[i][target.bucketIndex].count
		if tmpCount < minCount && tmpCount < bucketHeight {
			minCount = dlcbf.tables[i][target.bucketIndex].count
			tableI = uint(i)
		}
	}

	if minCount == uint8(math.MaxUint8) {
		return false
	}
	bucket := &dlcbf.tables[tableI][targets[tableI].bucketIndex]
	bucket.entries[minCount] = targets[tableI].fingerprint
	bucket.count++
	return true
}

/*
Delete data to filter return true if deletion was successful,
returns false if data not in filter
*/
func (dlcbf *Dlcbf) Delete(data []byte) bool {
	deleted := false
	targets := dlcbf.getTargets(data)
	for i, target := range targets {
		for j, fp := range dlcbf.tables[i][target.bucketIndex].entries {
			if fp == target.fingerprint {
				if dlcbf.tables[i][target.bucketIndex].count == 0 {
					continue
				}
				dlcbf.tables[i][target.bucketIndex].count--
				k := 0
				for l, fp := range dlcbf.tables[i][target.bucketIndex].entries {
					if j == l {
						continue
					}
					dlcbf.tables[i][target.bucketIndex].entries[k] = fp
					k++
				}
				lastindex := dlcbf.tables[i][target.bucketIndex].count
				dlcbf.tables[i][target.bucketIndex].entries[lastindex] = 0
				deleted = true
			}
		}
	}
	return deleted
}

func (dlcbf *Dlcbf) lookup(targets []target) (uint, uint, *target) {
	for i, target := range targets {
		for j, fp := range dlcbf.tables[i][target.bucketIndex].entries {
			if fp == target.fingerprint {
				return uint(i), uint(j), &target
			}
		}
	}
	return 0, 0, nil
}

/*
IsMember returns true if data is in filter
*/
func (dlcbf *Dlcbf) IsMember(data []byte) bool {
	targets := dlcbf.getTargets(data)
	_, _, bfp := dlcbf.lookup(targets)
	return bfp != nil
}

/*
GetCount returns cardinlaity count of current filter
GetCount返回当前过滤器的卡片数量
*/
func (dlcbf *Dlcbf) GetCount() uint {
	count := uint(0)
	for _, table := range dlcbf.tables {
		for _, bucket := range table {
			count += uint(bucket.count)
		}
	}
	return count
}

func main()  {
	dlcbf, _ := NewDlcbfForCapacity(1000000)
	// 插入
	add_num := 0
	for i := 1001; i < 2000; i++ {
		add_f := dlcbf.Add([]byte(strconv.Itoa(i)))
		if add_f == false{
			add_num++
		}
	}
	// 查询
	ls_num:=0
	for j := 1001; j < 2000; j++ {
		is_member := dlcbf.IsMember([]byte(strconv.Itoa(j)))
		if is_member == false{
			ls_num++
		}
	}
	// 删除
	de_num:=0
	for k := 1001; k < 2000; k++ { 
		de_member := dlcbf.IsMember([]byte(strconv.Itoa(k)))
		if de_member == false{
			de_num++
		}
	// 误报率
    }
	fmt.Printf("insert is err is %d",add_num)
	fmt.Printf("find is err is %d",ls_num)
	fmt.Printf("delete is err is %d",de_num)
}
	
// func TestDlcbf(t *testing.T) {
// 	dlcbf, _ := NewDlcbfForCapacity(1000000)
// 	fd, err := os.Open("/usr/share/dict/web2")
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}
// 	scanner := bufio.NewScanner(fd)

// 	for scanner.Scan() {
// 		s := []byte(scanner.Text())
// 		// 插入字节byte
// 		dlcbf.Add(s)
// 	}

// 	count := dlcbf.GetCount()
// 	if float64(count)*100/235886 < 1 {
// 		t.Error("Expected error < 1 percent, got", float64(count)*100/235886)
// 	}

// 	fd, err = os.Open("/usr/share/dict/web2")
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}
// 	scanner = bufio.NewScanner(fd)

// 	for scanner.Scan() {
// 		s := []byte(scanner.Text())
// 		dlcbf.Delete(s)
// 	}

// 	count = dlcbf.GetCount()
// 	if count != 0 {
// 		t.Error("Expected count == 0, got", count)
// 	}

