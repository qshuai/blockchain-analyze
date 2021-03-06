package main

import (
	"fmt"
	"log"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

var max int
var nulldata int

func main() {
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
		log.Fatal(err)
	}

	blockChan := make(chan *wire.MsgBlock, 200)
	go func() {
		for i := 230001; i < 567297; i++ {
			if i%1000 == 0 {
				fmt.Printf("%s Handle block height: %d, channel size: %d, max length: %d, nulldata: %d\n",
					time.Now().String(), i, len(blockChan), max, nulldata)
			}

			blockhash, err := client.GetBlockHash(int64(i))
			if err != err {
				panic(err)
			}

			block, err := client.GetBlock(blockhash)
			if err != nil {
				panic(err)
			}

			blockChan <- block
		}

		close(blockChan)
	}()

	for block := range blockChan {
		for _, tx := range block.Transactions {
			for idx, out := range tx.TxOut {
				// null data
				if len(out.PkScript) == 0 {
					fmt.Printf("transaction output without pkScript, %s:%d\n", tx.TxHash(), idx)
					continue
				}

				if out.PkScript[0] == 0x6a {
					nulldata++
					continue
				}

				class, _, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
				if class != txscript.PubKeyHashTy &&
					class != txscript.WitnessV0PubKeyHashTy &&
					class != txscript.PubKeyTy &&
					class != txscript.ScriptHashTy &&
					class != txscript.WitnessV0ScriptHashTy &&
					class != txscript.MultiSigTy {

					continue
				}

				if len(out.PkScript) > max {
					fmt.Printf("transaction: %s:%d\n", tx.TxHash(), idx)
					max = len(out.PkScript)
				}
			}
		}
	}

	client.Shutdown()
	fmt.Printf("most length in pkScript(byte): %d; nulldata output: %d", max, nulldata)
}