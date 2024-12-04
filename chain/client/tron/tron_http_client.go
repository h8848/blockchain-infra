package tron

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/h8848/blockchain-infra/chain/client/ethevent"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func hex2UInt(hexStr string) (uint64, error) {
	cleaned := strings.TrimPrefix(hexStr, "0x")
	if cleaned == "" {
		return 0, nil
	}

	result, err := strconv.ParseUint(cleaned, 16, 64)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// hex2BigInt converts a hex string to a big.Int
func hex2BigInt(hexStr string) (*big.Int, error) {
	// remove "0x" prefix if it exists
	if len(hexStr) > 1 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	// create a new big.Int and set it from the hex string
	value := new(big.Int)
	_, ok := value.SetString(hexStr, 16)
	if !ok {
		return nil, fmt.Errorf("invalid hex string: %s", hexStr)
	}
	return value, nil
}

// extractNumber extract a number from the returned string
// such as constant_result part, which are consists of hexed value
// padding with zeros
func extractNumber(data string) (*big.Int, error) {
	if common.Has0xPrefix(data) {
		data = data[2:]
	}
	if len(data) == 64 {
		var n big.Int
		_, ok := n.SetString(data, 16)
		if ok {
			return &n, nil
		}
	}
	return nil, fmt.Errorf("Cannot parse %s", data)
}

// extractString extract a string from the returned string
// such as constant_result, sometimes there's an encoding there
func extractString(data string) (string, error) {
	if common.Has0xPrefix(data) {
		data = data[2:]
	}
	if len(data) > 128 {
		n, _ := extractNumber(data[64:128])
		if n != nil {
			l := n.Uint64()
			if 2*int(l) <= len(data)-128 {
				b, err := hex.DecodeString(data[128 : 128+2*l])
				if err == nil {
					return string(b), nil
				}
			}
		}
	} else if len(data) == 64 {
		// allow string properties as 32 bytes of UTF-8 data
		b, err := hex.DecodeString(data)
		if err == nil {
			i := bytes.Index(b, []byte{0})
			if i > 0 {
				b = b[:i]
			}
			if utf8.Valid(b) {
				return string(b), nil
			}
		}
	}
	return "", fmt.Errorf("Cannot parse %s,", data)
}

// jsonRPCRequest is the structure for calling tron json-rpc apis
type jsonRPCRequest struct {
	JsonRPC string        `json:"jsonrpc,omitempty"`
	Method  string        `json:"method,omitempty"`
	Params  []interface{} `json:"params,omitempty"`
	ID      uint64        `json:"id,omitempty"`
}

type jsonRPCReponse struct {
	JsonRPC string `json:"jsonrpc,omitempty"`
	Result  string `json:"result,omitempty"`
	ID      uint64 `json:"id,omitempty"`
	Error   struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`
}

// walletRequest is the structure for calling contract related apis
type walletRequest struct {
	OwnerAddress     string   `json:"owner_address,omitempty"`
	ContractAddress  string   `json:"contract_address,omitempty"`
	FunctionSelector string   `json:"function_selector,omitempty"`
	Parameter        string   `json:"parameter,omitempty"`
	Visible          bool     `json:"visible"`
	FeeLimit         *big.Int `json:"fee_limit,omitempty"`
}

// contractRequest is the structure for calling create contract
type contractRequest struct {
	OwnereAddress           string `json:"owner_address"`
	ABI                     string `json:"abi"`
	Bytecode                string `json:"bytecode"`
	FeeLimit                int32  `json:"fee_limit"`
	Parameter               string `json:"parameter"`
	OriginEnergyLimit       int32  `json:"origin_energy_limit"`
	Name                    string `json:"name"`
	ConsumerResourcePercent int32  `json:"consume_user_resource_percent"`
	CallValue               int32  `json:"call_value"`
}

type deployResponse struct {
	Txid            string          `json:"txID,omitempty"`
	ContractAddress string          `json:"contract_address,omitempty"`
	RawData         *TransactionRaw `json:"raw_data,omitempty"`
	RawDataHex      string          `json:"raw_data_hex,omitempty"`
	Visible         bool            `json:"visible"`
}

// walletTransactionRequest is the structure for getting transaction related information
type walletTransactionRequest struct {
	Value string `json:"value"`
}

type walletResult struct {
	Result         walletResultMessage `json:"result"`
	ConstantResult []string            `json:"constant_result"`
	EnergyUsed     int64               `json:"energy_used"`
	EnergyPenalty  int64               `json:"energy_penalty"`
}

type walletResultMessage struct {
	Ok      bool   `json:"result"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func initJsonRequest(method string, r *jsonRPCRequest) {
	r.ID = 2023
	r.JsonRPC = "2.0"
	r.Method = method
}

// HTTPClient is the client to call tron http apis
type HTTPClient struct {
	client   *http.Client
	endPoint string
	APIKey   string
	fullnode string
	trongrid string
}

// NewHTTPClient creates the client
// Endpoint is the node address for http apis
func NewHTTPClient(Endpoint, FullNode, TronGrid string) *HTTPClient {
	c := http.Client{}
	return &HTTPClient{client: &c, endPoint: Endpoint, fullnode: FullNode, trongrid: TronGrid}
}

// rpcGet used for json-rpc
func (c *HTTPClient) rpcGet() ([]byte, error) {
	url := c.endPoint
	return c.get(url)
}

func (c *HTTPClient) fullnodeGet(path string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.fullnode, path)
	return c.get(url)
}

