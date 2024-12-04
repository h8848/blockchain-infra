package xaes

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
)

func ECBEncrypt(word string, key []byte) (string, error) {
	encryptString := []byte(word)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	paddedPlaintext := PKCS7Padding(encryptString, block.BlockSize())
	blockMode := NewECBEncrypter(block)
	ciphertext := make([]byte, len(paddedPlaintext))
	blockMode.CryptBlocks(ciphertext, paddedPlaintext)
	return hex.EncodeToString(ciphertext), nil
}

func ECBDecrypt(ciphertext string, key []byte) (string, error) {
	decodeString, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockMode := NewECBDecrypter(block)
	word := make([]byte, len(decodeString))
	blockMode.CryptBlocks(word, decodeString)
	unpaddedPlaintext, err := PKCS7UnPadding(word, block.BlockSize())
	if err != nil {
		return "", err
	}
	return string(unpaddedPlaintext), nil
}

type ecb struct {
	b         cipher.Block
	blockSize int
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

type ecbEncrypter ecb

func NewECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(newECB(b))
}

func (x *ecbEncrypter) BlockSize() int { return x.blockSize }

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.b.BlockSize() != 0 {
		panic("input not full blocks")
	}
	if len(dst) < len(src) {
		panic("output smaller than input")
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.b.BlockSize()])
		src = src[x.b.BlockSize():]
		dst = dst[x.b.BlockSize():]
	}
}

type ecbDecrypter ecb

func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypter)(newECB(b))
}
func (x *ecbDecrypter) BlockSize() int { return x.blockSize }

func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.b.BlockSize() != 0 {
		panic("input not full blocks")
	}
	if len(dst) < len(src) {
		panic("output smaller than input")
	}
	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.b.BlockSize()])
		src = src[x.b.BlockSize():]
		dst = dst[x.b.BlockSize():]
	}
}
