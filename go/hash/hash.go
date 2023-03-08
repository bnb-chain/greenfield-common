package hash

import (
	"bytes"
	"io"
	"os"
	"sync"

	"github.com/bnb-chain/greenfield-common/go/redundancy"
	"github.com/rs/zerolog/log"
)

// ComputerHash split the reader into segment, ec encode the data, compute the hash roots of pieces
// return the hash result array list and data size
func ComputerHash(reader io.Reader, segmentSize int64, dataShards, parityShards int) ([][]byte, int64, error) {
	var segChecksumList [][]byte
	var result [][]byte
	ecShards := dataShards + parityShards
	encodeData := make([][][]byte, ecShards)

	for i := 0; i < ecShards; i++ {
		encodeData[i] = make([][]byte, 0)
	}

	contentLen := int64(0)
	// read the data by segment size
	for {
		seg := make([]byte, segmentSize)
		n, err := reader.Read(seg)
		if err != nil {
			if err != io.EOF {
				log.Error().Msg("content read failed:" + err.Error())
				return nil, 0, err
			}
			break
		}
		if n > 0 {
			contentLen += int64(n)
			data := seg[:n]
			// compute segment hash
			checksum := GenerateChecksum(data)
			segChecksumList = append(segChecksumList, checksum)
			// get erasure encode bytes
			encodeShards, err := redundancy.EncodeRawSegment(data, dataShards, parityShards)
			if err != nil {
				return nil, 0, err
			}

			for index, shard := range encodeShards {
				encodeData[index] = append(encodeData[index], shard)
			}
		}
	}

	// combine the hash root of pieces of the PrimarySP
	segmentRootHash := GenerateIntegrityHash(segChecksumList)
	result = append(result, segmentRootHash)

	// compute the hash root of pieces of the SecondarySP
	wg := &sync.WaitGroup{}
	spLen := len(encodeData)
	wg.Add(spLen)
	hashList := make([][]byte, spLen)
	for spID, content := range encodeData {
		go func(data [][]byte, id int) {
			defer wg.Done()
			var checksumList [][]byte
			for _, pieces := range data {
				piecesHash := GenerateChecksum(pieces)
				checksumList = append(checksumList, piecesHash)
			}

			hashList[id] = GenerateIntegrityHash(checksumList)
		}(content, spID)
	}

	wg.Wait()

	for i := 0; i < spLen; i++ {
		result = append(result, hashList[i])
	}
	return result, contentLen, nil
}

// ComputerHashFromFile open a local file and compute hash result
func ComputerHashFromFile(filePath string, segmentSize int64, dataShards, parityShards int) ([][]byte, int64, error) {
	f, err := os.Open(filePath)
	// If any error fail quickly here.
	if err != nil {
		log.Error().Msg("open file fail:" + err.Error())
		return nil, 0, err
	}
	defer f.Close()

	return ComputerHash(f, segmentSize, dataShards, parityShards)
}

// ComputerHashFromBuffer support compute hash from byte buffer
func ComputerHashFromBuffer(content []byte, segmentSize int64, dataShards, parityShards int) ([][]byte, int64, error) {
	reader := bytes.NewReader(content)
	return ComputerHash(reader, segmentSize, dataShards, parityShards)
}
