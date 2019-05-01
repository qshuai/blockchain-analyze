package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync/atomic"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/toorop/go-bitcoind"
)

var (
	// control
	shutdown uint32
)

type BlockInfo struct {
	block *wire.MsgBlock
	height int
}

func main() {
	// bc, err := bitcoind.New("127.0.0.1", 8332, "root", "root", false)
	bc, err := bitcoind.New("127.0.0.1", 9332, "okcoin", "lZWMxOThhODEyNDQyYTg0NjY", false)
	if err != nil {
		panic(err)
	}

	blockChan := make(chan *BlockInfo, 20)
	go func() {
		for i := 0; i < 566970; i++ {
			// for i := 0; i < 300000; i++ {
			if atomic.LoadUint32(&shutdown) != 0 {
				fmt.Println("has requested shutdown")
				close(blockChan)
				return
			}

			blockHash, err := bc.GetBlockHash(uint64(i))
			if err != nil {
				panic(err)
			}

			blockRawString, err := bc.GetRawBlock(blockHash)
			if err != nil {
				panic(err)
			}

			var block wire.MsgBlock
			blockBytes, err := hex.DecodeString(blockRawString)
			if err != nil {
				panic(err)
			}
			err = block.Deserialize(bytes.NewReader(blockBytes))
			if err != nil {
				panic(err)
			}

			blockChan <- &BlockInfo {
				block: &block,
				height:i,
			}
		}

		close(blockChan)
	}()

	var txes int
	var inputs int
	var outputs int
	var nulldata int
	var nulldataWithValue int
	var nonePkscript int

	// script type count
	var p2pk int
	var p2pkh int
	var p2sh int
	var multisig int
	var witnessPubkeyHash int
	var witnessScriptHash int
	var nonStandard int

	fmt.Println("block height\ttxes\tinputs\toutputs\tnulldata\tnulldata with value\tnon pkScript\tp2pk\tp2pkh\tp2sh\tmultisig\twitnessPubkeyHash\twitnessScriptHash\tnonStandard")
	for blockInfo := range blockChan {
		txes += len(blockInfo.block.Transactions)

		var tmp_inputs int
		var tmp_outputs int
		var tmp_nulldata int
		var tmp_nulldataWithValue int
		var tmp_nonePkscript int

		// script type count
		var tmp_p2pk int
		var tmp_p2pkh int
		var tmp_p2sh int
		var tmp_multisig int
		var tmp_witnessPubkeyHash int
		var tmp_witnessScriptHash int
		var tmp_nonStandard int

		for _, tx := range blockInfo.block.Transactions {
			// containing coinbase inputs
			inputs += len(tx.TxIn)
			tmp_inputs += len(tx.TxIn)

			outputs += len(tx.TxOut)
			tmp_outputs += len(tx.TxOut)
			for idx, out := range tx.TxOut {
				if len(out.PkScript) == 0 {
					fmt.Printf("txout without pkscript: %s:%d\n", tx.TxHash(), idx)
					nonePkscript++
					tmp_nonePkscript++
				} else {
					if out.PkScript[0] == 0x6a {
						fmt.Printf("nulldata transaction: %s:%d\n", tx.TxHash(), idx)
						nulldata++
						tmp_nulldata++

						if out.Value != 0 {
							fmt.Printf("nulldata transaction with value: %s:%d\n", tx.TxHash(), idx)
							nulldataWithValue++
							tmp_nulldataWithValue++
						}
					}
				}

				class := txscript.GetScriptClass(out.PkScript)
				switch class {
				case txscript.PubKeyTy:
					p2pk++
					tmp_p2pk++
				case txscript.PubKeyHashTy:
					p2pkh++
					tmp_p2pkh++
				case txscript.ScriptHashTy:
					p2sh++
					tmp_p2sh++
				case txscript.WitnessV0PubKeyHashTy:
					witnessPubkeyHash++
					tmp_witnessPubkeyHash++
				case txscript.WitnessV0ScriptHashTy:
					witnessScriptHash++
					tmp_witnessScriptHash++
				case txscript.MultiSigTy:
					multisig++
					tmp_multisig++
				case txscript.NonStandardTy:
					nonStandard++
					tmp_nonStandard++
				}
			}
		}

		fmt.Printf("%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n",
			blockInfo.height, len(blockInfo.block.Transactions), tmp_inputs, tmp_outputs, tmp_nulldata, tmp_nulldataWithValue, tmp_nonePkscript,
			tmp_p2pk, tmp_p2pkh, tmp_p2sh, tmp_multisig, tmp_witnessPubkeyHash, tmp_witnessScriptHash, tmp_nonStandard)
	}

	fmt.Printf("total txes: %d, total inputs: %d, total outputs: %d, null data outpus: %d, null data with value: %d, non pkScript: %d, "+
		"p2pk: %d, p2pkh: %d, p2sh: %d, multisig: %d, witnessPubkeyHash: %d, witnessScriptHash: %d, nonstandard: %d\n",
		txes, inputs, outputs, nulldata, nulldataWithValue, nonePkscript, p2pk, p2pkh, p2sh, multisig, witnessPubkeyHash, witnessScriptHash, nonStandard)
}
