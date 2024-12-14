package ethevent

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type IEventType interface {
	SetContractAddr(contractAddr string)
	SetEventName(eventName string)
	SetEventID(eventID string)

	GetContractAddr() string
	GetEventName() string
	GetEventID() string
}

const (
	EventNameTransfer       = "Transfer"
	EventNameApproval       = "Approval"
	EventNameApprovalForAll = "ApprovalForAll"
)

type Origin struct {
	contractAddr string
	eventName    string
	eventID      string
}

type Transfer struct {
	Origin
	From  common.Address `json:"from"`
	To    common.Address `json:"to"`
	Value *big.Int       `json:"value"`
}

type Approval struct {
	Origin
	Owner   common.Address `json:"owner"`
	Spender common.Address `json:"spender"`
	Value   *big.Int       `json:"value"`
}

type ApprovalForAll struct {
	Origin
	Owner    common.Address `json:"owner"`
	Operator common.Address `json:"operator"`
	Approved bool           `json:"approved"`
}

func (e *Origin) SetContractAddr(contractAddr string) {
	e.contractAddr = contractAddr
}

func (e *Origin) SetEventName(eventName string) {
	e.eventName = eventName
}

func (e *Origin) SetEventID(eventID string) {
	e.eventID = eventID
}

func (e *Origin) GetContractAddr() string {
	return e.contractAddr
}

func (e *Origin) GetEventName() string {
	return e.eventName
}

func (e *Origin) GetEventID() string {
	return e.eventID
}
