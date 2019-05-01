package main

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"log"
	"time"
)

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
	defer client.Shutdown()

	for {
		asyncChainResult := client.GetBlockChainInfoAsync()
		// get pending transaction list at this timestamp
		mempoolList, err := client.GetRawMempool()
		if err != nil {
			panic(err)
		}
		chain, err := asyncChainResult.Receive()
		if err != nil {
			panic(err)
		}

		// record utxo dependency firstly because transaction list in mempool
		// maybe disorder for dependency relationship.
		deps := make(map[chainhash.Hash]struct{})
		for _, txHash := range  mempoolList {
			deps[*txHash] = struct{}{}
		}

		depsCount := 0
		depTx := 0
		inputs := 0
		flag := false
		for _, txHash := range mempoolList {
			if err != nil {
				panic(err)
			}
			tx, err := client.GetRawTransaction(txHash)
			if err != nil {
				fmt.Println(err, txHash)
				continue
			}

			inputs += len(tx.MsgTx().TxIn)
			for _, input := range tx.MsgTx().TxIn {
				if _, ok := deps[input.PreviousOutPoint.Hash]; ok {
					depsCount++

					if !flag {
						depTx++

						// only once
						flag = true
					}
				}
			}

			flag = false
		}

		fmt.Printf("%s: block: %d,  deps transaction: %d, total transactions: %d, rate: %.2f; %d spend entry in mempool, total inputs: %d, rate: %2f\n",
			time.Now().Format("2006-01-02 15:04:05"), chain.Blocks, depTx, len(mempoolList), float64(depTx) / float64(len(mempoolList)),
			depsCount, inputs, float64(depsCount) / float64(inputs))

		time.Sleep(20 * time.Minute)
	}
}
