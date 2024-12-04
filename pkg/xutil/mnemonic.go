package xutil

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/hashicorp/vault/shamir"
	"github.com/tyler-smith/go-bip39"
	"strconv"
)

func SplitNumber(num int64) (string, string, error) {
	// 将数字转换为字符串
	numStr := strconv.FormatInt(num, 10)
	// 计算中间位置
	mid := len(numStr) / 2
	// 拆分字符串
	leftPart := numStr[:mid]
	rightPart := numStr[mid:]
	// 去掉左边部分开头的所有 0
	for len(leftPart) > 0 && leftPart[0] == '0' {
		leftPart = leftPart[1:]
	}
	// 去掉右边部分开头的所有 0
	for len(rightPart) > 0 && rightPart[0] == '0' {
		rightPart = rightPart[1:]
	}

	return leftPart, rightPart, nil
}

func SplitToShares(privateKey string, minimumShares int, totalShares int) ([][]byte, error) {
	if privateKey == "" {
		return nil, fmt.Errorf("private key cannot be nil")
	}
	if minimumShares < 1 || minimumShares > totalShares {
		return nil, fmt.Errorf("invalid share count: minimum %d, total %d", minimumShares, totalShares)
	}

	shares, err := shamir.Split([]byte(privateKey), totalShares, minimumShares)
	if err != nil {
		return nil, fmt.Errorf("failed to split private key: %v", err)
	}
	return shares, nil
}

// GenMnemonic  生成12、24个的助记词；默认生成12个
func GenMnemonic(lang int) (string, error) {
	var entropy []byte
	var err error

	switch lang {
	case 24:
		entropy, err = bip39.NewEntropy(256)
		if err != nil {
			return "", err
		}
	case 12:
		entropy, err = bip39.NewEntropy(128)
		if err != nil {
			return "", err
		}
	default:
		entropy, err = bip39.NewEntropy(128)
		if err != nil {
			return "", err
		}
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	return mnemonic, err
}

// IsValidMnemonic 判断助记词是否有效
func IsValidMnemonic(mnemonic string) bool {
	_, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	return err == nil
}

// CombineShares shamir合并私钥
func CombineShares(shares [][]byte) ([]byte, error) {
	return shamir.Combine(shares)
}

// MnemonicToPrivateKey 助记词转私钥
func MnemonicToPrivateKey(mnemonic, salt, pathStr string) (*ecdsa.PrivateKey, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, salt)
	if err != nil {
		return nil, err
	}
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	path, err := accounts.ParseDerivationPath(pathStr)
	if err != nil {
		return nil, err
	}
	for _, n := range path {
		masterKey, err = masterKey.Derive(n)
		if err != nil {
			return nil, err
		}
	}
	privateKey, err := masterKey.ECPrivKey()
	if err != nil {
		return nil, err
	}
	return privateKey.ToECDSA(), nil
}
