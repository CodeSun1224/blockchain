package core

import (
	"bytes"
	"time"
	"strconv"
	"crypto/sha256"
)

type Block struct {
	Timestamp int64
	Data []byte
	PrevBlockHash []byte
	Hash []byte
}

// 计算block的Hash
func (b *Block) SetHash() {
	// 把int64类型的改成[]byte类型，方便计算哈希
	timeStamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	// []byte{}代表当前block的Hash。当前block的Hash是空的
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timeStamp}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}}
	block.SetHash()
	return block
}