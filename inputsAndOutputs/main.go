package main

import "fmt"

func main() {
	//height ,err := strconv.Atoi(os.Args[1])
	//if err != nil {
	//	panic(err)
	//}
	//
	//bc ,err := bitcoind.New("127.0.0.1", 8332, "KAuCgqk0gwgP9LWtDnu", "EGQFjJu81Ck3j7lFvU8cPW2jALopF", false)
	//if err != nil {
	//	panic(err)
	//}
	//
	//blockHash ,err := bc.GetBlockHash(uint64(height))
	//if err != nil {
	//	fmt.Printf("get block hash failed: %s", err)
	//	return
	//}
	//
	//rawBlock ,err := bc.GetRawBlock(blockHash)
	//if err != nil {
	//	fmt.Printf("get block failed: %s", err)
	//	return
	//}
	//var block wire.MsgBlock
	//blockBytes, err := hex.DecodeString(rawBlock)
	//if err != nil {
	//	fmt.Printf("decode block failed: %s\n", err)
	//	return
	//}
	//err = block.Deserialize(bytes.NewReader(blockBytes))
	//if err != nil {
	//	fmt.Printf("deserialize block failed: %s\n", err)
	//	return
	//}
	//
	//if err != nil {
	//	fmt.Printf("get block failed: %s\n", err)
	//	return
	//}
	//
	//txes := block.Transactions
	//inputs := 0
	//outputs := 0
	//for _, tx := range txes {
	//	inputs += len(tx.TxIn)
	//	outputs += len(tx.TxOut)
	//}
	//
	//fmt.Println(inputs, outputs, inputs + outputs)

	for i:= 0; i< 63; i++ {
		fmt.Printf("delete from btc_address_%d;\n", i)
	}
}
