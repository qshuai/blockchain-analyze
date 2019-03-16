package max_script_lenght

import (
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"log"
	"time"
)

var max int

func main() {
	connCfg := &rpcclient.ConnConfig{
		Host:         "127.0.0.1:8332",
		User:         "hello",
		Pass:         "world",
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

	blockChan := make(chan *wire.MsgBlock, 200)
	go func() {
		for i := 0; i < 567297; i++ {
			if i%1000 == 0 {
				fmt.Printf("%s Handle block height: %d, channel size: %d", time.Now().String(), i, len(blockChan))
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
			for _, out := range tx.TxOut {
				// null data
				if out.PkScript[0] == 0x6a {
					continue
				}

				if len(out.PkScript) > max {
					max = len(out.PkScript)
				}
			}
		}
	}

	fmt.Println("most length in pkScript(byte):", max)
}