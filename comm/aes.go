package comm

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"

	"github.com/solozyx/seckill/conf"
)

// 高级加密标准 Adevanced Encryption Standard ,AES

// 16 24 32位字符串 分别对应AES-128 AES-192 AES-256 加密方法
// key不能泄露
var PwdKey = []byte(conf.AESKey)

// PKCS7 填充模式 (还有 PKCS5 和 零填充模式 共3种填充模式)
// plaintext 要加密的明文
func pkcs7padding(plaintext []byte, blockSize int) []byte {
	// 计算需要填充的个数
	padding := blockSize - len(plaintext)%blockSize
	// Repeat()函数功能是把切片[]byte{byte(padding)}复制padding个 然后合并成新的字节切片返回
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, paddingText...)
}

// 填充反向操作 删除填充字符串
func pkcs7unpadding(paddingData []byte) ([]byte, error) {
	length := len(paddingData)
	if length == 0 {
		return nil, errors.New("padding data error")
	} else {
		// 获取填充字符串长度
		unPadding := int(paddingData[length-1])
		// 截取切片 删除填充字节
		return paddingData[:(length - unPadding)], nil
	}
}

// AES加密
func aesEncrypt(data []byte, key []byte) ([]byte, error) {
	// 创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 获取块大小 即获取 PwdKey的大小 16位对应 AES-128 ,24位 AES-192 ,32位 AES-256
	blockSize := block.BlockSize()
	// 对数据进行填充 让数据长度满足需求
	data = pkcs7padding(data, blockSize)
	// 采用AES加密方法中CBC加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	ciphertext := make([]byte, len(data))
	// 执行加密
	blockMode.CryptBlocks(ciphertext, data)
	return ciphertext, nil
}

// AES解密
func aesDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// 创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 获取块大小
	blockSize := block.BlockSize()
	// 创建加密实例
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	plaintext := make([]byte, len(ciphertext))
	// 解密
	blockMode.CryptBlocks(plaintext, ciphertext)
	// 去除填充字符串
	plaintext, err = pkcs7unpadding(plaintext)
	if err != nil {
		return nil, err
	}
	return plaintext, err
}

// AES加密得到切片,不方便写入客户端cookie,密文做base64编码
func AesEncryptBase64Encode(data []byte) (string, error) {
	result, err := aesEncrypt(data, PwdKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(result), err
}

func Base64DecodeAesDecrypt(base64string string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(base64string)
	if err != nil {
		return nil, err
	}
	// AES解密
	return aesDecrypt(data, PwdKey)
}
