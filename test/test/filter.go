package cuckoo

import (
	"fmt"
	"math/bits"
	"math/rand"
	"bytes"
	"encoding/gob"
	metro "github.com/dgryski/go-metro"
)

func main() {
	cf := NewFilter(1000000)
	for i := 1; i < 1000001; i++ {
		fmt.Println(i)
		InsertUnique([]byte(strconv.Itoa(i)))
	}

//	fd, err := os.Open("/home/jie/go/filter_exp/github_filter/github.com/seiflotfy/jiaoben/num.txt")
//	if err != nil {
//		panic(err)
//	}
//	scanner := bufio.NewScanner(fd)

//	var values [][]byte
//	var lineCount uint
//	for scanner.Scan() {
//		s := []byte(scanner.Text())
//		if cf.InsertUnique(s) {
//			lineCount++
//		}
//		values = append(values, s)
//	}

	count := cf.Count()
	fmt.Println(count)
	if count != lineCount {
		t.Errorf("Expected count = %d, instead count = %d", lineCount, count)
	}

//	for _, v := range values {
//		cf.Delete(v)
//	}

//	count = cf.Count()
//	if count != 0 {
//		t.Errorf("Expected count = 0, instead count == %d", count)
//	}
}


const maxCuckooCount = 500

// Filter is a probabilistic counter
type Filter struct {
	buckets   []bucket
	count     uint
	bucketPow uint
}

// NewFilter returns a new cuckoofilter with a given capacity.
// A capacity of 1000000 is a normal default, which allocates
// about ~1MB on 64-bit machines.
func NewFilter(capacity uint) *Filter {
	capacity = getNextPow2(uint64(capacity)) / bucketSize
	if capacity == 0 {
		capacity = 1
	}
	buckets := make([]bucket, capacity)
	return &Filter{
		buckets:   buckets,
		count:     0,
		bucketPow: uint(bits.TrailingZeros(capacity)),
	}
}

// Lookup returns true if data is in the counter
func (cf *Filter) Lookup(data []byte) bool {
	i1, fp := getIndexAndFingerprint(data, cf.bucketPow)
	if cf.buckets[i1].getFingerprintIndex(fp) > -1 {
		return true
	}
	i2 := getAltIndex(fp, i1, cf.bucketPow)
	return cf.buckets[i2].getFingerprintIndex(fp) > -1
}

// Reset ...
func (cf *Filter) Reset() {
	for i := range cf.buckets {
		cf.buckets[i].reset()
	}
	cf.count = 0
}

func randi(i1, i2 uint) uint {
	if rand.Intn(2) == 0 {
		return i1
	}
	return i2
}

// Insert inserts data into the counter and returns true upon success
func (cf *Filter) Insert(data []byte) bool {
	i1, fp := getIndexAndFingerprint(data, cf.bucketPow)
	if cf.insert(fp, i1) {
		return true
	}
	i2 := getAltIndex(fp, i1, cf.bucketPow)
	if cf.insert(fp, i2) {
		return true
	}
	return cf.reinsert(fp, randi(i1, i2))
}

// InsertUnique inserts data into the counter if not exists and returns true upon success
func (cf *Filter) InsertUnique(data []byte) bool {
	if cf.Lookup(data) {
		return false
	}
	return cf.Insert(data)
}

func (cf *Filter) insert(fp fingerprint, i uint) bool {
	if cf.buckets[i].insert(fp) {
		cf.count++
		return true
	}
	return false
}

func (cf *Filter) reinsert(fp fingerprint, i uint) bool {
	for k := 0; k < maxCuckooCount; k++ {
		j := rand.Intn(bucketSize)
		oldfp := fp
		fp = cf.buckets[i][j]
		cf.buckets[i][j] = oldfp

		// look in the alternate location for that random element
		i = getAltIndex(fp, i, cf.bucketPow)
		if cf.insert(fp, i) {
			return true
		}
	}
	return false
}

