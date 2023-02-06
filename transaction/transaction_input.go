package transaction

import (

)

// 交易输入（一个输入引用了之前交易的一个输出）
type TXInput struct {
	// 之前交易的 ID
	Txid []byte
	// 之前交易输出的某一个索引（一个交易可以有多个输出，我们需要指出具体是哪一个）
	Vout int
	// 解锁脚本
	ScriptSig string
}