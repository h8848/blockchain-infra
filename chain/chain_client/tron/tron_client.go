package tron

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	eABI "github.com/ethereum/go-ethereum/accounts/abi"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ecrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/abi"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/h8848/blockchain-infra/chain/chain_client"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	addressPrefix      = byte(0x41)
	emptyAddressHex    = "410000000000000000000000000000000000000000"
	transactionSuccess = "SUCCESS"
)

var (
	emptyAddressBase58 = address.HexToAddress(emptyAddressHex).String()
	Trc20ABIName       = "trc20"
	trc20Abi           = "[{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
	trc20Function      = map[string]string{"transferFrom": "transferFrom(address,address,uint256)", "balanceOf": "balanceOf(address)"}
)

// TronClient implements BlockChainClient Interface
type TronClient struct {
	c      *HTTPClient
	abiMap sync.Map
	//abiMap  map[string]*eABI.ABI
	chainID *big.Int
}

// NewTronClient creates the chain_client
func NewTronClient(config *chain_client.ChainConfiguration) (*TronClient, error) {
	c := TronClient{}
	c.abiMap = sync.Map{}
	if err := c.RegisterABI(Trc20ABIName, trc20Abi); err != nil {
		return nil, fmt.Errorf("register trc20 abi failed, err=%s", err)
	}

	c.c = NewHTTPClient(config.Endpoints[0], config.Endpoints[1], config.Endpoints[2])
	c.chainID = config.ChainID
	c.c.APIKey = config.APIKey
	return &c, nil
}

// RegisterABI registe the abi with a name
func (tc *TronClient) RegisterABI(name, abiStr string) error {
	compiled, err := eABI.JSON(strings.NewReader(abiStr))
	if err != nil {
		return err
	}
	tc.abiMap.Store(name, &compiled)
	return nil
}

// BalanceAt returns the amount of trx
func (tc *TronClient) BalanceAt(address string) (*big.Int, error) {
	//以太坊地址 -> Tron地址
	if ecommon.IsHexAddress(address) {
		address = tc.c.convertETHAddress(address)
	}

	balance, err := tc.c.BalanceAt(address)
	if err != nil {
		return nil, fmt.Errorf("http call failed, err=%s", err)
	}
	return balance, nil
}

// BalanceOf returns the amount of a token
func (tc *TronClient) BalanceOf(contract, from string) (*big.Int, error) {
	//以太坊地址 -> Tron地址
	if ecommon.IsHexAddress(from) {
		from = tc.c.convertETHAddress(from)
	}

	params, err := tc.generateParams("balanceOf", Trc20ABIName, from)
	if err != nil {
		return nil, fmt.Errorf("generate request from abi failed, err=%s", err)
	}
	parameter, err := abi.GetPaddedParam(params)
	if err != nil {
		return nil, fmt.Errorf("pack request failed, err=%s", err)
	}
	return tc.c.BalanceOf(contract, from, common.BytesToHexString(parameter))
}

// DecimalsOf returns the decimals of an contract
func (tc *TronClient) DecimalsOf(contract string) (uint8, error) {
	decimals, err := tc.c.DecimalsOf(contract)
	return uint8(decimals.Uint64()), err
}

// TotalSupplyOf returns the total supply of a contract
func (tc *TronClient) TotalSupplyOf(contract string) (*big.Int, error) {
	return tc.c.TotalSupplyOf(contract)
}

// SymbolOf returns the symbol of a contract
func (tc *TronClient) SymbolOf(contract string) (string, error) {
	return tc.c.SymbolOf(contract)
}

func (tc *TronClient) TransferData(to string, value *big.Int) ([]byte, error) {
	method := "transfer"
	return tc.GetTransactionDataByABI(method, Trc20ABIName, to, value)
}

func (tc *TronClient) ApproveData(contract, owner, spender string, amount *big.Int) ([]byte, error) {
	method := "approve"
	return tc.GetTransactionDataByABI(method, Trc20ABIName, spender, amount)
}

