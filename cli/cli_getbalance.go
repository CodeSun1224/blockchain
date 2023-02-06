package cli

import (
	"fmt"
	"blockchain/core"
)

func (cli *CLI) getBalance(address string) {
	bc := core.NewBlockchain(address)
	defer bc.DB.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}