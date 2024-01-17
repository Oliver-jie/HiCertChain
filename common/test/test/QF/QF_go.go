// Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved

// Package qf implements a quotient filter data
// structure which supports:
//  1. external storage per entry
//  2. dynamic doubling
//  3. packed or unpacked representations (choose time or space)
//  4. a user overrideable hash function (default is murmur)
package qf

import (
	"fmt"
	"math"
	"unsafe"
)

//MaxLoadingFactor 指定我们将商过滤器哈希表加倍的边界，也用于初始大小表。
const MaxLoadingFactor = 0.65

// Filter is a quotient filter representation
type Filter struct {
	entries      uint64
	size         uint64
	filter       Vector  // Vector类 是可以实现自动增长的对象数组
	storage      Vector  // Vector类 是可以实现自动增长的对象数组
	rBits, qBits uint    // uint代表1位  
	rMask        uint64  // UInt16代表2个字节
	maxEntries   uint64
	config       Config
	hashfn       HashFn
	allocfn      VectorAllocateFn
}

// Len returns the number of entries in the quotient filter
func (qf *Filter) Len() uint64 {
	return qf.entries
}

// DebugDump prints a textual representation of the quotient filter to stdout
// DebugDump 将商过滤器的文本表示打印到标准输出
func (qf *Filter) DebugDump(full bool) {
	fmt.Printf("\nquotient filter is %d large (%d q bits) with %d entries (loaded %0.3f)\n",
		qf.size, qf.qBits, qf.entries, float64(qf.entries)/float64(qf.size))

	if full {
		fmt.Printf("  bucket  O C S remainder->\n")
		skipped := 0
		for i := uint64(0); i < uint64(qf.size); i++ {
			o, c, s := 0, 0, 0
			sd := qf.read(i)
			if sd.occupied() {
				o = 1
			}
			if sd.continuation() {
				c = 1
			}
			if sd.shifted() {
				s = 1
			}
			if sd.empty() {
				skipped++
			} else {
				if skipped > 0 {
					fmt.Printf("          ...\n")
					skipped = 0
				}
				r := sd.r()
				v := uint64(0)
				if qf.storage != nil {
					v = qf.storage.Get(i)
				}
				fmt.Printf("%8d  %d %d %d %x (%d)\n", i, o, c, s, r, v)
			}
		}
		if skipped > 0 {
			fmt.Printf("          ...\n")
		}
	}
}

// iterate the qf and call the callback once for each hash value present
// 迭代 qf 并为存在的每个哈希值调用一次回调
func (qf *Filter) eachHashValue(cb func(uint64, uint64)) {
	// a stack of q values
	stack := []uint64{}
	// let's start from an unshifted value
	start := uint64(0)
	for qf.read(start).shifted() {
		right(&start, qf.size)
	}
	end := start
	left(&end, qf.size)
	for i := start; true; right(&i, qf.size) {
		sd := qf.read(i)
		if !sd.continuation() && len(stack) > 0 {
			stack = stack[1:]
		}
		if sd.occupied() {
			stack = append(stack, i)
		}
		if len(stack) > 0 {
			r := sd.r()
			cb((stack[0]<<qf.rBits)|(r&qf.rMask), i)
		}
		if i == end {
			break
		}
	}
}

// New allocates a new quotient filter with default initial sizing and no external storage configured.
//New 分配一个具有默认初始大小且未配置外部存储的新商过滤器。
func New() *Filter {
	return NewWithConfig(Config{})
}

// NewWithConfig allocates a new quotient filter based on the supplied configuration.
// NewWithConfig 根据提供的配置分配一个新的商过滤器。
func NewWithConfig(c Config) *Filter {
	var qf Filter
	if c.BitPacked {
		qf.allocfn = BitPackedVectorAllocate
	} else {
		qf.allocfn = UnpackedVectorAllocate
	}
	if c.HashFn == nil {
		c.HashFn = murmurhash64
	}
	qf.hashfn = c.HashFn

	qbits := c.QBits()

	qf.initForQuotientBits(uint(qbits))

	qf.config = c

	qf.allocStorage()

	if qf.maxEntries > qf.size {
		panic("internal inconsistency")
	}
	return &qf
}

// BitsOfStoragePerEntry reports the configured external storage for the quotient filter
// BitsOfStoragePerEntry 报告为商过滤器配置的外部存储
func (qf *Filter) BitsOfStoragePerEntry() uint {
	return qf.config.BitsOfStoragePerEntry
}

func (qf *Filter) allocStorage() {
	qf.filter = qf.allocfn(3+bitsPerWord-qf.qBits, qf.size)
	if qf.config.BitsOfStoragePerEntry > 0 {
		qf.storage = qf.allocfn(qf.config.BitsOfStoragePerEntry, qf.size)
	}
}

// 初始化大小
func (qf *Filter) initForQuotientBits(qBits uint) {
	qf.qBits = qBits
	qf.rBits, qf.rMask, qf.size = initForQuotientBits(qBits)
	qf.rBits = (bitsPerWord - qBits)
	qf.rMask = 0
	for i := uint(0); i < qf.rBits; i++ {
		qf.rMask |= 1 << i
	}
	qf.maxEntries = uint64(math.Ceil(float64(qf.size) * MaxLoadingFactor))
}

