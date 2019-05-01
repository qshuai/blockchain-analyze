package main

//
//import (
//	"encoding/binary"
//	"fmt"
//	"log"
//	"time"
//
//	"github.com/btcsuite/btcd/blockchain"
//	"github.com/btcsuite/btcd/chaincfg/chainhash"
//	"github.com/btcsuite/btcd/rpcclient"
//	"github.com/btcsuite/btcd/wire"
//	"github.com/syndtr/goleveldb/leveldb"
//	"github.com/syndtr/goleveldb/leveldb/opt"
//)
//
//type Block struct {
//	block  *wire.MsgBlock
//	height int
//}
//
//func main() {
//	connCfg := &rpcclient.ConnConfig{
//		Host:         "127.0.0.1:9332",
//		User:         "okcoin",
//		Pass:         "lZWMxOThhODEyNDQyYTg0NjY",
//		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
//		DisableTLS:   true, // Bitcoin core does not provide TLS by default
//	}
//	// Notice the notification parameter is nil since notifications are
//	// not supported in HTTP POST mode.
//	client, err := rpcclient.New(connCfg, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	db, err := leveldb.OpenFile("utxodb", &opt.Options{
//		BlockCacheCapacity: 1000 * opt.MiB,
//	})
//	if err != nil {
//		panic(err)
//	}
//
//	blockChan := make(chan *Block, 200)
//	go func() {
//		for i := 0; i < 300000; i++ {
//			blockhash, err := client.GetBlockHash(int64(i))
//			if err != err {
//				panic(err)
//			}
//
//			block, err := client.GetBlock(blockhash)
//			if err != nil {
//				panic(err)
//			}
//
//			blockChan <- &Block{
//				block:  block,
//				height: i,
//			}
//		}
//
//		close(blockChan)
//	}()
//
//	for block := range blockChan {
//		for _, tx := range block.block.Transactions {
//			txHash := tx.TxHash()
//
//			if !blockchain.IsCoinBaseTx(tx) {
//				for _, in := range tx.TxIn {
//					key := geneUtxoDBKey(in.PreviousOutPoint.Hash, int(in.PreviousOutPoint.Index))
//					err := db.Delete(key, nil)
//					if err != nil {
//						panic(err)
//					}
//				}
//			}
//
//			for idx, out := range tx.TxOut {
//				if len(out.PkScript) != 0 && out.PkScript[0] == 0x6a {
//					continue
//				}
//
//				utxo := Utxo{
//					isCoinbase: blockchain.IsCoinBaseTx(tx),
//					height:     block.height,
//					amount:     out.Value,
//					pkScript:   out.PkScript,
//				}
//				err := db.Put(geneUtxoDBKey(txHash, idx), utxo.encode(), nil)
//				if err != nil {
//					panic(err)
//				}
//			}
//		}
//
//		if block.height%10000 == 0 {
//			fmt.Printf("%s Handle block height: %d\n", time.Now().String(), block.height)
//		}
//	}
//
//	client.Shutdown()
//	fmt.Println("completed")
//}
//
//func geneUtxoDBKey(hash chainhash.Hash, index int) []byte {
//	r := make([]byte, 10)
//	copy(r[:6], hash[:6])
//	binary.LittleEndian.PutUint32(r[6:], uint32(index))
//
//	return r
//}
