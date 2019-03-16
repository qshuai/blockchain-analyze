package main

import (
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/qshuai/tcolor"
	"math/rand"
	"os"
	"time"
)

var (
	r *rand.Rand
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

	source := rand.NewSource(time.Now().Unix())
	r = rand.New(source)

	rpc,err := getRPCinstance(*host, *user, *passwd)
	if err != nil {
		fmt.Println(tcolor.WithColor(tcolor.Red, "initial rpc connection failed"))
		os.Exit(1)
	}

	hash ,err := chainhash.NewHashFromStr("0000000000000000001cec15110040b1792c1f68a2d9d6cd1436ac2e3cbf21af")
	if err != nil {
		panic(err)
	}

	block ,err := rpc.GetBlock(hash)
	if err != nil {
		panic(err)
	}

	for _, tx := range block.Transactions {
		for idx, in := range tx.TxIn {
			if len(in.SignatureScript) == 23 {
				fmt.Printf("transaction: %s, in index: %d\n", tx.TxHash(), idx)
			}
		}
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