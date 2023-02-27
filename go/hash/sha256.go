package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
)

// CalcSHA256Hex compute checksum of sha256 hash and encode it to hex
func CalcSHA256Hex(buf []byte) string {
	sum := CalcSHA256(buf)
	return hex.EncodeToString(sum)
}

// CalcSHA256 compute checksum of sha256 from byte array
func CalcSHA256(buf []byte) []byte {
	h := sha256.New()
	h.Write(buf)
	sum := h.Sum(nil)
	return sum
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
