package alg

import "fmt"

func CompressGraph(ch *[]CHNode) ([]byte, error) {
	bb := Encode(ch)
	bbCompressed, err := Compress(bb)
	if err != nil {
		return []byte{}, err
	}
	return bbCompressed, nil
}

func LoadCHGraph(bbCompressed []byte) ([]CHNode, error) {
	var ch []CHNode
	bb, err := Decompress(bbCompressed)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Decompressed size: %d\n", len(bb))
	ch, err = Decode(bb)

	return ch, err
}

// func CompressCHGraph(ch []alg.CHNode) ([]byte, error) {
// 	bb := Encode(ch)
// 	bbCompressed, err := Compress(bb.Bytes())
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	return bbCompressed, nil
// }

// func LoadCHGraph(bbCompressed []byte, ch *[]alg.CHNode) error {
// 	bb, err := Decompress(bbCompressed)
// 	if err != nil {
// 		return err
// 	}
// 	DecodeCHGraph(bb, ch)

// 	return err
// }
