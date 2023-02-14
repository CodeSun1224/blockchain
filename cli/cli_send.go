package cli

import (
	"fmt"
	"blockchain/core"
	"blockchain/transaction"
	"blockchain/wallet"
	"log"
)

func (cli *CLI) send(from, to string, amount int) {
	if !wallet.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !wallet.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}
	// 创建一个新区块
	bc := core.NewBlockchain()
	utxoset := core.UTXOSet{bc}

	defer bc.DB.Close()

	// 创建一个新交易
	tx := core.NewUTXOTransaction(from, to, amount, &utxoset)
	cbTx := transaction.NewCoinbaseTX(from, "")
    txs := []*transaction.Transaction{cbTx, tx}
	// 将交易打包进区块中，并加入区块链，写入数据库中
	newBlock := bc.MineBlock(txs)
	utxoset.Update(newBlock)
	fmt.Println("Success!")
}