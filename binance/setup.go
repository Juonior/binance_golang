package binance

import "fmt"

func GetInfo() ([]string, int, int, int, float64) {
	var sleepTime int
	fmt.Print("Enter cooldown requests (In Millisecond): ")
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

	var user_min_money int
	fmt.Print("Enter your min balance for order (RUB): ")
	fmt.Scan(&user_min_money)

	var user_max_money int
	fmt.Print("Enter your max balance for order (RUB): ")
	fmt.Scan(&user_max_money)

	var need_spread float64
	fmt.Print("Enter min spread (%): ")
	fmt.Scan(&need_spread)

	return proxies, sleepTime, user_min_money, user_max_money, need_spread
}
