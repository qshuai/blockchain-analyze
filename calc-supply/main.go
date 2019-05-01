package main

import (
	"flag"
	"fmt"

	"github.com/btcsuite/btcutil"
)

const (
	initReward     = 5000000000
	reduceInterval = 210000
)

func main() {
	height := flag.Int("height", 0, "please specify block height")
	flag.Parse()

	i := *height / reduceInterval
	j := *height%reduceInterval + 1

	var supply int64
	var reward int64 = initReward
	for i > 0 {
		supply += reduceInterval * reward
		reward /= 2
		i--
	}

	supply += reward * int64(j)

	fmt.Printf("total supply: %d(%f), up to block: %d\n", supply, btcutil.Amount(supply).ToBTC(), *height)
}
