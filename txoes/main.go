package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/toorop/go-bitcoind"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	levelPath = "sdk/txos"

	mapSize = 40000000
)

var (
	cache map[string]*SpendOutpoint

	db *leveldb.DB

	// debug:
	blocks int
	txes int
	inputs int
	outputs int
	cached int
	nulldata int

	// control
	shutdown uint32
	mtx sync.Mutex
)

type SpendOutpoint struct {
	height int
	value int64
	spendHeight int
	spendTxHash *chainhash.Hash
	spendTxIndex int
	pkScript []byte
}

type Blocks struct {
	block *wire.MsgBlock
	height int
}

func main()  {
	fmt.Println("start process transaction outputs...")
	var err error
	db ,err = leveldb.OpenFile(levelPath, &opt.Options{
		BlockCacheCapacity: 200,
	})
	if err != nil {
		panic(err)
	}

	cache = make(map[string]*SpendOutpoint, mapSize)
	start := time.Now()

	blockChan := make(chan *Blocks, 20)
	bc ,err := bitcoind.New("127.0.0.1", 8332, "root", "root", false)
	go func() {
		//for i := 0; i < 2; i++ {
		for i := 0; i < 300000; i++ {
			if atomic.LoadUint32(&shutdown) != 0 {
				fmt.Println("has requested shutdown")
				close(blockChan)
				return
			}

			blockHash ,err := bc.GetBlockHash(uint64(i))
			if err != nil {
				fmt.Printf("get block hash failed: %s", err)
				close(blockChan)
				return
			}

			rawBlock ,err := bc.GetRawBlock(blockHash)
			if err != nil {
				fmt.Printf("get block failed: %s", err)
				close(blockChan)
				return
			}
			var block wire.MsgBlock
			blockBytes, err := hex.DecodeString(rawBlock)
			if err != nil {
				fmt.Printf("decode block failed: %s\n", err)
				close(blockChan)
				return
			}
			err = block.Deserialize(bytes.NewReader(blockBytes))
			if err != nil {
				fmt.Printf("deserialize block failed: %s\n", err)
				close(blockChan)
				return
			}

			if err != nil {
				fmt.Printf("get block failed: %s\n", err)
				atomic.StoreUint32(&shutdown, 1)
				close(blockChan)
				return
			}

			blockChan <- &Blocks{
				block: &block,
				height: i,
			}
		}

		close(blockChan)
	}()

	defer func() {
		fmt.Println("start flushing to leveldb....")
		flush()
		fmt.Println("completed flush!!!")
	}()

	//for i := 0; i < 566970; i++ {
	for blockinfo := range blockChan {
		if atomic.LoadUint32(&shutdown) != 0 {
			break
		}

		batch := leveldb.Batch{}
		block := blockinfo.block
		txes += len(block.Transactions)
		for _, tx := range block.Transactions {
			inputs += len(tx.TxIn)
			outputs += len(tx.TxOut)

			hash := tx.TxHash()

			err = removeOldEntry(len(tx.TxOut))
			if err != nil {
				fmt.Println("remove old entry failed")
				return
			}
			for idx, out := range tx.TxOut {
				// skip NULLDATA output
				if len(out.PkScript) != 0 && out.PkScript[0] == 0x6a {
					nulldata++
					continue
				}

				cache[generateKey(&hash, idx)] = &SpendOutpoint{
					height: blockinfo.height,
					value:out.Value,
					pkScript:out.PkScript,
				}
			}

			if !blockchain.IsCoinBaseTx(tx) {
				for idx, in := range tx.TxIn {
					prevHash := in.PreviousOutPoint.Hash
					prevIndex := in.PreviousOutPoint.Index
					key := generateKey(&prevHash, int(prevIndex))
					outpoint, ok := cache[key]
					if !ok {
						value, err := db.Get([]byte(key), &opt.ReadOptions{
							DontFillCache: true,
						})
						if err != nil {
							fmt.Println("db get failed")
							return
						}

						spendOutpoint, err := decodeValue(value)
						if err != nil {
							fmt.Println("decode value failed")
							return
						}

						// update
						spendOutpoint.spendHeight = blockinfo.height
						spendOutpoint.spendTxIndex = idx
						spendOutpoint.spendTxHash = &hash
						batch.Put([]byte(key), generateValue(spendOutpoint, true))
					} else {
						cached++

						outpoint.spendHeight = blockinfo.height
						outpoint.spendTxHash = &hash
						outpoint.spendTxIndex = idx
						batch.Put([]byte(key), generateValue(outpoint, true))

						delete(cache, key)
					}
				}
			}
		}

		err = db.Write(&batch, nil)
		if err != nil {
			fmt.Println("error while write batch in progress")
			return
		}

		blocks++
		end := time.Now()
		diff := end.Sub(start).Seconds()
		if diff > 10 {
			fmt.Printf("Handled %d blocks in %f seconds, txes: %d(ins: %d[%d cached]), outs: %d, map size: %d, channel size: %d, nulldata: %d; current block: %d\n",
				blocks, diff, txes, inputs, cached, outputs, len(cache), len(blockChan), nulldata, blockinfo.height)

			start = end
			blocks = 0
			txes = 0
			inputs = 0
			outputs = 0
			cached = 0
		}
	}
}

