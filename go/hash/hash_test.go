package hash

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
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
	testSize             = 32 * 1024 * 1024
)

func TestHash(t *testing.T) {
	length := int64(testSize)
	contentToHash := createTestData(length)
	start := time.Now()

	hashResult, size, redundancyType, err := ComputeIntegrityHash(contentToHash, int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks, true)
	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Println("hash cost time:", time.Since(start).Milliseconds(), "ms")
	if size != length {
		t.Errorf("compute segmentSize error")
	}
	if redundancyType != types.REDUNDANCY_EC_TYPE {
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
	hashList, _, _, err := ComputeIntegrityHash(bytes.NewReader(buffer.Bytes()), int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks, true)
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
			t.Errorf("compare hash error, id: %d, hash1, %s, hash2 %s \n", id, base64.StdEncoding.EncodeToString(hash), expectedHashList[id])
		}
	}
}

func TestParallelHashResult(t *testing.T) {
	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890`

	// generate 98 buffer
	for i := 0; i < 1024*1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}

	hashList, _, _, err := ComputeIntegrityHashParallel(bytes.NewReader(buffer.Bytes()), int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks)
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
			t.Errorf("compare hash error, id: %d, hash1, %s, hash2 %s \n", id, base64.StdEncoding.EncodeToString(hash), expectedHashList[id])
		}
	}
}

// TestCompareHashResult compare serial and parallel version function hash results with different file size,
// it is expected that the hash result are same with different version.
func TestCompareHashResult(t *testing.T) {
	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,123456789012`
	// test file less than 16M
	for i := 0; i < 1024*100; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	compareHashResult(buffer, t)

	buffer.Reset()
	// test file 100M
	for i := 0; i < 1024*1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	compareHashResult(buffer, t)

	buffer.Reset()
	// test file 500M
	for i := 0; i < 1024*1024*5; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	compareHashResult(buffer, t)

	// test file 1G
	for i := 0; i < 1024*1024*10; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	compareHashResult(buffer, t)
}

// compareHashResult compare serial and parallel version function hash results
func compareHashResult(buffer bytes.Buffer, t *testing.T) {
	start := time.Now()
	expectedHashList, _, _, err := ComputeIntegrityHash(bytes.NewReader(buffer.Bytes()), int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks, false)
	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Println(" serial computing hash cost time:", time.Since(start).Milliseconds(), "ms")

	start = time.Now()
	paralleResult, _, _, err := ComputeIntegrityHashParallel(bytes.NewReader(buffer.Bytes()), int64(segmentSize), redundancy.DataBlocks, redundancy.ParityBlocks)
	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Println(" parallel computing hash cost time:", time.Since(start).Milliseconds(), "ms")

	for id, hash := range paralleResult {
		if !bytes.Equal(hash, expectedHashList[id]) {
			t.Errorf("compare hash error, id: %d, hash1, %s, hash2 %s \n", id, hash, expectedHashList[id])
		}
	}
}

func TestIntegrityHasher(t *testing.T) {
	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890`

	// generate 98 buffer
	for i := 0; i < 1024*1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
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

	contentlen := len(buffer.Bytes())
	hashHandler := NewHasher(segmentSize, 4, 2)
	hashHandler.Init()
	bufferCopy := make([]byte, contentlen)
	copy(bufferCopy, buffer.Bytes())

	reader := bytes.NewReader(buffer.Bytes())
	for {
		seg := make([]byte, segmentSize/2+50)
		n, err := reader.Read(seg)
		if err != nil {
			if err != io.EOF {
				t.Errorf(err.Error())
			}
			break
		}
		err = hashHandler.Append(seg[:n])
		if err != nil {
			t.Errorf(err.Error())
		}
	}

	if err := verifyHashResult(hashHandler, expectedHashList, int64(contentlen)); err != nil {
		t.Errorf(err.Error())
	}

	hashHandler.Init()
	reader = bytes.NewReader(bufferCopy)
	// change segment read chunk size and test again
	for {
		seg := make([]byte, segmentSize/2+100)
		n, err := reader.Read(seg)
		if err != nil {
			if err != io.EOF {
				t.Errorf(err.Error())
			}
			break
		}
		err = hashHandler.Append(seg[:n])
		if err != nil {
			t.Errorf(err.Error())
		}
	}

	if err := verifyHashResult(hashHandler, expectedHashList, int64(contentlen)); err != nil {
		t.Errorf(err.Error())
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

func verifyHashResult(hashHandler *IntegrityHasher, expectedResult []string, expectedSize int64) error {
	hashList, size, _, err := hashHandler.Finish()
	if err != nil {
		return err
	}

	if size != expectedSize {
		return errors.New("get error size")
	}

	for id, hash := range hashList {
		if base64.StdEncoding.EncodeToString(hash) != expectedResult[id] {
			return errors.New("failed to compare hash")
		}
	}

	return nil
}
