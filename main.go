package main

import (
	"fmt"
	"p2p/binance"
	"time"
)

func main() {
	ipAddresses := binance.GetLocalAddresses()[1:]
	// ipAddresses := []string{"192.168.2.211", "192.168.2.211", "192.168.2.211", "192.168.2.211", "192.168.2.211", "192.168.2.211"}
	binance.CheckToken()
	fmt.Println(len(ipAddresses), "Local IPS:", ipAddresses)
	sleepTime, user_min_money, user_max_money, need_spread := binance.GetInfo()
	fmt.Println(sleepTime, user_min_money, user_max_money, need_spread)
	assets := []string{"BNB", "BTC", "ETH"}
	go binance.CheckSell(assets, []string{"PostBankNew", "RussianStandardBank"}, "http://zUCkzixB:BFRHq5ne@77.90.160.64:62140")
	current_ip_num := 0
	for {
		if current_ip_num < len(ipAddresses)-len(assets) {
			current_ip_num = current_ip_num + 1
		} else {
			current_ip_num = 0
		}
		k := 0
		for _, asset := range assets {

			go binance.CheckAsset(user_min_money, user_max_money, need_spread, asset, []string{"PostBankNew", "RussianStandardBank"}, ipAddresses[current_ip_num+k])
			k += 1
		}
		duration := time.Duration(sleepTime) * time.Millisecond
		time.Sleep(duration)
	}

}