func (tc *TronClient) Allowance(contract, owner, spender string) (*big.Int, error) {
	method := "allowance"
	data, err := tc.GetTransactionDataByABI(method, Trc20ABIName, owner, spender)
	if err != nil {
		return nil, fmt.Errorf("get transaction data failed, err=%s", err)
	}
	result, err := tc.c.EthCall(owner, contract, big.NewInt(0), data)
	if err != nil {
		return nil, fmt.Errorf("http call failed, err=%s", err)
	}
	fields, err := tc.UnpackByABI(method, Trc20ABIName, result)
	if err != nil {
		return nil, fmt.Errorf("unpack failed, err=%s", err)
	}
	if len(fields) != 1 {
		return nil, fmt.Errorf("unpack result failed, fields=%d", len(fields))
	}
	return tc.AbiConvertToInt(fields[0]), nil
}

func (tc *TronClient) generateParams(method string, abiName string, args ...interface{}) (
	data []abi.Param, err error) {
	compiled, err := tc.GetABIByName(abiName)
	if err != nil {
		return nil, fmt.Errorf("get abi failed, err=%s", err)
	}
	return tc.generateFromAbi(method, compiled, args...)
}

func (tc *TronClient) generateFromAbi(method string, compiledAbi *eABI.ABI, args ...interface{}) (
	data []abi.Param, err error) {
	m, ok := compiledAbi.Methods[method]
	if !ok {
		return nil, fmt.Errorf("method=%s not found in abi", method)
	}
	if len(m.Inputs) != len(args) {
		return nil, fmt.Errorf("args=%d not match inputs+%d", len(args), len(m.Inputs))
	}
	result := make([]abi.Param, 0, len(m.Inputs))
	for i := range m.Inputs {
		argument := m.Inputs[i]
		typeName := argument.Type.String()
		param := abi.Param{}
		param[typeName] = args[i]
		result = append(result, param)
	}
	return result, nil
}

// GetTransactionData generate the data in transaction
func (tc *TronClient) GetTransactionData(method string, abiStr string, args ...interface{}) ([]byte, error) {
	compiled, err := eABI.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, fmt.Errorf("parse abi failed, err=%s", err)
	}
	params, err := tc.generateFromAbi(method, &compiled, args...)
	if err != nil {
		return nil, fmt.Errorf("generate params failed, err=%s", err)
	}
	data, err := abi.GetPaddedParam(params)
	if err != nil {
		return nil, fmt.Errorf("pack failed, err=%s", err)
	}
	return data, nil
}

// GetTransactionDataByABI GetERC20Transaction is similar with GetTransactionData, except it is using the build in abi
func (tc *TronClient) GetTransactionDataByABI(method, abiName string, args ...interface{}) (data []byte, err error) {
	compiled, err := tc.GetABIByName(abiName)
	if err != nil {
		return nil, fmt.Errorf("get abi failed, err=%s", err)
	}

	methodAbi, ok := compiled.Methods[method]
	if !ok {
		return nil, fmt.Errorf("method=%s not found in abi", method)
	}
	if len(methodAbi.Inputs) != len(args) {
		return nil, fmt.Errorf("args=%d not match inputs+%d", len(args), len(methodAbi.Inputs))
	}
	requests := make([]interface{}, 0, len(args))
	for i, input := range compiled.Methods[method].Inputs {
		if input.Type.String() == "address" {
			v := reflect.ValueOf(args[i])
			if v.Kind() == reflect.String {
				addr, err := address.Base58ToAddress(args[i].(string))
				if err != nil {
					return nil, fmt.Errorf("parse address failed, err=%s", err)
				}
				requests = append(requests, ecommon.HexToAddress(addr.Hex()))
			} else {
				requests = append(requests, args[i])
			}
		} else {
			requests = append(requests, args[i])
		}
	}
	return compiled.Pack(method, requests...)
}

func (tc *TronClient) GetFunctionSelectorByData(abiName string, data []byte) (sig string, err error) {
	compiled, err := tc.GetABIByName(abiName)
	if err != nil {
		return "", fmt.Errorf("get abi failed, err=%s", err)
	}

	if len(data) < 4 {
		return "", fmt.Errorf("data is too short, len=%d <= 4", len(data))
	}

	sigData := data[:4]
	method, err := compiled.MethodById(sigData)
	if err != nil {
		return "", fmt.Errorf("method not found, err=%s", err)
	}
	return method.Sig, nil
}

func (tc *TronClient) getFunctionSelector(abiName, method string) string {
	compiled, err := tc.GetABIByName(abiName)
	if err != nil {
		fmt.Printf("get abi failed, err=%s", err)
		return ""
	}
	methodAbi, ok := compiled.Methods[method]
	if !ok {
		return ""
	}
	return methodAbi.Sig
}

