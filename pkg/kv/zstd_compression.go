package kv

import (
	// jsoniter "github.com/json-iterator/go"
	"lintang/navigatorx/pkg/concurrent"

	"github.com/DataDog/zstd"
	"github.com/kelindar/binary"
)

// yang dipakai di road snap cuma intersection node sama centerLoc
type SmallWay struct {
	CenterLoc           []float64 // [lat, lon]
	IntersectionNodesID []int64
}

func (s *SmallWay) toConcurrentWay() concurrent.SmallWay {
	return concurrent.SmallWay{
		CenterLoc:           s.CenterLoc,
		IntersectionNodesID: s.IntersectionNodesID,
	}	
}

func Encode(sw []SmallWay) []byte {
	encoded, _ := binary.Marshal(sw)
	return encoded
}

func Decode(bb []byte) ([]SmallWay, error) {
	var ch []SmallWay
	binary.Unmarshal(bb, &ch)
	return ch, nil
}

func Compress(bb []byte) ([]byte, error) {
	var bbCompressed []byte
	bbCompressed, err := zstd.Compress(bbCompressed, bb)
	if err != nil {
		return []byte{}, err
	}
	return bbCompressed, nil
}

func Decompress(bbCompressed []byte) ([]byte, error) {
	var bb []byte
	bb, err := zstd.Decompress(bb, bbCompressed)
	if err != nil {
		return []byte{}, err
	}

	return bb, nil
}
