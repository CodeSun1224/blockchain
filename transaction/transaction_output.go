package transaction

import (
	"blockchain/util"
	"bytes"
)

// 交易输出
type TXOutput struct {
	// 一定数量的币
	Value int
	// 输出公钥哈希
	PubKeyHash []byte
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
