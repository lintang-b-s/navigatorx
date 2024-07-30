package filesystemstorage

// func SaveToDisk(ch alg.ContractedGraph, filename string) error {
// 	buf := new(bytes.Buffer)
// 	enc := gob.NewEncoder(buf)
// 	enc.Encode(ch)

// 	f, err := os.Create(filename)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()
// 	var bb []byte
// 	bb, err = zstd.Compress(bb, buf.Bytes())
// 	f.Write(bb)

// 	return nil
// }
