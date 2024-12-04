package xaes

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

func CBCEncryptPlus(key, iv []byte, plaintext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	paddedPlaintext := PKCS7Padding(plaintext, block.BlockSize())
	ciphertext := make([]byte, len(paddedPlaintext))

	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(ciphertext, paddedPlaintext)
	// 拆分密文
	arr := arrayInGroupsOf(ciphertext, 4)
	if len(arr) != 4 {
		return "", errors.New("invalid ciphertext")
	}
	// 交换位置
	// 0 1 2 3 => 2 3 0 1
	// 2 3 0 1 => 3 2 1 0
	// 3 2 1 0 => 3 0 1 2
	// 3 0 1 2 => 3 1 0 2
	// nolint
	arr[0], arr[1], arr[2], arr[3] = arr[3], arr[1], arr[0], arr[2]
	return base64.StdEncoding.EncodeToString(arrayTo(arr)), nil
}

func CBCDecryptPlus(key, iv []byte, ciphertext string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	arr := arrayInGroupsOf(b, 4)
	if len(arr) != 4 {
		return nil, errors.New("invalid ciphertext")
	}
	// 交换密文
	// 0 1 2 3 => 0 2 1 3
	// 0 2 1 3 => 0 3 1 2
	// 0 3 1 2 => 3 0 2 1
	// 3 0 2 1 => 2 1 3 0
	// nolint
	arr[0], arr[1], arr[2], arr[3] = arr[2], arr[1], arr[3], arr[0]
	ciarr := arrayTo(arr)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(ciarr))
	cbc.CryptBlocks(decrypted, ciarr)

	unpaddedPlaintext, err := PKCS7UnPadding(decrypted, block.BlockSize())
	if err != nil {
		return nil, err
	}

	return unpaddedPlaintext, nil
}

func arrayInGroupsOf(arr []byte, num int) [][]byte {
	partSize := (len(arr) + num - 1) / num
	var result [][]byte

	for i := 0; i < len(arr); i += partSize {
		end := i + partSize
		if end > len(arr) {
			end = len(arr)
		}
		a := arr[i:end]
		result = append(result, a)
	}

	return result
}

func arrayTo(arr [][]byte) []byte {
	b := make([]byte, 0)
	for i := 0; i < len(arr); i++ {
		b = append(b, arr[i]...)
	}
	return b
}
