package xutil

import "time"

// CompleteNPeriodsBetween 计算两个时间戳相差的N倍数
func CompleteNPeriodsBetween(start time.Time, end time.Time, N float64) float64 {
	duration := end.Sub(start)
	return duration.Hours() / N
}
