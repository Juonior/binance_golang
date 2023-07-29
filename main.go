package main

import (
	"fmt"
	"time"

	"p2p/binance"
)

func main() {
	binance.CheckToken()
	proxies, sleepTime, user_min_money, user_max_money, need_spread := binance.GetInfo()
	fmt.Println(proxies, sleepTime, user_min_money, user_max_money, need_spread)
	go binance.CheckSell("USDT", []string{"RosBankNew"}, "http://D6kBsL:7Kqr3J@194.226.60.132:8000")
	assets := []string{"USDT"}
	current_proxy_num := 0
	for {
		for _, asset := range assets {
			if current_proxy_num < len(proxies)-1 {
				current_proxy_num = current_proxy_num + 1
			} else {
				current_proxy_num = 0
			}
			go binance.CheckAsset(user_min_money, user_max_money, need_spread, asset, []string{"RosBankNew"}, proxies[current_proxy_num])
		}
		duration := time.Duration(sleepTime) * time.Millisecond
		time.Sleep(duration)
	}

}
