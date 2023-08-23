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
	"sort"
	"strconv"
	"strings"
	"time"
)

const discordSuccessURL = "https://discord.com/api/webhooks/1131334900272857089/DzxJQ8wD-EMcnl65Ev4ww5I6aVcYxSw25LCihHIloRwU-1anlLKpEzV_1b6w0mMBXY1l"
const discordMonitorURL = "https://discord.com/api/webhooks/1142911381134393447/Z-YLp-wI-ObOWmYsEMzPapl7IA-M6kM-Rl4dRIExxYj1PZRpWnGcroK54u8PjP0C4MOG"

var sellData []map[string]interface{}
var last_order_id string
var sellUpdate = time.Now().Format("15:04:05.000000")
var ipAddresses_api []string

type Message struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds"`
}

type Embed struct {
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	Fields      []Field    `json:"fields,omitempty"`
	Color       int        `json:"color,omitempty"`
	Thumbnail   *Thumbnail `json:"thumbnail,omitempty"`
	Footer      *Footer    `json:"footer,omitempty"`
}
type Thumbnail struct {
	URL string `json:"url,omitempty"`
}

type Footer struct {
	Text string `json:"text,omitempty"`
}
type Field struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

func formatNum(numFloat float64) string {
	num := fmt.Sprintf("%v", numFloat)
	formattedStr := ""
	for i := 0; i < len(num); i++ {
		formattedStr += string(num[i])
		if (len(num)-i-1)%3 == 0 && i != len(num)-1 {
			formattedStr += " "
		}
	}
	return formattedStr
}
func SendWebhook(status string, amount string, profit float64, spread float64, price string, orderTime, requestTime, ip string, color string) {
	embed := Embed{
		Title: fmt.Sprintf("[JP] %v", status),
		Color: parseColor(color),
		Fields: []Field{
			{
				Name:  "Buy Amount",
				Value: fmt.Sprintf("%v руб.", amount),
			},
			{
				Name:  "Profit",
				Value: fmt.Sprintf("%v руб.", profit),
			},
			{
				Name:  "Spread",
				Value: fmt.Sprintf("%v%%", spread),
			},
			{
				Name:  "Price",
				Value: fmt.Sprintf("%v руб.", price),
			},
		},
		Thumbnail: &Thumbnail{
			URL: "https://cdn-icons-png.flaticon.com/512/6163/6163319.png",
		},
		Footer: &Footer{
			Text: fmt.Sprintf("%v  | %v | %v", orderTime, requestTime, ip),
		},
	}
	message := Message{
		Embeds: []Embed{embed},
	}

	body, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	resp, err := http.Post(discordSuccessURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		fmt.Println("Error sending webhook:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Webhook sent successfully.")
}

func SendWebhookMonitor(amount float64, spread float64, price string, fiat string, minlim float64, maxlim float64, trader string, color string) {
	embed := Embed{
		Title: "Binance Order",
		Color: parseColor(color),
		Fields: []Field{
			{
				Name:  "Amount",
				Value: fmt.Sprintf("%v руб.", formatNum(amount)),
			},
			{
				Name:  "Spread",
				Value: fmt.Sprintf("%v%%", spread),
			},
			{
				Name:  "Price",
				Value: fmt.Sprintf("%v руб.", price),
			},
			{
				Name:  "Limits",
				Value: fmt.Sprintf("%v - %v", formatNum(minlim), formatNum(maxlim)),
			},
			{
				Name:  "Trader",
				Value: fmt.Sprintf("%v", trader),
			},
			{
				Name:  "Crypto-Fiat",
				Value: fmt.Sprintf("%v-RUB", fiat),
			},
		},
		Thumbnail: &Thumbnail{
			URL: "https://cdn-icons-png.flaticon.com/512/6163/6163319.png",
		},
	}
	message := Message{
		Embeds: []Embed{embed},
	}

	body, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	resp, err := http.Post(discordMonitorURL, "application/json", strings.NewReader(string(body)))
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

func GetLocalAddresses() []string {
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
				ipAddresses_api = append(ipAddresses, ipNet.IP.String())
			}
		}
	}
	return ipAddresses
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