func (c *HTTPClient) gridGet(path string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.trongrid, path)
	return c.get(url)
}

func (c *HTTPClient) get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed, err=%s", err)
	}
	req.Header.Set("TRON-PRO-API-KEY", c.APIKey)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call url=%s failed, err=%s", url, err)
	}
	defer resp.Body.Close()
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed, err=%s", err)
	}
	return res, nil
}

func (c *HTTPClient) rpcPost(body interface{}) ([]byte, error) {
	url := c.endPoint
	return c.post(url, body)
}

func (c *HTTPClient) gridPost(path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.trongrid, path)
	return c.post(url, body)
}

func (c *HTTPClient) fullnodePost(path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.fullnode, path)
	return c.post(url, body)
}

func (c *HTTPClient) post(url string, body interface{}) ([]byte, error) {
	js, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encode json failed, err=%s", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(js))
	if err != nil {
		return nil, fmt.Errorf("create request failed, err=%s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("TRON-PRO-API-KEY", c.APIKey)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call url=%s failed, req=%s, err=%s", url, js, err)
	}
	defer resp.Body.Close()
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed, url=%s, req=%s, err=%s", url, js, err)
	}

	//打印请求相应参数的日志
	//log.Printf("post_request: url=%s, req=%s, res=%s \n", url, js, res)
	return res, nil
}

type transferJsonRequest struct {
	From    string   `json:"owner_address"`
	To      string   `json:"to_address"`
	Amount  *big.Int `json:"amount"`
	Visible bool     `json:"visible"`
}

func (c *HTTPClient) TriggerTransfer(from, to string, amount *big.Int) (*TransactionExtention, error) {
	//以太坊地址 -> Tron地址
	if ecommon.IsHexAddress(from) {
		from = c.convertETHAddress(from)
	}

	if ecommon.IsHexAddress(to) {
		to = c.convertETHAddress(to)
	}

	//from <> to
	if from == to {
		return nil, fmt.Errorf("from address[%s] == to address", from)
	}

	//地址校验
	fromAddr, err := address.Base58ToAddress(from)
	if err != nil {
		return nil, fmt.Errorf("from address invalid , err=%s", err)
	}

	toAddr, err := address.Base58ToAddress(to)
	if err != nil {
		return nil, fmt.Errorf("to address invalid, err=%s", err)
	}

	//若amount为0，则报错返回
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("amount <= 0")
	}

	req := transferJsonRequest{
		From:    fromAddr.String(),
		To:      toAddr.String(),
		Amount:  amount,
		Visible: true,
	}
	response, err := c.fullnodePost("wallet/createtransaction", req)
	if err != nil {
		return nil, fmt.Errorf("post request failed, err=%s", err)
	}
	tx := TronTransaction{}
	d := json.NewDecoder(bytes.NewReader(response))
	d.UseNumber()
	if err := d.Decode(&tx); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s, resp=%s", err, string(response))
	}
	if tx.Txid == "" {
		return nil, fmt.Errorf("parse result failed, js=%s", string(response))
	}
	txe := TransactionExtention{
		Transaction: &tx,
		Txid:        tx.Txid,
	}
	return &txe, nil
}

