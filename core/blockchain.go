package core

import (
	"github.com/boltdb/bolt"
	"log"
	"blockchain/transaction"
	"encoding/hex"
	"fmt"
	"os"
	"blockchain/wallet"
	"bytes"
	"crypto/ecdsa"
	"errors"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type Blockchain struct {
	// tip：数据库中存储的最后一个区块的哈希
	Tip []byte
	DB *bolt.DB
}

func (bc *Blockchain) AddBlock(transactions []*transaction.Transaction) {
	// 获取数据库中最后一个区块的哈希
	var lastHash []byte
	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	// 创建要增加的新块
	newBlock := NewBlock(transactions, lastHash)
	// 将新块加入数据库中，更新tip
	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), newBlock.Hash)
		bc.Tip = newBlock.Hash

		return nil
	})
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// 如果数据库找不到区块链，则需要调用CreateBlockchain创建一个，否则取出tip，构造一个新块
func NewBlockchain() *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// 创建一个创世块，写入数据库中
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	cbtx := transaction.NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// 查看余额：找到未包含在任何交易输入中的输出
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []transaction.Transaction {
	var unspentTXs []transaction.Transaction
	// spentTXOs：交易ID->该笔交易中哪些输出已经被包含在其它交易的输入中了
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	// 遍历区块链中的每一个区块
	for {
		block := bci.Next()

		// 遍历一个区块中的所有交易
	    for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
			// 对于每一笔交易，遍历所有的输出
			Outputs:
			for outIdx, out := range tx.Vout {
				// 查看spentTXOs中是否包括当前交易ID，是，则继续遍历，否则，即找到一个未包含在任何交易输入中的输出
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				// 我们不需要关注不能被解锁的交易
				if out.IsLockedWithKey(pubKeyHash) {
					// 此时将该笔交易加入结果集中
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			// 如果当前交易不是coinbase交易（有1个或多个交易输入），则遍历当前交易的所有输入
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					// 我们不需要关注不能被解锁的交易
					if in.UsesKey(pubKeyHash) {
						// 上一笔交易的输出，就是这笔交易的输入，此时将上一笔交易ID和对应的输出索引，记录在spentTXOs中
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
	  	}

		// PrevBlockHash长度为0，代表已经遍历到初始块，此时遍历结束，退出循环
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
  
	return unspentTXs
}

// 普通交易：from给to发amount个币
func NewUTXOTransaction(from, to string, amount int, utxoset *UTXOSet) *transaction.Transaction {
	var inputs []transaction.TXInput
	var outputs []transaction.TXOutput

	wallets, err := wallet.NewWallets()
	if err != nil {
		log.Panic(err)
	}
	wallet_from := wallets.GetWallet(from)
	pubKeyHash := wallet.HashPubKey(wallet_from.PublicKey)

	// 找到所有未花费的输出，并计算它们的value和是否足够
	acc, validOutputs := utxoset.FindSpendableOutputs(pubKeyHash, amount)
	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// 对所有未花费的输出进行遍历，构造交易输入
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if(err != nil) {
			log.Panic(err)
		}
		//  一个交易ID对应多个交易输出，所以还要遍历一次
		for _, out := range outs {
			// 构造交易输入
			input := transaction.TXInput{txID, out, nil, wallet_from.PublicKey}
			inputs = append(inputs, input)
		}
	}

	// 构造一个交易输出
	outputs = append(outputs, *transaction.NewTXOutput(amount, to))
	// 如果未花费的币的数量超过了新交易的输入数量，多余的币还要退还给from，因此还要构造一个交易输出，输出地址是from
	if acc > amount {
		outputs = append(outputs, *transaction.NewTXOutput(acc - amount, from)) // a change
	}

	tx := transaction.Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	utxoset.Blockchain.SignTransaction(&tx, wallet_from.PrivateKey)

	return &tx
}
// 找到所有未花费的输出，并计算它们的value和是否足够
func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	// 记录所有未花费交易输出
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

	Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

// 所有交易打包为一个区块，写入数据库中
func (bc *Blockchain) MineBlock(transactions []*transaction.Transaction) *Block {
	var lastHash []byte

	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	
	newBlock := NewBlock(transactions, lastHash)

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.Tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

// 查余额：找到所有未花费的输出，计算所有输出的value和
func (bc *Blockchain) FindUTXO() map[string]transaction.TXOutputs {
	UTXO := make(map[string]transaction.TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

// 根据ID获取交易
func (bc *Blockchain) FindTransaction(ID []byte) (transaction.Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return transaction.Transaction{}, errors.New("Transaction is not found")
}

// 对交易输入引用的之前的交易进行签名
func (bc *Blockchain) SignTransaction(tx *transaction.Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]transaction.Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// 验证交易
func (bc *Blockchain) VerifyTransaction(tx *transaction.Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTXs := make(map[string]transaction.Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}