// GetTransaction returns the unsigned transaction and the hash value
func (tc *TronClient) GetTransaction(td *chain_client.Transaction) ([]byte, []byte, error) {
	var tx *TransactionExtention
	var err error
	if len(td.Data) == 0 {
		tx, err = tc.c.TriggerTransfer(td.From, td.To, td.Amount)
	} else {
		energyLimit := big.NewInt(0)
		if td.Fee != nil && td.Fee.Gas != nil && td.Fee.GasFeeCap != nil {
			energyLimit = big.NewInt(1).Mul(td.Fee.Gas, td.Fee.GasFeeCap)
			energyLimit = energyLimit.Add(energyLimit, big.NewInt(int64(len(td.Data)/2)))
		}
		tx, err = tc.c.TriggerSmartContract(td.To, td.From, td.Data, energyLimit)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("triggersmartcontract failed, err=%s", err)
	}
	return tc.getTransactionExtensionData(tx)
}

// TODO, this function is not finished yet, need to refer to
// https://github.com/tronprotocol/sun-network/blob/develop/js-sdk/src/index.js
// to change the sign functions to make it work
func (tc *TronClient) DeployContract(contractAbi, contractBin string, td *chain_client.Transaction) (
	[]byte, []byte, string, error) {
	tx, addr, err := tc.c.DeployContract(contractAbi, contractBin, td.From, "Migrations")
	if err != nil {
		return nil, nil, "", fmt.Errorf("try deploy failed, err=%s", err)
	}
	message, err := json.Marshal(tx)
	if err != nil {
		return nil, nil, "", fmt.Errorf("encode transaction failed, err=%s", err)
	}
	hash, err := hex.DecodeString(tx.Txid)
	if err != nil {
		return nil, nil, "", fmt.Errorf("decode txid to hash failed, err=%s", err)
	}
	return message, hash, addr, nil
}

// BroadcastTransaction broadcasts the transaction to chain
func (tc *TronClient) BroadcastTransaction(trans []byte, signature []byte) ([]byte, error) {
	tx := TransactionExtention{}
	d := json.NewDecoder(bytes.NewReader(trans))
	d.UseNumber()
	if err := d.Decode(&tx); err != nil {
		return nil, fmt.Errorf("transaction format is incorrect, err=%s", err)
	}
	txid, err := hex.DecodeString(tx.Txid)
	if err != nil {
		return nil, fmt.Errorf("decode txid failed, err=%s", err)
	}
	sig := hex.EncodeToString(signature)
	transaction := TronTransaction{}
	transaction.RawData = tx.Transaction.RawData
	transaction.RawDataHex = tx.Transaction.RawDataHex
	transaction.Signature = []string{sig}
	transaction.ContractAddress = tx.Transaction.ContractAddress
	transaction.Visible = tx.Transaction.Visible
	transaction.Txid = string(tx.Txid)
	return txid, tc.c.BroadCastTransaction(&transaction)
}

// GetNonce is not implemented for Tron
// And Tron is not used by Tron
func (tc *TronClient) GetNonce(address string) (uint64, error) {
	return 0, nil
}

func (tc *TronClient) GetNonceByNumber(address string, blockNumber *big.Int) (uint64, error) {
	return 0, nil
}

// GetSuggestFee returns the estimated fee for a transaction
func (tc *TronClient) GetSuggestFee(td *chain_client.Transaction) (*chain_client.FeeLimit, error) {
	gas, err := tc.EstimateGas(td)
	if err != nil {
		return nil, fmt.Errorf("estimate gas failed, err=%s", err)
	}
	gasPrice, _, err := tc.GetGasPrice()
	if err != nil {
		return nil, fmt.Errorf("get gas price failed, err=%s", err)
	}
	fee := chain_client.FeeLimit{}
	fee.Gas = new(big.Int).SetUint64(gas)
	fee.GasFeeCap = gasPrice
	fee.GasTipCap = big.NewInt(0)
	return &fee, nil
}

func (tc *TronClient) EstimateGas(td *chain_client.Transaction) (uint64, error) {
	if len(td.Data) <= 0 {
		return 0, nil
	}

	gasLimit, err := tc.c.EstimateGas(td.From, td.To, "0x"+td.Amount.Text(16), td.Data)
	if err != nil {
		return 0, fmt.Errorf("eth estimategas failed, err=%s", err)
	}
	return gasLimit.Uint64(), nil
}

