package xutil

import "regexp"

func IsHttpURL(url string) bool {
	// 使用正则表达式匹配 HTTP 或 HTTPS URL
	pattern := `^(http|https|wss|ws)://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=]+$`
	match, err := regexp.MatchString(pattern, url)
	return err == nil && match
}