func (c *HTTPClient) GetBlockByLastNumber() (*big.Int, error) {
	url := "wallet/getblockbylatestnum?num=1"
	response, err := c.fullnodeGet(url)
	if err != nil {
		return nil, fmt.Errorf("get request failed, err=%s", err)
	}
	info := struct {
		Block []struct {
			BlockID     string `json:"blockID"`
			BlockHeader struct {
				RawData struct {
					Number    int64 `json:"number"`
					Timestamp int64 `json:"timestamp"`
				} `json:"raw_data"`
			} `json:"block_header"`
		} `json:"block"`
	}{}
	if err := json.Unmarshal(response, &info); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}
	if len(info.Block) == 0 {
		return nil, fmt.Errorf("parse result failed, js=%s", string(response))
	}
	return big.NewInt(info.Block[0].BlockHeader.RawData.Number), nil
}

// GetBlockByNumber Returns information about a block by hash.
func (c *HTTPClient) GetBlockByNumber(num *big.Int) (map[string]interface{}, error) {
	jrpc := jsonRPCRequest{}
	initJsonRequest("eth_getBlockByNumber", &jrpc)
	jrpc.Params = []interface{}{num, true}

	body, err := c.rpcPost(&jrpc)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}
	type blockResult struct {
		Result map[string]interface{} `json:"result"`
	}
	result := blockResult{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}
	return result.Result, nil
}

func (c *HTTPClient) EthCall(from, to string, value *big.Int, data []byte) ([]byte, error) {
	fromAddr, err := address.Base58ToAddress(from)
	if err != nil {
		fromAddr = address.Address{}
	}
	toAddr, err := address.Base58ToAddress(to)
	if err != nil {
		return nil, fmt.Errorf("wrong to address, err=%s", err)
	}
	strval := "0"
	if value != nil {
		strval = value.Text(16)
	}
	request := map[string]string{"from": fromAddr.Hex()[2:], "to": toAddr.Hex()[2:], "data": common.ToHex(data), "value": "0x" + strval}
	jrpc := jsonRPCRequest{}
	initJsonRequest("eth_call", &jrpc)
	jrpc.Params = []interface{}{request, "latest"}
	result, err := c.rpcPost(&jrpc)
	if err != nil {
		return nil, fmt.Errorf("post request failed, err=%s", err)
	}
	resp := jsonRPCReponse{}
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}
	if resp.Error.Code != 0 || resp.Error.Message != "" {
		return nil, fmt.Errorf("eth_call failed, code=%d, err=%s", resp.Error.Code, resp.Error.Message)
	}
	if strings.HasPrefix(resp.Result, "0x") {
		return hex.DecodeString(resp.Result[2:])
	}
	return []byte(resp.Result), nil
}

// GetEnergyPrice calls https://api.shasta.trongrid.io/wallet/getenergyprices and gets the latest
// energy price in sun
// Can use GetGasPrice method, since the price is the same
func (c *HTTPClient) GetEnergyPrice() (uint64, error) {
	response, err := c.fullnodeGet("wallet/getenergyprices")
	if err != nil {
		return 0, fmt.Errorf("get request failed, err=%s", err)
	}
	info := struct {
		Prices string `json:"prices"`
	}{}
	if err := json.Unmarshal(response, &info); err != nil {
		return 0, fmt.Errorf("parse json failed, err=%s", err)
	}
	// prices are in the format of "timestamp:price,timestamp:price,..."
	prices := strings.Split(info.Prices, ",")
	nowMillis := time.Now().UnixNano() / 1000000
	lastest, latestValue := nowMillis, ""
	// find the latest price
	for _, price := range prices {
		priceInfo := strings.Split(price, ":")
		if len(priceInfo) != 2 {
			continue
		}
		statTime, err := strconv.ParseInt(priceInfo[0], 10, 64)
		if err != nil {
			continue
		}
		if nowMillis-statTime <= lastest {
			lastest = statTime
			latestValue = priceInfo[1]
		}
	}
	value, _ := strconv.ParseInt(latestValue, 10, 64)
	return uint64(value), nil
}

