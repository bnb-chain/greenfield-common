package hash

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-common/go/redundancy"
)

const (
	segmentSize          = 16 * 1024 * 1024
	expectedHashBytesLen = 32
)

func TestHash(t *testing.T) {
	length := int64(32 * 1024 * 1024)
	contentToHash := createTestData(length)
	start := time.Now()

	hashResult, size, redundnacyType, err := ComputeIntegrityHash(contentToHash, int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks)
	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Println("hash cost time:", time.Since(start).Milliseconds(), "ms")
	if size != length {
		t.Errorf("compute size error")
	}
	if redundnacyType != types.REDUNDANCY_EC_TYPE {
		t.Errorf("compare  redundnacy type error")
	}

	if len(hashResult) != redundancy.DataBlocks+redundancy.ParityBlocks+1 {
		t.Errorf("compute hash num not right")
	}

	for _, hash := range hashResult {
		if len(hash) != expectedHashBytesLen {
			t.Errorf("hash length not right")
		}
	}
	hashList, _, _, err := ComputerHashFromFile("hash.go", int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks)
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
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890`

	// generate 98 buffer
	for i := 0; i < 1024*1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	hashList, _, _, err := ComputeIntegrityHash(bytes.NewReader(buffer.Bytes()), int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks)
	if err != nil {
		t.Errorf(err.Error())
	}

	// this is generated from sp side
	expectedHashList := []string{
		"6YA/kt2H0pS6+/tyR20LCqqeWmNCelS4wQcEUIhnAko=",
		"C00Wks+pfo6NBQkG8iRGN5M0EtTvUAwMyaQ8+RsG4rA=",
		"Z5AW9CvNIsDo9jtxeQysSpn2ayNml3Kr4ksm/2WUu8s=",
		"dMlsKDw2dGRUygEgkyHJvOHYn9jVtycpUb7zvIGvEEk=",
		"v7vNLlbIg+27zFAOYfT2UDkoAId53Z1gDkcTA7VWT5A=",
		"1b7QsyQ8QT+7UoMU7K1SRhKOfIylogIfrSFsKJUfi4U=",
		"/7A2gwAnaJ5jFuK6sbov6iFAkhfOga4wdAK/NlCuJBo=",
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
