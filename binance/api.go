package binance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const discordWebhookURL = "https://discord.com/api/webhooks/1131334900272857089/DzxJQ8wD-EMcnl65Ev4ww5I6aVcYxSw25LCihHIloRwU-1anlLKpEzV_1b6w0mMBXY1l"

var sellData []map[string]interface{}
var last_order_id string
var sellUpdate = time.Now().Format("15:04:05.000000")

type Message struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds"`
}

type Embed struct {
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	Fields      []Field `json:"fields,omitempty"`
	Color       int     `json:"color,omitempty"`
}

type Field struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

func SendWebhook(title string, color string) {
	embed := Embed{
		Title: title,
		Color: parseColor(color),
	}

	message := Message{
		Embeds: []Embed{embed},
	}

	body, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	resp, err := http.Post(discordWebhookURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		fmt.Println("Error sending webhook:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Webhook sent successfully.")
}
func parseColor(color string) int {
	color = strings.TrimPrefix(color, "#")
	var value int
	_, err := fmt.Sscanf(color, "%x", &value)
	if err != nil {
		fmt.Println("Error parsing color:", err)
		return 0
	}
	return value
}
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

func MakeOrder(OrderNumber string, matchPrice string, totalAmount string, asset string, spread float64, profit float64) {
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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 YaBrowser/23.5.2.625 Yowser/2.5 Safari/537.36")
	req.Header.Set("clienttype", "web")
	req.Header.Set("csrftoken", settings["csrftoken"].(string))
	req.Header.Set("Cookie", settings["cookie"].(string))
	req.Header.Set("Content-Type", "application/json")

	// Parse proxy URL
	// proxyURL, err := url.Parse(proxy)
	// if err != nil {
	// 	return nil, err, 0
	// }
	// t := http.DefaultTransport.(*http.Transport).Clone()
	// t.Proxy = http.ProxyURL(proxyURL)
	// client := &http.Client{Transport: t}
	client := &http.Client{}

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
	// fmt.Printf("Execution time: %s", elapsed)

	time_after_response := time.Now().Format("15:04:05.000000")
	if response["success"] == true {
		go SendWebhook(fmt.Sprintf("[JP][%v%%] %v | Profit: %v руб. | Amount: %v RUB | Successfully created! | Request time: %v", spread, time_after_response, math.Round(profit), totalAmount, elapsed), "#46b000")
	} else {
		go SendWebhook(fmt.Sprintf("[JP] [%v%%] %v | Profit: %v руб. | Amount: %v RUB | Message: %v | Request time: %v", spread, time_after_response, math.Round(profit), totalAmount, response["message"], elapsed), "#510A1F")
	}
}
func BuyInfo(proxy string, asset string, transAmount string, payTypes []string) ([]map[string]interface{}, error) {
	priceInfoURL := "https://p2p.binance.com/bapi/c2c/v2/friendly/c2c/adv/search"
	proxies := map[string]string{
		"http":  proxy,
		"https": proxy,
	}

	proxyURL, err := url.Parse(proxies["http"])
	if err != nil {
		log.Fatal(err)
	}

	payload := map[string]interface{}{
		"asset":         asset,
		"transAmount":   transAmount,
		"payTypes":      payTypes,
		"page":          1,
		"rows":          3,
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

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	response, err := httpClient.Post(priceInfoURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&data)
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

func SellInfo(proxy string, asset string, transAmount string, payTypes []string) ([]map[string]interface{}, error) {
	priceInfoURL := "https://p2p.binance.com/bapi/c2c/v2/friendly/c2c/adv/search"
	proxies := map[string]string{
		"http":  proxy,
		"https": proxy,
	}

	proxyURL, err := url.Parse(proxies["http"])
	if err != nil {
		log.Fatal(err)
	}

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

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	response, err := httpClient.Post(priceInfoURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&data)
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

func CheckSell(asset string, bank interface{}, currentSellProxy string) {
	for {
		sellData, _ = SellInfo(currentSellProxy, asset, "0", bank.([]string))
		if len(sellData) > 0 {
			sellData = sellData[2:]
		}
		sellUpdate = time.Now().Format("15:04:05.000000")
		time.Sleep(5 * time.Second)
	}
}
func CheckAsset(user_min_limit int, user_max_limit int, need_spread float64, asset string, bank interface{}, currentBuyProxy string) {
	buyData, _ := BuyInfo(currentBuyProxy, asset, "0", bank.([]string))
	// buyData = buyData[:1]
	if len(buyData) > 0 {
		if len(sellData) > 0 {
			if asset == "USDT" {
				now := time.Now().Format("15:04:05.000000")
				fmt.Println(now, "|", asset, "| BUY:", buyData[0]["price"], "| SELL:", sellData[0]["price"], "| SELL UPDATE:", sellUpdate)
			}
			// fmt.Println(sellData)
			resultOptions := []interface{}{}
			for _, buyOffer := range buyData {
				buyPrice, _ := strconv.ParseFloat(buyOffer["price"].(string), 64)
				// buyPrice := 80.00
				buyMinLimit, _ := strconv.ParseFloat(buyOffer["minLimit"].(string), 64)
				buyMaxLimit, _ := strconv.ParseFloat(buyOffer["maxLimit"].(string), 64)
				if buyMinLimit > float64(user_max_limit) {
					continue
				}
				possiblyBuyAmount := []interface{}{}
				for _, sellOffer := range sellData {
					sellPrice, _ := strconv.ParseFloat(sellOffer["price"].(string), 64)
					if sellPrice > buyPrice {
						sellMinLimit, _ := strconv.ParseFloat(sellOffer["minLimit"].(string), 64)
						sellMaxLimit, _ := strconv.ParseFloat(sellOffer["maxLimit"].(string), 64)

						if sellMinLimit > float64(user_max_limit) || buyMinLimit > sellMaxLimit || buyMaxLimit < sellMinLimit || sellMaxLimit < float64(user_min_limit) {
							continue
						}

						if buyMaxLimit == sellMinLimit && (sellMinLimit > float64(user_min_limit) || sellMinLimit < float64(user_max_limit)) {
							canBuy := buyMaxLimit
							spread := math.Round((sellPrice/buyPrice*100-100)*100) / 100
							result := []interface{}{((canBuy / buyPrice) * sellPrice) - canBuy, canBuy, buyOffer, sellOffer, spread}
							resultOptions = append(resultOptions, result)
						} else if buyMinLimit == sellMaxLimit && (sellMaxLimit > float64(user_min_limit) || sellMaxLimit < float64(user_max_limit)) {
							canBuy := buyMinLimit
							spread := math.Round((sellPrice/buyPrice*100-100)*100) / 100
							result := []interface{}{((canBuy / buyPrice) * sellPrice) - canBuy, canBuy, buyOffer, sellOffer, spread}
							resultOptions = append(resultOptions, result)
						} else {
							possiblyBuyAmount = append(possiblyBuyAmount, []interface{}{buyMinLimit, 'b'})
							possiblyBuyAmount = append(possiblyBuyAmount, []interface{}{buyMaxLimit, 'b'})
							possiblyBuyAmount = append(possiblyBuyAmount, []interface{}{sellMinLimit, 's'})
							possiblyBuyAmount = append(possiblyBuyAmount, []interface{}{sellMaxLimit, 's'})
							sort.Slice(possiblyBuyAmount, func(i, j int) bool {
								return possiblyBuyAmount[i].([]interface{})[0].(float64) < possiblyBuyAmount[j].([]interface{})[0].(float64)
							})
							possibly_buy_interval := []float64{possiblyBuyAmount[1].([]interface{})[0].(float64), possiblyBuyAmount[2].([]interface{})[0].(float64)}
							if float64(user_min_limit) <= possibly_buy_interval[0] && possibly_buy_interval[0] <= float64(user_max_limit) {
								canBuy := math.Min(float64(user_max_limit), possibly_buy_interval[1])
								spread := math.Round((sellPrice/buyPrice*100-100)*100) / 100
								result := []interface{}{((canBuy / buyPrice) * sellPrice) - canBuy, canBuy, buyOffer, sellOffer, spread}
								resultOptions = append(resultOptions, result)
							}

						}

					}
				}
			}
			sort.Slice(resultOptions, func(i, j int) bool {
				return resultOptions[i].([]interface{})[4].(float64) > resultOptions[j].([]interface{})[4].(float64)
			})
			if len(resultOptions) > 0 {
				profit := resultOptions[0].([]interface{})[0].(float64)
				spread := resultOptions[0].([]interface{})[4].(float64)
				order_info := resultOptions[0].([]interface{})[2].(map[string]interface{})
				if last_order_id != order_info["id"].(string) {
					if spread >= float64(need_spread) {
						canBuy := resultOptions[0].([]interface{})[1].(float64)
						canBuyStr := strconv.FormatFloat(canBuy, 'f', -1, 64)
						last_order_id = order_info["id"].(string)
						for i := 0; i < 5; i++ {
							go MakeOrder(order_info["id"].(string), order_info["price"].(string), canBuyStr, asset, spread, profit)
						}
					}
				}
			}
		}
	} else {
		fmt.Println("BuyData None", currentBuyProxy)
	}
}
