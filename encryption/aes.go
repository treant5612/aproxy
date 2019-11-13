package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
)

type Encipher interface {
	Encrypt(data []byte) (cipherText []byte)
	Decrypt(data []byte) (plainText []byte)
}

type AesCfbEncipher struct {
	key          []byte
	block        cipher.Block
	cfbEncrypter cipher.Stream
	cfbDecrypter cipher.Stream
}

func NewAesCfbEncipher(keyStr string, digit ...int) (*AesCfbEncipher, error) {
	key := convertKey(keyStr)
	if len(digit) == 1 {
		//密钥长度16,24,32分别对应AES-128,AES-192,AES-256算法
		if keyLen := digit[0]; keyLen == 16 || keyLen == 24 {
			key = key[:keyLen]
		}
	}
	cfb := new(AesCfbEncipher)
	cfb.key = make([]byte, len(key))
	copy(cfb.key, key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cfb.block = block
	cfb.cfbEncrypter = cipher.NewCFBEncrypter(cfb.block, cfb.key[:block.BlockSize()])
	cfb.cfbDecrypter = cipher.NewCFBDecrypter(cfb.block, cfb.key[:block.BlockSize()])
	return cfb, nil
}

func (cfb *AesCfbEncipher) Encrypt(data []byte) (cipherText []byte) {
	data = PCKS7Padding(data, cfb.block.BlockSize())
	cipherText = make([]byte, len(data))
	cfb.cfbEncrypter.XORKeyStream(cipherText, data)
	return cipherText
}
func (cfb *AesCfbEncipher) Decrypt(data []byte) (plainText []byte) {
	plainText = make([]byte, len(data))
	cfb.cfbDecrypter.XORKeyStream(plainText, data)
	plainText = PCKS7UnPadding(plainText)
	return
}

/*
PCKS7Padding填充算法用于填充不足的位数
填充由一个字节序列组成，其中每个字节的值等于其长度。
AES要求明文长度为128bit即16byte的整数倍。
故补位长度为 16 - len%16。
*/
func PCKS7Padding(cipher []byte, blockSize int) []byte {
	b := blockSize - len(cipher)%blockSize
	paddingBytes := bytes.Repeat([]byte{byte(b)}, b)
	return append(cipher, paddingBytes...)
}
func PCKS7UnPadding(plain []byte) []byte {
	b := plain[len(plain)-1]
	return plain[:len(plain)-int(b)]
}

func convertKey(keyStr string) []byte {
	key := sha256.Sum256([]byte(keyStr))
	return key[:]
}