// DeployContract will call deploycontract api, this api will generate the unsigned transaction
func (c *HTTPClient) DeployContract(strABI, strBIN, owner, name string) (*TransactionExtention, string, error) {
	ownerAddr, err := address.Base58ToAddress(owner)
	if err != nil {
		return nil, "", fmt.Errorf("owner address is not base58")
	}
	req := contractRequest{
		OwnereAddress:           ownerAddr.Hex()[2:],
		ABI:                     strABI,
		Bytecode:                strBIN,
		Name:                    name,
		FeeLimit:                1000000000,
		ConsumerResourcePercent: 0,
		CallValue:               0,
		Parameter:               "",
		OriginEnergyLimit:       1000000000,
	}
	response, err := c.fullnodePost("wallet/deploycontract", req)
	if err != nil {
		return nil, "", fmt.Errorf("http request failed, err=%s", err)
	}
	resp := deployResponse{}
	d := json.NewDecoder(bytes.NewReader(response))
	d.UseNumber()
	if err := d.Decode(&resp); err != nil {
		return nil, "", fmt.Errorf("parse json failed, err=%s", err)
	}
	trans := TronTransaction{
		RawData:         resp.RawData,
		RawDataHex:      resp.RawDataHex,
		Txid:            resp.Txid,
		Visible:         resp.Visible,
		ContractAddress: resp.ContractAddress,
	}
	tx := TransactionExtention{
		Transaction: &trans,
		Txid:        resp.Txid,
	}
	return &tx, resp.ContractAddress, nil
}

func (c *HTTPClient) convertETHAddress(addr string) string {
	s := "41" + strings.TrimPrefix(addr, "0x")
	return address.HexToAddress(s).String()
}

func (c *HTTPClient) triggerConstantContractResult(parameter, selector, contract, from string) (rest *walletResult, err error) {
	if common.Has0xPrefix(parameter) {
		parameter = parameter[2:]
	}

	//以太坊地址 -> Tron地址
	if ecommon.IsHexAddress(from) {
		from = c.convertETHAddress(from)
	}

	if ecommon.IsHexAddress(contract) {
		contract = c.convertETHAddress(contract)
	}

	//地址进行校验
	if _, err := address.Base58ToAddress(from); err != nil {
		return nil, fmt.Errorf("from address[%s] invalid", from)
	}
	if _, err := address.Base58ToAddress(contract); err != nil {
		return nil, fmt.Errorf("contract address[%s] invalid", contract)
	}

	req := walletRequest{
		OwnerAddress:     from,
		ContractAddress:  contract,
		FunctionSelector: selector,
		Parameter:        parameter,
		Visible:          true,
	}
	response, err := c.fullnodePost("wallet/triggerconstantcontract", req)
	if err != nil {
		return nil, fmt.Errorf("call wallet/triggerconstantcontract failed, err=%s", err)
	}

	result := &walletResult{}
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}

	if !result.Result.Ok {
		return nil, fmt.Errorf("call wallet/triggerconstantcontract failed, result[%+v]", result)
	}
	return result, nil
}

func (c *HTTPClient) triggerConstantContract(parameter, selector, contract, from string) (string, error) {
	result, err := c.triggerConstantContractResult(parameter, selector, contract, from)
	if err != nil {
		return "", err
	}

	if len(result.ConstantResult) > 0 {
		return result.ConstantResult[0], nil
	}
	return "", fmt.Errorf("call wallet/triggerconstantcontract failed, result[%+v]", result)
}

