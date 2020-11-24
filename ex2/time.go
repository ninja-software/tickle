package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now()
	fmt.Println(now)
	nowTc := now.Truncate(time.Hour * 24)
	fmt.Println(nowTc)

	nowUtc := now.UTC()
	fmt.Println(nowUtc)
	nowUtcTc := nowUtc.Truncate(time.Hour * 24)
	fmt.Println(nowUtcTc)
}
