package core

import (
	"github.com/boltdb/bolt"
	"log"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"

type Blockchain struct {
	// tip：数据库中存储的最后一个区块的哈希
	Tip []byte
	DB *bolt.DB
}

func (bc *Blockchain) AddBlock(data string) {
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
	newBlock := NewBlock(data, lastHash)
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

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

func NewBlockchain() *Blockchain {
	var tip []byte
	// 打开bolt数据库文件
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		// 打开bucket
		b := tx.Bucket([]byte(blocksBucket))
		// bucket为空，代表没有创建区块链，此时创建bucket，并在bucket中添加创世区块
		//（key：区块哈希，value：序列化后的区块），同时更新一下l和tip
		if b == nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			// l中保存着最后一个区块的哈希（l : hash of the last block）
			err = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})
	// 创建区块链实例
	bc := Blockchain{tip, db}

	return &bc
}