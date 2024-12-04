package xutil

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	SolClient "github.com/blocto/solana-go-sdk/client"
	SolCommon "github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/pkg/hdwallet"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/h8848/blockchain-infra/chain/client/tron"
	"github.com/h8848/blockchain-infra/chain/ethereum/eth_abi"
	"github.com/shopspring/decimal"
	"github.com/tyler-smith/go-bip39"
	"strconv"
)

func MnemonicGetEthAddress(mnemonic string, path string) (address string, err error) {
	privateKeyECDSA, err := MnemonicToPrivateKey(mnemonic, "", path)
	if err != nil {
		return "", err
	}
	pub := crypto.FromECDSAPub(&privateKeyECDSA.PublicKey)
	pubHex := hex.EncodeToString(pub)
	ethAddress, err := PublicKeyToEthAddress(pubHex)
	if err != nil {
		return "", err
	}
	return ethAddress, nil
}

func PublicKeyToEthAddress(publicKey string) (string, error) {
	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return "", err
	}
	uPublicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return "", err
	}
	return crypto.PubkeyToAddress(*uPublicKey).Hex(), nil
}

func MnemonicGetTronAddress(mnemonic string, path string) (address string, err error) {
	if path == "" {
		path = "m/44'/195'/0'/0/0"
	}
	privateKeyECDSA, err := MnemonicToPrivateKey(mnemonic, "", path)
	if err != nil {
		return "", err
	}
	pub := crypto.FromECDSAPub(&privateKeyECDSA.PublicKey)
	pubHex := hex.EncodeToString(pub)
	ethAddress, err := PublicKeyToEthAddress(pubHex)
	if err != nil {
		return "", err
	}
	return ETHToTronAddress(ethAddress), nil
}

func EthAddrToTronAddr(ethAddr string) (tronAddr string) {
	tronAddress := ETHToTronAddress(ethAddr)
	return tronAddress
}

func TronAddrToEthAddr(tronAddr string) (ethAddr string) {
	ethAddress, _ := TronToEthAddress(tronAddr)
	return ethAddress.Hex()
}

func MnemonicGetSolAddress(mnemonic string, path string) (address string, err error) {
	if path == "" {
		path = `m/44'/501'/0'/0'`
	}
	seed := bip39.NewSeed(mnemonic, "")
	derivedKey, err := hdwallet.Derived(path, seed)
	if err != nil {
		return "", err
	}
	account, err := types.AccountFromSeed(derivedKey.PrivateKey)
	if err != nil {
		return "", err
	}
	return account.PublicKey.ToBase58(), nil
}

func PrivateKeyGetEthAddress(privateKeyECDSA *ecdsa.PrivateKey) (address string, err error) {
	pub := crypto.FromECDSAPub(&privateKeyECDSA.PublicKey)
	pubHex := hex.EncodeToString(pub)
	ethAddress, err := PublicKeyToEthAddress(pubHex)
	if err != nil {
		return "", err
	}
	return ethAddress, nil
}

func PrivateKeyGetTronAddress(privateKeyECDSA *ecdsa.PrivateKey) (address string, err error) {
	pub := crypto.FromECDSAPub(&privateKeyECDSA.PublicKey)
	pubHex := hex.EncodeToString(pub)
	ethAddress, err := PublicKeyToEthAddress(pubHex)
	if err != nil {
		return "", err
	}
	return ETHToTronAddress(ethAddress), nil
}

func PrivateKeyGetSolAddress(privateKeyECDSA *ecdsa.PrivateKey) (address string, err error) {
	privateKeyBytes := privateKeyECDSA.D.Bytes()
	privateKeyEd25519 := ed25519.NewKeyFromSeed(privateKeyBytes)
	//fmt.Println("Ed25519私钥", base58.Encode(privateKeyEd25519))
	solPublicKey := make([]byte, ed25519.PublicKeySize)
	copy(solPublicKey, privateKeyEd25519.Public().(ed25519.PublicKey))
	//fmt.Println("Ed25519公钥", hex.EncodeToString(solPublicKey))
	solAddress := SolCommon.PublicKeyFromBytes(solPublicKey).ToBase58()
	return solAddress, nil
}

