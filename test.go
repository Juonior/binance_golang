package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func getLocalAddresses() []string {
	var ipAddresses []string
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Failed to get network interfaces:", err)
		return nil
	}

	// Перебираем каждый сетевой интерфейс
	for _, i := range interfaces {
		// Получаем адреса для текущего интерфейса
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Println("Failed to get addresses for interface", i.Name, ":", err)
			continue
		}

		// Перебираем каждый адрес для текущего интерфейса
		for _, addr := range addrs {
			// Проверяем, является ли адрес IP-адресом
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			// Проверяем, является ли адрес локальным
			if !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
				// Добавляем IP-адрес в массив
				ipAddresses = append(ipAddresses, ipNet.IP.String())
			}
		}
	}
	return ipAddresses
}

func main() {
	ipAddresses := getLocalAddresses()
	// fmt.Println(ipAddresses)
	settings := make(map[string]interface{})
	file, err := ioutil.ReadFile("settings.json")
	if err != nil {
		return
	} else {
		_ = json.Unmarshal(file, &settings)
	}
	payload := map[string]interface{}{
		"asset":         "USDT",
		"transAmount":   "0",
		"payTypes":      []interface{}{},
		"page":          1,
		"rows":          3,
		"countries":     []interface{}{},
		"publisherType": nil,
		"fiat":          "RUB",
		"tradeType":     "SELL",
		"merchantCheck": false,
	}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return
	}
	payloadbuy := map[string]interface{}{
		"advOrderNumber": 1,
		"matchPrice":     1,
		"totalAmount":    1,
		"asset":          "USDT",
		"fiatUnit":       "RUB",
		"buyType":        "BY_MONEY",
		"tradeType":      "BUY",
		"origin":         "MAKE_TAKE",
	}
	requestBodybuy, err := json.Marshal(payloadbuy)
	if err != nil {
		return
	}
	fastestIPs := make([]string, 0)
	fastestTimes := make([]time.Duration, 0)
	for _, ip := range ipAddresses {
		totalElapsedSearch := time.Duration(0)
		totalElapsedBuy := time.Duration(0)

		for i := 0; i < 3; i++ {
			start := time.Now()
			req, err := http.NewRequest("POST", "https://p2p.binance.com/bapi/c2c/v2/friendly/c2c/adv/search", bytes.NewBuffer(requestBody))
			if err != nil {
				return
			}
			req.Header.Add("clienttype", "web")
			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 YaBrowser/23.7.1.1140 Yowser/2.5 Safari/537.36")
			req.Header.Set("Content-Type", "application/json")
			dialer := &net.Dialer{
				LocalAddr: &net.TCPAddr{
					IP: net.ParseIP(ip),
				},
			}

			// Create an HTTP client with the custom dialer
			client := &http.Client{
				Transport: &http.Transport{
					Dial: dialer.Dial,
				},
			}

			// Send the request
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			// Process the response
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				return
			}

			// Return the response
			elapsed := time.Since(start)
			totalElapsedSearch += elapsed
			// fmt.Println("[SEARCH] Execution time: %s", elapsed)
		}
		for i := 0; i < 3; i++ {
			start := time.Now()
			req, err := http.NewRequest("POST", "https://p2p.binance.com/bapi/c2c/v2/private/c2c/order-match/makeOrder", bytes.NewBuffer(requestBodybuy))
			if err != nil {
				return
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 YaBrowser/23.5.2.625 Yowser/2.5 Safari/537.36")
			req.Header.Set("clienttype", "web")
			req.Header.Set("csrftoken", settings["csrftoken"].(string))
			req.Header.Set("Cookie", settings["cookie"].(string))
			req.Header.Set("Content-Type", "application/json")
			dialer := &net.Dialer{
				LocalAddr: &net.TCPAddr{
					IP: net.ParseIP(ip),
				},
			}

			// Create an HTTP client with the custom dialer
			client := &http.Client{
				Transport: &http.Transport{
					Dial: dialer.Dial,
				},
			}

			// Send the request
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			// Process the response}
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				return
			}

			// Return the response
			elapsed := time.Since(start)
			totalElapsedBuy += elapsed
			// fmt.Println("[BUY] Execution time: %s", elapsed)
		}
		averageElapsedBuy := totalElapsedBuy / 3
		fmt.Println(fmt.Sprintf("[%v]", ip), "[SEARCH] Average:", totalElapsedSearch/3, "| [BUY] Average:", averageElapsedBuy)
		if len(fastestIPs) < 3 || averageElapsedBuy < fastestTimes[2] {
			// Insert the IP and time at the correct position in the sorted arrays
			j := 0
			for j < len(fastestTimes) && averageElapsedBuy > fastestTimes[j] {
				j++
			}
			if j < 3 {
				fastestIPs = append(fastestIPs[:j], append([]string{ip}, fastestIPs[j:]...)...)
				fastestTimes = append(fastestTimes[:j], append([]time.Duration{averageElapsedBuy}, fastestTimes[j:]...)...)
			}
			// Keep only the top 3 fastest IP addresses
			if len(fastestIPs) > 3 {
				fastestIPs = fastestIPs[:3]
				fastestTimes = fastestTimes[:3]
			}
		}
	}
	fmt.Println("TOP-3 HIGH SPEED IP")
	for i := 0; i < len(fastestIPs); i++ {
		fmt.Println(i+1, "IP:", fastestIPs[i], "| Average Buy Speed:", fastestTimes[i])
	}

	var input string
	fmt.Scanln(&input)
}
