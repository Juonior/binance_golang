package main

import (
	"fmt"
	"time"

	"p2p/binance"
)

func main() {
	binance.CheckToken()
	proxies, sleepTime, user_min_money, user_max_money, need_spread := binance.GetInfo()
	fmt.Println(len(proxies), "Proxies:", proxies)
	fmt.Println(sleepTime, user_min_money, user_max_money, need_spread)
	go binance.CheckSell("BTC", []string{"PostBankNew"}, "http://user129953:97w4lt@45.135.248.216:2045")
	assets := []string{"BTC"}
	current_proxy_num := 0
	for {
		for _, asset := range assets {
			if current_proxy_num < len(proxies)-1 {
				current_proxy_num = current_proxy_num + 1
			} else {
				current_proxy_num = 0
			}
			go binance.CheckAsset(user_min_money, user_max_money, need_spread, asset, []string{"PostBankNew"}, proxies[current_proxy_num])
		}
		duration := time.Duration(sleepTime) * time.Millisecond
		time.Sleep(duration)
	}

}
