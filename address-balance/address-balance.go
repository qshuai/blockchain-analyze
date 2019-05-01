package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/toorop/go-bitcoind"
	"sync/atomic"
	"time"
)

const (
	addressLevelPath = "sdk/address"
	utxoLevelPath = "sdk/utxo"

	mapSize = 450000000
)

var (
	addressDb *leveldb.DB
	utxoDB *leveldb.DB

	// 0: running
	// 1: shutdown requesting
	shutdown uint32

	addressCollector map[string]*AddressBalanceInfo

	emptyFlag = []byte{0x12, 0xbe, 0x14, 0xaa}
)

type AddressBalanceInfo struct {
	received int64
	send int64
	txes uint32
	unspentTxes uint32
	created uint32
	updated uint32

	bestHeight uint32
}

func (info *AddressBalanceInfo) Encode() []byte {
	r := make([]byte, 32)
	binary.LittleEndian.PutUint64(r[0:8], uint64(info.received))
	binary.LittleEndian.PutUint64(r[8:16], uint64(info.send))
	binary.LittleEndian.PutUint64(r[16:20], uint64(info.txes))
	binary.LittleEndian.PutUint64(r[20:24], uint64(info.unspentTxes))
	binary.LittleEndian.PutUint64(r[24:28], uint64(info.created))
	binary.LittleEndian.PutUint64(r[28:32], uint64(info.updated))

	return r
}

func (info *AddressBalanceInfo) receiveCoin(height uint32, blockTime uint32, out *wire.TxOut, isFirst bool) {
	if info.bestHeight != height {
		info.txes++
		info.bestHeight = height
	}

	if isFirst {
		info.created = blockTime
	}

	info.received += out.Value
	info.unspentTxes++
	info.updated = blockTime
}

func (info *AddressBalanceInfo) spendCoin(height uint32, blockTime uint32, value int64)  {
	if info.bestHeight != height {
		info.txes++
		info.bestHeight = height
	}

	info.updated = blockTime
	info.send += value
	info.unspentTxes--
}

type UtxoView struct {
	isFromCoinbase bool
	height uint32
	amount int64
	pkScript []byte
}

func (u UtxoView) encode() []byte {
	r := make([]byte, 12 +len(u.pkScript))
	binary.LittleEndian.PutUint32(r[0:4], uint32(u.compactCoinbaseAndHeight()))
	binary.LittleEndian.PutUint64(r[4:12], uint64(u.amount))
	copy(r[12:], u.pkScript)

	return r
}

func decodeUtxoView(value []byte) *UtxoView {
	coinbaseAndHeight := binary.LittleEndian.Uint32(value[0:4])
	amount := binary.LittleEndian.Uint64(value[4:12])
	return &UtxoView{
		isFromCoinbase: coinbaseAndHeight & 0x01 == 1,
		height:coinbaseAndHeight / 2,
		amount:int64(amount),
		pkScript:value[12:],
	}
}

func (u UtxoView) compactCoinbaseAndHeight() uint32 {
	if u.isFromCoinbase {
		return u.height * 2 | 1
	}

	return u.height * 2
}

type Blocks struct {
	block *wire.MsgBlock
	height uint32
}

func main() {
	fmt.Println("start process transaction outputs...")
	blockChan := make(chan *Blocks, 10)
	bc ,err := bitcoind.New("127.0.0.1", 8332, "KAuCgqk0gwgP9LWtDnu", "EGQFjJu81Ck3j7lFvU8cPW2jALopF", false)
	if err != nil {
		panic(err)
	}

	go func() {
		for i := uint32(0); i < 566970; i++ {
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

	start := time.Now()
	var blocks int
	for block := range blockChan {
		txes := block.block.Transactions
		for _, tx := range txes {
			txHash := tx.TxHash()

			for index, out := range tx.TxOut {
				// ignore OP_RETURN
				if out.PkScript != nil && len(out.PkScript) != 0 && out.PkScript[0] == 0x6a {
					continue
				}

				utxo := &UtxoView{
					isFromCoinbase:blockchain.IsCoinBaseTx(tx),
					height:block.height,
					amount:out.Value,
					pkScript:out.PkScript,
				}
				err := utxoDB.Put(generateUtxoKey(&txHash, index), utxo.encode(), nil)
				if err != nil {
					panic(err)
				}

				addressKey := generateAddressKey(out.PkScript)
				item, ok := addressCollector[addressKey]
				if ok {
					item.receiveCoin(block.height, uint32(block.block.Header.Timestamp.Unix()), out, false)
				} else {
					item = new(AddressBalanceInfo)
					item.receiveCoin(block.height, uint32(block.block.Header.Timestamp.Unix()), out, false)
					addressCollector[addressKey] = item
				}
			}

			for _, in := range tx.TxIn {
				if blockchain.IsCoinBaseTx(tx) {
					break
				}

				utxoKey := generateUtxoKey(&in.PreviousOutPoint.Hash, int(in.PreviousOutPoint.Index))
				value ,err := utxoDB.Get(utxoKey, &opt.ReadOptions{DontFillCache: true})
				if err != nil {
					panic(err)
				}
				utxo := decodeUtxoView(value)

				addressKey := generateAddressKey(utxo.pkScript)
				info, ok := addressCollector[addressKey]
				if !ok {
					panic("absent address item" + addressKey)
				}
				info.spendCoin(block.height, uint32(block.block.Header.Timestamp.Unix()), utxo.amount)

				err = utxoDB.Delete(utxoKey, nil)
				if err != nil {
					panic(err)
				}
			}
		}

		blocks++

		end := time.Now()
		if end.Sub(start).Seconds() > 10 {
			fmt.Printf("Handle blocks: %d during %f seconds, map size: %d, current block: %d\n", blocks, end.Sub(start).Seconds(), len(addressCollector), block.height)
			blocks = 0
			start = end
		}
	}

	flush()
}

func generateAddressKey(pkScript []byte) string {
	var addressKey string
	if len(pkScript) == 0 {
		addressKey = hex.EncodeToString(emptyFlag)
	} else {
		addressKey = hex.EncodeToString(pkScript)
	}

	return addressKey
}

func generateUtxoKey(hash *chainhash.Hash, index int) []byte {
	r := make([]byte, 36)
	copy(r[:32], hash[:])
	binary.LittleEndian.PutUint32(r[32:], uint32(index))

	return r
}

func flush() {
	fmt.Printf("start flush address cache to leveldb, total items: %d\n", len(addressCollector))

	var i int
	for key, item := range addressCollector {
		dbKey, err := hex.DecodeString(key)
		if err != nil {
			panic(err)
		}

		err = addressDb.Put(dbKey, item.Encode(), nil)
		if err != nil {
			panic(err)
		}

		if i%10000 == 0 {
			fmt.Printf("Has Handled %d entries\n", i)
		}

		i++
	}

	fmt.Println("completed flush!!!")
}

func init() {
	var err error
	addressDb ,err = leveldb.OpenFile(addressLevelPath, &opt.Options{
		BlockCacheCapacity: 200 * 1048576, // 200 MB
		Compression:opt.SnappyCompression,
	})
	if err != nil {
		panic(err)
	}

	utxoDB ,err = leveldb.OpenFile(utxoLevelPath, &opt.Options{
		BlockCacheCapacity: 400 * 1048576, // 200 MB
		Compression:opt.SnappyCompression,
	})
	if err != nil {
		panic(err)
	}

	addressCollector = make(map[string]*AddressBalanceInfo, mapSize)
}
