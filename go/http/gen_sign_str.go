package http

import (
	"bytes"
	"net/http"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

var supportHeads = []string{
	HTTPHeaderContentSHA256, HTTPHeaderTransactionHash, HTTPHeaderObjectID, HTTPHeaderRedundancyIndex, HTTPHeaderResource,
	HTTPHeaderDate, HTTPHeaderRange, HTTPHeaderPieceIndex, HTTPHeaderContentType, HTTPHeaderContentMD5, HTTPHeaderUnsignedMsg, HTTPHeaderUserAddress,
	HTTPHeaderExpiryTimestamp,
}

// getCanonicalHeaders generate a list of request headers with their values
func getCanonicalHeaders(req *http.Request, supportHeaders map[string]struct{}) string {
	var content bytes.Buffer
	var containHostHeader bool
	sortHeaders := getSortedHeaders(req, supportHeaders)
	headerMap := make(map[string][]string)
	for key, data := range req.Header {
		headerMap[strings.ToLower(key)] = data
	}

	for _, header := range sortHeaders {
		content.WriteString(strings.ToLower(header))
		content.WriteByte(':')

		if header != "host" {
			for i, v := range headerMap[header] {
				if i > 0 {
					content.WriteByte(',')
				}
				trimVal := strings.Join(strings.Fields(v), " ")
				content.WriteString(trimVal)
			}
			content.WriteByte('\n')
		} else {
			containHostHeader = true
			content.WriteString(GetHostInfo(req))
			content.WriteByte('\n')
		}
	}

	if !containHostHeader {
		content.WriteString(GetHostInfo(req))
		content.WriteByte('\n')
	}
	return content.String()
}

// getSignedHeaders return the sorted header array
func getSortedHeaders(req *http.Request, supportMap map[string]struct{}) []string {
	var signHeaders []string
	for k := range req.Header {
		if _, ok := supportMap[k]; ok {
			signHeaders = append(signHeaders, strings.ToLower(k))
		}
	}
	sort.Strings(signHeaders)
	return signHeaders
}

// getSignedHeaders return the alphabetically sorted, semicolon-separated list of lowercase request header names.
func getSignedHeaders(req *http.Request, supportHeaders map[string]struct{}) string {
	return strings.Join(getSortedHeaders(req, supportHeaders), ";")
}

// GetCanonicalRequest generate the canonicalRequest base on aws s3 sign without payload hash. t
func GetCanonicalRequest(req *http.Request) string {
	supportHeaders := initSupportHeaders()
	req.URL.RawQuery = strings.ReplaceAll(req.URL.Query().Encode(), "+", "%20")
	canonicalRequest := strings.Join([]string{
		req.Method,
		EncodePath(req.URL.Path),
		req.URL.RawQuery,
		getCanonicalHeaders(req, supportHeaders),
		getSignedHeaders(req, supportHeaders),
	}, "\n")
	return canonicalRequest
}

// GetMsgToSign generate the msg bytes from canonicalRequest to sign
func GetMsgToSign(req *http.Request) []byte {
	return crypto.Keccak256([]byte(GetCanonicalRequest(req)))
}

// GetMsgToSignForPreSignedURL is only used in SP get Object API.  This util method can be used in by SP side and client side to construct the MsgToSign
func GetMsgToSignForPreSignedURL(req *http.Request) []byte {
	queryValues := req.URL.Query()
	queryValues.Del(HTTPHeaderAuthorization)
	req.URL.RawQuery = queryValues.Encode()
	return GetMsgToSign(req)
}

func initSupportHeaders() map[string]struct{} {
	supportMap := make(map[string]struct{})
	for _, header := range supportHeads {
		emptyStruct := new(struct{})
		supportMap[header] = *emptyStruct
	}
	return supportMap
}
