package transaction

import (

)

// 交易输出
type TXOutput struct {
	// 一定数量的币
	Value int
	// 锁定脚本：要花这笔钱，必须要解锁该脚本
	ScriptPubKey string
}