// CallContract call eth_call
func (tc *TronClient) CallContract(td *chain_client.Transaction) ([]byte, error) {
	return tc.c.EthCall(td.From, td.To, td.Amount, td.Data)
}

func (tc *TronClient) UnpackByABI(method, name string, data []byte) ([]interface{}, error) {
	compiled, err := tc.GetABIByName(name)
	if err != nil {
		return nil, fmt.Errorf("get abi by name failed, err=%s", err)
	}
	return compiled.Unpack(method, data)
}

func (tc *TronClient) GetABIByName(name string) (*eABI.ABI, error) {
	cabi, ok := tc.abiMap.Load(name)
	if !ok {
		return nil, fmt.Errorf("abi=%s not found", name)
	}
	compiled := cabi.(*eABI.ABI)
	return compiled, nil
}

func (tc *TronClient) AbiConvertToInt(v interface{}) *big.Int {
	return *eABI.ConvertType(v, new(*big.Int)).(**big.Int)
}

func (tc *TronClient) AbiConvertToString(v interface{}) string {
	return *eABI.ConvertType(v, new(string)).(*string)
}

func (tc *TronClient) AbiConvertToBytes(v interface{}) []byte {
	value := eABI.ConvertType(v, new([]byte)).(*[]byte)
	return *value
}

func (tc *TronClient) AbiConvertToAddress(v interface{}) string {
	value := eABI.ConvertType(v, new(ecommon.Address)).(*ecommon.Address)
	return value.Hex()
}

// extract string from event log
func getString(value any) string {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.String {
		return v.String()
	}
	return ""
}

func (tc *TronClient) AddressFromPrivateKey(privateKey string) (string, error) {
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}
	key, err := ecrypto.HexToECDSA(privateKey)
	if err != nil {
		return "", fmt.Errorf("wrong private key=%s, err=%s", privateKey, err)
	}
	return tc.AddressFromPublicKey(&key.PublicKey)
}

func (tc *TronClient) AddressFromPublicKey(pubKey *ecdsa.PublicKey) (string, error) {
	addrByte := ecrypto.PubkeyToAddress(*pubKey)
	addr := make([]byte, 0, 32)
	addr = append(addr, addressPrefix)
	addr = append(addr, addrByte.Bytes()...)
	return address.HexToAddress(hex.EncodeToString(addr)).String(), nil
}

func hexToBase58(hexAddr string) string {
	hexForm := ecommon.FromHex(hexAddr)
	if len(hexForm) == 0 {
		return emptyAddressBase58
	}
	addressHex := hexForm
	if hexForm[0] != addressPrefix {
		addressHex = append([]byte{addressPrefix}, hexForm...)
	}
	return address.HexToAddress(hex.EncodeToString(addressHex)).String()
}

func (tc *TronClient) GetLatestBlockNumber() (*big.Int, error) {
	return tc.c.GetBlockByLastNumber()
}

// GetBlockByNumber returns block info
func (tc *TronClient) GetBlockByNumber(num *big.Int) (map[string]interface{}, error) {
	//以太坊地址 -> Tron地址
	return tc.c.GetBlockByNumber(num)
}

