package hash

import (
	"bytes"
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-common/redundancy"
)

// ComputerHash split the reader into segment, ec encode the data, compute the hash roots of pieces,
// and return the hash result array list and data size
func ComputerHash(reader io.Reader, segmentSize int64, dataShards, parityShards int) ([]string, int64, error) {
	var segChecksumList [][]byte
	var result []string
	encodeData := make([][][]byte, dataShards+parityShards)
	seg := make([]byte, segmentSize)

	contentLen := int64(0)
	// read the data by segment size
	for {
		n, err := reader.Read(seg)
		if err != nil {
			if err != io.EOF {
				log.Error().Msg("content read fail:" + err.Error())
				return nil, 0, err
			}
			break
		}
		if n > 0 {
			contentLen += int64(n)
			// compute segment hash
			segmentReader := bytes.NewReader(seg[:n])
			if segmentReader != nil {
				checksum, err := CalcSHA256HashByte(segmentReader)
				if err != nil {
					log.Error().Msg("compute checksum fail:" + err.Error())
					return nil, 0, err
				}
				segChecksumList = append(segChecksumList, checksum)
			}

			// get erasure encode bytes
			encodeShards, err := redundancy.EncodeRawSegment(seg[:n], dataShards, parityShards)
			if err != nil {
				return nil, 0, err
			}

			for index, shard := range encodeShards {
				encodeData[index] = append(encodeData[index], shard)
			}
		}
	}

	// combine the hash root of pieces of the PrimarySP
	segBytesTotal := bytes.Join(segChecksumList, []byte(""))
	segmentRootHash := CalcSHA256Hex(segBytesTotal)
	result = append(result, segmentRootHash)

	// compute the hash root of pieces of the SecondarySP
	var wg = &sync.WaitGroup{}
	spLen := len(encodeData)
	wg.Add(spLen)
	hashList := make([]string, spLen)
	for spID, content := range encodeData {
		go func(data [][]byte, id int) {
			defer wg.Done()
			var checksumList [][]byte
			for _, pieces := range data {
				piecesHash := CalcSHA256(pieces)
				checksumList = append(checksumList, piecesHash)
			}

			piecesBytesTotal := bytes.Join(checksumList, []byte(""))
			hashList[id] = CalcSHA256Hex(piecesBytesTotal)
		}(content, spID)
	}
	wg.Wait()

	for i := 0; i < spLen; i++ {
		result = append(result, hashList[i])
	}
	return result, contentLen, nil
}

// ComputerHashFromFile open a local file and compute hash result
func ComputerHashFromFile(filePath string, segmentSize int64, dataShards, parityShards int) ([]string, int64, error) {
	f, err := os.Open(filePath)
	// If any error fail quickly here.
	if err != nil {
		log.Error().Msg("open file fail:" + err.Error())
		return nil, 0, err
	}
	defer f.Close()

	return ComputerHash(f, segmentSize, dataShards, parityShards)
}
