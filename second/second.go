package main

import (
	"fmt"
	"io"
	"os"
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

func main() {
	start := time.Now()
	lastCharIsSpace := true
	wordCount := 0
	lineCount := 0

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
		count := GetCount(Chunk{lastCharIsSpace, buffer[:bytes]})
		lastCharIsSpace = IsSpace(buffer[bytes-1])
		wordCount += count.WordCount
		lineCount += count.LineCount
	}

	fileStat, err := file.Stat()
	if err != nil {
		panic(err)
	}
	byteCount := fileStat.Size()
	fmt.Println(time.Since(start), file.Name(), byteCount, wordCount, lineCount)
}