// Delete data from counter if exists and return if deleted or not
func (cf *Filter) Delete(data []byte) bool {
	i1, fp := getIndexAndFingerprint(data, cf.bucketPow)
	if cf.delete(fp, i1) {
		return true
	}
	i2 := getAltIndex(fp, i1, cf.bucketPow)
	return cf.delete(fp, i2)
}

func (cf *Filter) delete(fp fingerprint, i uint) bool {
	if cf.buckets[i].delete(fp) {
		if cf.count > 0 {
			cf.count--
		}
		return true
	}
	return false
}

// Count returns the number of items in the counter
func (cf *Filter) Count() uint {
	return cf.count
}

// Encode returns a byte slice representing a Cuckoofilter
func (cf *Filter) Encode() []byte {
	bytes := make([]byte, len(cf.buckets)*bucketSize)
	for i, b := range cf.buckets {
		for j, f := range b {
			index := (i * len(b)) + j
			bytes[index] = byte(f)
		}
	}
	return bytes
}

// Decode returns a Cuckoofilter from a byte slice
func Decode(bytes []byte) (*Filter, error) {
	var count uint
	if len(bytes)%bucketSize != 0 {
		return nil, fmt.Errorf("expected bytes to be multiple of %d, got %d", bucketSize, len(bytes))
	}
	if len(bytes) == 0 {
		return nil, fmt.Errorf("bytes can not be empty")
	}
	buckets := make([]bucket, len(bytes)/4)
	for i, b := range buckets {
		for j := range b {
			index := (i * len(b)) + j
			if bytes[index] != 0 {
				buckets[i][j] = fingerprint(bytes[index])
				count++
			}
		}
	}
	return &Filter{
		buckets:   buckets,
		count:     count,
		bucketPow: uint(bits.TrailingZeros(uint(len(buckets)))),
	}, nil
}




type fingerprint byte

type bucket [bucketSize]fingerprint

const (
	nullFp     = 0
	bucketSize = 4
)

func (b *bucket) insert(fp fingerprint) bool {
	for i, tfp := range b {
		if tfp == nullFp {
			b[i] = fp
			return true
		}
	}
	return false
}

func (b *bucket) delete(fp fingerprint) bool {
	for i, tfp := range b {
		if tfp == fp {
			b[i] = nullFp
			return true
		}
	}
	return false
}

func (b *bucket) getFingerprintIndex(fp fingerprint) int {
	for i, tfp := range b {
		if tfp == fp {
			return i
		}
	}
	return -1
}

func (b *bucket) reset() {
	for i := range b {
		b[i] = nullFp
	}
}






const (
	DefaultLoadFactor = 0.9
	DefaultCapacity   = 10000
)

type ScalableCuckooFilter struct {
	filters    []*Filter
	loadFactor float32
	//when scale(last filter size * loadFactor >= capacity) get new filter capacity
	scaleFactor func(capacity uint) uint
}

type option func(*ScalableCuckooFilter)

type Store struct {
	Bytes      [][]byte
	LoadFactor float32
}

/*
 by default option the grow capacity is:
 capacity , total
 4096  4096
 8192  12288
16384  28672
32768  61440
65536  126,976
*/
func NewScalableCuckooFilter(opts ...option) *ScalableCuckooFilter {
	sfilter := new(ScalableCuckooFilter)
	for _, opt := range opts {
		opt(sfilter)
	}
	configure(sfilter)
	return sfilter
}

func (sf *ScalableCuckooFilter) Lookup(data []byte) bool {
	for _, filter := range sf.filters {
		if filter.Lookup(data) {
			return true
		}
	}
	return false
}

func (sf *ScalableCuckooFilter) Reset() {
	for _, filter := range sf.filters {
		filter.Reset()
	}
}

