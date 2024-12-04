package xutil

import (
	"context"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
)

type JwtParam struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

func NewJwtService(token, secret string) *JwtParam {
	return &JwtParam{
		Token:  token,
		Secret: secret,
	}
}

// ParseToken 解析token
func (j *JwtParam) ParseToken() (auth *jwt.Token, err error) {
	auth, err = jwt.Parse(j.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.Secret), nil
	})
	return
}

// GetUid 获取uid
func (j *JwtParam) GetUid() (uid int64, err error) {
	auth, err := j.ParseToken()
	if err != nil {
		return
	}
	claims := auth.Claims.(jwt.MapClaims)
	uid = int64(claims["uid"].(float64))
	return
}

// GetExp 获取过期时间
func (j *JwtParam) GetExp() (exp int64, err error) {
	auth, err := j.ParseToken()
	if err != nil {
		return
	}
	claims := auth.Claims.(jwt.MapClaims)
	exp = int64(claims["exp"].(float64))
	return
}

func GetToken(secretKey, account string, iat, seconds, uid, kaiId int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["uid"] = uid
	claims["account"] = account
	claims["kai_id"] = strconv.FormatInt(kaiId, 10)
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

func GetUid(ctx context.Context) int64 {
	data := ctx.Value("jwt_info").(map[string]interface{})
	return int64(data["uid"].(int))
}

// GetKaiId 获取kai-id
func GetKaiId(ctx context.Context) (kaiId int64) {
	data := ctx.Value("jwt_info").(map[string]interface{})
	if data["kai_id"] != nil {
		kaiId, _ = strconv.ParseInt(data["kai_id"].(string), 10, 64)
		return
	}
	return
}

func GetAccount(ctx context.Context) (account string) {
	data := ctx.Value("jwt_info").(map[string]interface{})
	if data["account"] != nil {
		account = data["account"].(string)
	}
	return
}

func VerifyToken(token, secret string) (*jwt.Token, error) {
	auth, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	return auth, err
}
