package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

type FileReader struct {
	File            *os.File
	LastCharIsSpace bool
	mutex           sync.Mutex
}

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

func (fileReader *FileReader) ReadChunk(buffer []byte) (Chunk, error) {
	fileReader.mutex.Lock()
	defer fileReader.mutex.Unlock()

	bytes, err := fileReader.File.Read(buffer)
	if err != nil {
		return Chunk{}, err
	}
	chunk := Chunk{PreCharIsSpace: fileReader.LastCharIsSpace, Buffer: buffer[:bytes]}
	fileReader.LastCharIsSpace = IsSpace(buffer[bytes-1])
	return chunk, nil
}

func FileReaderCount(reader *FileReader, counts chan<- Count) {
	const bufferSize = 1024 * 16
	buffer := make([]byte, bufferSize)
	totalCount := Count{}
	for {
		chunk, err := reader.ReadChunk(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}
		count := GetCount(chunk)
		totalCount.LineCount += count.LineCount
		totalCount.WordCount += count.WordCount
	}
	counts <- totalCount
}

func main() {
	start := time.Now()
	counts := make(chan Count)
	workersNum := runtime.NumCPU()
	if len(os.Args) < 2 {
		panic("no args")
	}

	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileReader := &FileReader{
		File:            file,
		LastCharIsSpace: true,
		mutex:           sync.Mutex{},
	}
	for i := 0; i < workersNum; i++ {
		go FileReaderCount(fileReader, counts)
	}

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
