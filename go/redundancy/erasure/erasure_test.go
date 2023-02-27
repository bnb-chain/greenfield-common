package erasure

import (
	"bytes"
	"log"
	"math/rand"
	"testing"
)

const (
	dataShards   = 4
	parityShards = 2
)

func TestRSEncoder(t *testing.T) {
	blockSize := 16 * 1024 * 1024
	RSEncoderStorage, err := NewRSEncoder(dataShards, parityShards, int64(blockSize))
	if err != nil {
		t.Errorf("new RSEncoder failed")
	}

	// generate encode source data
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	originData := make([]byte, blockSize)
	for i := range originData {
		originData[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	shards, err := RSEncoderStorage.EncodeData(originData)
	if err != nil {
		t.Errorf("encode failed")
	}
	log.Println("encode successfully")

	// set 2 dataBlock of origin as empty block
	shardsToRecover := make([][]byte, 6)
	shardsToRecover[0] = shards[0]
	shardsToRecover[1] = shards[1]
	shardsToRecover[2] = []byte("")
	shardsToRecover[3] = []byte("")
	shardsToRecover[4] = shards[4]
	shardsToRecover[5] = shards[5]

	err = RSEncoderStorage.DecodeDataShards(shardsToRecover)
	if err != nil {
		t.Errorf("decode failed")
	} else {
		log.Println("decode successfully")
	}

	// compare data Blocks
	var buffer bytes.Buffer
	for i := 0; i < 4; i++ {
		buffer.Write(shardsToRecover[i])
	}

	shardSize := RSEncoderStorage.ShardSize()
	deCodeBytes := buffer.Bytes()
	// ignore padding content
	if int(shardSize*4) >= len(originData) {
		deCodeBytes = deCodeBytes[:len(originData)]
	}

	if !bytes.Equal(deCodeBytes, originData) {
		t.Errorf("decode data failed")
	}

	// delete 2 dataBlock of origin
	shardsToRecover[0] = nil
	shardsToRecover[1] = nil

	deCodeContent, err := RSEncoderStorage.GetOriginalData(shardsToRecover, int64(len(originData)))
	if err != nil {
		t.Errorf("decode failed")
	}

	if !bytes.Equal(deCodeContent, originData) {
		t.Errorf("decode data failed")
	}

	// set 2 priorityBlock of origin as empty block
	shardsToRecover[4] = nil
	shardsToRecover[5] = nil
	deCodeContent, err = RSEncoderStorage.GetOriginalData(shardsToRecover, int64(len(originData)))
	if err != nil {
		t.Errorf("decode failed")
	}

	if !bytes.Equal(deCodeContent, originData) {
		t.Errorf("decode data failed")
	}

	// set 3 dataBlock of origin as empty block, decode should be fail
	shardsToRecover[0] = nil
	shardsToRecover[1] = nil
	shardsToRecover[2] = nil
	err = RSEncoderStorage.DecodeDataShards(shardsToRecover)
	if err == nil {
		t.Errorf("decode should failed")
	}

}
