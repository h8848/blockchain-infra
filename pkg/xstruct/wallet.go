package xstruct

type UserOrder struct {
	OrderType int8
	TxHash    string
	ChainId   string
	OrderId   uint32
	ToUID     int64
}