func PrivateECDSAToSolEd25519(privateKeyECDSA *ecdsa.PrivateKey) (edPrivate ed25519.PrivateKey, err error) {
	privateKeyBytes := privateKeyECDSA.D.Bytes()
	privateKeyEd25519 := ed25519.NewKeyFromSeed(privateKeyBytes)
	//fmt.Println("Ed25519私钥", base58.Encode(privateKeyEd25519))
	//solPublicKey := make([]byte, ed25519.PublicKeySize)
	return privateKeyEd25519, nil
}

func EcdsaPrivateKeyGetSolPrivateKey(privateKeyECDSA *ecdsa.PrivateKey) (key ed25519.PrivateKey, address string, err error) {
	privateKeyBytes := privateKeyECDSA.D.Bytes()
	privateKeyEd25519 := ed25519.NewKeyFromSeed(privateKeyBytes)
	solPublicKey := make([]byte, ed25519.PublicKeySize)
	copy(solPublicKey, privateKeyEd25519.Public().(ed25519.PublicKey))
	solAddress := SolCommon.PublicKeyFromBytes(solPublicKey).ToBase58()
	return privateKeyEd25519, solAddress, nil
}

func EthBalanceAtAndOf(address, contract, tokenDecimals string, ethClient *ethclient.Client) (bAt, bOf decimal.Decimal, err error) {
	balanceAt, err := ethClient.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	token, err := eth_abi.NewErc20Token(common.HexToAddress(contract), ethClient)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	balanceOf, err := token.BalanceOf(nil, common.HexToAddress(address))
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	decimalAt := ToDecimal(balanceAt.String(), 18)
	tDec, _ := strconv.ParseInt(tokenDecimals, 10, 64)
	decimalOf := ToDecimal(balanceOf.String(), int(tDec))

	return decimalAt, decimalOf, nil
}

func TrxBalanceAtAndOf(address, contract, tokenDecimals string, trxClient *tron.TronClient) (bAt, bOf decimal.Decimal, err error) {
	balanceAt, err := trxClient.BalanceAt(address)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	balanceOf, err := trxClient.BalanceOf(contract, address)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	decimalAt := ToDecimal(balanceAt.String(), 6)
	tDec, _ := strconv.ParseInt(tokenDecimals, 10, 64)
	decimalOf := ToDecimal(balanceOf.String(), int(tDec))

	return decimalAt, decimalOf, nil

}

func SolBalanceAtAndOf(address, contract, tokenDecimals string, solC *SolClient.Client) (bAt, bOf decimal.Decimal, err error) {
	// 检查 gas
	mainSolAccount := SolCommon.PublicKeyFromString(address)
	mainBalanceAt, err := solC.GetBalance(context.Background(), mainSolAccount.ToBase58())
	if err != nil {
		fmt.Printf("Error fetching balance for address: %s, error: %v", address, err)
		return decimal.Zero, decimal.Zero, err
	}
	// 检测主账户 USDT balance
	mainATA, _, err := SolCommon.FindAssociatedTokenAddress(mainSolAccount, SolCommon.PublicKeyFromString(contract))
	if err != nil {
		fmt.Printf("Error finding associated token address for contract: %s, error: %v", contract, err)
		return decimal.Zero, decimal.Zero, err
	}
	mainTokenBalance, err := solC.GetTokenAccountBalance(context.Background(), mainATA.ToBase58())
	if err != nil {
		fmt.Printf("Error fetching token balance for address: %s, error: %v", mainATA.ToBase58(), err)
		return decimal.Zero, decimal.Zero, err
	}
	mainBalanceAtDec := ToDecimal(fmt.Sprintf("%d", mainBalanceAt), 9)
	mainTokenAmount, _ := decimal.NewFromString(mainTokenBalance.UIAmountString)

	return mainBalanceAtDec, mainTokenAmount, nil
}
