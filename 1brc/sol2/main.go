package sol2

import (
	"bytes"
	"fmt"
	"github.com/draculaas/1brc/common"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type node struct {
	min, max, sum, count int64
}

func Run(fileName string) string {
	data := common.Mmap(fileName)

	workers := runtime.NumCPU()
	chunkSize := len(data) / workers
	if chunkSize == 0 {
		chunkSize = len(data)
	}

	chunks := make([]int, 0, chunkSize)
	offset := 0

	for offset < len(data) {
		offset += chunkSize

		if offset >= len(data) {
			chunks = append(chunks, len(data))
			break
		}
		chunkEndPos := bytes.IndexByte(data[offset:], '\n')
		if chunkEndPos == -1 {
			chunks = append(chunks, len(data))
			break
		} else {
			offset += chunkEndPos + 1
			chunks = append(chunks, offset)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(chunks))

	intermediate := make([]map[string]*node, len(chunks))
	start := 0

	for i, end := range chunks {
		dataSlice := data[start:end]
		go func() {
			intermediate[i] = handleChunk(dataSlice)
			wg.Done()
		}()
		start = end
	}

	wg.Wait()

	mapping := make(map[string]*node)

	for _, mp := range intermediate {
		for key, r := range mp {
			if item, ok := mapping[key]; !ok {
				mapping[key] = r
			} else {
				item.min = min(item.min, r.min)
				item.max = max(item.max, r.max)
				item.sum += r.sum
				item.count += r.count
			}
		}
	}

	cities := make([]string, 0, len(mapping))
	for city := range mapping {
		cities = append(cities, city)
	}
	sort.Strings(cities)

	var stringsBuilder strings.Builder

	stringsBuilder.WriteString(fmt.Sprintf("{"))
	for i, city := range cities {
		if i > 0 {
			stringsBuilder.WriteString(", ")
		}
		m := mapping[city]
		stringsBuilder.WriteString(fmt.Sprintf("%s=%.1f/%.1f/%.1f", city,
			common.Round(float64(m.min)/10.0),
			common.Round(float64(m.sum)/10.0/float64(m.count)),
			common.Round(float64(m.max)/10.0)))
	}
	stringsBuilder.WriteString(fmt.Sprintf("}\n"))

	return stringsBuilder.String()
}

func handleChunk(data []byte) map[string]*node {
	pos := 0
	mapping := make(map[string]*node)

	for len(data) > 0 {
		for i, b := range data {
			if b == ';' {
				pos = i
				break
			}
		}
		key := data[:pos]
		data = data[pos+1:]

		var tmp int64
		{
			isNegative := data[0] == '-'
			if isNegative {
				data = data[1:]
			}

			_ = data[3]
			if data[1] == '.' {
				// 1.2\n
				tmp = int64(data[0])*10 + int64(data[2]) - '0'*(10+1)
				data = data[4:]
				// 12.3\n
			} else {
				_ = data[4]
				tmp = int64(data[0])*100 + int64(data[1])*10 + int64(data[3]) - '0'*(100+10+1)
				data = data[5:]
			}

			if isNegative {
				tmp = -tmp
			}
		}

		if item, ok := mapping[string(key)]; !ok {
			mapping[string(key)] = &node{tmp, tmp, tmp, 1}
		} else {
			item.min = min(item.min, tmp)
			item.max = max(item.max, tmp)
			item.sum += tmp
			item.count++
		}
	}

	return mapping
}