func (tc *TronClient) GetTransactionByHash(transactionHash string) (*chain_client.TransactionInfo, error) {
	if strings.HasPrefix(transactionHash, "0x") {
		transactionHash = transactionHash[2:]
	}
	info := chain_client.TransactionInfo{}
	tx := chain_client.Transaction{}
	info.Tx = &tx

	txInfo, err := tc.c.GetTransactionInfoByID(transactionHash)
	if err != nil {
		return nil, fmt.Errorf("get transaction info failed, err=%s", err)
	}
	transaction, err := tc.c.GetTransactionByID(transactionHash)
	if err != nil {
		return nil, fmt.Errorf("get transaction failed, err=%s", err)
	}
	if transaction.RawData == nil || len(transaction.Ret) == 0 {
		// maybe not finished, but we don't know
		return &info, nil
	}
	if len(transaction.RawData.Contract) == 0 {
		info.Status, info.Error = chain_client.TransactionStatusInvalid, "no_contract_found_in_raw_data"
		return nil, fmt.Errorf("transaction contract not found")
	}
	fee := chain_client.FeeLimit{
		Gas:       txInfo.Fee,
		GasFeeCap: big.NewInt(1),
		GasTipCap: big.NewInt(0),
	}
	tx.Fee = &fee

	info.Gas = &chain_client.TxGasInfo{
		Fee: txInfo.Fee,
	}

	callData, ok := transaction.RawData.Contract[0].Parameter.Value["data"]
	if ok {
		callDataStr, success := callData.(string)
		if !success {
			info.Status, info.Error = chain_client.TransactionStatusInvalid, "call_data_is_not_string"
			return nil, fmt.Errorf("transaction data format is incorrect")
		}
		tx.Data, err = hex.DecodeString(callDataStr)
		if err != nil {
			info.Status, info.Error = chain_client.TransactionStatusInvalid, "call_data_decode_failed"
			return nil, fmt.Errorf("transaction data decode failed, err=%s", err)
		}
	}
	info.IsPending = true
	if len(transaction.RawData.Contract) > 0 {
		value := transaction.RawData.Contract[0].Parameter.Value
		tx.From = hexToBase58(getString(value["owner_address"]))
		tx.To = hexToBase58(getString(value["to_address"]))
		if tx.To == emptyAddressBase58 {
			tx.To = hexToBase58(getString(value["contract_address"]))
		}
	}
	tx.ChainID = tc.chainID
	if txInfo.BlockNumber != nil {
		info.IsPending = false
	}
	if strings.EqualFold(transaction.Ret[0].ContractRet, "SUCCESS") {
		info.Status = chain_client.TransactionStatusSuccess
	} else {
		info.Status = chain_client.TransactionStatusFailed
	}
	logs, err := tc.c.GetTransactionEventsByID(transactionHash)
	if err != nil {
		return nil, fmt.Errorf("get event logs failed, err=%s", err)
	}
	for i := range logs.Data {
		if logs.Data[i] == nil {
			continue
		}
		tronEvent, err := json.Marshal(logs.Data[i].Results)
		if err != nil {
			return nil, fmt.Errorf("encode json failed, err=%s", err)
		}
		logData := logs.Data[i]
		topic := []byte(logData.EventName)
		event := chain_client.EventLog{
			Address: logData.ContractAddress,
			Topics:  [][]byte{topic},
			Data:    tronEvent,
		}
		info.Logs = append(info.Logs, &event)
	}
	return &info, nil
}

func (tc *TronClient) ParseEventLog(abiName string, eventLog *chain_client.EventLog) ([]interface{}, error) {
	event := TronEvent{}
	if err := json.Unmarshal(eventLog.Data, &event); err != nil {
		return nil, fmt.Errorf("decode event failed, err=%s", err)
	}
	results := make([]interface{}, 0, len(event.Results))
	for i := 0; i < len(event.Results); i++ {
		key := strconv.Itoa(i)
		if value, ok := event.Results[key]; ok {
			results = append(results, value)
		} else {
			break
		}
	}
	return results, nil
}

func (tc *TronClient) IsValidAddress(addr string) bool {
	if addr == emptyAddressBase58 {
		return false
	}
	_, err := address.Base58ToAddress(addr)
	return err == nil
}

// AddressFromString convert base58 address hexed address
// 兼容以太坊地址
func (tc *TronClient) AddressFromString(addr string) (ecommon.Address, error) {
	if ecommon.IsHexAddress(addr) {
		return ecommon.HexToAddress(addr), nil
	}

	value, err := address.Base58ToAddress(addr)
	if err != nil {
		return ecommon.Address{}, fmt.Errorf("invalid address, err=%s", err)
	}
	return ecommon.HexToAddress(value.Hex()), nil
}

func (tc *TronClient) AddressToString(addr ecommon.Address) string {
	addressTron := make([]byte, 0)
	addressTron = append(addressTron, address.TronBytePrefix)
	addressTron = append(addressTron, addr.Bytes()...)
	a := address.Address(addressTron)
	return a.String()
}

func (tc *TronClient) ContractAddress(addr ecommon.Address) (bool, error) {
	code, err := tc.c.GetCode(addr)
	if err != nil {
		return false, fmt.Errorf("get code failed, err=%s", err)
	}
	return len(code) > 0, nil
}

