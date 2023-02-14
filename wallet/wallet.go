package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"blockchain/util"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"bytes"
)

const version = byte(0x00)
const addressChecksumLen = 4

// Wallet stores private and public keys
type Wallet struct {
	// 公钥私钥使用椭圆曲线数字签名算法（ecdsa）生成
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	// 生成公钥私钥，创建一个新的钱包
	private, public := newKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	// 使用elliptic算法生成一个椭圆曲线
	curve := elliptic.P256()
	// 使用ecdsa算法，根据curve和随机数rand.Reader，生成一个私钥
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	// 公钥是椭圆曲线上的点，由X、Y坐标组成。相关数据从私钥中取出
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}

// 根据公钥生成钱包地址
func (w Wallet) GetAddress() []byte {
	// 对钱包公钥进行哈希计算，获取公钥哈希
	pubKeyHash := HashPubKey(w.PublicKey)
	// 这里的版本号version是自定义的
	versionedPayload := append([]byte{version}, pubKeyHash...)
	//checksum由versionedPayload经过两次sha256算法计算得出
	checksum := checksum(versionedPayload)
	fullPayload := append(versionedPayload, checksum...)
	// 地址由 版本号version + 公钥哈希pubKeyHash + checksum 三部分，通过Base58算法计算得出
	address := util.Base58Encode(fullPayload)

	return address
}

// 对钱包公钥进行哈希计算，获取公钥哈希
func HashPubKey(pubKey []byte) []byte {
	// 首先对公钥进行一次sha256计算，
	publicSHA256 := sha256.Sum256(pubKey)
	// 之后再进行一次ripemd160计算，
	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	// 最后将结果转为[]byte输出
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// 检查地址是否有效
func ValidateAddress(address string) bool {
	// 首先对传入的地址进行解码
	pubKeyHash := util.Base58Decode([]byte(address))
	// Checksum在最后，占4个字节。因此我们取出后4个字节的数据actualChecksum
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	// version占1个字节
	version := pubKeyHash[0]
	// 剩下的就是该地址的公钥哈希
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	// 如果一个地址有效，
	// 则它的 version+pubKeyHash 计算出的checksum肯定和它的后4个字节（actualChecksum）相等
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}