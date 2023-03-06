package hash

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-common/go/redundancy"
)

const (
	segmentSize = 16 * 10
)

func TestHash(t *testing.T) {
	length := int64(32 * 1024 * 1024)
	contentToHash := createTestData(length)
	start := time.Now()

	hashResult, size, err := ComputerHash(contentToHash, int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks)
	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Println("hash cost time:", time.Since(start).Milliseconds(), "ms")
	if size != length {
		t.Errorf("compute size error")
	}

	if len(hashResult) != redundancy.DataBlocks+redundancy.ParityBlocks+1 {
		t.Errorf("compute hash num not right")
	}
	for _, hash := range hashResult {
		if len(hash) != 64 {
			t.Errorf("hash length not right")
		}
	}
	hashList, _, err := ComputerHashFromFile("hash.go", int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks)
	assert.Nil(t, err)
	if len(hashList) != redundancy.DataBlocks+redundancy.ParityBlocks+1 {
		t.Errorf("compute hash num not right")
	}
	for _, hash := range hashResult {
		if len(hash) != 64 {
			t.Errorf("hash length not right")
		}
	}
}

func TestHash2(t *testing.T) {
	length := int64(17 * 10)
	contentToHash := createTestData2(length)

	stringReader1 := strings.NewReader(contentToHash)
	stringReader2 := strings.NewReader(contentToHash)
	hashResult, _, err := ComputerHash(stringReader1, int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks)
	if err != nil {
		t.Errorf(err.Error())
	}

	for idx, hash := range hashResult {
		fmt.Println("hash", idx, "：", hash)
	}

	_, _, _ = ComputerHashNoParallel2(stringReader2, int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks)
	if err != nil {
		t.Errorf(err.Error())
	}

	hashResult2, _, err := ComputerHashNoParallel(stringReader2, int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks)

	for idx, hash := range hashResult2 {
		if hash != hashResult[idx] {
			t.Errorf("compare hash error")
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

func createTestData2(size int64) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(buf)
}
