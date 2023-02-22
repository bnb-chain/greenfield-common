package redundancy

import (
	"strconv"
	"strings"

	"github.com/bnb-chain/greenfield-common/go/redundancy/erasure"
	"github.com/rs/zerolog/log"
)

// PieceObject - details of the erasure encoded piece
type PieceObject struct {
	Key       string
	ECdata    []byte
	ECIndex   int
	PieceSize int
}

// Segment - detail of segment split from objects
type Segment struct {
	SegName     string
	SegmentSize int64
	SegmentID   int
	Data        []byte
}

const (
	DataBlocks   int = 4
	ParityBlocks int = 2
)

var defaultEcConfig = ECConfig{
	dataBlocks:   DataBlocks,
	parityBlocks: ParityBlocks,
}

// NewSegment creates a new Segment object
func NewSegment(size int64, content []byte, segmentId int, objectId string) *Segment {
	return &Segment{
		SegName:     objectId + "_s" + strconv.Itoa(segmentId),
		SegmentSize: size,
		SegmentID:   segmentId,
		Data:        content,
	}
}

// EncodeSegment encode to segment, return the piece content and the meta of pieces
func EncodeSegment(s *Segment) ([]*PieceObject, error) {
	erasure, err := erasure.NewRSEncoder(defaultEcConfig.dataBlocks, defaultEcConfig.parityBlocks, s.SegmentSize)
	if err != nil {
		log.Error().Msg("new RSEncoder fail" + err.Error())
		return nil, err
	}
	shards, err := erasure.EncodeData(s.Data)
	if err != nil {
		log.Error().Msg("encode data fail :" + err.Error() + "segment name:" + s.SegName)
		return nil, err
	}

	pieceObjectList := make([]*PieceObject, DataBlocks+ParityBlocks)
	for index, shard := range shards {
		piece := &PieceObject{
			Key:       s.SegName + "_p" + strconv.Itoa(index),
			ECdata:    shard,
			ECIndex:   index,
			PieceSize: len(shard),
		}
		pieceObjectList[index] = piece
	}

	return pieceObjectList, nil
}

// DecodeSegment decode with the pieceObjects and reconstruct the original segment
func DecodeSegment(pieces []*PieceObject, segmentSize int64) (*Segment, error) {
	encoder, err := erasure.NewRSEncoder(defaultEcConfig.dataBlocks, defaultEcConfig.parityBlocks, segmentSize)
	if err != nil {
		log.Error().Msg("new RSEncoder fail" + err.Error())
		return nil, err
	}

	pieceObjectsData := make([][]byte, DataBlocks+ParityBlocks)
	for i := 0; i < DataBlocks+ParityBlocks; i++ {
		pieceObjectsData[i] = pieces[i].ECdata
	}

	deCodeBytes, err := encoder.GetOriginalData(pieceObjectsData, segmentSize)
	if err != nil {
		log.Error().Msg("reconstruct segment content fail:" + err.Error())
		return nil, err
	}

	// construct the segmentId and segmentName from piece key
	pieceName := pieces[0].Key
	segIndex := strings.Index(pieceName, "_s")
	ecIndex := strings.Index(pieceName, "_p")

	segIdStr := pieceName[segIndex+2 : ecIndex]
	segId, err := strconv.Atoi(segIdStr)
	if err != nil {
		log.Error().Msg("fetch segment ID fail: " + err.Error())
		return nil, err
	}

	return &Segment{
		SegmentSize: segmentSize,
		SegName:     pieceName[:ecIndex],
		SegmentID:   segId,
		Data:        deCodeBytes,
	}, nil
}

// EncodeRawSegment encode a raw byte array and return erasure encoded shards in orders
func EncodeRawSegment(content []byte, dataShards, parityShards int) ([][]byte, error) {
	erasure, err := erasure.NewRSEncoder(dataShards, parityShards, int64(len(content)))
	if err != nil {
		log.Error().Msg("new RSEncoder fail:" + err.Error())
		return nil, err
	}
	shards, err := erasure.EncodeData(content)
	if err != nil {
		return nil, err
	}
	return shards, nil
}

// DecodeRawSegment decode the erasure encoded data and return original content
// If the piece data has lost, need to pass an empty bytes array as one piece
func DecodeRawSegment(piecesData [][]byte, segmentSize int64, dataShards, parityShards int) ([]byte, error) {
	encoder, err := erasure.NewRSEncoder(dataShards, parityShards, segmentSize)
	if err != nil {
		log.Error().Msg("new RSEncoder fail:" + err.Error())
		return nil, err
	}

	deCodeBytes, err := encoder.GetOriginalData(piecesData, segmentSize)
	if err != nil {
		return nil, err
	}
	return deCodeBytes, nil
}
