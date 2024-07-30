package alg

import (
	"bytes"
	"encoding/gob"

	"github.com/DataDog/zstd"
)

func Encode(sw []SurakartaWay) []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(sw)
	return buf.Bytes()
}

func Decode(bb []byte) ([]SurakartaWay, error) {
	var ch []SurakartaWay
	dec := gob.NewDecoder(bytes.NewReader(bb))
	err := dec.Decode(&ch)
	return ch, err
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
