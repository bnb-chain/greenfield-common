# greenfield-common
Support common libs for different repos of greenfield


## Disclaimer
**The software and related documentation are under active development, all subject to potential future change without
notification and not ready for production use. The code and security audit have not been fully completed and not ready
for any bug bounty. We advise you to be careful and experiment on the network at your own risk. Stay safe out there.**

## Supported Common Functions

### 1. Erasure encode/decode algorithm 

- Erasure package support RSEncoder which contain basic Encode and Decode ReedSolomon APIs. Function as follows:

```go
// first step, create a new rs encoder, the blockSize indicate the data size to be encoded
func NewRSEncoder(dataShards, parityShards int, blockSize int64) (r RSEncoder, err error) 
// encode data and return the encoded shard number
func (r *RSEncoder) EncodeData(content []byte) ([][]byte, error) 
// decode the input data and reconstruct the data shards data (not include the parity shards).
func (r *RSEncoder) DecodeDataShards(content [][]byte) error 
// decode the input data and reconstruct the data shards and parity Shards
func (r *RSEncoder) DecodeShards(data [][]byte) error
```

- Redundancy package support methods to encode/decode segments data using RSEncoder. Function as follows:

```go
// encode one segment 
func EncodeRawSegment(content []byte, dataShards, parityShards int) ([][]byte, error) 

// decode the segment and reconstruct the original segment content
func DecodeRawSegment(pieceData [][]byte, segmentSize int64, dataShards, parityShards int) ([]byte, error) 
```

### 2. Compute integrity hash of file content

Hash package support methods to compute the integrity hash of greenfield objects , the computed methods is based on 
redundancy strategy of greenfield. Function as follows:

```go
// ComputeIntegrityHash compute the integrity hash of file, return the integrity hashes, redundancy type  and file size.
// the parameters of segment size, dataShards and parityShards should fetch from chain
func ComputeIntegrityHash(reader io.Reader, segmentSize int64, dataShards, parityShards int, isSerial bool) ([][]byte, int64,
storageTypes.RedundancyType, error)

// ComputerHashFromFile compute the integrity hash based on file path
func ComputerHashFromFile(filePath string, segmentSize int64, dataShards, parityShards int) ([]string, int64, error)

// IntegrityHasher is used to calculate integrityHash in a stream way. It contains Init, Append, and Finish functions.
IntegrityHasher := NewHasher(segmentSize, dataShards, parityShards)
IntegrityHasher.Init()

// append the data chunks to IntegrityHasher, the data size should be less than segment size
func (i *IntegrityHasher) Append(data []byte) error

// compute the result of the Integrity hashes
func (i *IntegrityHasher) Finish() ([][]byte, int64, storageTypes.RedundancyType, error) {
```

### 3. Generate checksum and integrity hash
Common library supports generating checksum and integrity hash. `GenerateChecksum` uses sha256 algorithm to compute hash. 
`GenerateIntegrityHash` is used to compute all checksum to get an integrity hash. Function as follows:

```go
// GenerateChecksum generates the checksum of one piece data
func GenerateIntegrityHash(checksumList [][]byte) []byte

// GenerateIntegrityHash generates integrity hash of all piece data checksum
func GenerateIntegrityHash(checksumList [][]byte) []byte

// ChallengePieceHash challenge integrity hash and checksum list
// integrityHash represents the integrity hash of one piece list, this piece list may be ec piece data list or
// segment piece data list; if piece data list is ec, this list is all ec1 piece data; if piece list is segment, all
// piece data is all segments of an object
// checksumList is composed of  one piece list's individual checksum
// index represents which piece is used to challenge
// pieceData represents piece physical data that user want to challenge
func ChallengePieceHash(integrityHash []byte, checksumList [][]byte, index int, pieceData []byte) error

// VerifyIntegrityHash verify integrity hash if right
func VerifyIntegrityHash(integrityHash []byte, checksumList [][]byte) error
```
