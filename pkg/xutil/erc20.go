package xutil

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

type Erc20Transfer struct {
	Client     *ethclient.Client
	PrivateKey *ecdsa.PrivateKey
	Contract   *common.Address
}

func NewErc20Transfer(client *ethclient.Client, privateKey *ecdsa.PrivateKey, contract common.Address) *Erc20Transfer {
	return &Erc20Transfer{
		Client:     client,
		PrivateKey: privateKey,
		Contract:   &contract,
	}
}

func (e *Erc20Transfer) Close() {
	e.Client.Close()
	e.Contract = nil
	e.PrivateKey = nil
}

func (e *Erc20Transfer) SenToErc20(toAddress common.Address, amount *big.Int) (string, error) {

	publicKey := e.PrivateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := e.Client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}
	chainID, err := e.Client.ChainID(context.Background())
	if err != nil {
		return "", err
	}

	gasPrice, err := e.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}
	gasTipPrice, err := e.Client.SuggestGasTipCap(context.Background())
	if err != nil {
		return "", err
	}

	// 编码 ERC20 `transfer` 方法
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := crypto.Keccak256Hash(transferFnSignature).Hex()
	methodID := common.HexToHash(hash).Bytes()[:4]

	// 添加方法 ID 和参数（地址和金额）
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	txTata := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        e.Contract,
		Gas:       uint64(21000 * 3),
		GasFeeCap: gasPrice,
		GasTipCap: gasTipPrice,
		Data:      data,
	}
	tx := types.NewTx(txTata)

	signerTX := types.NewLondonSigner(chainID)
	signedTx, err := types.SignTx(tx, signerTX, e.PrivateKey)
	if err != nil {
		return "", err
	}
	//send
	err = e.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}
	return signedTx.Hash().Hex(), nil
}

func Erc20SingAndSend(client *ethclient.Client, privateKey *ecdsa.PrivateKey, contract, toAddress common.Address, amount *big.Int) (string, error) {

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return "", err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}
	gasTipPrice, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return "", err
	}

	// 编码 ERC20 `transfer` 方法
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := crypto.Keccak256Hash(transferFnSignature).Hex()
	methodID := common.HexToHash(hash).Bytes()[:4]

	// 添加方法 ID 和参数（地址和金额）
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	txTata := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &contract,
		Gas:       uint64(21000 * 3),
		GasFeeCap: gasPrice,
		GasTipCap: gasTipPrice,
		Data:      data,
	}
	tx := types.NewTx(txTata)

	signerTX := types.NewLondonSigner(chainID)
	signedTx, err := types.SignTx(tx, signerTX, privateKey)
	if err != nil {
		return "", err
	}
	//send
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}
	return signedTx.Hash().Hex(), nil
}
