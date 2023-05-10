package hash

import (
	"bytes"
	"errors"
	"io"
	"os"
	"sync"

	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-common/go/redundancy"
)

// IntegrityHasher compute integrityHash
type IntegrityHasher struct {
	ecDataHashes [][][]byte
	segHashes    [][]byte
	buffer       []byte
	segmentSize  int64
	dataShards   int
	parityShards int
	contentLen   int64
}

func NewHasher(size int64, data, parity int) *IntegrityHasher {
	return &IntegrityHasher{
		buffer:       make([]byte, 0),
		segmentSize:  size,
		dataShards:   data,
		parityShards: parity,
	}
}

// Init the integrityHash fields
func (i *IntegrityHasher) Init() {
	ecShards := i.dataShards + i.parityShards
	encodeDataHash := make([][][]byte, ecShards)
	for i := 0; i < ecShards; i++ {
		encodeDataHash[i] = make([][]byte, 0)
	}
	segChecksumList := make([][]byte, 0)
	i.ecDataHashes = encodeDataHash
	i.segHashes = segChecksumList
	if len(i.buffer) > 0 {
		i.buffer = i.buffer[:0]
	}
	i.contentLen = 0
}

// Append the data chunks to IntegrityHasher , the data size should be less than segment size
func (i *IntegrityHasher) Append(data []byte) error {
	dataSize := len(data)
	if dataSize > int(i.segmentSize) {
		return errors.New("the length of data size should be less than segmentSize")
	}
	if len(i.buffer) >= int(i.segmentSize) {
		return errors.New("the buffer of handler should be less than segmentSize")
	}
	originBuffer := make([]byte, len(i.buffer))
	copy(originBuffer, i.buffer)
	// use tempBuffer to store exceed data
	var tempBuffer []byte
	totalSize := int64(dataSize + len(i.buffer))
	if totalSize > i.segmentSize {
		index := dataSize - int(totalSize-i.segmentSize)
		tempBuffer = make([]byte, dataSize-index)
		copy(tempBuffer, data[index:])
		// buffer should be equal with segment size
		i.buffer = append(i.buffer, data[:index]...)
	} else {
		i.buffer = append(i.buffer, data...)
		// if buffer size less than segment size, just store the data
		if totalSize < i.segmentSize {
			return nil
		}
	}

	// compute segment hash
	if err := i.computeBufferHash(); err != nil {
		return err
	}

	// copy exceed data to buffer if exist
	if len(tempBuffer) > 0 {
		i.buffer = i.buffer[:0]
		i.buffer = append(i.buffer, tempBuffer...)
		return nil
	}

	i.buffer = i.buffer[:0]
	return nil
}

// Finish return the result of the Integrity hashes
func (i *IntegrityHasher) Finish() ([][]byte, int64, storageTypes.RedundancyType, error) {
	// deal with  remain content tot be computed
	if len(i.buffer) > 0 {
		if err := i.computeBufferHash(); err != nil {
			return nil, 0, storageTypes.REDUNDANCY_EC_TYPE, err
		}
	}

	hashList := make([][]byte, i.parityShards+i.dataShards+1)

	hashList[0] = GenerateIntegrityHash(i.segHashes)

	wg := &sync.WaitGroup{}
	spLen := len(i.ecDataHashes)
	wg.Add(spLen)
	for spID, content := range i.ecDataHashes {
		go func(data [][]byte, id int) {
			defer wg.Done()
			hashList[id+1] = GenerateIntegrityHash(data)
		}(content, spID)
	}
	wg.Wait()

	return hashList, i.contentLen, storageTypes.REDUNDANCY_EC_TYPE, nil
}

