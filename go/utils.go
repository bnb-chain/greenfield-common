package redundancy

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"
)

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

// EncodePath encode the strings from UTF-8 byte representations to HTML hex escape sequences
func EncodePath(pathName string) string {
	reservedNames := regexp.MustCompile("^[a-zA-Z0-9-_.~/]+$")
	// no need to encode
	if reservedNames.MatchString(pathName) {
		return pathName
	}
	var encodedPathName strings.Builder
	for _, s := range pathName {
		if 'A' <= s && s <= 'Z' || 'a' <= s && s <= 'z' || '0' <= s && s <= '9' { // §2.3 Unreserved characters (mark)
			encodedPathName.WriteRune(s)
			continue
		}
		switch s {
		case '-', '_', '.', '~', '/':
			encodedPathName.WriteRune(s)
			continue
		default:
			length := utf8.RuneLen(s)
			if length < 0 {
				// if utf8 cannot convert return the same string as is
				return pathName
			}
			u := make([]byte, length)
			utf8.EncodeRune(u, s)
			for _, r := range u {
				hexStr := hex.EncodeToString([]byte{r})
				encodedPathName.WriteString("%" + strings.ToUpper(hexStr))
			}
		}
	}
	return encodedPathName.String()
}

// GetHostInfo returns host header from the request
func GetHostInfo(req *http.Request) string {
	host := req.Header.Get("host")
	if host != "" {
		return host
	}
	if req.Host != "" {
		return req.Host
	}
	return req.URL.Host
}
