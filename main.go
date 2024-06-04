package main

import (
	"fmt"
	"p2p/binance"
	"time"
)

var sleepTime float64

func main() {
	ipAddresses := binance.GetLocalAddresses()
	binance.CheckToken()
	fmt.Println(len(ipAddresses), "Local IPS:", ipAddresses)

	fmt.Print("Enter cooldown requests (In Millisecond): ")
	fmt.Scanln(&sleepTime)
	go binance.CheckSell("ASSET_EXAMPLE", []string{"BANK_EXAMPLE1", "BANK_EXAPMLE2"}, "YOUR PROXIES")
	assets := []string{"ASSET_EXAMPLE"}
	current_ip_num := 0
	for {
		for _, asset := range assets {
			if current_ip_num < len(ipAddresses)-1 {
				current_ip_num = current_ip_num + 1
			} else {
				current_ip_num = 0
			}
			go binance.CheckAsset(USERMINLIMIT, USERMAXLIMIT, asset, []]string{"BANK_EXAMPLE1", "BANK_EXAPMLE2"}, ipAddresses[current_ip_num])
		}
		duration := time.Duration(sleepTime) * time.Millisecond
		time.Sleep(duration)
	}

}