func (sf *ScalableCuckooFilter) Insert(data []byte) bool {
	needScale := false
	lastFilter := sf.filters[len(sf.filters)-1]
	if (float32(lastFilter.count) / float32(len(lastFilter.buckets))) > sf.loadFactor {
		needScale = true
	} else {
		b := lastFilter.Insert(data)
		needScale = !b
	}
	if !needScale {
		return true
	}
	newFilter := NewFilter(sf.scaleFactor(uint(len(lastFilter.buckets))))
	sf.filters = append(sf.filters, newFilter)
	return newFilter.Insert(data)
}

func (sf *ScalableCuckooFilter) InsertUnique(data []byte) bool {
	if sf.Lookup(data) {
		return false
	}
	return sf.Insert(data)
}

func (sf *ScalableCuckooFilter) Delete(data []byte) bool {
	for _, filter := range sf.filters {
		if filter.Delete(data) {
			return true
		}
	}
	return false
}

func (sf *ScalableCuckooFilter) Count() uint {
	var sum uint
	for _, filter := range sf.filters {
		sum += filter.count
	}
	return sum

}

func (sf *ScalableCuckooFilter) Encode() []byte {
	slice := make([][]byte, len(sf.filters))
	for i, filter := range sf.filters {
		encode := filter.Encode()
		slice[i] = encode
	}
	store := &Store{
		Bytes:      slice,
		LoadFactor: sf.loadFactor,
	}
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(store)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

func (sf *ScalableCuckooFilter) DecodeWithParam(fBytes []byte, opts ...option) (*ScalableCuckooFilter, error) {
	instance, err := DecodeScalableFilter(fBytes)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt(instance)
	}
	return instance, nil
}

func DecodeScalableFilter(fBytes []byte) (*ScalableCuckooFilter, error) {
	buf := bytes.NewBuffer(fBytes)
	dec := gob.NewDecoder(buf)
	store := &Store{}
	err := dec.Decode(store)
	if err != nil {
		return nil, err
	}
	filterSize := len(store.Bytes)
	instance := NewScalableCuckooFilter(func(filter *ScalableCuckooFilter) {
		filter.filters = make([]*Filter, filterSize)
	}, func(filter *ScalableCuckooFilter) {
		filter.loadFactor = store.LoadFactor
	})
	for i, oneBytes := range store.Bytes {
		filter, err := Decode(oneBytes)
		if err != nil {
			return nil, err
		}
		instance.filters[i] = filter
	}
	return instance, nil

}

func configure(sfilter *ScalableCuckooFilter) {
	if sfilter.loadFactor == 0 {
		sfilter.loadFactor = DefaultLoadFactor
	}
	if sfilter.scaleFactor == nil {
		sfilter.scaleFactor = func(currentSize uint) uint {
			return currentSize * bucketSize * 2
		}
	}
	if sfilter.filters == nil {
		initFilter := NewFilter(DefaultCapacity)
		sfilter.filters = []*Filter{initFilter}
	}
}





	


var (
	altHash = [256]uint{}
	masks   = [65]uint{}
)

func init() {
	for i := 0; i < 256; i++ {
		altHash[i] = (uint(metro.Hash64([]byte{byte(i)}, 1337)))
	}
	for i := uint(0); i <= 64; i++ {
		masks[i] = (1 << i) - 1
	}
}

func getAltIndex(fp fingerprint, i uint, bucketPow uint) uint {
	mask := masks[bucketPow]
	hash := altHash[fp] & mask
	return (i & mask) ^ hash
}

func getFingerprint(hash uint64) byte {
	// Use least significant bits for fingerprint.
	fp := byte(hash%255 + 1)
	return fp
}

// getIndicesAndFingerprint returns the 2 bucket indices and fingerprint to be used
func getIndexAndFingerprint(data []byte, bucketPow uint) (uint, fingerprint) {
	hash := metro.Hash64(data, 1337)
	fp := getFingerprint(hash)
	// Use most significant bits for deriving index.
	i1 := uint(hash>>32) & masks[bucketPow]
	return i1, fingerprint(fp)
}

func getNextPow2(n uint64) uint {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return uint(n)
}