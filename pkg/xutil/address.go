package xutil

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/cc14514/go-geoip2"
	geoIp "github.com/cc14514/go-geoip2-db"
)

type Resp struct {
	Address Address `json:"address"`
}

type Address struct {
	City        string `json:"city"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	State       string `json:"state"`
}

// GetAddress 根据经纬度获取国家信息
func GetAddress(latitude, longitude string) (data Resp, err error) {
	client := &http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=jsonv2&lat=%s&lon=%s", latitude, longitude)
	// nolint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	res, err := client.Do(req)

	// 增加错误&空处理
	if err != nil {
		return
	}
	if res == nil || res.Body == nil {
		return
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return
	}
	return
}

// GetNationByIp 根据ip获取国家信息
func GetNationByIp(ip string) (record *geoip2.City, err error) {
	db, _ := geoIp.NewGeoipDbByStatik()
	defer func(db *geoip2.DBReader) {
		err = db.Close()
	}(db)
	record, _ = db.City(net.ParseIP(ip))
	return
}
