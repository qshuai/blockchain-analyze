package main

import (
	"encoding/binary"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"os"
)

func main() {
	txindexDB ,err := leveldb.OpenFile(os.Args[1], &opt.Options{
		Compression:opt.SnappyCompression,
	})
	if err != nil {
		panic(err)
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         "127.0.0.1:9332",
		User:         "okcoin",
		Pass:         "lZWMxOThhODEyNDQyYTg0NjY",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		panic(err)
	}
	defer client.Shutdown()

	batch := &leveldb.Batch{}
	for i := 0; i < 566970; i++ {
		blockHash, err := client.GetBlockHash(int64(i))
		if err != nil {
			panic(err)
		}

		ret ,err := client.GetBlockVerbose(blockHash)
		if err != nil {
			panic(err)
		}

		for _, item := range ret.Tx {
			txHash, err := chainhash.NewHashFromStr(item)
			if err != nil {
				panic(err)
			}

			height := make([]byte, 4)
			binary.LittleEndian.PutUint32(height, uint32(i))
			batch.Put(txHash[:6], height)
		}

		if batch.Len() > 100000 {
			err = txindexDB.Write(batch, nil)
			if err != nil {
				panic(err)
			}

			batch = &leveldb.Batch{}
		}

		if i != 0 && i % 10000 == 0 {
			fmt.Printf("Handle block: %d\n", i)
		}
	}

}
