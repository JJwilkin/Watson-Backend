package main

import (
	"time"
)

func GetCurrentMonthYear() int {
	now := time.Now()
	return int(now.Month())*10000 + now.Year()
}
