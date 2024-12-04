package xgorm

type GormConf struct {
	DB           string
	Addr         string
	User         string
	Passwd       string
	TimeoutSec   int    `json:",default=0"`
	MaxIdleConns int    `json:",default=10"`
	MaxOpenConns int    `json:",default=10"`
	Metric       bool   `json:",default=true"`
	Trace        bool   `json:",default=true"`
	Options      string `json:",default=''"`
}
