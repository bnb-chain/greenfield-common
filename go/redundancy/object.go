package redundancy

type RedundancyConfig struct {
	BlockNumber uint64
	SegmentSize uint64
	ECCfg       ECConfig
}

type ECConfig struct {
	dataBlocks   int
	parityBlocks int
}

var Redundancy map[int]RedundancyConfig

type Object struct {
	ObjectInfo *ObjectInfo
	ObjectData []byte
	// ObjectData ObjectPayloadReader
}

type ObjectInfo struct {
	ID            uint64
	ObjectName    string
	ObjectSize    uint64
	RedundancyCfg RedundancyConfig
}
