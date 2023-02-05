package main

import "blockchain/cli"
import "blockchain/core"

func main() {
	bc := core.NewBlockchain()
	defer bc.DB.Close()

	cli := cli.CLI{bc}
	cli.Run()
}