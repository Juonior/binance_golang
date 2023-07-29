package binance

import "fmt"

func GetInfo() ([]string, float64, int, int) {
	var sleepTime float64
	fmt.Print("Enter cooldown requests (In Seconds): ")
	fmt.Scan(&sleepTime)

	var proxyAmount int
	fmt.Print("Enter count of proxy: ")
	fmt.Scan(&proxyAmount)

	proxies := make([]string, proxyAmount)
	for i := 0; i < proxyAmount; i++ {
		var proxy string
		fmt.Scan(&proxy)
		proxies[i] = proxy
	}

	fmt.Println("Your proxies:", proxies)

	var userMoney int
	fmt.Print("Enter your max balance for order (RUB): ")
	fmt.Scan(&userMoney)

	var needProfit int
	fmt.Print("Enter min profit: ")
	fmt.Scan(&needProfit)

	return proxies, sleepTime, userMoney, needProfit
}
