package cli

import (
	"fmt"
	"blockchain/core"
	"blockchain/transaction"
)

func (cli *CLI) send(from, to string, amount int) {
	// 创建一个新区块
	bc := core.NewBlockchain(from)
	defer bc.DB.Close()
	// 创建一个新交易
	tx := core.NewUTXOTransaction(from, to, amount, bc)
	// 将交易打包进区块中，并加入区块链，写入数据库中
	bc.MineBlock([]*transaction.Transaction{tx})
	fmt.Println("Success!")
}