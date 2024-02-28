package sol1

import (
	"bufio"
	"fmt"
	"github.com/draculaas/1brc/common"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type node struct {
	min, max, sum, count int64
}

func Run(fileName string) string {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open the file %v", err)
	}

	defer func() {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}()

	s := bufio.NewScanner(f)

	mapping := make(map[string]*node)

	for s.Scan() {
		line := s.Text()
		data := strings.Split(line, ";")
		key := data[0]
		val := convertStringToInt64(data[1])
		if err != nil {
			log.Fatalf("Failed to parse temp %v", err)
		}

		if item, ok := mapping[data[0]]; !ok {
			mapping[key] = &node{min: val, max: val, sum: val, count: 1}
		} else {
			item.max = max(item.max, val)
			item.min = min(item.min, val)
			item.count += 1
			item.sum += val
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

func convertStringToInt64(input string) int64 {
	input = input[:len(input)-2] + input[len(input)-1:]
	output, _ := strconv.ParseInt(input, 10, 64)
	return output
}
