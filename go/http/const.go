package http

const (
	HTTPHeaderContentType     = "Content-Type"
	HTTPHeaderTransactionHash = "X-Gnfd-Txn-Hash"
	HTTPHeaderResource        = "X-Gnfd-Resource"
	HTTPHeaderPieceIndex      = "X-Gnfd-Piece-Index"
	HTTPHeaderObjectID        = "X-Gnfd-Object-ID"
	HTTPHeaderRedundancyIndex = "X-Gnfd-Redundancy-Index"
	HTTPHeaderUnsignedMsg     = "X-Gnfd-Unsigned-Msg"

	HTTPHeaderContentMD5    = "Content-MD5"
	HTTPHeaderRange         = "Range"
	HTTPHeaderContentSHA256 = "X-Gnfd-Content-Sha256"

	HTTPHeaderUserAddress = "X-Gnfd-User-Address"
	// HTTPHeaderDate The date and time format must follow the ISO 8601 standard, and must be formatted with the "yyyyMMddTHHmmssZ" format. For example if the date and time was "08/01/2016 15:32:41.982-700" then it must first be converted to UTC (Coordinated Universal Time) and then submitted as "20160801T223241Z".
	HTTPHeaderDate = "X-Gnfd-Date"
	// HTTPHeaderExpiryTimestamp defines the expiry timestamp, which is the ISO 8601 datetime string (e.g. 2021-09-30T16:25:24Z), and the maximum Timestamp since the request sent must be less than MaxExpiryAgeInSec (seven days).
	HTTPHeaderExpiryTimestamp = "X-Gnfd-Expiry-Timestamp"
	HTTPHeaderAuthorization   = "Authorization"
	// MaxExpiryAgeInSec
	MaxExpiryAgeInSec = 3600 * 24 * 7 // 7 days
)
