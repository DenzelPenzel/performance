package sol4

import (
	"bytes"
	"fmt"
	"github.com/draculaas/1brc/common"
	"math/bits"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"unsafe"
)

var file *os.File

const (
	chunkSize  = 1 << 20
	bucketSize = 1 << 16
)

type split struct {
	offset, len int64
}

type chunk struct {
	offset int64
	start  bool
	raw    string
}

type worker struct {
	m      mapping
	chunks []chunk
}

func (w *worker) exec(wg *sync.WaitGroup, ch <-chan split) {
	buf := make([]byte, chunkSize)
	chunks := make([]chunk, 0, 100)

	for r := range ch {
		b := buf[0:r.len]
		_, err := file.ReadAt(b, r.offset)
		if err != nil {
			return
		}

		firstEndLine := bytes.IndexByte(b, '\n')
		chunks = append(chunks, chunk{
			offset: r.offset,
			start:  false,
			raw:    string(b[:firstEndLine+1]),
		})

		lastEndLine := bytes.LastIndexByte(b, '\n')
		if lastEndLine < len(b)-1 {
			chunks = append(chunks, chunk{
				offset: r.offset + r.len,
				start:  true,
				raw:    string(b[lastEndLine+1:]),
			})
		}

		startPtr := uintptr(unsafe.Pointer(&b[0]))
		start := startPtr + uintptr(firstEndLine) + 1
		end := startPtr + uintptr(lastEndLine) + 1

		for start < end {
			hash, val, nameLen, lineLen := parse(start)
			// find item in map
			ok, item := w.m.find(hash)
			if !ok {
				item.hash = hash
				item.name = string(b[start-startPtr : start-startPtr+nameLen])
				item.count = 1
				item.min = val
				item.max = val
				item.sum = val
			} else {
				item.min = min(item.min, val)
				item.max = max(item.max, val)
				item.sum += val
				item.count++
			}
			start += lineLen
		}
	}

	w.chunks = chunks
	wg.Done()
}

func parse(ptr uintptr) (hash uint64, val int64, nameLen, lineLen uintptr) {
	sep := ptr + 1
	for ; *(*byte)(unsafe.Pointer(sep)) != ';'; sep++ {
	}
	nameLen = sep - ptr

	for ; ptr+8 < sep; ptr += 8 {
		hash ^= *(*uint64)(unsafe.Pointer(ptr))
		hash *= 7
	}
	hash ^= *(*uint64)(unsafe.Pointer(ptr)) & ((1 << ((sep - ptr) * 8)) - 1)

	// Let's try to parse without any conditionals.
	//
	// Four possibilities:
	//
	//   a.b\n         ?? ?? 0A bb 2E aa
	//   ab.c\n        ?? 0A cc 2E bb aa
	//   -a.b\n        ?? 0A bb 2E aa 2D
	//   -ab.c\n       0A cc 2E bb aa 2D

	// ASCII values:
	//  -    0x2D      0b00101101
	//  .    0x2E      0b00101110
	//  \n   0x0A      0b00001010
	//  0-9  0x30-0x39 0b0011....

	// Restrict to the lower 5 bytes.
	x := *(*uint64)(unsafe.Pointer(sep + 1)) & 0xFFFFFFFFFF

	// Digits have the 5th bit (0x10) set to 1. The decimal point
	// can be in byte 1 (0x1000), 2 (0x100000) or 3 (0x10000000).
	n := bits.TrailingZeros64((^x) & 0x10101000)
	// Byte 1: n=12, format is "a.b"
	// Byte 2: n=20, format is "ab.c" or "-a.b"
	// Byte 3: n=28, format is "-ab.c"

	// Byte 0 is either a digit or '-'. Again we can check the 5th bit.
	minus := ((^x) >> 4) & 1        // 0 if no minus, or 1 if minus.
	minusMask := (minus - 1) & 0xFF // 0xFF if no minus, or 0 of minus.
	x = (x & (0xFFFFFFFF00 | minusMask)) << (28 - n)
	valUnsigned := ((x>>8)&0x0F)*100 + ((x>>16)&0x0F)*10 + (x>>32)&0x0F
	val = int64(valUnsigned ^ (-minus) + minus)

	return hash, val, nameLen, nameLen + 4 + uintptr(n)>>3
}

type record struct {
	name                 string
	hash                 uint64
	min, max, sum, count int64
}

type mapping struct {
	bucket [bucketSize]record
}

func (m *mapping) find(hash uint64) (bool, *record) {
	for i := hash % bucketSize; ; i = (i + 1) % bucketSize {
		if m.bucket[i].hash == hash {
			return true, &m.bucket[i]
		}
		if m.bucket[i].hash == 0 {
			return false, &m.bucket[i]
		}
	}
}

func Run(fileName string) string {
	var err error
	file, err = os.Open(fileName)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	info, err := file.Stat()
	if err != nil {
		panic(err)
	}
	size := info.Size()

	n := int((size + chunkSize - 1) / chunkSize)
	ch := make(chan split, n)

	for offset := int64(0); offset < size; offset += chunkSize {
		ch <- split{offset: offset, len: min(chunkSize, size-offset)}
	}
	close(ch)
	numGoroutines := runtime.GOMAXPROCS(0)

	// run workers
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	workers := make([]*worker, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		workers[i] = new(worker)
		go workers[i].exec(&wg, ch)
	}
	wg.Wait()

	// merge all chunks
	var chunks []chunk
	for _, w := range workers {
		chunks = append(chunks, w.chunks...)
	}

	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].offset < chunks[j].offset ||
			(chunks[i].offset == chunks[j].offset && chunks[i].start && !chunks[j].start)
	})

	buf := make([]byte, 1024)
	ptr := uintptr(unsafe.Pointer(&buf[0]))

	for i := 0; i < len(chunks); i++ {
		buf = append(buf[:0], []byte(chunks[i].raw)...)
		if i+1 < len(chunks) && chunks[i+1].offset == chunks[i].offset {
			i++
			buf = append(buf, []byte(chunks[i].raw)...)
		}

		hash, val, nameLen, _ := parse(ptr)

		ok, item := workers[0].m.find(hash)
		if !ok {
			item.hash = hash
			item.name = string(buf[:nameLen])
			item.count = 1
			item.min = val
			item.max = val
			item.sum = val
		} else {
			item.count++
			item.min = min(item.min, val)
			item.max = max(item.max, val)
			item.sum += val
		}
	}

	for _, w := range workers {
		for _, x := range w.m.bucket {
			if x.hash != 0 {
				ok, xx := workers[0].m.find(x.hash)
				if !ok {
					*xx = x
				} else {
					xx.sum += x.sum
					xx.count += x.count
					xx.min = min(xx.min, x.min)
					xx.max = max(xx.max, x.max)
				}
			}
		}
	}

	type stats struct {
		name, min, avg, max string
	}

	ss := make([]stats, 0, 1024)

	for _, item := range workers[0].m.bucket {
		if item.hash != 0 {
			ss = append(ss, stats{
				name: item.name,
				min:  fmt.Sprintf("%.1f", common.Round(float64(item.min)/10.0)),
				avg:  fmt.Sprintf("%.1f", common.Round(float64(item.sum)/10.0/float64(item.count))),
				max:  fmt.Sprintf("%.1f", common.Round(float64(item.max)/10.0)),
			})
		}
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].name < ss[j].name
	})

	res := make([]string, 0, len(ss))

	for _, i := range ss {
		res = append(res, fmt.Sprintf("%s=%s/%s/%s", i.name, i.min, i.avg, i.max))
	}

	return "{" + strings.Join(res, ", ") + "}\n"
}
