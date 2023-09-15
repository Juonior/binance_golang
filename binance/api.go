package binance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var sellData []map[string]interface{}
var last_order_id string
var sellUpdate = time.Now().Format("15:04:05.000000")
var ipAddresses_api []string
var foundMatch bool

func CheckToken() {
	banks_exists := [3]string{"RosBankNew", "TinkoffNew", "PostBankNew"}
	cookie := ""
	settings := make(map[string]interface{})

	file, err := ioutil.ReadFile("settings.json")
	if err != nil {
		fmt.Println("File settings.json not found.")
		csrftoken := ""
		fmt.Print("Enter csrftoken: ")
		fmt.Scan(&csrftoken)
		for {
			cookie := ""
			fmt.Print("Enter cookie: ")
			fmt.Scan(&cookie)
			if strings.Contains(cookie, "p20t") {
				cookie = "p20t=" + strings.Split(strings.Split(cookie, "p20t=")[1], ";")[0]
				break
			} else {
				fmt.Println("Not found key \"p20t\"")
			}
		}
		banks := ""
		fmt.Println("Choose you bank:")
		fmt.Println("- 1. Rosbank")
		fmt.Println("- 2. Tinkoff")
		fmt.Println("- 3. Pochtabank")
		fmt.Println("If you want more banks than 1 please enter banks without vsego. Example: \"12\"")
		fmt.Scan(&banks)
		banksArray := []string{}
		for _, char := range banks {
			i, _ := strconv.Atoi(string(char))
			if i >= 1 && i <= len(banks_exists) {
				banksArray = append(banksArray, banks_exists[i-1])
			} else {
				fmt.Printf("Not valid bank number: %d", i)
			}
		}
		settings["banks"] = banksArray
		settings["cookie"] = cookie
		settings["csrftoken"] = csrftoken

		jsonData, _ := json.Marshal(settings)
		_ = ioutil.WriteFile("settings.json", jsonData, 0644)
	} else {
		_ = json.Unmarshal(file, &settings)
	}

	checkToken_url := "https://p2p.binance.com/bapi/composite/v1/private/inbox/user/token/get"
	req, err := http.NewRequest("POST", checkToken_url, nil)
	if err != nil {
		fmt.Println("Error while reqeust:", err)
		os.Exit(1)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 YaBrowser/23.5.2.625 Yowser/2.5 Safari/537.36")
	req.Header.Set("clienttype", "web")
	req.Header.Set("csrftoken", settings["csrftoken"].(string))
	req.Header.Set("Cookie", settings["cookie"].(string))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error while reqeust:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Token is valid.")
	} else {
		fmt.Println("TOKEN INVALID")
		_, _ = fmt.Scanln()
		os.Exit(1)
	}
}