// GetAccountResource returns the resource of this account
func (c *HTTPClient) GetAccountResource(address string) (*big.Int, *big.Int, error) {
	req := struct {
		Address string `json:"address"`
		Visible bool   `json:"visible"`
	}{
		Address: address,
		Visible: true,
	}
	response, err := c.fullnodePost("wallet/getaccountresource", req)
	if err != nil {
		return nil, nil, fmt.Errorf("http request failed, err=%s", err)
	}
	resp := struct {
		FreeNetUsed  uint64 `json:"freeNetUsed"`
		FreeNetLimit uint64 `json:"freeNetLimit"`
		EnergyUsed   uint64 `json:"EnergyUsed"`
		EnergyLimit  uint64 `json:"EnergyLimit"`
	}{}
	if err := json.Unmarshal(response, &resp); err != nil {
		return nil, nil, fmt.Errorf("parse json failed, err=%s", err)
	}
	netLeft := big.NewInt(int64(resp.FreeNetLimit - resp.FreeNetUsed))
	if netLeft.Cmp(big.NewInt(0)) < 0 {
		netLeft = big.NewInt(0)
	}
	energyLeft := big.NewInt(int64(resp.EnergyLimit - resp.EnergyUsed))
	if energyLeft.Cmp(big.NewInt(0)) < 0 {
		energyLeft = big.NewInt(0)
	}
	return netLeft, energyLeft, nil
}

func (c *HTTPClient) triggerSmartContract(data []byte, selector, contract, from string, feeLimit *big.Int) ([]byte, error) {
	if len(data) > 4 {
		data = data[4:]
	}
	parameter := hex.EncodeToString(data)

	//以太坊地址 -> Tron地址
	if ecommon.IsHexAddress(contract) {
		contract = c.convertETHAddress(contract)
	}

	if ecommon.IsHexAddress(from) {
		from = c.convertETHAddress(from)
	}

	//地址进行校验
	if _, err := address.Base58ToAddress(from); err != nil {
		return nil, fmt.Errorf("from address[%s] invalid", from)
	}
	if _, err := address.Base58ToAddress(contract); err != nil {
		return nil, fmt.Errorf("contract address[%s] invalid", contract)
	}

	req := walletRequest{
		OwnerAddress:     from,
		ContractAddress:  contract,
		FunctionSelector: selector,
		Parameter:        parameter,
		Visible:          true,
		FeeLimit:         feeLimit,
	}
	return c.fullnodePost("wallet/triggersmartcontract", req)
}

// TriggerSmartContract calls TriggerSmartContract
// the details of this api can be found here: https://developers.tron.network/reference/triggersmartcontract
// This api will not run the contract, it just returns the transactions generated, but unsigned
func (c *HTTPClient) TriggerSmartContract(contract, from string, data []byte, feeLimit *big.Int) (*TransactionExtention, error) {
	method, err := ethevent.GetMethodByData(data)
	if err != nil {
		return nil, fmt.Errorf("get method by data failed, err=%s", err)
	}

	response, err := c.triggerSmartContract(data, method.Sig, contract, from, feeLimit)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}
	tx := TransactionExtention{}
	d := json.NewDecoder(bytes.NewReader(response))
	d.UseNumber()
	if err := d.Decode(&tx); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s, resp=%s", err, response)
	}
	if tx.Transaction == nil || tx.Transaction.Txid == "" {
		return nil, fmt.Errorf("wrong result, %s", string(response))
	}
	tx.Txid = tx.Transaction.Txid
	return &tx, nil
}

// GetTransactionByID returns the transaction information, such as from, to, calldata
func (c *HTTPClient) GetTransactionByID(txHash string) (*TronTransaction, error) {
	req := walletTransactionRequest{
		Value: txHash,
	}
	response, err := c.fullnodePost("wallet/gettransactionbyid", req)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}
	tx := TronTransaction{}
	d := json.NewDecoder(bytes.NewReader(response))
	d.UseNumber()
	if err := d.Decode(&tx); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}
	return &tx, nil
}

// GetTransactionInfo returns the transaction receipt and status
func (c *HTTPClient) GetTransactionInfoByID(txHash string) (*TransactionInfo, error) {
	req := walletTransactionRequest{
		Value: txHash,
	}
	response, err := c.fullnodePost("wallet/gettransactioninfobyid", req)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}
	tx := TransactionInfo{}
	d := json.NewDecoder(bytes.NewReader(response))
	d.UseNumber()
	if err := d.Decode(&tx); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}
	return &tx, nil
}

