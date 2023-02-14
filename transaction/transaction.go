package transaction

import (
	"fmt"
	"crypto/sha256"
	"bytes"
	"encoding/gob"
	"log"
)

// 挖出新块的奖励金
const subsidy = 10

// 一个交易包含了交易ID、多个交易输入、多个交易输出
type Transaction struct {
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// 当矿工挖出一个新的块时，它会向新的块中添加一个 coinbase 交易。
// coinbase 交易只有一个输出，没有输入。
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	// 由于没有输入，所以 Txid 为空，Vout 等于 -1
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	// 输出的 锁定脚本 暂时用地址to代替
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}

// 交易哈希为交易ID
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}