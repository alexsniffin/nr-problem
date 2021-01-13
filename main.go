package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
)

const (
	defaultChannelBuffer = 10000
	streamBatchSize      = 500
	filterFanOutSize     = 1000
	filterPattern        = `[^\w]`
	topNCount            = 100
)

type wordSequenceCount struct {
	words string
	count int
}

type flags []string

func (i *flags) String() string {
	return strings.Join(*i, "")
}

func (i *flags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// asynchronously reads in files, filters symbols, counts 3-word sequences and outputs the top 100 sequences and their counts
func main() {
	var myFlags flags
	flag.Var(&myFlags, "fpath", "input file")
	flag.Parse()

	// read in each stream and merge
	wordStreams := merge(func() []<-chan []string {
		var wordStreams []<-chan []string
		if len(myFlags) == 0 {
			// create stream from stdin if no flags are passed in
			wordStreams = append(wordStreams, readFromStream(os.Stdin, streamBatchSize))
		}
		for _, fileName := range myFlags {
			f, err := os.Open(fileName)
			if err != nil {
				log.Fatal(err)
			}

			// create stream of the file
			wordStreams = append(wordStreams, readFromStream(f, streamBatchSize))
		}
		return wordStreams
	}()...)

	// fan-out and parallelize the filtering of words
	filteredWords := merge(fanOut(filterFanOutSize, func() <-chan []string {
		return filter(filterPattern, wordStreams)
	})...)

	// count all of the words
	countedSequences := count(filteredWords)

	// filter to the topN words
	topSequences := topN(topNCount, countedSequences)

	// print the result
	for _, wsc := range topSequences {
		fmt.Println(fmt.Sprintf("%s - %d", wsc.words, wsc.count))
	}
}

// readFromStream asynchronously reads from an `io.Reader` and streams the content in a batches of strings, returns a channel
// of []string where each array will be the length of `batchSize`
func readFromStream(reader io.Reader, batchSize int) <-chan []string {
	ch := make(chan []string, defaultChannelBuffer)
	go func() {
		curWords := make([]string, 0)
		s := bufio.NewScanner(reader)
		s.Split(bufio.ScanWords)
		curBatch := 0
		for s.Scan() {
			if err := s.Err(); err != nil {
				log.Fatal(err)
			}
			if curBatch <= batchSize {
				curWords = append(curWords, s.Text())
				curBatch++
			} else {
				curWords = append(curWords, s.Text())
				ch <- curWords
				curWords = curWords[len(curWords)-2:] // retain the last 2 words so that the 3 word sequence is counted
				curBatch = 0
			}
		}
		ch <- curWords
		close(ch)
	}()
	return ch
}

// filter asynchronously filters a stream of []string based on the regex `pattern`, returns a channel of []string
func filter(pattern string, wordsBatch <-chan []string) <-chan []string {
	ch := make(chan []string, defaultChannelBuffer)
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for words := range wordsBatch {
			for i, word := range words {
				w := re.ReplaceAllString(word, "")
				words[i] = w
			}
			ch <- words
		}
		close(ch)
	}()
	return ch
}

// count asynchronously aggregates a stream of []string counting the occurrence of all 3 word sequences, returns a
// channel of `wordSequenceCount`
func count(wordsBatch <-chan []string) <-chan wordSequenceCount {
	ch := make(chan wordSequenceCount, defaultChannelBuffer)
	go func() {
		wscMap := make(map[string]int)
		for words := range wordsBatch {
			for i := 2; i < len(words); i++ {
				wordsSeq := fmt.Sprintf("%s %s %s", words[i-2], words[i-1], words[i])
				if c, ok := wscMap[wordsSeq]; ok {
					wscMap[wordsSeq] = c + 1
				} else {
					wscMap[wordsSeq] = 1
				}
			}
		}
		for k, v := range wscMap {
			ch <- wordSequenceCount{
				words: k,
				count: v,
			}
		}
		close(ch)
	}()
	return ch
}

// topN aggregates a stream of `wordSequenceCount` counting the topN in descending order, returns an array once the
// channel closes
func topN(n int, wscBatch <-chan wordSequenceCount) []wordSequenceCount {
	arr := make([]wordSequenceCount, 0)
	for wsc := range wscBatch {
		if len(arr) < n {
			arr = insertDescSort(arr, wsc)
		} else {
			arr = insertDescSort(arr, wsc)[:n]
		}
	}
	return arr
}

// insertDescSort finds the index to insert an element in descending order
func insertDescSort(wscArr []wordSequenceCount, wsc wordSequenceCount) []wordSequenceCount {
	if len(wscArr) == 0 {
		return []wordSequenceCount{wsc}
	}
	i := sort.Search(len(wscArr), func(i int) bool { return wscArr[i].count < wsc.count })
	wscArr = append(wscArr[:i], append([]wordSequenceCount{wsc}, wscArr[i:]...)...)
	return wscArr
}

// fanOut asynchronously executes a function returning a `<-chan []string` and parallelizes it by `count`,
// returning an array of channels
func fanOut(count int, fn func() <-chan []string) []<-chan []string {
	var arr []<-chan []string
	for i := 0; i < count; i++ {
		arr = append(arr, fn())
	}
	return arr
}

// merge asynchronously takes an array of channels and will merge them together until all channels are closed
func merge(cs ...<-chan []string) <-chan []string {
	var wg sync.WaitGroup
	out := make(chan []string)

	output := func(c <-chan []string) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
