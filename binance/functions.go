package binance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

const discordSuccessURL = "https://discord.com/api/webhooks/1131334900272857089/DzxJQ8wD-EMcnl65Ev4ww5I6aVcYxSw25LCihHIloRwU-1anlLKpEzV_1b6w0mMBXY1l"
const discordMonitorURL = "https://discord.com/api/webhooks/1142911381134393447/Z-YLp-wI-ObOWmYsEMzPapl7IA-M6kM-Rl4dRIExxYj1PZRpWnGcroK54u8PjP0C4MOG"

func containsString(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}
func loadSettings() (map[string]interface{}, error) {
	file, err := ioutil.ReadFile("settings.json")
	if err != nil {
		return nil, err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(file, &settings); err != nil {
		return nil, err
	}

	return settings, nil
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

func SendWebhookMonitor(amount float64, spread float64, price string, fiat string, minlim float64, maxlim float64, trader string, banks string, color string) {
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
			{
				Name:  "Banks",
				Value: fmt.Sprintf("%v", banks),
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