// computeBufferHash erasure encode the buffer of IntegrityHasher and compute the hash
func (i *IntegrityHasher) computeBufferHash() error {
	i.contentLen += int64(len(i.buffer))
	originBuffer := make([]byte, len(i.buffer))
	copy(originBuffer, i.buffer)
	// compute segment hash
	checksum := GenerateChecksum(i.buffer)
	i.segHashes = append(i.segHashes, checksum)
	// get erasure encoded bytes and compute pieces hashes
	encodeShards, err := redundancy.EncodeRawSegment(i.buffer, i.dataShards, i.parityShards)
	if err != nil {
		// recover buffer content if encode error
		i.buffer = i.buffer[:0]
		i.buffer = append(i.buffer, originBuffer...)
		return err
	}

	for index, shard := range encodeShards {
		// compute hash of pieces
		piecesHash := GenerateChecksum(shard)
		i.ecDataHashes[index] = append(i.ecDataHashes[index], piecesHash)
	}

	return nil
}

// ComputeIntegrityHash split the reader into segment, ec encode the data, compute the hash roots of pieces
// return the hash result array list and data segmentSize
func ComputeIntegrityHash(reader io.Reader, segmentSize int64, dataShards, parityShards int) ([][]byte, int64,
	storageTypes.RedundancyType, error,
) {
	var segChecksumList [][]byte
	ecShards := dataShards + parityShards

	encodeDataHash := make([][][]byte, ecShards)
	for i := 0; i < ecShards; i++ {
		encodeDataHash[i] = make([][]byte, 0)
	}

	hashList := make([][]byte, ecShards+1)
	contentLen := int64(0)
	// read the data by segment segmentSize
	for {
		seg := make([]byte, segmentSize)
		n, err := reader.Read(seg)
		if err != nil {
			if err != io.EOF {
				log.Error().Msg("failed to read content:" + err.Error())
				return nil, 0, storageTypes.REDUNDANCY_EC_TYPE, err
			}
			break
		}

		if n > 0 && n <= int(segmentSize) {
			contentLen += int64(n)
			data := seg[:n]
			// compute segment hash
			checksum := GenerateChecksum(data)
			segChecksumList = append(segChecksumList, checksum)
			// get erasure encode bytes
			encodeShards, err := redundancy.EncodeRawSegment(data, dataShards, parityShards)
			if err != nil {
				return nil, 0, storageTypes.REDUNDANCY_EC_TYPE, err
			}

			for index, shard := range encodeShards {
				// compute hash of pieces
				piecesHash := GenerateChecksum(shard)
				encodeDataHash[index] = append(encodeDataHash[index], piecesHash)
			}
		}
	}

	// combine the hash root of pieces of the PrimarySP
	hashList[0] = GenerateIntegrityHash(segChecksumList)

	// compute the integrity hash of the SecondarySP
	wg := &sync.WaitGroup{}
	spLen := len(encodeDataHash)
	wg.Add(spLen)
	for spID, content := range encodeDataHash {
		go func(data [][]byte, id int) {
			defer wg.Done()
			hashList[id+1] = GenerateIntegrityHash(data)
		}(content, spID)
	}

	wg.Wait()

	return hashList, contentLen, storageTypes.REDUNDANCY_EC_TYPE, nil
}

// ComputerHashFromFile open a local file and compute hash result and segmentSize
func ComputerHashFromFile(filePath string, segmentSize int64, dataShards, parityShards int) ([][]byte, int64, storageTypes.RedundancyType, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Error().Msg("failed to open file:" + err.Error())
		return nil, 0, storageTypes.REDUNDANCY_EC_TYPE, err
	}
	defer f.Close()

	return ComputeIntegrityHash(f, segmentSize, dataShards, parityShards)
}

// ComputerHashFromBuffer support computing hash and segmentSize from byte buffer
func ComputerHashFromBuffer(content []byte, segmentSize int64, dataShards, parityShards int) ([][]byte, int64, storageTypes.RedundancyType, error) {
	reader := bytes.NewReader(content)
	return ComputeIntegrityHash(reader, segmentSize, dataShards, parityShards)
}
