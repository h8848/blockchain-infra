package xutil

import (
	"context"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/mr-tron/base58"
)

// IsSolAddress 判断是否为合法的 SOL 地址
func IsSolAddress(address string) bool {
	// SOL 地址应该是 Base58 编码，长度为 32 字节
	pubKeyBytes, err := base58.Decode(address)
	if err != nil {
		return false
	}
	return len(pubKeyBytes) == 32
}

func SolIsTransactionSuccessful(txHash string, solClient *client.Client) (bool, error) {
	// 获取交易信息
	solTx, err := solClient.GetTransaction(context.Background(), txHash)
	if err != nil {
		return false, fmt.Errorf("failed to get transaction: %v", err)
	}

	// 判断交易的 meta 信息中的 Err 字段是否为 nil
	if solTx.Meta.Err == nil {
		return true, nil
	}

	// 交易失败
	return false, fmt.Errorf("transaction failed with error: %v", solTx.Meta.Err)
}