func initForQuotientBits(qBits uint) (rBits uint, rMask, size uint64) {
	size = 1 << (uint64(qBits))
	rBits = (bitsPerWord - qBits)
	for i := uint(0); i < rBits; i++ {
		rMask |= 1 << i
	}
	return
}
// 一个槽中的数据占8个字节
type slotData uint64

const (
	occupiedMask     = slotData(1) // 强制性类型转换
	continuationMask = slotData(1 << 1) // a << 2 将a的二进制位左移2位
	shiftedMask      = slotData(1 << 2)
	bookkeepingMask  = slotData(0x7) // 十六进制的7
)

func (sd slotData) empty() bool {
	return (sd & bookkeepingMask) == 0
}

func (sd slotData) occupied() bool {
	return (sd & occupiedMask) != 0
}

func (sd *slotData) setOccupied(on bool) {
	if on {
		*sd |= occupiedMask
	} else {
		*sd &= ^occupiedMask
	}
}

func (sd slotData) continuation() bool {
	return (sd & continuationMask) != 0
}

func (sd *slotData) setContinuation(on bool) {
	if on {
		*sd |= continuationMask
	} else {
		*sd &= ^continuationMask
	}
}

func (sd slotData) shifted() bool {
	return (sd & shiftedMask) != 0
}

func (sd *slotData) setShifted(on bool) {
	if on {
		*sd |= shiftedMask
	} else {
		*sd &= ^shiftedMask
	}
}

func (sd slotData) r() uint64 {
	return uint64(sd >> 3)
}

func (sd *slotData) setR(r uint64) {
	*sd = (*sd & bookkeepingMask) | slotData(r<<3)
}

func (qf *Filter) read(slot uint64) slotData {
	return slotData(qf.filter.Get(slot))
}

func (qf *Filter) write(slot uint64, sd slotData) {
	qf.filter.Set(slot, uint64(sd))
}

func (qf *Filter) swap(slot uint64, sd slotData) slotData {
	return slotData(qf.filter.Swap(slot, uint64(sd)))
}

func (qf *Filter) countEntries() (count uint64) {
	for i := uint64(0); i < qf.size; i++ {
		if !qf.read(i).empty() {
			count++
		}
	}
	return
}

// InsertStringWithValue stores the string key and an associated
// integer value in the quotient filter it returns whether the
// key was already present in the quotient filter.
// InsertStringWithValue 将字符串键和关联的整数值存储在商过滤器中，
// 它返回键是否已经存在于商过滤器中。
func (qf *Filter) InsertStringWithValue(s string, value uint64) bool {
	return qf.InsertWithValue(*(*[]byte)(unsafe.Pointer(&s)), value)
}

// InsertString stores the string key in the quotient filter and
// returns whether this string was already present
// InsertString 将字符串键存储在商过滤器中，并返回该字符串是否已经存在
func (qf *Filter) InsertString(s string) bool {
	return qf.InsertStringWithValue(s, 0)
}

// InsertRawHash inserts a pre-calculated raw hash value with associated
// external data into the quotient filter.  The hash calculation algorithm
// must be the very same used internally by the quotient filter, otherwise
// lookups will fail.  This is a very low level insertion, use with care
// InsertRawHash 将预先计算的原始哈希值与相关的外部数据插入商过滤器。
// 哈希计算算法必须与商过滤器内部使用的完全相同，否则查找将失败。
// 这是一个非常低级别的插入，请小心使用
func (qf *Filter) InsertRawHash(hv uint64, value uint64) (update bool) {
	if qf.maxEntries <= qf.entries {
		qf.double()
	}
	dq := hv >> qf.rBits
	dr := hv & qf.rMask
	return qf.insertByHash(dq, dr, value)
}

func (qf *Filter) double() {
	// start with a shallow coppy
	cpy := *qf
	cpy.entries = 0
	cpy.initForQuotientBits(cpy.qBits + 1)
	cpy.allocStorage()
	qf.eachHashValue(func(hv uint64, slot uint64) {
		dq := hv >> cpy.rBits
		dr := hv & cpy.rMask
		var v uint64
		if qf.storage != nil {
			v = qf.storage.Get(slot)
		}
		cpy.insertByHash(dq, dr, v)
	})

	// shallow copy back over self
	*qf = cpy
}

// InsertWithValue stores the key (byte slice) and an integer value in
// the quotient filter.  It returns whether a value already existed.
// InsertWithValue 将键（字节切片）和一个整数值存储在商过滤器中。
// 它返回一个值是否已经存在。
func (qf *Filter) InsertWithValue(v []byte, value uint64) (update bool) {
	if qf.maxEntries <= qf.entries {
		qf.double()
	}
	dq, dr := hash(qf.hashfn, v, qf.rBits, qf.rMask)
	return qf.insertByHash(uint64(dq), uint64(dr), value)
}

// Insert stores the key (byte slice) in the quotient filter it
// returns whether it already existed
func (qf *Filter) Insert(v []byte) (update bool) {
	return qf.InsertWithValue(v, 0)
}