func MakeOrder(OrderNumber string, matchPrice string, totalAmount string, asset string, spread float64, profit float64, localIP string) {
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

	// Process the response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return
	}

	// Return the response
	elapsed := time.Since(start)
	// fmt.Printf("Execution time: %s", elapsed)
	// status string, amount string, profit string, spread string, price string, orderTime, requestTime, color string
	time_after_response := time.Now().Format("15:04:05.000000")
	if response["success"] == true {
		go SendWebhook("Successful creation order Binance", totalAmount, math.Round(profit), spread, matchPrice, time_after_response, fmt.Sprintf("%v", elapsed), localIP, "#46b000")
	} else {
		go SendWebhook(fmt.Sprintf("%v", response["message"]), totalAmount, math.Round(profit), spread, matchPrice, time_after_response, fmt.Sprintf("%v", elapsed), localIP, "#510A1F")
		// go SendWebhook(fmt.Sprintf("[JP] [%v%%] %v | Profit: %v руб. | Price: %v RUB | Amount: %v RUB | Message: %v | Request time: %v", spread, time_after_response, math.Round(profit), matchPrice, totalAmount, response["message"], elapsed), "#510A1F")
	}
}
func BuyInfo(localIP string, asset string, transAmount string, payTypes []string) []map[string]interface{} {
	priceInfoURL := "https://p2p.binance.com/bapi/c2c/v2/friendly/c2c/adv/search"

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
		return nil
	}

	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP: net.ParseIP(localIP),
		},
		Timeout: time.Second * 1,
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
		Timeout: time.Second * 1,
	}

	response, err := httpClient.Post(priceInfoURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil
	}
	defer response.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return nil
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

	return binanceSellInfo
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

	// dialer := &net.Dialer{
	// 	LocalAddr: &net.TCPAddr{
	// 		IP: net.ParseIP(localIP),
	// 	},
	// 	Timeout: time.Second * 1,
	// }

	// // Create an HTTP client with the custom dialer
	// httpClient := &http.Client{
	// 	Transport: &http.Transport{
	// 		Dial: dialer.Dial,
	// 	},
	// 	Timeout: time.Second * 1,
	// }

	response, err := client.Post(priceInfoURL, "application/json", bytes.NewBuffer(requestBody))
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
func CheckAsset(user_min_limit int, user_max_limit int, need_spread float64, asset string, bank interface{}, currentBuyProxy string) {
	buyData := BuyInfo(currentBuyProxy, asset, "0", bank.([]string))
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
				for _, sellOffer := range sellData {
					sellPrice, _ := strconv.ParseFloat(sellOffer["price"].(string), 64)
					if sellPrice > buyPrice {
						if buyMaxLimit < float64(user_min_limit) || buyMinLimit > float64(user_max_limit) {
							continue
						}
						spread := math.Round((sellPrice/buyPrice*100-100)*100) / 100
						canBuy := math.Min(float64(user_max_limit), buyMaxLimit)
						if spread > 5 {
							canBuy = math.Min(canBuy, 90000)
						}
						result := []interface{}{((canBuy / buyPrice) * sellPrice) - canBuy, canBuy, buyOffer, sellOffer, spread}
						if (spread < 5) || (spread > 5 && canBuy <= 90000) {
							resultOptions = append(resultOptions, result)
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
						fmt.Println(order_info)
						last_order_id = order_info["id"].(string)
						for i := 0; i < 3; i++ {
							go MakeOrder(order_info["id"].(string), order_info["price"].(string), canBuyStr, asset, spread, profit, ipAddresses_api[i])
						}
						amount, _ := strconv.ParseFloat(order_info["amount"].(string), 64)
						price, _ := strconv.ParseFloat(order_info["price"].(string), 64)
						minlim, _ := strconv.ParseFloat(order_info["minLimit"].(string), 64)
						maxlim, _ := strconv.ParseFloat(order_info["maxLimit"].(string), 64)
						amount_in_rub := math.Round(amount) * math.Round(price)
						merchant_name := order_info["name"].(string)
						SendWebhookMonitor(math.Round(amount_in_rub), spread, order_info["price"].(string), asset, math.Round(minlim), math.Round(maxlim), merchant_name, "67008c")
					}
				}
			}
		}
	} else {
		fmt.Println("BuyData None", currentBuyProxy)
	}
}
