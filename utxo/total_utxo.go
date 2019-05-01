package main

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	db, err := leveldb.OpenFile("utxodb", nil)
	if err != nil {
		panic(err)
	}

	var totalAmount int64
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		var utxo Utxo
		err = utxo.decode(iter.Value())
		if err != nil {
			fmt.Println("decode utxo failed")
			break
		}

		totalAmount += utxo.amount
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("total utxo amount:", totalAmount)
}
