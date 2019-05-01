package main

import (
	"log"

	"github.com/gcash/bchd/rpcclient"
	"github.com/gin-gonic/gin"
)

func main() {
	connCfg := &rpcclient.ConnConfig{
		Host:         "127.0.0.1:8332",
		User:         "KAuCgqk0gwgP9LWtDnu",
		Pass:         "EGQFjJu81Ck3j7lFvU8cPW2jALopF",
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

	r := gin.Default()
	r.GET("/mempool", func(c *gin.Context) {
		ret, err := client.GetMempoolInfo()
		if err != nil {
			c.JSON(200, gin.H{
				"message": err.Error(),
			})

			return
		}

		c.JSON(200, ret)
	})

	r.GET("/mempool/detail", func(c *gin.Context) {
		txList, err := client.GetRawMempool()
		ret := make([]string, 0, len(txList))
		for _, tx := range txList {
			ret = append(ret, tx.String())
		}

		if err != nil {
			c.JSON(200, gin.H{
				"message": err.Error(),
			})

			return
		}

		c.JSON(200, ret)
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
