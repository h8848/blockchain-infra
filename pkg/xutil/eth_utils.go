package xutil

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shopspring/decimal"
	"math/big"
	"reflect"
	"regexp"
	"time"
)

func EthZeroAddress() string {
	return "0x0000000000000000000000000000000000000000"
}
func IsEthAddress(address string) bool {
	return common.IsHexAddress(address)
}

// IsValidAddress validate hex address
func IsValidAddress(iaddress interface{}) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	switch v := iaddress.(type) {
	case string:
		return re.MatchString(v)
	case common.Address:
		return re.MatchString(v.Hex())
	default:
		return false
	}
}

// IsZeroAddress validate if it's a 0 address
func IsZeroAddress(iaddress interface{}) bool {
	var address common.Address
	switch v := iaddress.(type) {
	case string:
		address = common.HexToAddress(v)
	case common.Address:
		address = v
	default:
		return false
	}

	zeroAddressBytes := common.FromHex("0x0000000000000000000000000000000000000000")
	addressBytes := address.Bytes()
	return reflect.DeepEqual(addressBytes, zeroAddressBytes)
}

// ToDecimal wei to decimals
func ToDecimal(ivalue interface{}, decimals int) decimal.Decimal {
	value := new(big.Int)
	switch v := ivalue.(type) {
	case string:
		if _, ok := value.SetString(v, 10); !ok {
			return decimal.Zero // Return zero if the string cannot be parsed
		}
	case *big.Int:
		value = v
	default:
		return decimal.Zero // Return zero if the type is unsupported
	}

	// Convert 10^decimals to decimal.Decimal for division
	mul := decimal.NewFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil), 0)
	num := decimal.NewFromBigInt(value, 0)
	result := num.Div(mul)

	return result
}

// ToWei decimals to wei
func ToWei(value interface{}, decimals int) *big.Int {
	var amount decimal.Decimal

	// 处理输入的不同类型
	switch v := value.(type) {
	case string:
		amount, _ = decimal.NewFromString(v) // 从字符串创建 decimal
	case float64:
		amount = decimal.NewFromFloat(v) // 从 float64 创建 decimal
	case float32:
		amount = decimal.NewFromFloat32(v)
	case int64:
		amount = decimal.NewFromInt(v) // 使用整数类型创建 decimal，避免浮点数转换
	case int:
		amount = decimal.NewFromInt(int64(v)) // 使用整数类型创建 decimal
	case decimal.Decimal:
		amount = v // 直接使用 decimal
	case *decimal.Decimal:
		amount = *v // 指针形式的 decimal
	default:
		return nil // 无法识别的类型，返回 nil
	}

	// 计算10^decimals
	mul := decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(decimals)))

	// 将代币数量乘以10^decimals，得到的结果转为 *big.Int
	result := amount.Mul(mul)
	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}

func ChainNameToChainID(chainName string) string {
	chainMap := make(map[string]string)
	chainMap["ETH"] = "1"
	chainMap["ETH Sepolia"] = "11155111"
	chainMap["MATIC"] = "137"
	chainMap["XDAI"] = "100"
	chainMap["BNB"] = "56"
	chainMap["TRX"] = "1000001"
	chainMap["TRX TestNet"] = "3448148188"

	chainMap["AVAX"] = "43114"
	chainMap["FTM"] = "250"
	chainMap["ARB"] = "42161"
	chainMap["OPT"] = "10"
	return chainMap[chainName]
}

func ChainIdToChainName(chainId string) string {
	chainMap := make(map[string]string)
	chainMap["1"] = "ETH"
	chainMap["11155111"] = "ETH Sepolia"
	chainMap["137"] = "MATIC"
	chainMap["100"] = "XDAI"
	chainMap["56"] = "BNB"
	chainMap["1000001"] = "TRX"
	chainMap["3448148188"] = "TRX TestNet"
	chainMap["43114"] = "AVAX"
	chainMap["250"] = "FTM"
	chainMap["42161"] = "ARB"
	chainMap["10"] = "OPT"
	return chainMap[chainId]
}

func EthUrlConnTest(client *rpc.Client) (time.Duration, error) {
	if client == nil {
		return time.Millisecond * 99999, errors.New("chain_client is nil")
	}
	startTime := time.Now()
	//TODO 拼url， url+token，username，password
	var blockNumber string
	err := client.Call(&blockNumber, "eth_blockNumber")
	if err != nil {
		return time.Millisecond * 99999, err
	}

	elapsedTime := time.Since(startTime)
	return elapsedTime, nil
}
