package ethevent

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"log"
	"reflect"
	"strings"
)

type EventLog struct {
	// Consensus fields:
	// address of the contract that generated the event
	Address common.Address `json:"address" gencodec:"required"`
	// list of topics provided by the contract.
	Topics []common.Hash `json:"topics" gencodec:"required"`
	// supplied by the contract, usually ABI-encoded
	Data hexutil.Bytes `json:"data" gencodec:"required"`
}

var parsedABI abi.ABI

func init() {
	var err error
	parsedABI, err = abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		log.Panicf("abi.JSON read error[%v]", err)
	}
}

func getOutTypeFromEventName(name string) (out IEventType, err error) {
	if name == "" {
		return nil, errors.New("event name is empty")
	}

	switch name {
	case EventNameTransfer:
		out = &Transfer{}
	case EventNameApproval:
		out = &Approval{}
	default:
		return nil, fmt.Errorf("event name[%s] not supported", name)
	}
	return out, nil
}

func ParseEventToStruct(output IEventType, eventLog *EventLog) (out IEventType, err error) {
	if output != nil && reflect.TypeOf(output).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("output must be a pointer")
	}

	if len(eventLog.Topics) < 1 {
		return nil, errors.New("log topics = 0")
	}

	findEvent, err := parsedABI.EventByID(eventLog.Topics[0])
	if err != nil {
		return nil, fmt.Errorf("call EventByID error[%v]", err)
	}

	if findEvent == nil {
		return nil, fmt.Errorf("event[%s] not found", eventLog.Topics[0])
	}

	//若output输入为nil，则自动推断数据类型并返回
	if output == nil {
		output, err = getOutTypeFromEventName(findEvent.Name)
		if err != nil {
			return nil, fmt.Errorf("event type auto parsed error[%v]", err)
		}
	}

	output.SetContractAddr(eventLog.Address.String())
	output.SetEventName(findEvent.Name)
	output.SetEventID(findEvent.ID.Hex())

	err = parsedABI.UnpackIntoInterface(output, findEvent.Name, eventLog.Data)
	if err != nil {
		return nil, fmt.Errorf("UnpackIntoInterface error[%v]]", err)
	}

	//构造topic args
	args := make([]abi.Argument, 0, len(findEvent.Inputs))
	for _, arg := range findEvent.Inputs {
		if arg.Indexed {
			args = append(args, arg)
		}
	}

	//构造topics
	topics := eventLog.Topics[1:]
	err = abi.ParseTopics(output, args, topics)
	if err != nil {
		return nil, fmt.Errorf("failed to parse topics into TransferEvent: %v", err)
	}
	return output, nil
}
