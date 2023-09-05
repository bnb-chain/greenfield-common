package redundancy

// RedundancyConfig redundancy config
type RedundancyConfig struct {
	BlockNumber uint64
	SegmentSize uint64
	ECCfg       ECConfig
}

// ECConfig contains data blocks and parity blocks
type ECConfig struct {
	dataBlocks   int
	parityBlocks int
}

var Redundancy map[int]RedundancyConfig

// Object describes an object
type Object struct {
	ObjectInfo *ObjectInfo
	ObjectData []byte
	// ObjectData ObjectPayloadReader
}

// ObjectInfo describes basic object info
type ObjectInfo struct {
	ID            uint64
	ObjectName    string
	ObjectSize    uint64
	RedundancyCfg RedundancyConfig
}
