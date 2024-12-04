package xutil

import (
	"crypto/rand"
	"math/big"
	rand2 "math/rand"
	"time"
	"unicode"
)

// GetRandStr 获取A-Z-16位随机字符串
func GetRandStr(strSize int) string {
	dictionary := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bt := make([]byte, strSize)
	_, _ = rand.Read(bt)
	for k, v := range bt {
		bt[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bt)
}

const (
	charset       = "aAbBcCdDeEfFgGhHiIjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyYzZ0123456789"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand2.NewSource(time.Now().UnixNano())

// GenerateSeed 生成种子
func GenerateSeed(dividend int64) int64 {
	divisor := int64(300)
	remainder := dividend % divisor
	min := dividend - remainder
	max := min + divisor
	average := float64(min+max) / float64(2)
	return int64(average)
}

// GenerateRandomKey 生成随机字符串
func GenerateRandomKey(seed int64, length int) string {
	result := make([]byte, length)
	a, c, m := int64(1664525), int64(1013904223), int64(0x100000000)
	x := seed
	for i := range result {
		x = (a*x + c) % m
		result[i] = charset[x%int64(len(charset))]
	}
	return string(result)
}

func ReverseRandomKey(partialStr string) int64 {
	a, m := int64(1664525), int64(0x100000000)
	aInverse := new(big.Int).ModInverse(big.NewInt(a), big.NewInt(m))
	partialBytes := []byte(partialStr)
	var seed int64
	for _, char := range partialBytes {
		charIndex := indexOf(char, charset)
		seed = (seed - int64(charIndex) + m) * aInverse.Int64() % m
		if seed < 0 {
			seed += m
		}
	}
	return seed
}

func indexOf(char byte, charset string) int {
	for i, c := range charset {
		if byte(c) == char {
			return i
		}
	}
	return -1
}

// GenerateRandomString 生成指定长度的字符串
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	for i, cache, remain := length-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(charset) {
			b[i] = charset[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

func GenerateRandomDigits(length int) string {
	numbers := extractDigits(charset)
	b := make([]byte, length)
	for i, cache, remain := length-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(numbers) {
			b[i] = numbers[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

func extractDigits(s string) string {
	var result []rune
	for _, r := range s {
		if unicode.IsDigit(r) {
			result = append(result, r)
		}
	}
	return string(result)
}
