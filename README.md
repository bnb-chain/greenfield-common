# greenfield-common
Support common libs for different repos of greenfield

## Supported Common Functions

### 1. Erasure encode/decode algorithm 

(1) erasure package support RSEncoder which contain basic Encode and Decode reedSolomon APIs 
```
// first step, create a new rs encoder, the blockSize indicate the data size to be encoded
func NewRSEncoder(dataShards, parityShards int, blockSize int64) (r RSEncoder, err error) {
// encode data and return the encoded shard number
func (r *RSEncoder) EncodeData(content []byte) ([][]byte, error) 
// decode the input data and reconstruct the data shards data (not include the parity shards).
func (r *RSEncoder) DecodeDataShards(content [][]byte) error 
// decode the input data and reconstruct the data shards and parity Shards
func (r *RSEncoder) DecodeShards(data [][]byte) error
```
(2) redundancy package support methods to encode/decode segments data using RSEncoder 
```
// encode one segment 
func EncodeRawSegment(content []byte, dataShards, parityShards int) ([][]byte, error) 

// decode the segment and reconstruct the original segment content
func DecodeRawSegment(pieceData [][]byte, segmentSize int64, dataShards, parityShards int) ([]byte, error) 
```

### 2. Compute sha256 hash of file content

hash package support methods to compute hash roots of greenfield objects , the computed methods is based on 
redundancy strategy of greenfield

```
// compute hash roots fromm io reader, the parameters should fetch from chain besides reader
func ComputeIntegrityHash(reader io.Reader, segmentSize int64, dataShards, parityShards int) ([]string, int64, error)

// compute hash roots based on file path
func ComputerHashFromFile(filePath string, segmentSize int64, dataShards, parityShards int) ([]string, int64, error)
```


### 3. Generate checksum and integrity hash
