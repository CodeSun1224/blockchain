package transaction

import (
	"bytes"
	"blockchain/wallet"
)

// 交易输入（一个输入引用了之前交易的一个输出）
type TXInput struct {
	// 之前交易的 ID
	Txid []byte
	// 之前交易输出的某一个索引（一个交易可以有多个输出，我们需要指出具体是哪一个）
	Vout int
	// 输入数字签名和公钥
	Signature []byte
	PubKey []byte
}

// 检查输出的公钥哈希pubKeyHash是否是由输入的公钥生成的
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}