package http

import (
	"bytes"
	"net/http"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	utils "github.com/bnb-chain/greenfield-common/go"
)

var supportHeads = []string {
	HTTPHeaderContentSHA256, HTTPHeaderTransactionHash, HTTPHeaderObjectId, HTTPHeaderSPAddr, HTTPHeaderResource,
	HTTPHeaderDate, HTTPHeaderRange, HTTPHeaderPieceIndex, HTTPHeaderContentType, HTTPHeaderContentMD5
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
			content.WriteString(utils.GetHostInfo(req))
			content.WriteByte('\n')
		}
	}

	if !containHostHeader {
		content.WriteString(utils.GetHostInfo(req))
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
func GetCanonicalRequest(req *http.Request, supportHeaders map[string]struct{}) string {
	req.URL.RawQuery = strings.ReplaceAll(req.URL.Query().Encode(), "+", "%20")
	canonicalRequest := strings.Join([]string{
		req.Method,
		utils.EncodePath(req.URL.Path),
		req.URL.RawQuery,
		getCanonicalHeaders(req, supportHeaders),
		getSignedHeaders(req, supportHeaders),
	}, "\n")

	return canonicalRequest
}

// GetMsgToSign generate the msg bytes from canonicalRequest to sign
func GetMsgToSign(req *http.Request) []byte {
	headers := initSupportHeaders()
	signBytes := utils.CalcSHA256([]byte(GetCanonicalRequest(req, headers)))
	return crypto.Keccak256(signBytes)
}


func initSupportHeaders() map[string]struct{} {
	supportMap := make(map[string]struct{})
	for _, header := range supportHeads {
		emptyStruct := new(struct{})
		supportMap[header] = *emptyStruct
	}
	return supportMap
}

