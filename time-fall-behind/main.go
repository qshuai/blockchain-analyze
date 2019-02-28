package main

import (
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/qshuai/tcolor"
	"os"
	"strings"
	"time"
)

func main() {
	host := flag.String("host", "127.0.0.1:8332", "rpc server ip + port")
	user := flag.String("user", "", "rpc authorized username")
	passwd := flag.String("passwd", "", "rpc authorized password")
	flag.Parse()

	if user == nil || passwd == nil {
		fmt.Println(tcolor.WithColor(tcolor.Red, "Please input rpc username and password for rpc authorization"))
		os.Exit(1)
	}

	rpc,err := getRPCinstance(*host, *user, *passwd)
	if err != nil {
		fmt.Println(tcolor.WithColor(tcolor.Red, "initial rpc connection failed"))
		os.Exit(1)
	}

	var height int64
	var prevTime time.Time
	var total int
	for {
		hash ,err := rpc.GetBlockHash(height)
		if err != nil {
			if strings.Contains(err.Error(), "Block height out of range") {
				fmt.Printf("Reach the last block, Finished! Collection %d block meeting the condition", total)

				os.Exit(0)
			}

			panic(err)
		}

		blockHeader, err := rpc.GetBlockHeader(hash)
		if err != nil {
			panic(err)
		}

		if prevTime.After(blockHeader.Timestamp) {
			fmt.Printf("block: %s:%d with timestamp: %d less than the previous block timestamp: %d\n",
				hash, height, blockHeader.Timestamp.Unix(), prevTime.Unix())

			total++
		}

		prevTime = blockHeader.Timestamp
		height++
	}
}

func getRPCinstance(host string, user string, passwd string) (*rpcclient.Client, error){
	connCfgService := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         passwd,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	rpc, err := rpcclient.New(connCfgService, nil)
	if err != nil {
		return nil, err
	}

	return rpc, nil
}