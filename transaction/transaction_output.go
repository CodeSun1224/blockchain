package transaction

import (
	"blockchain/util"
	"bytes"
	"log"
	"encoding/gob"
)

// 交易输出
type TXOutput struct {
	// 一定数量的币
	Value int
	// 输出公钥哈希
	PubKeyHash []byte
}

type TXOutputs struct {
	Outputs []TXOutput
}

// 对于一笔发往address的交易，需要对该地址进行锁定（即计算该地址对应的公钥哈希，存入交易输出中）
// ，从而对这笔交易进行唯一性标记
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := util.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// 通过对比接收方的公钥哈希与接收到的交易输出中的公钥哈希是否一致，可以验证该笔交易的目的地址是否正确
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}