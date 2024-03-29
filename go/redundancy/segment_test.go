package redundancy

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"
)

func TestSegmentPieceEncode(t *testing.T) {
	segmentSize := 16*1024*1024 - 2
	// generate encode source data
	segmentData := initSegmentData(segmentSize)
	const ObjectID string = "testabc"
	log.Printf("origin data size :%d", len(segmentData))
	segment := NewSegment(int64(segmentSize), segmentData, 1, ObjectID)

	piecesObjects, err := EncodeSegment(segment)
	if err != nil {
		t.Errorf("segment encode failed")
	}

	log.Print("encode result:")
	for _, piece := range piecesObjects {
		log.Printf("piece Object name: %s, index:%d \n", piece.Key, piece.ECIndex)
	}

	// set 2 dataBlocks as empty, decode should success
	shardsToRecover := make([]*PieceObject, 6)

	shardsToRecover[0] = piecesObjects[0]
	shardsToRecover[1] = piecesObjects[1]
	shardsToRecover[2] = &PieceObject{}
	shardsToRecover[3] = &PieceObject{}
	shardsToRecover[4] = piecesObjects[4] // priority block
	shardsToRecover[5] = piecesObjects[5] // priority block

	start := time.Now()
	decodeSegment, err := DecodeSegment(shardsToRecover, int64(segmentSize))
	if err != nil {
		t.Errorf("segment Reconstruct failed")
	}

	fmt.Printf("decode cost time: %d", time.Since(start).Milliseconds())
	if !bytes.Equal(decodeSegment.Data, segmentData) {
		t.Errorf("compare segment data failed")
	}
	if decodeSegment.SegmentID != segment.SegmentID {
		t.Errorf("compare segment id failed")
	}
	if decodeSegment.SegmentName != segment.SegmentName {
		t.Errorf("compare segment name failed")
	}

	// set 1 data block and 1 priority block as empty, decode should success
	shardsToRecover[0] = piecesObjects[0]
	shardsToRecover[1] = &PieceObject{}
	shardsToRecover[2] = piecesObjects[2]
	shardsToRecover[3] = piecesObjects[3]
	shardsToRecover[4] = &PieceObject{}   // priority block
	shardsToRecover[5] = piecesObjects[5] // priority block

	decodeSegment, err = DecodeSegment(shardsToRecover, int64(segmentSize))
	if err != nil {
		t.Errorf("segment Reconstruct failed")
	}
	if !bytes.Equal(decodeSegment.Data, segmentData) {
		t.Errorf("compare fail")
	}
	if decodeSegment.SegmentID != segment.SegmentID {
		t.Errorf("compare segment id failed")
	}
	if decodeSegment.SegmentName != segment.SegmentName {
		t.Errorf("compare segment name failed")
	}

	// set 2 data block and 1 priority block as empty, decode should fail
	shardsToRecover[0] = piecesObjects[0]
	shardsToRecover[1] = &PieceObject{}
	shardsToRecover[2] = &PieceObject{}
	shardsToRecover[3] = piecesObjects[3]
	shardsToRecover[4] = &PieceObject{}   // priority block
	shardsToRecover[5] = piecesObjects[5] // priority block

	_, err = DecodeSegment(shardsToRecover, int64(segmentSize))
	if err == nil {
		t.Errorf("segment decode should fail")
	}
}

func TestRawSegmentEncode(t *testing.T) {
	segmentSize := 16*1024*1024 - 2
	segmentData := initSegmentData(segmentSize)

	piecesShards, err := EncodeRawSegment(segmentData, DataBlocks, ParityBlocks)
	if err != nil {
		t.Errorf("segment encode failed")
	}

	// set 2 dataBlock of origin as empty block
	shardsToRecover := make([][]byte, 6)
	shardsToRecover[0] = piecesShards[0]
	shardsToRecover[1] = piecesShards[1]
	shardsToRecover[2] = []byte("")
	shardsToRecover[3] = []byte("")
	shardsToRecover[4] = piecesShards[4]
	shardsToRecover[5] = piecesShards[5]

	deCodeBytes, err := DecodeRawSegment(shardsToRecover, int64(segmentSize), DataBlocks, ParityBlocks)
	if err != nil {
		t.Errorf("decode failed")
	} else {
		log.Println("decode successfully")
	}

	// compare decode data with original data
	if !bytes.Equal(deCodeBytes, segmentData) {
		t.Errorf("decode data failed")
	}

	// set 2 data block and 1 priority block as empty, decode should fail
	shardsToRecover[2] = []byte("")
	shardsToRecover[3] = []byte("")
	shardsToRecover[4] = []byte("")

	_, err = DecodeRawSegment(shardsToRecover, int64(segmentSize), DataBlocks, ParityBlocks)
	if err == nil {
		t.Errorf("segment decode should fail")
	}
}

func initSegmentData(segmentSize int) []byte {
	// generate encode source data
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	segmentData := make([]byte, segmentSize)
	for i := range segmentData {
		segmentData[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return segmentData
}
