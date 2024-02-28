package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/trace"
	"strings"
	"sync"
	"sync/atomic"
)

type (
	item struct {
		XMLName     xml.Name `xml:"item"`
		Title       string   `xml:"title"`
		Description string   `xml:"description"`
	}
	channel struct {
		XMLName xml.Name `xml:"channel"`
		Items   []item   `xml:"item"`
	}
	document struct {
		XMLName xml.Name `xml:"rss"`
		Channel channel  `xml:"channel"`
	}
)

func fn(target string, docs []string) int {
	var counter int32

	g := runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	wg.Add(g)

	log.Printf("GOMAX [%v]", g)

	ch := make(chan string, g)

	for i := 0; i < g; i++ {
		go func() {
			var localCounter int32

			defer func() {
				//	Performance issue due when the copy inside of any core is incremented
				//	it will mark all the other copies in all the other cores dirty
				atomic.AddInt32(&counter, localCounter)
				wg.Done()
			}()

			for doc := range ch {
				file := fmt.Sprintf("../data/%s.xml", doc[:8])
				f, err := os.OpenFile(file, os.O_RDONLY, 0)
				if err != nil {
					log.Printf("Opening Document [%s] : ERROR : %v", doc, err)
					return
				}

				data, err := io.ReadAll(f)
				f.Close()
				if err != nil {
					log.Printf("Reading Document [%s] : ERROR : %v", doc, err)
					return
				}

				var d document
				if err := xml.Unmarshal(data, &d); err != nil {
					log.Printf("Decoding Document [%s] : ERROR : %v", doc, err)
					return
				}

				for _, item := range d.Channel.Items {
					if strings.Contains(item.Title, target) || strings.Contains(item.Description, target) {
						localCounter++
					}
				}
			}
		}()
	}

	for _, doc := range docs {
		ch <- doc
	}
	close(ch)

	wg.Wait()

	return int(counter)
}

func main() {
	trace.Start(os.Stdout)
	defer trace.Stop()

	docs := make([]string, 4000)
	for i := range docs {
		docs[i] = fmt.Sprintf("newsfeed-%.4d.xml", i)
	}
	target := "president"
	n := fn(target, docs)
	log.Printf("Searching %d files, found %s %d times.", len(docs), target, n)
}
