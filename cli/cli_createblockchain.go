package cli

import (
	"fmt"
	"blockchain/core"
)

func (cli *CLI) createBlockchain(createBlockchainAddress string) {
	// if !ValidateAddress(address) {
	// 	log.Panic("ERROR: Address is not valid")
	// }
	bc := core.CreateBlockchain(createBlockchainAddress)
	defer bc.DB.Close()

	fmt.Println("Done!")
}