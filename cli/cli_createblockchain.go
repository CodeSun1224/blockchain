package cli

import (
	"fmt"
	"blockchain/core"
	"blockchain/wallet"
	"log"
)

func (cli *CLI) createBlockchain(createBlockchainAddress string, nodeID string) {
	if !wallet.ValidateAddress(createBlockchainAddress) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := core.CreateBlockchain(createBlockchainAddress, nodeID)
	defer bc.DB.Close()

	UTXOSet := core.UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}