func MakeOrder(wg *sync.WaitGroup, OrderNumber string, matchPrice string, totalAmount string, asset string, spread float64, profit float64, localIP string, order_responses *[]map[string]interface{}) {
	defer wg.Done()
	start := time.Now()
	settings := make(map[string]interface{})
	file, err := ioutil.ReadFile("settings.json")
	if err != nil {
		return
	} else {
		_ = json.Unmarshal(file, &settings)
	}
	payload := map[string]interface{}{
		"advOrderNumber": OrderNumber,
		"matchPrice":     matchPrice,
		"totalAmount":    totalAmount,
		"asset":          asset,
		"fiatUnit":       "RUB",
		"buyType":        "BY_MONEY",
		"tradeType":      "BUY",
		"origin":         "MAKE_TAKE",
	}
	// Convert payload to JSON
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return
	}
	// Create POST request to makeOrder endpoint
	req, err := http.NewRequest("POST", "https://p2p.binance.com/bapi/c2c/v2/private/c2c/order-match/makeOrder", bytes.NewBuffer(requestBody))
	if err != nil {
		return
	}
	// Set necessary headers
	req.Header.Set("User-Agent", "Windows")
	req.Header.Set("clienttype", "web")
	req.Header.Set("csrftoken", settings["csrftoken"].(string))
	req.Header.Set("Cookie", settings["cookie"].(string))
	req.Header.Set("Content-Type", "application/json")

	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP: net.ParseIP(localIP),
		},
		Timeout: time.Minute * 1,
	}

	// Create an HTTP client with the custom dialer
	client := &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
		Timeout: time.Minute * 1,
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	client.Transport.(*http.Transport).CloseIdleConnections()
	// Process the response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return
	}

	// Return the response
	elapsed := time.Since(start)
	time_after_response := time.Now().Format("15:04:05.000000")
	if response["success"] == true {
		successfulCreation := map[string]interface{}{
			"message":     "Successful creation order Binance",
			"success":     true,
			"totalAmount": totalAmount,
			"profit":      math.Round(profit),
			"spread":      spread,
			"matchPrice":  matchPrice,
			"time":        time_after_response,
			"elapsed":     fmt.Sprintf("%v", elapsed),
			"localIP":     localIP,
			"color":       "#46b000",
		}
		*order_responses = append(*order_responses, successfulCreation)
	} else {
		successfulCreation := map[string]interface{}{
			"message":     fmt.Sprintf("%v", response["message"]),
			"success":     false,
			"totalAmount": totalAmount,
			"profit":      math.Round(profit),
			"spread":      spread,
			"matchPrice":  matchPrice,
			"time":        time_after_response,
			"elapsed":     fmt.Sprintf("%v", elapsed),
			"localIP":     localIP,
			"color":       "#510A1F",
		}
		*order_responses = append(*order_responses, successfulCreation)
	}
}
func BuyInfo(localIP string, asset string, transAmount string, payTypes []string) ([]map[string]interface{}, error) {
	priceInfoURL := "https://p2p.binance.com/bapi/c2c/v2/friendly/c2c/adv/search"

	settings, err := loadSettings()
	if err != nil {
		return nil, err
	}
	payload := map[string]interface{}{
		"asset":         asset,
		"transAmount":   transAmount,
		"payTypes":      payTypes,
		"page":          1,
		"rows":          1,
		"countries":     []interface{}{},
		"publisherType": nil,
		"fiat":          "RUB",
		"tradeType":     "BUY",
		"merchantCheck": false,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Millisecond * 5000,
			LocalAddr: &net.TCPAddr{IP: net.ParseIP(localIP), Port: 0},
		}).DialContext,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Millisecond * 5000,
	}

	req, err := http.NewRequest("POST", priceInfoURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Windows")
	req.Header.Set("clienttype", "web")
	req.Header.Set("csrftoken", settings["csrftoken"].(string))
	req.Header.Set("Cookie", settings["cookie"].(string))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	client.Transport.(*http.Transport).CloseIdleConnections()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	var binanceSellInfo []map[string]interface{}
	for _, info := range data["data"].([]interface{}) {
		adv := info.(map[string]interface{})["adv"].(map[string]interface{})
		advertiser := info.(map[string]interface{})["advertiser"].(map[string]interface{})

		orderInfo := map[string]interface{}{
			"id":           adv["advNo"].(string),
			"amount":       adv["tradableQuantity"].(string),
			"minLimit":     adv["minSingleTransAmount"].(string),
			"price":        adv["price"].(string),
			"maxLimit":     adv["dynamicMaxSingleTransAmount"].(string),
			"name":         advertiser["nickName"].(string),
			"tradeMethods": adv["tradeMethods"].([]interface{}),
			"link":         "https://p2p.binance.com/ru/advertiserDetail?advertiserNo=" + advertiser["userNo"].(string),
		}

		binanceSellInfo = append(binanceSellInfo, orderInfo)
	}
	return binanceSellInfo, err
}

