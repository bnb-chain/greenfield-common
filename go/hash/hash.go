package hash

import (
	"bytes"
	"io"
	"os"
	"sync"

	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-common/go/redundancy"
)

// ComputeIntegrityHash split the reader into segment, ec encode the data, compute the hash roots of pieces
// return the hash result array list and data size
func ComputeIntegrityHash(reader io.Reader, segmentSize int64, dataShards, parityShards int) ([][]byte, int64, storageTypes.RedundancyType, error) {
	var segChecksumList [][]byte
	ecShards := dataShards + parityShards

	encodeData := make([][][]byte, ecShards)
	for i := 0; i < ecShards; i++ {
		encodeData[i] = make([][]byte, 0)
	}

	hashList := make([][]byte, ecShards+1)
	contentLen := int64(0)
	// read the data by segment size
	for {
		seg := make([]byte, segmentSize)
		n, err := reader.Read(seg)
		if err != nil {
			if err != io.EOF {
				log.Error().Msg("content read failed:" + err.Error())
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
				encodeData[index] = append(encodeData[index], shard)
			}
		}
	}

	// combine the hash root of pieces of the PrimarySP
	hashList[0] = GenerateIntegrityHash(segChecksumList)

	// compute the hash root of pieces of the SecondarySP
	wg := &sync.WaitGroup{}
	spLen := len(encodeData)
	wg.Add(spLen)
	for spID, content := range encodeData {
		go func(data [][]byte, id int) {
			defer wg.Done()
			var checksumList [][]byte
			for _, pieces := range data {
				piecesHash := GenerateChecksum(pieces)
				checksumList = append(checksumList, piecesHash)
			}

			hashList[id+1] = GenerateIntegrityHash(checksumList)
		}(content, spID)
	}

	wg.Wait()

	return hashList, contentLen, storageTypes.REDUNDANCY_EC_TYPE, nil
}

// ComputerHashFromFile open a local file and compute hash result and size
func ComputerHashFromFile(filePath string, segmentSize int64, dataShards, parityShards int) ([][]byte, int64, storageTypes.RedundancyType, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Error().Msg("open file fail:" + err.Error())
		return nil, 0, storageTypes.REDUNDANCY_EC_TYPE, err
	}
	defer f.Close()

	return ComputeIntegrityHash(f, segmentSize, dataShards, parityShards)
}

// ComputerHashFromBuffer support computing hash and size from byte buffer
func ComputerHashFromBuffer(content []byte, segmentSize int64, dataShards, parityShards int) ([][]byte, int64, storageTypes.RedundancyType, error) {
	reader := bytes.NewReader(content)
	return ComputeIntegrityHash(reader, segmentSize, dataShards, parityShards)
}
