package main

import (
	"fmt"
	"p2p/binance"
	"time"
)

func main() {
	ipAddresses := binance.GetLocalAddresses()
	binance.CheckToken()
	fmt.Println(len(ipAddresses), "Local IPS:", ipAddresses)
	sleepTime, user_min_money, user_max_money, need_spread := binance.GetInfo()
	fmt.Println(sleepTime, user_min_money, user_max_money, need_spread)
	go binance.CheckSell("USDT", []string{"PostBankNew", "RussianStandardBank"}, "http://zUCkzixB:BFRHq5ne@77.90.160.64:62140")
	assets := []string{"USDT"}
	current_ip_num := 0
	for {
		for _, asset := range assets {
			if current_ip_num < len(ipAddresses)-1 {
				current_ip_num = current_ip_num + 1
			} else {
				current_ip_num = 0
			}
			go binance.CheckAsset(user_min_money, user_max_money, need_spread, asset, []string{"PostBankNew", "RussianStandardBank"}, ipAddresses[current_ip_num])
		}
		duration := time.Duration(sleepTime) * time.Millisecond
		time.Sleep(duration)
	}

}
