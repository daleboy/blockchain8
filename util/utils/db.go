package utils

import (
	"os"
	"runtime"
	"syscall"

	"github.com/vrecan/death/v3"
	blockchain "sd-chain/blockchain8/core"
)

func CloseDB(chain *blockchain.Blockchain) {
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)
		defer runtime.Goexit()
		chain.Database.Close()
	})
}
