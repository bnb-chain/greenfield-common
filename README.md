# greenfield-common
Support common libs for different repos of greenfield

## Supported Common Functions

### 1. Erasure encode/decode algorithm 

(1) erasure package support RSEncoder which contain basic Encode and Decode reedSolomon APIs 
```
RSEncoderStorage, err := NewRSEncoder(dataShards, parityShards, int64(blockSize))
// encode data and return the encoded shard number
func (r *RSEncoder) EncodeData(content []byte) ([][]byte, error) 
// decodes the input erasure encoded data shards data.
func (r *RSEncoder) DecodeDataShards(content [][]byte) error {
```
(2) redundancy package support methods to encode/decode segments data using RSEncoder 
```
// encode segment
func EncodeRawSegment(content []byte, dataShards, parityShards int) ([][]byte, error) 

// decode segment
func DecodeRawSegment(pieceData [][]byte, segmentSize int64, dataShards, parityShards int) ([]byte, error) 
```

### 2. Compute sha256 hash of file content

hash package support methods to compute hash roots of greenfield objects , the computed methods is based on 
redundancy Strategy of greenfield

```
// compute hash roots fromm io reader, the parameters should fetch from chain besides reader
func ComputerHash(reader io.Reader, segmentSize int64, dataShards, parityShards int) ([]string, int64, error)

// compute hash roots based on file path
func ComputerHashFromFile(filePath string, segmentSize int64, dataShards, parityShards int) ([]string, int64, error)
```


### 3. Generate checksum and integrity hash
