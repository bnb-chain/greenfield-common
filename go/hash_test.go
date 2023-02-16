package redundancy

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

const (
	segmentSize = 16 * 1024 * 1024
	ecShards    = 6
)

func TestHash(t *testing.T) {
	length := int64(32 * 1024 * 1024)
	contentToHash := createTestData(length)
	start := time.Now()

	hashResult, size, err := ComputerHash(contentToHash, int64(segmentSize), ecShards)
	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Println("hash cost time:", time.Since(start).Milliseconds(), "ms")

	if size != length {
		t.Errorf("compute size error")
	}

	if len(hashResult) != ecShards+1 {
		t.Errorf("cimpute hash num not right")
	}

	for _, hash := range hashResult {
		if len(hash) != 64 {
			t.Errorf("hash length not right")
		}
	}

	hashList, _, err := ComputerHashFromFile("hash.go", int64(segmentSize), ecShards)

	if len(hashList) != ecShards+1 {
		t.Errorf("cimpute hash num not right")
	}

	for _, hash := range hashResult {
		if len(hash) != 64 {
			t.Errorf("hash length not right")
		}
	}
}

func createTestData(size int64) *strings.Reader {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	r := strings.NewReader(string(buf))
	return r
}
