package core

import (
	"math/big"
	"bytes"
	"blockchain/util"
	"fmt"
	"math"
	"crypto/sha256"
)

// 区块哈希值前面至少有targetBits个0
const targetBits = 16
const maxNonce = math.MaxInt64

type ProofOfWork struct {
	block *Block
	target *big.Int
}

// 设置区块b和target
func NewProofOfWork(b *Block) *ProofOfWork {
	// 计算哈希使用的是SHA-256算法，哈希值一共256位
	target := big.NewInt(1)
	// 1左移256-targetBits位
	target.Lsh(target, uint(256 - targetBits))

	pow := &ProofOfWork{b, target}
	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			util.IntToHex(pow.block.Timestamp),
			util.IntToHex(int64(targetBits)),
			util.IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	// 256位哈希，即32比特
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		// 为了方便和target比较，将hash转为big.Int
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
		    fmt.Printf("\r%x", hash)
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// 工作量证明
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