func (qf *Filter) insertByHash(dq, dr, value uint64) bool {
	sd := qf.read(dq)

	// case 1, the slot is empty
	if sd.empty() {
		qf.entries++
		sd.setOccupied(true)
		sd.setR(dr)
		qf.write(uint64(dq), sd)
		if qf.storage != nil {
			qf.storage.Set(dq, value)
		}
		return false
	}

	// if the occupied bit is set for this dq, then we are
	// extending an existing run
	// 如果为此 dq 设置了占用位，那么我们正在扩展现有运行
	extendingRun := sd.occupied()

	// mark occupied if we are not extending a run
	// 如果我们不延长运行，则标记已占用
	if !extendingRun {
		sd.setOccupied(true)
		qf.write(dq, sd)
	}

	// ok, let's find the start
	runStart := dq
	if sd.shifted() {
		runStart = findStart(dq, qf.size, qf.filter.Get)
	}
	// now let's find the spot within the run
	slot := runStart
	if extendingRun {
		sd = qf.read(slot)
		for {
			if sd.empty() || sd.r() >= dr {
				break
			}
			right(&slot, qf.size)
			sd = qf.read(slot)
			if !sd.continuation() {
				break
			}
		}
	}

	// case 2, the value is already in the filter
	if dr == sd.r() {
		// update value
		if qf.storage != nil {
			qf.storage.Set(slot, value)
		}
		return true
	}
	qf.entries++

	// case 3: we have to insert into an existing run
	// we are writing remainder <dr> into <slot>
	shifted := (slot != uint64(dq))
	continuation := slot != runStart

	for {
		// dr -> the remainder to write here
		if qf.storage != nil {
			value = qf.storage.Swap(slot, value)
		}
		var new slotData
		new.setShifted(shifted)
		new.setContinuation(continuation)
		old := qf.read(slot)
		new.setOccupied(old.occupied())
		new.setR(dr)
		qf.write(slot, new)
		if old.empty() {
			break
		}
		if ((slot == runStart) && extendingRun) || old.continuation() {
			continuation = true
		} else {
			continuation = false
		}
		dr = old.r()
		right(&slot, qf.size)
		shifted = true
	}
	return false
}

func right(i *uint64, size uint64) {
	*i++
	if *i >= size {
		*i = 0
	}
}

func left(i *uint64, size uint64) {
	if *i == 0 {
		*i += size
	}
	*i--
}

// XXX: error
func findStart(dq uint64, size uint64, read readFn) uint64 {
	// scan left to figure out how much to skip
	runs, complete := 1, 0
	for i := dq; true; left(&i, size) {
		sd := slotData(read(i))
		if !sd.continuation() {
			complete++
		}
		if !sd.shifted() {
			break
		} else if sd.occupied() {
			runs++
		}
	}
	// scan right to find our run
	for runs > complete {
		right(&dq, size)
		if !slotData(read(dq)).continuation() {
			complete++
		}
	}
	return dq
}

// Contains returns whether the byte slice is contained
// within the quotient filter
// Contains返回字节切片是否包含在商过滤器中
func (qf *Filter) Contains(v []byte) bool {
	found, _ := qf.Lookup(v)
	return found
}

// ContainsString returns whether the string is contained
// within the quotient filter
// ContainsString 返回字符串是否包含在商过滤器中
func (qf *Filter) ContainsString(s string) bool {
	found, _ := qf.Lookup(*(*[]byte)(unsafe.Pointer(&s)))
	return found
}

// Lookup searches for key and returns whether it
// exists, and the value stored with it (if any)
// Lookup 搜索key并返回它是否存在，以及与它一起存储的值（如果有）
func (qf *Filter) Lookup(key []byte) (bool, uint64) {
	dq, dr := hash(qf.hashfn, key, qf.rBits, qf.rMask)
	var storageFn readFn
	if qf.storage != nil {
		storageFn = qf.storage.Get
	}
	return lookupByHash(dq, dr, qf.size, qf.filter.Get, storageFn)
}

func lookupByHash(dq, dr, size uint64, read, storage readFn) (bool, uint64) {
	sd := slotData(read(dq))
	if !sd.occupied() {
		return false, 0
	}
	slot := dq
	if sd.shifted() {
		slot = findStart(dq, size, read)
		sd = slotData(read(slot))
	}
	for {
		if sd.r() == dr {
			value := uint64(0)
			if storage != nil {
				value = storage(slot)
			}
			return true, value
		}
		if sd.r() > dr {
			break
		}
		right(&slot, size)
		sd = slotData(read(slot))
		if !sd.continuation() {
			break
		}
	}
	return false, 0
}

// LookupString searches for key and returns whether it
// exists, and the value stored with it (if any)
func (qf *Filter) LookupString(key string) (bool, uint64) {
	return qf.Lookup(*(*[]byte)(unsafe.Pointer(&key)))
}

func hash(fn HashFn, v []byte, rBits uint, rMask uint64) (q, r uint64) {
	hv := fn(v)
	dq := hv >> rBits
	dr := hv & rMask
	return uint64(dq), uint64(dr)
}