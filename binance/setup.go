package binance

import "fmt"

func GetInfo() (float64, int, int, float64) {
	var sleepTime float64
	fmt.Print("Enter cooldown requests (In Millisecond): ")
	fmt.Scan(&sleepTime)

	var user_min_money int
	fmt.Print("Enter your min balance for order (RUB): ")
	fmt.Scan(&user_min_money)

	var user_max_money int
	fmt.Print("Enter your max balance for order (RUB): ")
	fmt.Scan(&user_max_money)

	var need_spread float64
	fmt.Print("Enter min spread (%): ")
	fmt.Scan(&need_spread)

	return sleepTime, user_min_money, user_max_money, need_spread
}