// GetTransactionEventsByID returns the events log generated by a transaction
func (c *HTTPClient) GetTransactionEventsByID(txHash string) (*EventLogs, error) {
	url := fmt.Sprintf("v1/transactions/%s/events", txHash)
	response, err := c.gridGet(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get events, tx=%s, err=%s", txHash, err)
	}
	logs := EventLogs{}
	if err := json.Unmarshal(response, &logs); err != nil {
		return nil, fmt.Errorf("failed to parse events, err=%s", err)
	}
	return &logs, nil
}

// TotalSupplyOf implements totalSupply of an TRC20 contract
func (c *HTTPClient) TotalSupplyOf(contract string) (*big.Int, error) {
	selector := "totalSupply()"
	response, err := c.triggerConstantContract("", selector, contract, emptyAddressBase58)
	if err != nil {
		return nil, fmt.Errorf("triggerconstantcontract failed, contract=%s, selecotr=%s, err=%s",
			contract, selector, err)
	}
	return extractNumber(response)
}

// DecimalsOf calls decimals of TRC20
func (c *HTTPClient) DecimalsOf(contract string) (*big.Int, error) {
	selector := "decimals()"
	response, err := c.triggerConstantContract("", selector, contract, emptyAddressBase58)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}
	return extractNumber(response)
}

// SymbolOf calls symbol of TRC20
func (c *HTTPClient) SymbolOf(contract string) (string, error) {
	selector := "symbol()"
	response, err := c.triggerConstantContract("", selector, contract, emptyAddressBase58)
	if err != nil {
		return "", fmt.Errorf("http request failed, err=%s", err)
	}
	return extractString(response)
}

// BroadCastTransaction broads the signed transaction to tron
func (c *HTTPClient) BroadCastTransaction(transaction *TronTransaction) error {
	r, err := c.fullnodePost("wallet/broadcasttransaction", transaction)
	if err != nil {
		return fmt.Errorf("http request failed, err=%s", err)
	}
	type broadcastResult struct {
		Result bool   `json:"result"`
		Txid   string `json:"txid"`
	}
	var result broadcastResult
	if err := json.Unmarshal(r, &result); err != nil {
		return fmt.Errorf("parse json result failed, json=%s, err=%s", string(r), err)
	}
	if !result.Result {
		return fmt.Errorf("result failed, json=%s", string(r))
	}
	return nil
}

// BalanceOf calls balanceOf of TRC20
func (c *HTTPClient) BalanceOf(contract, addr, body string) (*big.Int, error) {
	selector := "balanceOf(address)"
	response, err := c.triggerConstantContract(body, selector, contract, addr)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}
	balance, err := hex2BigInt(response)
	if err != nil {
		return nil, fmt.Errorf("parse balance failed, err=%s", err)
	}
	return balance, nil
}

// BalanceAt returns the trx of an address
func (c *HTTPClient) BalanceAt(addr string) (*big.Int, error) {
	addrHex, err := address.Base58ToAddress(addr)
	if err != nil {
		return nil, fmt.Errorf("addr is not base58")
	}
	jrpc := jsonRPCRequest{}
	initJsonRequest("eth_getBalance", &jrpc)
	jrpc.Params = []interface{}{addrHex.Hex(), "latest"}

	body, err := c.rpcPost(&jrpc)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}
	type balanceResult struct {
		Result string `json:"result"`
	}
	result := balanceResult{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}
	balance, err := hex2BigInt(result.Result)
	if err != nil {
		return nil, fmt.Errorf("parse balance failed, err=%s", err)
	}
	return balance, nil
}

// EstimateGas calls eth_estimateGas api
// This api is not used for now, we just use TriggerConstantContract to get the estimated energy
func (c *HTTPClient) EstimateGas(from, to string, hexValue string, data []byte) (*big.Int, error) {
	fromAddr, err := address.Base58ToAddress(from)
	if err != nil {
		return nil, fmt.Errorf("from address not base58")
	}
	toAddr, err := address.Base58ToAddress(to)
	if err != nil {
		return nil, fmt.Errorf("to address not base58")
	}
	request := map[string]string{"from": fromAddr.Hex(), "to": toAddr.Hex(), "data": common.ToHex(data), "value": hexValue}
	jrpc := jsonRPCRequest{}
	initJsonRequest("eth_estimateGas", &jrpc)
	jrpc.Params = []interface{}{request}
	body, err := c.rpcPost(&jrpc)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}
	type gasResult struct {
		Result  string `json:"result"`
		Jsonrpc string `json:"jsonrpc"`
		Id      int    `json:"id"`
	}
	response := gasResult{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}
	gas, err := hex2BigInt(response.Result)
	if err != nil {
		return nil, fmt.Errorf("parse gas failed, err=%s", err)
	}
	return gas, nil
}