func SellInfo(proxy string, asset string, transAmount string, payTypes []string) ([]map[string]interface{}, error) {
	priceInfoURL := "https://p2p.binance.com/bapi/c2c/v2/friendly/c2c/adv/search"
	payload := map[string]interface{}{
		"asset":         asset,
		"transAmount":   transAmount,
		"payTypes":      payTypes,
		"page":          1,
		"rows":          20,
		"countries":     []interface{}{},
		"publisherType": nil,
		"fiat":          "RUB",
		"tradeType":     "SELL",
		"merchantCheck": false,
	}
	settings := make(map[string]interface{})
	file, err := ioutil.ReadFile("settings.json")
	if err != nil {
		return nil, err
	} else {
		_ = json.Unmarshal(file, &settings)
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	req, err := http.NewRequest("POST", priceInfoURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 YaBrowser/23.5.2.625 Yowser/2.5 Safari/537.36")
	req.Header.Set("clienttype", "web")
	req.Header.Set("csrftoken", settings["csrftoken"].(string))
	req.Header.Set("Cookie", settings["cookie"].(string))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	client.Transport.(*http.Transport).CloseIdleConnections()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	var binanceSellInfo []map[string]interface{}
	for _, info := range data["data"].([]interface{}) {
		adv := info.(map[string]interface{})["adv"].(map[string]interface{})
		advertiser := info.(map[string]interface{})["advertiser"].(map[string]interface{})

		orderInfo := map[string]interface{}{
			"id":           adv["advNo"].(string),
			"amount":       adv["tradableQuantity"].(string),
			"minLimit":     adv["minSingleTransAmount"].(string),
			"price":        adv["price"].(string),
			"maxLimit":     adv["dynamicMaxSingleTransAmount"].(string),
			"name":         advertiser["nickName"].(string),
			"tradeMethods": adv["tradeMethods"].([]interface{}),
			"link":         "https://p2p.binance.com/ru/advertiserDetail?advertiserNo=" + advertiser["userNo"].(string),
		}

		binanceSellInfo = append(binanceSellInfo, orderInfo)
	}

	return binanceSellInfo, nil
}

func CheckSell(asset string, bank interface{}, currentSellIp string) {
	for {
		sellData, _ = SellInfo(currentSellIp, asset, "0", bank.([]string))
		if len(sellData) > 0 {
			sellData = sellData[2:]
		}
		sellUpdate = time.Now().Format("15:04:05.000000")
		time.Sleep(5 * time.Second)
	}
}

func CheckAsset(user_min_limit int, user_max_limit int, need_spread float64, asset string, bank interface{}, currentBuyIp string) {
	buyData, BuyError := BuyInfo(currentBuyIp, asset, "0", bank.([]string))
	if len(buyData) > 0 {
		if len(sellData) > 0 {
			if asset == "USDT" {
				now := time.Now().Format("15:04:05.000000")
				fmt.Println(now, "|", asset, "| BUY:", buyData[0]["price"], "| SELL:", sellData[0]["price"], "| SELL UPDATE:", sellUpdate)
			}
			for _, buyOffer := range buyData {
				buyPrice, _ := strconv.ParseFloat(buyOffer["price"].(string), 64)
				// buyPrice := 80.00
				buyMinLimit, _ := strconv.ParseFloat(buyOffer["minLimit"].(string), 64)
				buyMaxLimit, _ := strconv.ParseFloat(buyOffer["maxLimit"].(string), 64)
				if buyMinLimit > float64(user_max_limit) {
					continue
				}
				for _, sellOffer := range sellData {
					sellPrice, _ := strconv.ParseFloat(sellOffer["price"].(string), 64)
					if sellPrice > buyPrice {
						banks_buy := []string{}
						banks_sell := []string{}
						for _, bank_b := range buyOffer["tradeMethods"].([]interface{}) {
							banks_buy = append(banks_buy, bank_b.(map[string]interface{})["identifier"].(string))
						}
						for _, bank_s := range sellOffer["tradeMethods"].([]interface{}) {
							banks_sell = append(banks_sell, bank_s.(map[string]interface{})["identifier"].(string))
						}

						for _, bank_b := range buyOffer["tradeMethods"].([]interface{}) {
							identifier := bank_b.(map[string]interface{})["identifier"].(string)
							if !containsString(banks_sell, identifier) || !containsString(banks_buy, identifier) {
								foundMatch = false
								break
							}
							foundMatch = true
						}
						if foundMatch {
							if buyMaxLimit < float64(user_min_limit) || buyMinLimit > float64(user_max_limit) {
								continue
							}
							spread := math.Round((sellPrice/buyPrice*100-100)*100) / 100
							canBuy := math.Min(float64(user_max_limit), buyMaxLimit)
							if spread > 5 {
								canBuy = math.Min(canBuy, 90000)
							}
							if (spread < 5) || (spread > 5 && canBuy <= 90000) {
								profit := ((canBuy / buyPrice) * sellPrice) - canBuy
								if last_order_id != buyOffer["id"].(string) {
									if spread >= float64(need_spread) {
										last_order_id = buyOffer["id"].(string)
										var wg sync.WaitGroup
										wg.Add(6)
										order_responses := make([]map[string]interface{}, 0)
										for i := 0; i < 6; i++ {
											go MakeOrder(&wg, buyOffer["id"].(string), buyOffer["price"].(string), strconv.FormatFloat(canBuy, 'f', -1, 64), asset, spread, profit, ipAddresses_api[i], &order_responses)
										}
										amount, _ := strconv.ParseFloat(buyOffer["amount"].(string), 64)
										price, _ := strconv.ParseFloat(buyOffer["price"].(string), 64)
										minlim, _ := strconv.ParseFloat(buyOffer["minLimit"].(string), 64)
										maxlim, _ := strconv.ParseFloat(buyOffer["maxLimit"].(string), 64)
										amount_in_rub := math.Round(amount) * math.Round(price)
										merchant_name := buyOffer["name"].(string)
										SendWebhookMonitor(math.Round(amount_in_rub), spread, buyOffer["price"].(string), asset, math.Round(minlim), math.Round(maxlim), merchant_name, fmt.Sprintf("%v", banks_buy), "67008c")
										wg.Wait()

										orderInfo := make(map[string]interface{})
										for _, order := range order_responses {
											if order["success"].(bool) {
												orderInfo = order
												break
											}
										}

										if len(orderInfo) > 0 {
											go SendWebhook(orderInfo["message"].(string), orderInfo["totalAmount"].(string), orderInfo["profit"].(float64), orderInfo["spread"].(float64), orderInfo["matchPrice"].(string), orderInfo["time"].(string), orderInfo["elapsed"].(string), orderInfo["localIP"].(string), orderInfo["color"].(string))
										} else {
											go SendWebhook(order_responses[0]["message"].(string), order_responses[0]["totalAmount"].(string), order_responses[0]["profit"].(float64), order_responses[0]["spread"].(float64), order_responses[0]["matchPrice"].(string), order_responses[0]["time"].(string), order_responses[0]["elapsed"].(string), order_responses[0]["localIP"].(string), order_responses[0]["color"].(string))
										}
										return
									}
								}
							}
						}
					}
				}
			}
		}
	} else {

		fmt.Println("BuyData None", currentBuyIp, BuyError)
		fmt.Println(buyData)
	}

	return
}
