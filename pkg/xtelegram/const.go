package xtelegram

type (
	LogLevel int
)

// 变量名会和方法名冲突， logt.Error()
//const (
//	Info LogLevel = iota
//	Warn
//	Error
//	Crash
//)

const (
	INFO LogLevel = iota
	WARN
	DEBUG
	ERROR
	CRASH
)
