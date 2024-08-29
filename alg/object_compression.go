package alg

func CompressWay(sw []SurakartaWay) ([]byte, error) {
	bb := Encode(sw)
	// bbCompressed, err := Compress(bb)
	// if err != nil {
	// 	return []byte{}, err
	// }
	bbCompressed := bb

	return bbCompressed, nil
}
func LoadWay(bbCompressed []byte) ([]SurakartaWay, error) {
	var sw []SurakartaWay
	// bb, err := Decompress(bbCompressed)
	// if err != nil {
	// 	return []SurakartaWay{}, err
	// }
	sw, err := Decode(bbCompressed)

	return sw, err
}
