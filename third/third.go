package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

type Chunk struct {
	PreCharIsSpace bool
	Buffer         []byte
}

type Count struct {
	WordCount int
	LineCount int
}

func GetCount(chunk Chunk) Count {
	count := Count{}
	preCharIsSpace := chunk.PreCharIsSpace
	for _, b := range chunk.Buffer {
		switch b {
		case '\n':
			count.LineCount++
		case ' ', '\t', '\r', '\v', '\f':
			preCharIsSpace = true
		default:
			if preCharIsSpace {
				count.WordCount++
				preCharIsSpace = false
			}
		}
	}
	return count
}

func IsSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\v' || b == '\f'
}

func ChunkCount(chunks <-chan Chunk, counts chan<- Count) {
	totalCount := Count{}
	for {
		chunk, ok := <-chunks
		if !ok {
			break
		}
		count := GetCount(chunk)
		totalCount.WordCount += count.WordCount
		totalCount.LineCount += count.LineCount

	}
	counts <- totalCount
}

func main() {
	start := time.Now()
	chunks := make(chan Chunk)
	counts := make(chan Count)
	workersNum := runtime.NumCPU()
	for i := 0; i < workersNum; i++ {
		go ChunkCount(chunks, counts)
	}

	lastCharIsSpace := true

	const bufferSize = 1024 * 16
	buffer := make([]byte, bufferSize)

	if len(os.Args) < 2 {
		panic("no args")
	}

	filePath := os.Args[1]

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	for {
		bytes, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}
		chunks <- Chunk{lastCharIsSpace, buffer[:bytes]}
		lastCharIsSpace = IsSpace(buffer[bytes-1])
	}
	close(chunks)

	wordCount := 0
	lineCount := 0
	for i := 0; i < workersNum; i++ {
		count := <-counts
		wordCount += count.WordCount
		lineCount += count.LineCount
	}
	close(counts)

	fileStat, err := file.Stat()
	if err != nil {
		panic(err)
	}
	byteCount := fileStat.Size()
	fmt.Println(time.Since(start), file.Name(), byteCount, wordCount, lineCount)
}
