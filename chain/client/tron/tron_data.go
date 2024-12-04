package tron

import "math/big"

type TransactionExtention struct {
	Transaction    *TronTransaction `json:"transaction,omitempty"`
	Txid           string           `json:"txID,omitempty"`
	ConstantResult [][]byte         `json:"constant_result,omitempty"`
	Energy         uint64           `json:"energy_used,omitempty"`
}

type TronTransaction struct {
	Visible         bool                 `json:"visible"`
	Txid            string               `json:"txID,omitempty"`
	ContractAddress string               `json:"contract_address,omitempty"`
	RawData         *TransactionRaw      `json:"raw_data,omitempty"`
	RawDataHex      string               `json:"raw_data_hex,omitempty"`
	Signature       []string             `json:"signature,omitempty"`
	Ret             []*TransactionResult `json:"ret,omitempty"`
}

type TransactionInfo struct {
	ID             string   `json:"id,omitempty"`
	Fee            *big.Int `json:"fee,omitempty"`
	BlockNumber    *big.Int `json:"blockNumber,omitempty"`
	BlockTimeStamp uint64   `json:"blockTimeStamp,omitempty"`
	Result         string   `json:"result,omitempty"`
	Message        string   `json:"message,omitempty"`
}

type TransactionResult struct {
	Fee                           int64  `json:"fee,omitempty"`
	Ret                           string `json:"ret,omitempty"`
	ContractRet                   string `json:"contractRet,omitempty"`
	AssetIssueID                  string `json:"assetIssueID,omitempty"`
	WithdrawAmount                int64  `json:"withdraw_amount,omitempty"`
	UnfreezeAmount                int64  `json:"unfreeze_amount,omitempty"`
	ExchangeReceivedAmount        int64  `json:"exchange_received_amount,omitempty"`
	ExchangeInjectAnotherAmount   int64  `json:"exchange_inject_another_amount,omitempty"`
	ExchangeWithdrawAnotherAmount int64  `json:"exchange_withdraw_another_amount,omitempty"`
	ExchangeId                    int64  `json:"exchange_id,omitempty"`
	ShieldedTransactionFee        int64  `json:"shielded_transaction_fee,omitempty"`
}

type TransactionRaw struct {
	//only support size = 1, repeated list here for extension
	Contract      []*TransactionContract `json:"contract,omitempty"`
	RefBlockBytes []byte                 `json:"ref_block_bytes,omitempty"`
	RefBlockNum   int64                  `json:"ref_block_num,omitempty"`
	RefBlockHash  []byte                 `json:"ref_block_hash,omitempty"`
	Expiration    int64                  `json:"expiration,omitempty"`
	Auths         []*Authority           `json:"auths,omitempty"`
	// transaction note
	Data []byte `json:"data,omitempty"`
	// scripts not used
	Scripts   []byte `json:"scripts,omitempty"`
	FeeLimit  int64  `json:"fee_limit,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

type Authority struct {
	Account        *AccountId `json:"account,omitempty"`
	PermissionName []byte     `json:"permission_name,omitempty"`
}

type AccountId struct {
	Name    []byte `json:"name,omitempty"`
	Address []byte `json:"address,omitempty"`
}

type TransactionContract struct {
	Type         string    `json:"type,omitempty"`
	Parameter    Parameter `json:"parameter,omitempty"`
	Provider     []byte    `json:"provider,omitempty"`
	ContractName []byte    `json:"ContractName,omitempty"`
	PermissionId int32     `json:"Permission_id,omitempty"`
}

type Parameter struct {
	Value   map[string]any `json:"value"`
	TypeUrl string         `json:"type_url"`
}

type TronEvent struct {
	BlockNumber     *big.Int          `json:"block_number"`
	BlockTimeStamp  uint64            `json:"block_timestamp"`
	ContractAddress string            `json:"contract_address"`
	EventName       string            `json:"event_name"`
	Event           string            `json:"event"`
	Results         map[string]string `json:"result"`
}
type EventLogs struct {
	Data    []*TronEvent `json:"data"`
	Success bool         `json:"success"`
}
