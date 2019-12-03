package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	start := time.Now()

	if len(os.Args) < 2 {
		panic("No file path specified.")
	}

	filePath := os.Args[1]

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	const bufferSize = 16 * 1024
	reader := bufio.NewReaderSize(file, bufferSize)

	lineCount := 0
	wordCount := 0
	byteCount := 0

	prevByteIsSpace := true

	for {
		b, err := reader.ReadByte()

		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}
		byteCount++
		switch b {
		case '\n':
			lineCount++
		case ' ', '\t', '\r', '\v', '\f':
			prevByteIsSpace = true
		default:
			if prevByteIsSpace {
				wordCount++
				prevByteIsSpace = false
			}
		}
	}
	fmt.Println(time.Since(start), file.Name(), byteCount, wordCount, lineCount)

}
