package redundancy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"sync"
)

// SplitAndComputerHash split the reader into segment, ec encode the data, compute the hash roots of pieces,
// and return the hash result array list and data size
func SplitAndComputerHash(reader io.Reader, segmentSize int64, ecShards int) ([]string, int64, error) {
	var segChecksumList [][]byte
	var result []string
	encodeData := make([][][]byte, ecShards)
	seg := make([]byte, segmentSize)

	contentLen := int64(0)
	// read the data by segment size
	for {
		n, err := reader.Read(seg)
		if err != nil {
			if err != io.EOF {
				log.Println("content read error:", err)
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
					log.Println("compute checksum err:", err)
					return nil, 0, err
				}
				segChecksumList = append(segChecksumList, checksum)
			}

			// get erasure encode bytes
			encodeShards, err := EncodeRawSegment(seg[:n])

			if err != nil {
				log.Println("erasure encode err:", err)
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

func SplitAndComputerHash2(reader io.Reader, segmentSize int64, ecShards int) ([]string, int64, error) {
	var segChecksumList [][]byte
	var result []string
	encodeData := make([][][]byte, ecShards)
	seg := make([]byte, segmentSize)

	contentLen := int64(0)
	// read the data by segment size
	for {
		n, err := reader.Read(seg)
		if err != nil {
			if err != io.EOF {
				log.Println("content read error:", err)
				return nil, 0, err
			}
			break
		}
		if n > 0 {
			contentLen += int64(n)
			// compute segment hash
			checksum := CalcSHA256(seg[:n])
			if err != nil {
				log.Println("compute checksum err:", err)
				return nil, 0, err
			}
			segChecksumList = append(segChecksumList, checksum)

			// get erasure encode bytes
			encodeShards, err := EncodeRawSegment(seg[:n])

			if err != nil {
				log.Println("erasure encode err:", err)
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
func ComputerHashFromFile(filePath string, segmentSize int64, ecShards int) ([]string, int64, error) {
	fReader, err := os.Open(filePath)
	// If any error fail quickly here.
	if err != nil {
		return nil, 0, err
	}
	defer fReader.Close()

	return SplitAndComputerHash(fReader, segmentSize, ecShards)
}

func ComputerHashFromFile2(filePath string, segmentSize int64, ecShards int) ([]string, int64, error) {
	fReader, err := os.Open(filePath)
	// If any error fail quickly here.
	if err != nil {
		return nil, 0, err
	}
	defer fReader.Close()

	return SplitAndComputerHash2(fReader, segmentSize, ecShards)
}

// CalcSHA256Hex compute checksum of sha256 hash and encode it to hex
func CalcSHA256Hex(buf []byte) (hexStr string) {
	sum := CalcSHA256(buf)
	hexStr = hex.EncodeToString(sum)
	return
}

// CalcSHA256 compute checksum of sha256 from byte array
func CalcSHA256(buf []byte) []byte {
	h := sha256.New()
	h.Write(buf)
	sum := h.Sum(nil)
	return sum[:]
}
func CalcSHA256Hex2(buf []byte) (hexStr string) {
	sum := CalcSHA256(buf)
	hexStr = hex.EncodeToString(sum)
	return
}

// CalcSHA256HashByte compute checksum of sha256 from io.reader
func CalcSHA256HashByte(body io.Reader) ([]byte, error) {
	if body == nil {
		return []byte(""), errors.New("body empty")
	}
	buf := make([]byte, 1024)
	h := sha256.New()
	if _, err := io.CopyBuffer(h, body, buf); err != nil {
		return []byte(""), err
	}
	hash := h.Sum(nil)
	return hash, nil
}
