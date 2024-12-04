package xrsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"sync"
)

// GenerateRsaKey Generate rsa Key pair, then save to a disk file
func GenerateRsaKey(keySize int, pathPrefix string) error {
	// 1. Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return err
	}
	// 2. x509 marshal
	derText := x509.MarshalPKCS1PrivateKey(privateKey)
	// 3.pem.Block
	block := pem.Block{
		Type:  "rsa private key", // string
		Bytes: derText,
	}
	// 4. pem encode
	file, err := os.Create(pathPrefix + "_private.pem")
	if err != nil {
		return err
	}
	_ = pem.Encode(file, &block)
	_ = file.Close()

	// ============ public key ==========
	// 1. public key
	publicKey := privateKey.PublicKey
	// 2. x509 marshal
	derStream, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return err
	}
	// 3. input pem.Block
	block = pem.Block{
		Type:  "rsa public key",
		Bytes: derStream,
	}
	// 4. pem encode
	file, err = os.Create(pathPrefix + "_public.pem")
	if err != nil {
		return err
	}
	_ = pem.Encode(file, &block)
	_ = file.Close()
	return nil
}

// RSAEncrypt RSA use rsa public key encode
func RSAEncrypt(plainText []byte, pk []byte) (encodeRes []byte, err error) {
	block, _ := pem.Decode(pk)
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	//public key
	pubKey := pubInterface.(*rsa.PublicKey)
	// encode
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, plainText)
	if err != nil {
		return nil, err
	}
	return cipherText, nil
}

// RSADecrypt RSA decode
func RSADecrypt(cipherText []byte, pk []byte) (decodeRes []byte, err error) {
	//private key
	block, _ := pem.Decode(pk)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	// decode
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
	if err != nil {
		return nil, err
	}
	return plainText, nil
}

// HighRSAEncrypt RSA Encode high version
func HighRSAEncrypt(plainText []byte, pk []byte) ([]byte, error) {
	length := 100
	var rSAEncryptRes = make([]byte, 0)
	total := len(plainText)/length + 1
	for i := 0; i <= total; i++ {
		start := i * length
		end := (i + 1) * length
		if end > len(plainText) {
			end = len(plainText)
		}
		if start >= end {
			break
		}
		data := plainText[start:end]
		rsaEncodeV, err := RSAEncrypt(data, pk)
		if err != nil {
			return nil, err
		}
		rSAEncryptRes = append(rSAEncryptRes, rsaEncodeV...)
	}
	return rSAEncryptRes, nil
}

// HighRSADecrypt RSA decode high version
func HighRSADecrypt(RSADecryptSrc []byte, pk []byte) ([]byte, error) {
	splitLen := 512
	total := len(RSADecryptSrc)/splitLen + 1
	var rSADecryptRes = make([]byte, 0)
	var sp = make([][]byte, total)
	wg := sync.WaitGroup{}
	for i := 0; i < total; i++ {
		wg.Add(1)
		start := i * splitLen
		end := (i + 1) * splitLen
		if end > len(RSADecryptSrc) {
			end = len(RSADecryptSrc)
		}
		if start >= end {
			wg.Done()
			break
		}
		go func(data []byte, i int) {
			defer wg.Done()
			resDecodeV, err := RSADecrypt(data, pk)
			if err != nil {
				return
			}
			sp[i] = resDecodeV
		}(RSADecryptSrc[start:end], i)
	}
	wg.Wait()
	for i := 0; i < len(sp); i++ {
		rSADecryptRes = append(rSADecryptRes, sp[i]...)
	}
	return rSADecryptRes, nil
}

var (
	ErrKeyMustBePEMEncoded = errors.New("invalid key: Key must be a PEM encoded PKCS1 or PKCS8 key")
	ErrNotRSAPrivateKey    = errors.New("key is not a valid RSA private key")
	ErrNotRSAPublicKey     = errors.New("key is not a valid RSA public key")
)

// ParseRSAPrivateKeyFromPEM parses a PEM encoded PKCS1 or PKCS8 private key
func ParseRSAPrivateKeyFromPEM(key []byte) (*rsa.PrivateKey, error) {
	var err error

	// Parse PEM block
	var block *pem.Block
	if block, _ = pem.Decode(key); block == nil {
		return nil, ErrKeyMustBePEMEncoded
	}

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
			return nil, err
		}
	}

	var pkey *rsa.PrivateKey
	var ok bool
	if pkey, ok = parsedKey.(*rsa.PrivateKey); !ok {
		return nil, ErrNotRSAPrivateKey
	}

	return pkey, nil
}

// ParseRSAPrivateKeyFromPEMWithPassword parses a PEM encoded PKCS1 or PKCS8 private key protected with password
//
// Deprecated: This function is deprecated and should not be used anymore. It uses the deprecated x509.DecryptPEMBlock
// function, which was deprecated since RFC 1423 is regarded insecure by design. Unfortunately, there is no alternative
// in the Go standard library for now. See https://github.com/golang/go/issues/8860.
func ParseRSAPrivateKeyFromPEMWithPassword(key []byte, password string) (*rsa.PrivateKey, error) {
	var err error

	// Parse PEM block
	var block *pem.Block
	if block, _ = pem.Decode(key); block == nil {
		return nil, ErrKeyMustBePEMEncoded
	}

	var parsedKey interface{}

	var blockDecrypted []byte
	if blockDecrypted, err = x509.DecryptPEMBlock(block, []byte(password)); err != nil {
		return nil, err
	}

	if parsedKey, err = x509.ParsePKCS1PrivateKey(blockDecrypted); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(blockDecrypted); err != nil {
			return nil, err
		}
	}

	var pkey *rsa.PrivateKey
	var ok bool
	if pkey, ok = parsedKey.(*rsa.PrivateKey); !ok {
		return nil, ErrNotRSAPrivateKey
	}

	return pkey, nil
}

// ParseRSAPublicKeyFromPEM parses a PEM encoded PKCS1 or PKCS8 public key
func ParseRSAPublicKeyFromPEM(key []byte) (*rsa.PublicKey, error) {
	var err error

	// Parse PEM block
	var block *pem.Block
	if block, _ = pem.Decode(key); block == nil {
		return nil, ErrKeyMustBePEMEncoded
	}

	// Parse the key
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKIXPublicKey(block.Bytes); err != nil {
		if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
			parsedKey = cert.PublicKey
		} else {
			return nil, err
		}
	}

	var pkey *rsa.PublicKey
	var ok bool
	if pkey, ok = parsedKey.(*rsa.PublicKey); !ok {
		return nil, ErrNotRSAPublicKey
	}

	return pkey, nil
}