func (c *HTTPClient) GetGasPrice() (*big.Int, error) {
	jrpc := jsonRPCRequest{}
	initJsonRequest("eth_gasPrice", &jrpc)
	jrpc.Params = []interface{}{}
	body, err := c.rpcPost(&jrpc)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}
	type gasResult struct {
		Result string `json:"result"`
	}
	response := gasResult{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}
	gas, err := hex2BigInt(response.Result)
	if err != nil {
		return nil, fmt.Errorf("parse gas failed, err=%s", err)
	}
	return gas, nil
}

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type jsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *jsonError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

func (c *HTTPClient) GetCode(addr ecommon.Address) (hexutil.Bytes, error) {
	jrpc := jsonRPCRequest{}
	initJsonRequest("eth_getCode", &jrpc)
	jrpc.Params = []interface{}{addr.Hex(), "latest"}
	body, err := c.rpcPost(&jrpc)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}
	response := jsonrpcMessage{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}

	var result hexutil.Bytes
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, fmt.Errorf("parse result failed, err=%s", err)
	}
	return result, nil
}

// TriggerStack generate a transaction to freeze trx
func (c *HTTPClient) TriggerStack(from string, resource string, amount *big.Int) (*TransactionExtention, error) {
	type jsonRequest struct {
		From     string   `json:"owner_address"`
		Amount   *big.Int `json:"frozen_balance"`
		Resource string   `json:"resource"`
	}
	fromAddr, err := address.Base58ToAddress(from)
	if err != nil {
		return nil, fmt.Errorf("from address not base58")
	}
	req := jsonRequest{From: fromAddr.Hex()[2:], Amount: amount, Resource: resource}
	resp, err := c.post("wallet/freezebalancev2", req)
	if err != nil {
		return nil, fmt.Errorf("post request failed, err=%s", err)
	}
	tx := TronTransaction{}
	d := json.NewDecoder(bytes.NewReader(resp))
	d.UseNumber()
	if err := d.Decode(&tx); err != nil {
		return nil, fmt.Errorf("decode json failed, err=%s", err)
	}
	if tx.Txid == "" {
		return nil, fmt.Errorf("parse result failed, js=%s", string(resp))
	}

	txe := TransactionExtention{Transaction: &tx, Txid: tx.Txid}
	return &txe, nil
}

// TriggerUnStack generate a transaction to unfreeze trx
func (c *HTTPClient) TriggerUnStack(from string, resource string, amount *big.Int) (*TransactionExtention, error) {
	type jsonRequest struct {
		From     string   `json:"owner_address"`
		Amount   *big.Int `json:"unfreeze_balance"`
		Resource string   `json:"resource"`
	}
	fromAddr, err := address.Base58ToAddress(from)
	if err != nil {
		return nil, fmt.Errorf("from address not base58")
	}
	req := jsonRequest{From: fromAddr.Hex()[2:], Amount: amount, Resource: resource}
	resp, err := c.post("wallet/unfreezebalancev2", req)
	if err != nil {
		return nil, fmt.Errorf("post request failed, err=%s", err)
	}
	tx := TronTransaction{}
	d := json.NewDecoder(bytes.NewReader(resp))
	d.UseNumber()
	if err := d.Decode(&tx); err != nil {
		return nil, fmt.Errorf("decode json failed, err=%s", err)
	}
	if tx.Txid == "" {
		return nil, fmt.Errorf("parse result failed, js=%s", string(resp))
	}

	txe := TransactionExtention{Transaction: &tx, Txid: tx.Txid}
	return &txe, nil
}