func (tc *TronClient) IsNativeAsset(asset string) bool {
	return asset == emptyAddressBase58
}

func (tc *TronClient) NormalizeAddress(addr string) string {
	if addr == "" {
		return ""
	}

	// 兼容以太坊地址 -> 自动转为tron地址
	if ecommon.IsHexAddress(addr) {
		return tc.c.convertETHAddress(addr)
	}

	// tron hex address -> base58 address
	if strings.HasPrefix(addr, "41") {
		return address.HexToAddress(addr).String()
	}

	// check tron address
	base58Addr, err := address.Base58ToAddress(addr)
	if err != nil {
		fmt.Printf("address[%s] base58 check failed, err=%s\n", addr, err)
		return ""
	}
	return base58Addr.String()
}

func (tc *TronClient) GetGasPrice() (*big.Int, *big.Int, error) {
	gasPrice, err := tc.c.GetGasPrice()
	return gasPrice, big.NewInt(0), err
}

func (tc *TronClient) GetSuggestGasPrice() (*big.Int, *big.Int, *big.Int, error) {
	gasPrice, err := tc.c.GetGasPrice()
	return big.NewInt(0), big.NewInt(0), gasPrice, err
}

func (tc *TronClient) NativeAssetAddress() string {
	return emptyAddressBase58
}

func (tc *TronClient) PublicKeyHexToAddress(key string) (string, error) {
	buffer, err := hex.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("decode public key failed, err=%s", err)
	}
	pubKey, err := ecrypto.UnmarshalPubkey(buffer)
	if err != nil {
		return "", fmt.Errorf("unmarshal public key failed, err=%s", err)
	}
	return address.PubkeyToAddress(*pubKey).String(), nil
}

func (tc *TronClient) GetLackedGas(address string, gas uint64, gasPrice *big.Int, txSize uint64) (*big.Int, error) {
	_, _, err := tc.c.GetAccountResource(address)
	return nil, err
}

func (tc *TronClient) NativeAssetDecimals() uint8 {
	return 6
}

func (tc *TronClient) GenerateStackTransactionData(from string, resource string, amount *big.Int) ([]byte,
	[]byte, error) {
	tx, err := tc.c.TriggerStack(from, resource, amount)
	if err != nil {
		return nil, nil, err
	}
	return tc.getTransactionExtensionData(tx)
}

func (tc *TronClient) GenerateUnStackTransactionData(from, resource string, amount *big.Int) ([]byte, []byte, error) {
	tx, err := tc.c.TriggerUnStack(from, resource, amount)
	if err != nil {
		return nil, nil, err
	}
	return tc.getTransactionExtensionData(tx)
}

func (tc *TronClient) GetWithdrawUnStackData(from string) ([]byte, []byte, error) {
	tx, err := tc.c.TriggerWithdrawUnStack(from)
	if err != nil {
		return nil, nil, err
	}
	return tc.getTransactionExtensionData(tx)
}

func (tc *TronClient) getTransactionExtensionData(tx *TransactionExtention) ([]byte, []byte, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("encode rawdata failed, err=%s", err)
	}
	hash, err := hex.DecodeString(tx.Txid)
	if err != nil {
		return nil, nil, fmt.Errorf("decode txid to hash failed, err=%s", err)
	}
	return data, hash, nil
}

func (tc *TronClient) GenerateDelegateResourceTransactionData(from, to, resource string, amount *big.Int) ([]byte, []byte,
	error) {
	tx, err := tc.c.TriggerDelegateResource(from, to, resource, amount)
	if err != nil {
		return nil, nil, err
	}
	return tc.getTransactionExtensionData(tx)
}

// ChainID returns chainID
func (tc *TronClient) ChainID() (*big.Int, error) {
	chainId, err := tc.c.ChainID()
	if err != nil {
		return nil, fmt.Errorf("http call failed, err=%s", err)
	}
	return chainId, nil
}

func (tc *TronClient) TransactionReceipt(hash ecommon.Hash) (*types.Receipt, error) {
	tx, err := tc.c.TransactionReceipt(hash)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (tc *TronClient) BlockNumber() (*big.Int, error) {
	blockNumber, err := tc.c.BlockNumber()
	if err != nil {
		return nil, fmt.Errorf("http call failed, err=%s", err)
	}
	return blockNumber, nil
}
