package hash

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

type SegmentInfo struct {
	SegmentID int
	Data      []byte
}

// GenerateChecksum generates the checksum of one piece data
func GenerateChecksum(pieceData []byte) []byte {
	hash := sha256.New()
	hash.Write(pieceData)
	return hash.Sum(nil)
}

// GenerateIntegrityHash generates integrity hash of all piece data checksum
func GenerateIntegrityHash(checksumList [][]byte) []byte {
	hash := sha256.New()
	checksumBytesTotal := bytes.Join(checksumList, []byte(""))
	hash.Write(checksumBytesTotal)
	return hash.Sum(nil)
}

// ChallengePieceHash challenge integrity hash and checksum list
// integrityHash represents the integrity hash of one piece list, this piece list may be ec piece data list or
// segment piece data list; if piece data list is ec, this list is all ec1 piece data; if piece list is segment, all
// piece data is all segments of an object
// checksumList is composed of  one piece list's individual checksum
// index represents which piece is used to challenge
// pieceData represents piece physical data that user want to challenge
func ChallengePieceHash(integrityHash []byte, checksumList [][]byte, index int, pieceData []byte) error {
	if len(checksumList) <= index {
		return fmt.Errorf("invalid checksum list")
	}
	if !bytes.Equal(checksumList[index], GenerateChecksum(pieceData)) {
		return fmt.Errorf("piece data and piece hash are inconsistent")
	}
	if err := VerifyIntegrityHash(integrityHash, checksumList); err != nil {
		return err
	}
	return nil
}

// VerifyIntegrityHash verify integrity hash if right
func VerifyIntegrityHash(integrityHash []byte, checksumList [][]byte) error {
	if !bytes.Equal(integrityHash, GenerateIntegrityHash(checksumList)) {
		return fmt.Errorf("invalid integrity hash")
	}
	return nil
}