func generateKey(hash *chainhash.Hash, idx int) string {
	return hash.String() + strconv.Itoa(idx)
}

func generateValue(outpoint *SpendOutpoint, completed bool) []byte {
	ret := make([]byte, 52 + len(outpoint.pkScript))

	binary.LittleEndian.PutUint32(ret[0:4], uint32(outpoint.height))
	binary.LittleEndian.PutUint64(ret[4:12], uint64(outpoint.value))
	if completed {
		binary.LittleEndian.PutUint32(ret[12:16], uint32(outpoint.spendHeight))
		copy(ret[16:48], outpoint.spendTxHash[:])
		binary.LittleEndian.PutUint32(ret[48:52], uint32(outpoint.spendTxIndex))
	} else {
		binary.LittleEndian.PutUint32(ret[12:16], uint32(0))
		binary.LittleEndian.PutUint32(ret[48:52], 0)
	}
	copy(ret[52:], outpoint.pkScript)

	return ret
}

func decodeValue(value []byte) (*SpendOutpoint, error ) {
	height := binary.LittleEndian.Uint32(value[0:4])
	amount := binary.LittleEndian.Uint64(value[4:12])
	spendHeight := binary.LittleEndian.Uint32(value[12:16])
	txhash, err := chainhash.NewHash(value[16:48])
	if err != nil {
		return nil ,err
	}
	index := binary.LittleEndian.Uint32(value[48:52])
	pkScript := value[52:]

	return &SpendOutpoint{
		height:int(height),
		value:int64(amount),
		spendHeight:int(spendHeight),
		spendTxHash:txhash,
		spendTxIndex:int(index),
		pkScript:pkScript,
	}, nil
}

func removeOldEntry(toBeInsert int) error {
	batch := leveldb.Batch{}
	toBeRemoved := (toBeInsert + len(cache)) - mapSize
	if toBeRemoved > 0 {
		if toBeRemoved < 20000 {
			toBeRemoved = 20000
		}

		i := 0
		for key, value := range cache {
			if i > toBeRemoved {
				break
			}

			batch.Put([]byte(key), generateValue(value, false))
			delete(cache, key)
			i++
		}

		err := db.Write(&batch, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func flush() {
	mtx.Lock()
	defer mtx.Unlock()

	batch := leveldb.Batch{}
	count := 0
	for key, value := range cache {
		count++
		if count >= 10000 {
			err := db.Write(&batch, nil)
			if err != nil {
				fmt.Println("panic while flush in progress")
				panic(err)
			}

			count = 0
		}

		batch.Put([]byte(key), generateValue(value, false))
		delete(cache, key)
	}

	err := db.Write(&batch, nil)
	if err != nil {
		fmt.Println("panic while flush in progress")
		panic(err)
	}
}

func initRPC() *rpcclient.Client {
	connCfg := &rpcclient.ConnConfig{
		Host:         "127.0.0.1:9332",
		User:         "okcoin",
		//User:         "root",
		//User:         "qQGq8VOZCwSe926W",
		Pass:         "lZWMxOThhODEyNDQyYTg0NjY",
		//Pass: "root",
		//Pass:         "cRRwsNVPcC4HUzxM88HYAliQ6GodFo1BuN6y9",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

