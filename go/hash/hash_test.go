package hash

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-common/go/redundancy"
)

const (
	segmentSize = 16 * 1024 * 1024
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
		if len(hash) != 32 {
			t.Errorf("hash length not right")
		}
	}
	hashList, _, err := ComputerHashFromFile("hash.go", int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks)
	assert.Nil(t, err)
	if len(hashList) != redundancy.DataBlocks+redundancy.ParityBlocks+1 {
		t.Errorf("compute hash num not right")
	}
	for _, hash := range hashResult {
		if len(hash) != 32 {
			t.Errorf("hash length not right")
		}
	}
}

func TestHashResult(t *testing.T) {
	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,11`
	// generate 100M+ buffer
	for i := 0; i < 1024*1000; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	hashList, _, err := ComputerHash(bytes.NewReader(buffer.Bytes()), int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks)
	if err != nil {
		t.Errorf(err.Error())
	}

	// this is generated from sp side
	expectedHashList := []string{
		"dhmPA471pRuKF95ln9VWEqpwtN8BanO+FhRbdIy0sM0=",
		"8mDetlm/ecGcNOcE5C7qsVsqp7S1eeB7wrRVF9nv32A=",
		"b8OEwaDv9D4joBOfLxtBFD2+GS5ut+HxvGCpkz6SymY=",
		"xh5A6s5pbC7/CbPjStIrlTSRRzl+kibWh+UJ5Xrtvm8=",
		"AzWZNSOqQfPI0Ti84imcNCfNpkUQ41qICjyYmPvn9xY=",
		"J2ekM038/tQMO5T6Zcf5JlpbXKkym6P9AdH6ozi0Wa0=",
	}

	for id, hash := range hashList {
		if base64.StdEncoding.EncodeToString(hash) != expectedHashList[id] {
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

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
