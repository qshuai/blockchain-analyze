package main

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	mod = 15

	levelPath = "sdk/txos"
)

var (
	db *leveldb.DB
	dbs [mod]*leveldb.DB

	roption = &opt.ReadOptions{DontFillCache:true}
)

type SpendOutpoint struct {
	height int
	value int64
	flag bool
	spendHeight int
	spendTxHash *chainhash.Hash
	spendTxIndex int
	pkScript []byte
}

func main() {
	start := time.Now()
	count := 0

	iter := db.NewIterator(nil, roption)
	for iter.Next() {
		count++

		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := iter.Key()
		value := iter.Value()
		outpoint ,err := decodeValue(value)
		if err != nil {
			panic(err)
		}
		k ,err := decodeAndGenerateKey(key, outpoint)
		if err != nil {
			panic(err)
		}
		// for the forth byte to decide which db to store
		shardingDB := getDB(k[4])

		value = generateValue(outpoint)
		err = shardingDB.Put(k, value, nil)
		if err != nil {
			panic(err)
		}

		now := time.Now()
		if now.Sub(start).Seconds() > 10 {
			fmt.Printf("%s Iterate over %d entries\n", now.Format("2016-01-02 15:04:05"), count)
			start = now
		}
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Complete iterate all entries, total: %d\n", count)
}

func decodeAndGenerateKey(key []byte, outpoint *SpendOutpoint) ([]byte, error) {
	str := string(key[:64])
	hash ,err := chainhash.NewHashFromStr(str)
	if err != nil {
		return nil, err
	}

	numStr := string(key[64:])
	num ,err := strconv.Atoi(numStr)
	if err != nil {
		return nil, err
	}

	r := make([]byte, 4)
	binary.BigEndian.PutUint32(r, uint32(num))

	// remove prefixed 0
	index := 0
	found := false
	for i := 0; i < 4; i++ {
		if r[i] != 0 {
			found = true
			index = i
			break
		}
	}

	// spendHeidht + txhash + index
	var k []byte
	if index == 0 && !found {
		k = make([]byte, 4 + 32 + 1)
		k[36] = 0
	} else {
		k = make([]byte, 4 + 32 + 4 - index)
		copy(k[36:], r[index:])
	}

	binary.LittleEndian.PutUint32(k[:4], uint32(outpoint.spendHeight))
	copy(k[4:36], hash[:])
	return k, nil
}

func generateValue(outpoint *SpendOutpoint) []byte {
	var ret []byte
	if outpoint.flag {
		ret = make([]byte, 52 + 1 + len(outpoint.pkScript))
	} else {
		// todo fixed
		ret = make([]byte, 12 + 1 + len(outpoint.pkScript))
	}

	binary.LittleEndian.PutUint32(ret[0:4], uint32(outpoint.height))
	binary.LittleEndian.PutUint64(ret[4:12], uint64(outpoint.value))
	if outpoint.flag {
		ret[12] = 1
		binary.LittleEndian.PutUint32(ret[13:17], uint32(outpoint.spendHeight))
		copy(ret[17:49], outpoint.spendTxHash[:])
		binary.LittleEndian.PutUint32(ret[49:53], uint32(outpoint.spendTxIndex))
		copy(ret[53:], outpoint.pkScript)
	} else {
		ret[12] = 0
		copy(ret[13:], outpoint.pkScript)
	}

	return ret
}

func decodeValue(value []byte) (*SpendOutpoint, error ) {
	height := binary.LittleEndian.Uint32(value[0:4])
	amount := binary.LittleEndian.Uint64(value[4:12])
	spendHeight := binary.LittleEndian.Uint32(value[12:16])
	pkScript := value[52:]
	if spendHeight == 0 {
		return &SpendOutpoint{
			height:int(height),
			value:int64(amount),
			spendHeight:int(spendHeight),
			flag:false,
		}, nil
	}

	txhash, err := chainhash.NewHash(value[16:48])
	if err != nil {
		return nil ,err
	}
	index := binary.LittleEndian.Uint32(value[48:52])

	return &SpendOutpoint{
		height:int(height),
		value:int64(amount),
		flag:true,
		spendHeight:int(spendHeight),
		spendTxHash:txhash,
		spendTxIndex:int(index),
		pkScript:pkScript,
	}, nil
}

func sharding(num byte) int {
	return int(num) % mod
}

func getDB(num byte) *leveldb.DB {
	n := sharding(num)
	return dbs[n]
}

func init() {
	var err error
	db ,err = leveldb.OpenFile(levelPath, &opt.Options{
		BlockCacheCapacity: 200 * opt.MiB,
	})
	if err != nil {
		panic(err)
	}

	for i := 0; i < mod; i++ {
		shardDB ,err := leveldb.OpenFile(levelPath+strconv.Itoa(i), &opt.Options{
			BlockCacheCapacity: 200 * opt.MiB,
		})
		if err != nil {
			panic(err)
		}

		dbs[i] = shardDB
	}
}
