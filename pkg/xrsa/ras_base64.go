package xrsa

import (
	"encoding/base64"
)

// HighRSAEncryptBase64 RSA encrypt high base64 version
func HighRSAEncryptBase64(plainText []byte, pk []byte) ([]byte, error) {
	length := 100
	var ret = make([]string, 0)

	for i := 0; i < len(plainText); i++ {
		if len(plainText)-(length*(i+1)) > 0 {
			data := plainText[i*length : length*(i+1)]
			ret = append(ret, string(data))
		} else {
			data := plainText[len(ret)*length:]
			ret = append(ret, string(data))
			break
		}
	}
	//encode
	var rsaDecryptRes string
	for _, v := range ret {
		rsaEncodeV, err := RSAEncrypt([]byte(v), pk)
		if err != nil {
			return nil, err
		}
		encodeV := base64.StdEncoding.EncodeToString(rsaEncodeV)
		if encodeV == "" {
			return nil, err
		}
		rsaDecryptRes = rsaDecryptRes + string(encodeV)
	}

	return []byte(rsaDecryptRes), nil
}

// HighRSADecryptBase64 RSA Decrypt high base64 version
func HighRSADecryptBase64(RSADecryptSrc []byte, pk []byte) ([]byte, error) {
	splitLen := 684
	var sa = make([]string, 0)
	for i := 0; i < len(RSADecryptSrc); i++ {
		if len(RSADecryptSrc)-(splitLen*(i+1)) > 0 {
			data := RSADecryptSrc[i*splitLen : splitLen*(i+1)]
			sa = append(sa, string(data))
		} else {
			data := RSADecryptSrc[len(sa)*splitLen:]
			sa = append(sa, string(data))
			break
		}
	}
	//decode
	var rsaDecryptRes string
	for _, v := range sa {
		decodeV, _ := base64.StdEncoding.DecodeString(v)
		if decodeV == nil {
			break
		}
		resDecodeV, err := RSADecrypt(decodeV, pk)
		if err != nil {
			return nil, err
		}
		rsaDecryptRes = rsaDecryptRes + string(resDecodeV)
	}

	return []byte(rsaDecryptRes), nil
}
