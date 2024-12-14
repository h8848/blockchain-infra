package ethevent

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

func GetMethodByData(data []byte) (method *abi.Method, err error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data is too short, len=%d <= 4", len(data))
	}
	sigData := data[:4]
	return parsedABI.MethodById(sigData)
}

func BuildMethodData(name string, args ...interface{}) ([]byte, error) {
	// Fetch the ABI of the requested method
	if name == "" {
		// constructor
		arguments, err := parsedABI.Constructor.Inputs.Pack(args...)
		if err != nil {
			return nil, err
		}
		return arguments, nil
	}
	method, exist := parsedABI.Methods[name]
	if !exist {
		return nil, fmt.Errorf("method '%s' not found", name)
	}
	arguments, err := method.Inputs.Pack(args...)
	if err != nil {
		return nil, err
	}
	// Pack up the method ID too if not a constructor and return
	return append(method.ID, arguments...), nil
}

func UnpackMethodDataToMap(v map[string]interface{}, name string, data []byte) error {
	if name == "" {
		return fmt.Errorf("abi name emtpy")
	}

	method, ok := parsedABI.Methods[name]
	if !ok {
		return fmt.Errorf("abi method '%s' not found", name)
	}
	return method.Inputs.UnpackIntoMap(v, data)
}
