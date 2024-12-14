package xutil

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/h8848/blockchain-infra/chain/chain_client/tron"
	"strconv"
	"strings"
)

const (
	addressPrefix   = byte(0x41)
	emptyAddressHex = "410000000000000000000000000000000000000000"
	tronZeroAddress = "T9yD14Nj9j7xAB4dbGeiX9h8unkKHxuWwb"
)

var (
	emptyAddressBase58 = address.HexToAddress(emptyAddressHex).String()
)

func ETHToTronAddress(addr string) string {
	s := "41" + strings.TrimPrefix(addr, "0x")
	return address.HexToAddress(s).String()
}

func TronToEthAddress(addr string) (common.Address, error) {
	if common.IsHexAddress(addr) {
		return common.HexToAddress(addr), nil
	}
	value, err := address.Base58ToAddress(addr)
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid address, err=%s", err)
	}
	return common.HexToAddress(value.Hex()), nil
}

func IsTronAddress(addr string) bool {
	if addr == emptyAddressBase58 {
		return false
	}
	_, err := address.Base58ToAddress(addr)
	return err == nil
}

func TronZeroAddress(asset string) string {
	return tronZeroAddress
}

func IsTronNativeAsset(asset string) bool {
	return asset == emptyAddressBase58
}

func TronNativeAssetDecimals() uint8 {
	return 6
}

func TrxIsTransactionSuccessful(client *tron.TronClient, txHash common.Hash) (bool, uint64) {
	receipt, err := client.TransactionReceipt(txHash)
	if err != nil {
		fmt.Println("Error fetching transaction receipt:", err)
		return false, 0
	}
	return receipt.Status == types.ReceiptStatusSuccessful, receipt.BlockNumber.Uint64()
}

func HexToDec(hexStr string) uint64 {
	if len(hexStr) > 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}
	decValue, err := strconv.ParseUint(hexStr, 16, 64)
	if err != nil {
		return 0
	}

	return decValue
}