// TriggerWithdrawUnStack generate a transaction to withdraw unfrozen trx
func (c *HTTPClient) TriggerWithdrawUnStack(from string) (*TransactionExtention, error) {
	type jsonRequest struct {
		From string `json:"owner_address"`
	}
	fromAddr, err := address.Base58ToAddress(from)
	if err != nil {
		return nil, fmt.Errorf("from address not base58")
	}
	req := jsonRequest{From: fromAddr.Hex()[2:]}
	resp, err := c.post("wallet/withdrawexpireunfreeze", req)
	if err != nil {
		return nil, fmt.Errorf("post request failed, err=%s", err)
	}
	tx := TronTransaction{}
	d := json.NewDecoder(bytes.NewReader(resp))
	d.UseNumber()
	if err := d.Decode(&tx); err != nil {
		return nil, fmt.Errorf("decode json failed, err=%s", err)
	}
	if tx.Txid == "" {
		return nil, fmt.Errorf("parse result failed, js=%s", string(resp))
	}

	txe := TransactionExtention{Transaction: &tx, Txid: tx.Txid}
	return &txe, nil
}

func (c *HTTPClient) TriggerDelegateResource(from string, to string, resource string, amount *big.Int) (*TransactionExtention, error) {
	type jsonRequest struct {
		From     string   `json:"owner_address"`
		To       string   `json:"receiver_address"`
		Amount   *big.Int `json:"balance"`
		Resource string   `json:"resource"`
	}
	fromAddr, err := address.Base58ToAddress(from)
	if err != nil {
		return nil, fmt.Errorf("from address not base58")
	}
	toAddr, err := address.Base58ToAddress(to)
	if err != nil {
		return nil, fmt.Errorf("to address not base58")
	}
	req := jsonRequest{From: fromAddr.Hex()[2:], To: toAddr.Hex()[2:], Amount: amount, Resource: resource}
	resp, err := c.post("wallet/delegateresource", req)
	if err != nil {
		return nil, fmt.Errorf("post request failed, err=%s", err)
	}
	tx := TronTransaction{}
	d := json.NewDecoder(bytes.NewReader(resp))
	d.UseNumber()
	if err := d.Decode(&tx); err != nil {
		return nil, fmt.Errorf("decode json failed, err=%s", err)
	}
	if tx.Txid == "" {
		return nil, fmt.Errorf("parse result failed, js=%s", string(resp))
	}

	txe := TransactionExtention{Transaction: &tx, Txid: tx.Txid}
	return &txe, nil
}

func (c *HTTPClient) ChainID() (*big.Int, error) {
	jRpc := jsonRPCRequest{}
	initJsonRequest("eth_chainId", &jRpc)
	jRpc.Params = []interface{}{}

	body, err := c.rpcPost(&jRpc)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}
	type chainIDResult struct {
		Result string `json:"result"`
	}
	var result chainIDResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}
	chainID := new(big.Int)
	chainID.SetString(result.Result[2:], 16) // Resolve after removing the prefix "0x"
	return chainID, nil
}

func (c *HTTPClient) TransactionReceipt(hash ecommon.Hash) (*types.Receipt, error) {
	jRpc := jsonRPCRequest{}
	initJsonRequest("eth_getTransactionReceipt", &jRpc)
	jRpc.Params = []interface{}{hash} // set parameters

	body, err := c.rpcPost(&jRpc)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}

	var result struct {
		Result types.Receipt `json:"result"`
	}
	// return JSON data
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}

	return &result.Result, nil
}

func (c *HTTPClient) BlockNumber() (*big.Int, error) {
	jRpc := jsonRPCRequest{}
	initJsonRequest("eth_blockNumber", &jRpc)
	jRpc.Params = []interface{}{}

	body, err := c.rpcPost(&jRpc)
	if err != nil {
		return nil, fmt.Errorf("http request failed, err=%s", err)
	}
	type chainIDResult struct {
		Result string `json:"result"`
	}
	var result chainIDResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse json failed, err=%s", err)
	}
	chainID := new(big.Int)
	chainID.SetString(result.Result[2:], 16) // Resolve after removing the prefix "0x"
	return chainID, nil
}
