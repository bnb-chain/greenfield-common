package hash

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-common/go/redundancy"
)

// ComputerHash split the reader into segment, ec encode the data, compute the hash roots of pieces
// return the hash result array list and data size
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
				log.Error().Msg("content read failed:" + err.Error())
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
					log.Error().Msg("compute checksum failed:" + err.Error())
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
	wg := &sync.WaitGroup{}
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

func ComputerHashNoParallel(reader io.Reader, segmentSize int64, dataShards, parityShards int) ([]string, int64, error) {
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
				log.Error().Msg("content read failed:" + err.Error())
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
					log.Error().Msg("compute checksum failed:" + err.Error())
					return nil, 0, err
				}
				segChecksumList = append(segChecksumList, checksum)
			}

			// get erasure encode bytes
			encodeShards, err := redundancy.EncodeRawSegment(seg[:n], dataShards, parityShards)
			if err != nil {
				return nil, 0, err
			}
			log.Error().Msg("piece hash info, ec_0 hash1: " + hex.EncodeToString(CalcSHA256(encodeShards[0])))
			//log.Error().Msg("piece hash info, ec_0 hash1: " + hex.EncodeToString(encodeShards[0]))
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
	spLen := len(encodeData)

	hashList := make([]string, spLen)
	for spID, content := range encodeData {
		var checksumList [][]byte
		for _, pieces := range content {
			if spID == 0 {
				log.Error().Msg("piece hash info, ec_0 hash2: " + hex.EncodeToString(CalcSHA256(pieces)))
			}
			piecesHash := CalcSHA256(pieces)

			checksumList = append(checksumList, piecesHash)
		}
		piecesBytesTotal := bytes.Join(checksumList, []byte(""))
		hashList[spID] = CalcSHA256Hex(piecesBytesTotal)
	}

	for i := 0; i < spLen; i++ {
		result = append(result, hashList[i])
	}
	return result, contentLen, nil
}

func ComputerHashNoParallel2(reader io.Reader, segmentSize int64, dataShards, parityShards int) ([]string, int64, error) {
	ecShards := dataShards + parityShards
	encodeData := make([][][]byte, ecShards)

	for i := 0; i < ecShards; i++ {
		encodeData[i] = make([][]byte, 0)
	}
	seg := make([]byte, segmentSize)
	// read the data by segment size
	segNum := 0
	for {
		n, err := reader.Read(seg)
		if err != nil {
			if err != io.EOF {
				log.Error().Msg("content read failed:" + err.Error())
				return nil, 0, err
			}
			break
		}
		if n > 0 {
			segNum++
			// get erasure encode bytes
			encodeShards, err := redundancy.EncodeRawSegment(seg[:n], dataShards, parityShards)
			if err != nil {
				return nil, 0, err
			}
			for spID := 0; spID < ecShards; spID++ {
				encodeData[spID] = append(encodeData[spID], encodeShards[spID])
				if spID == 0 {
					fmt.Println("piece hash info1", encodeData[spID])
				}
			}
		}
	}

	for spID := 0; spID < parityShards+dataShards; spID++ {
		for _, piecedata := range encodeData[spID] {
			if spID == 0 {
				fmt.Println("piece hash info2", piecedata)
			}
		}
	}

	return nil, 0, nil
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
