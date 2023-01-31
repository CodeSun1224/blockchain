package main

import "fmt"
import "blockchain/core" // go mod包管理模式，这里的blockchain是go mod init blockchain中生成的项目模块名称

func main() {
	bc := core.NewBlockchain()
	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 BTC to Ivan")
	// go语言中，方法、属性等  大写开头表示可以全局调用，小写开头表示局部调用
	for _, block := range bc.Blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}