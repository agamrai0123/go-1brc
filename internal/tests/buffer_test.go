package tests

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"testing"
)

// run using := go test -bench=BenchmarkFileRead -benchmem > benchmark_results.txt

var filePath = "../../measurements.txt"

func BenchmarkFileRead(b *testing.B) {
	bufferSizes := []int{
		1024,             // 1 KB
		2 * 1024,         // 2 KB
		4 * 1024,         // 4 KB
		8 * 1024,         // 8 KB
		16 * 1024,        // 16 KB
		32 * 1024,        // 32 KB
		64 * 1024,        // 64 KB
		128 * 1024,       // 128 KB
		256 * 1024,       // 256 KB
		512 * 1024,       // 512 KB
		1024 * 1024,      // 1 MB
		2 * 1024 * 1024,  // 2 MB
		4 * 1024 * 1024,  // 4 MB
		8 * 1024 * 1024,  // 8 MB
		16 * 1024 * 1024, // 16 MB
		32 * 1024 * 1024, // 32 MB
		64 * 1024 * 1024, // 64 MB
	}

	for _, bufferSize := range bufferSizes {
		b.Run("Direct_"+getBufferSizeLabel(bufferSize), func(b *testing.B) {
			for b.Loop() {
				readFileDirect(bufferSize)
			}
		})

		b.Run("Bufio_"+getBufferSizeLabel(bufferSize), func(b *testing.B) {
			for b.Loop() {
				readFileWithBufio(bufferSize)
			}
		})
	}
}

func readFileDirect(bufferSize int) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buffer := make([]byte, bufferSize)
	for {
		_, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
	}
}

func readFileWithBufio(bufferSize int) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, bufferSize)
	buffer := make([]byte, bufferSize)
	for {
		_, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
	}
}

func getBufferSizeLabel(size int) string {
	if size < 1024 {
		return strconv.Itoa(size) + "B"
	} else if size < 1024*1024 {
		return strconv.Itoa(size/1024) + "KB"
	}
	return strconv.Itoa(size/(1024*1024)) + "MB"
}
