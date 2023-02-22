package http

import (
	"encoding/hex"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"
)

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
