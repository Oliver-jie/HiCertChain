package main

import (
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
	"math"
	"math/rand"
	"time"
	"strconv"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// ErrFull is returned when Add is called while the filter is at max capacity.
// ErrFull 在过滤器处于最大容量时调用 Add 时返回。
var ErrFull = errors.New("filter is at its max capacity")

// QuotientFilter is a basic quotient filter implementation.
// None of the methods are thread safe.
type QuotientFilter struct {
	// quotient and remainder bits
	qbits uint8
	rbits uint8
	// total slot size, qbits + 3 metadata bits
	ssize uint8
	// how many elements does the filter contain and capacity 1 << qbits
	// 过滤器包含多少元素和容量 1 << qbits
	len uint64
	cap uint64
	// data
	data []uint64
	// precalculated masks for slot, quotient and remainder
	// 槽、商和余数的预计算掩码
	sMask uint64
	qMask uint64
	rMask uint64
	// hash function
	h hash.Hash64
}

// NewProbability returns a quotient filter that can accomidate capacity number of elements
// and maintain the probability passed.
// NewProbability 返回一个商过滤器，可以容纳元素的容量数量并保持通过的概率。
func NewProbability(capacity int, probability float64) *QuotientFilter {
	// size to double asked capacity so that probability is maintained
	// at capacity num keys (at 50% fill rate)
	// 将请求容量加倍，以便将概率保持在容量 num 个键(填充率为 50%)
	q := uint8(math.Ceil(math.Log2(float64(capacity * 2))))
	r := uint8(-math.Log2(probability)) //以2为底的对数
	return New(q, r)
}

// NewHash returns a QuotientFilter backed by a different hash function.
// Default hash function is FNV-64a
// NewHash 返回由不同哈希函数支持的 QuotientFilter。 
// 默认哈希函数为 FNV-64a
func NewHash(h hash.Hash64, q, r uint8) *QuotientFilter {
	qf := New(q, r)
	qf.h = h
	return qf
}

// New returns a QuotientFilter with q quotient bits and r remainder bits.
// it can hold 1 << q elements.
// New 返回具有 q 个商位和 r 个余数位的 QuotientFilter。
// 它可以容纳 1 << q 个元素。
func New(q, r uint8) *QuotientFilter {
	if q+r > 64 {
		panic("q + r has to be less 64 bits or less")
	}
	qf := &QuotientFilter{
		qbits: q,
		rbits: r,
		ssize: r + 3,
		len:   0,
		cap:   1 << q,
		h:     fnv.New64a(),
	}
	qf.qMask = maskLower(uint64(q))
	qf.rMask = maskLower(uint64(r))
	qf.sMask = maskLower(uint64(qf.ssize))
	qf.data = make([]uint64, uint64Size(q, r))
	return qf
}

// FPProbability returns the probability for false positive with the current fillrate
// FPProbability 返回当前填充率的误报概率
// n = length
// m = capacity
// a = n / m
// r = remainder bits
// then probability for false positive is
// 1 - e^(-a/2^r) <= 2^-r
func (qf *QuotientFilter) FPProbability() float64 {
	a := float64(qf.len) / float64(qf.cap)
	return 1.0 - math.Pow(math.E, -(a/math.Pow(2, float64(qf.rbits))))
}

func (qf *QuotientFilter) info() {
	fmt.Printf("Filter qbits: %d, rbits: %d, len: %d, capacity: %d, current fp rate: %f\n", qf.qbits, qf.rbits, qf.len, qf.cap, qf.FPProbability())
	fmt.Println("slot, (is_occopied:is_continuation:is_shifted): remainder")
	for i := uint64(0); i < qf.cap; i++ {
		s := qf.getSlot(i)
		if i%8 == 0 && i != 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("% 5d: (%b%b%b): % 6d | ", i, s&1, s&2>>1, s&4>>2, s.remainder())
	}
	fmt.Printf("\n")
}

func (qf *QuotientFilter) quotientAndRemainder(h uint64) (uint64, uint64) {
	return (h >> qf.rbits) & qf.qMask, h & qf.rMask
}

func (qf *QuotientFilter) hash(key string) uint64 {
	defer qf.h.Reset()
	qf.h.Write([]byte(key))
	return qf.h.Sum64()
}

func (qf *QuotientFilter) getSlot(index uint64) slot {
	_, sliceIndex, bitOffset, nextBits := qf.slotIndex(index)
	s := (qf.data[sliceIndex] >> bitOffset) & qf.sMask
	// does the slot span to next slice index, if so, capture rest of the bits from there
	// 插槽是否跨越到下一个切片索引，如果是，则从那里捕获其余位
	if nextBits > 0 {
		sliceIndex++
		s |= (qf.data[sliceIndex] & maskLower(uint64(nextBits))) << (uint64(qf.ssize) - uint64(nextBits))
	}
	return slot(s)
}

func (qf *QuotientFilter) setSlot(index uint64, s slot) {
	// slot starts at bit data[sliceIndex][bitoffset:]
	// if the slot crosses slice boundary, nextBits contains
	// the number of bits the slot spans over to next slice item.
	// 槽从位 data[sliceIndex][bitoffset:] 开始，如果槽跨越切片边界，
	// nextBits 包含槽跨越到下一个切片项的位数。
	_, sliceIndex, bitOffset, nextBits := qf.slotIndex(index)
	// 删除除剩余和元位之外的所有内容。
	s &= slot(qf.sMask)
	qf.data[sliceIndex] &= ^(qf.sMask << bitOffset)
	qf.data[sliceIndex] |= uint64(s) << bitOffset
	// the slot spans slice boundary, write the rest of the element to next index.
	// 插槽跨越切片边界，将元素的其余部分写入下一个索引。
	if nextBits > 0 {
		sliceIndex++
		qf.data[sliceIndex] &^= maskLower(uint64(nextBits))
		qf.data[sliceIndex] |= uint64(s) >> (uint64(qf.ssize) - uint64(nextBits))
	}
}

func (qf *QuotientFilter) slotIndex(index uint64) (uint64, uint64, uint64, int) {
	bitIndex := uint64(qf.ssize) * index
	bitOffset := bitIndex % 64
	sliceIndex := bitIndex / 64
	bitsInNextSlot := int(bitOffset) + int(qf.ssize) - 64
	return bitIndex, sliceIndex, bitOffset, bitsInNextSlot
}

func (qf *QuotientFilter) previous(index uint64) uint64 {
	return (index - 1) & qf.qMask
}
func (qf *QuotientFilter) next(index uint64) uint64 {
	return (index + 1) & qf.qMask
}

// Contains checks if key is present in the filter
// false positive probability is based on q, r and number of added keys
// false negatives are not possible, unless Delete is used in conjunction with a hash function
// that yields more that q+r bits.
// Contains 检查过滤器中是否存在键误报概率基于 q、r 和添加键的数量 误报是不可能的，
// 除非删除与产生更多 q+r 位的哈希函数结合使用。
func (qf *QuotientFilter) Contains(key string) bool {
	q, r := qf.quotientAndRemainder(qf.hash(key))

	if !qf.getSlot(q).isOccupied() {
		return false
	}

	index := qf.findRun(q)
	slot := qf.getSlot(index)
	for {
		remainder := slot.remainder()
		if remainder == r {
			return true
		} else if remainder > r {
			return false
		}
		index = qf.next(index)
		slot = qf.getSlot(index)
		if !slot.isContinuation() {
			break
		}
	}
	return false
}

// Add adds the key to the filter.
func (qf *QuotientFilter) Add(key string) error {
	if qf.len >= qf.cap {
		return ErrFull
	}
	q, r := qf.quotientAndRemainder(qf.hash(key))
	slot := qf.getSlot(q)
	new := newSlot(r)

	// if slot is empty, just set the new there and occupy it and return.
	if slot.isEmpty() {
		qf.setSlot(q, new.setOccupied())
		qf.len++
		return nil
	}

	if !slot.isOccupied() {
		qf.setSlot(q, slot.setOccupied())
	}

	start := qf.findRun(q)
	index := start

	if slot.isOccupied() {
		runSlot := qf.getSlot(index)
		for {
			remainder := runSlot.remainder()
			if r == remainder {
				return nil
			} else if remainder > r {
				break
			}
			index = qf.next(index)
			runSlot = qf.getSlot(index)
			if !runSlot.isContinuation() {
				break
			}
		}
		if index == start {
			old := qf.getSlot(start)
			qf.setSlot(start, old.setContinuation())
		} else {
			new = new.setContinuation()
		}
	}
	if index != q {
		new = new.setShifted()
	}
	qf.insertSlot(index, new)
	qf.len++

	return nil
}

func (qf *QuotientFilter) insertSlot(index uint64, s slot) {
	curr := s
	for {
		prev := qf.getSlot(index)
		empty := prev.isEmpty()
		if !empty {
			prev = prev.setShifted()
			if prev.isOccupied() {
				curr = curr.setOccupied()
				prev = prev.clearOccupied()
			}
		}
		qf.setSlot(index, curr)
		curr = prev
		index = qf.next(index)
		if empty {
			break
		}
	}
}

func (qf *QuotientFilter) findRun(quotient uint64) (run uint64) {
	var slot slot
	index := quotient
	for {
		slot = qf.getSlot(index)
		if !slot.isShifted() {
			break
		}
		index = qf.previous(index)
	}
	run = index
	for index != quotient {
		for {
			run = qf.next(run)
			slot = qf.getSlot(run)
			if !slot.isContinuation() {
				break
			}
		}
		for {
			index = qf.next(index)
			slot = qf.getSlot(index)
			if slot.isOccupied() {
				break
			}
		}
	}
	return
}

// AddAll adds multiple keys to the filter
func (qf *QuotientFilter) AddAll(keys []string) error {
	for _, k := range keys {
		if err := qf.Add(k); err != nil {
			return err
		}
	}
	return nil
}	

func maskLower(e uint64) uint64 {
	return (1 << e) - 1
}

func uint64Size(q, r uint8) int {
	var bits int = (1 << q) * int(r+3)
	bytes := bits / 8
	if bits%8 != 0 {
		bytes++
	}
	return int(bytes)
}

type slot uint64

func newSlot(remainder uint64) slot {
	// shift remained left to make room for 3 control bits.
	return slot((int64(remainder) << 3) & ^7)
}
func (s slot) isOccupied() bool {
	return s&1 == 1
}
func (s slot) setOccupied() slot {
	s |= 1
	return s
}
func (s slot) clearOccupied() slot {
	clrBits := int64(^1)
	return s & slot(clrBits)
}
func (s slot) isContinuation() bool {
	return s&2 == 2
}
func (s slot) setContinuation() slot {
	return s | 2
}
func (s slot) clearContinuation() slot {
	clrBits := int64(^2)
	return s & slot(clrBits)
}

func (s slot) isShifted() bool {
	return s&4 == 4
}
func (s slot) setShifted() slot {
	return s | 4
}

func (s slot) clearShifted() slot {
	clrBits := int64(^4)
	return s & slot(clrBits)
}

func (s slot) remainder() uint64 {
	return uint64(s) >> 3
}
func (s slot) isEmpty() bool {
	return s&7 == 0
}
func (s slot) isClusterStart() bool {
	return s.isOccupied() && !s.isContinuation() && !s.isShifted()
}
func (s slot) isRunStart() bool {
	return s.isContinuation() && (s.isOccupied() || s.isShifted())
}

func main()  {
	qf := NewProbability(1024*8*1024,1000)
	f_pos := qf.FPProbability()
	qf.info()
	// 插入过程
	f_num :=0
	num_not:=0
	for i := 1001; i < 2000; i++ {
		err := qf.Add(strconv.Itoa(i))
		if err != nil {
			f_num++
		}
	}
	// 查询过程
	for j := 1001; j < 2000; j++ {
		not_exist := qf.Contains(strconv.Itoa(j))
		if not_exist == false {
			num_not++
		}
	}
	fmt.Printf("the false is %v\n",f_pos)
	fmt.Printf("the Add err is %d\n",f_num)
	fmt.Printf("the Contains err is %d\n",num_not)
}