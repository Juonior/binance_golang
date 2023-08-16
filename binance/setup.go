package binance

import (
	"bufio"
	"fmt"
	"os"
)

func GetInfo() ([]string, int, int, int, float64) {
	var count int
	fmt.Print("Введите количество прокси: ")
	fmt.Scan(&count)
	proxies := make([]string, count)

	// Ввод адресов прокси
	scanner := bufio.NewScanner(os.Stdin)
	for i := 0; i < count; i++ {
		fmt.Printf("Введите прокси #%d: ", i+1)
		scanner.Scan()
		proxies[i] = scanner.Text()
	}
	var sleepTime int
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

	return proxies, sleepTime, user_min_money, user_max_money, need_spread
}
