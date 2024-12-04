package eth_abi

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func GetClient(url string) (*ethclient.Client, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetBlockNumber(client *rpc.Client) (int64, error) {
	ctx := context.Background()
	var result hexutil.Uint64
	err := client.CallContext(ctx, &result, "eth_blockNumber")
	return int64(result), err
}
