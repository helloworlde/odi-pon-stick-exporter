package main

import (
	"bytes"
	"github.com/antchfx/htmlquery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	str2duration "github.com/xhit/go-str2duration"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	uptimeMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pon_stick_uptime",
		Help: "运行时间",
	}, []string{})
	uptimeDescMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pon_stick_uptime_desc",
		Help: "运行时间",
	}, []string{"desc"})
	cpuUsageMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pon_stick_cpu_usage",
		Help: "CPU使用率(%)",
	}, []string{})
	memoryUsageMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pon_stick_memory_usage",
		Help: "内存使用率(%)",
	}, []string{})

	temperatureMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pon_stick_temperature",
		Help: "温度(C)",
	}, []string{})
	voltageMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pon_stick_voltage",
		Help: "电压(V)",
	}, []string{})
	transmitPowerMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pon_stick_transmit_power",
		Help: "发送功率(dBm)",
	}, []string{})
	receivePowerMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pon_stick_receive_power",
		Help: "接收功率(dBm)",
	}, []string{})
	currentMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pon_stick_current",
		Help: "偏置电流(mA)",
	}, []string{})
	llidStatusMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pon_stick_llid_status",
		Help: "EPON LLID Status",
	}, []string{})
)

func init() {
	log.Println("服务初始化")
	prometheus.MustRegister(uptimeMetrics)
	prometheus.MustRegister(uptimeDescMetrics)
	prometheus.MustRegister(cpuUsageMetrics)
	prometheus.MustRegister(memoryUsageMetrics)
	prometheus.MustRegister(temperatureMetrics)
	prometheus.MustRegister(voltageMetrics)
	prometheus.MustRegister(transmitPowerMetrics)
	prometheus.MustRegister(receivePowerMetrics)
	prometheus.MustRegister(currentMetrics)
	prometheus.MustRegister(llidStatusMetrics)
}

func main() {
	log.Println("服务开始启动")
	routerAddr := getEnv("ROUTER_ADDR", "192.168.1.1")
	username := getEnv("USERNAME", "admin")
	password := getEnv("PASSWORD", "admin")
	log.Println("ROUTER_ADDR:", routerAddr, ", USERNAME:", username, ", PASSWORD:", password)
	go func() {
		intervalStr := getEnv("SCRAP_INTERVAL", "60")
		interval, err := strconv.Atoi(intervalStr)
		if err != nil {
			log.Fatal("无法将环境变量转换为整数:", err)
		}
		log.Println("SCRAP_INTERVAL: ", interval)
		for range time.Tick(time.Duration(interval) * time.Second) {
			log.Println("开始获取 Pon Stick 状态")
			login(routerAddr, username, password)
			ponStatus(routerAddr)
			status(routerAddr)
		}
	}()
	log.Println("服务启动完成")
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9001", nil))
}

func getEnv(name, defaultValue string) string {
	value := os.Getenv(name)
	if value != "" {
		return value
	} else {
		return defaultValue
	}
}

func login(routerAddr, username, password string) {
	log.Println("开始登陆")
	url := "http://" + routerAddr + "/boaform/admin/formLogin"
	payload := "challenge=&username=" + username + "&password=" + password + "&save=%E7%99%BB%E5%BD%95&submit-url=%2Fadmin%2Flogin.asp"

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(payload))
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh,en;q=0.9,zh-CN;q=0.8,en-US;q=0.7")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("DNT", "1")
	req.Header.Set("Origin", "http://192.168.1.1")
	req.Header.Set("Referer", "http://192.168.1.1/admin/login.asp")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	req.Header.Set("sec-gpc", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request:", err)
	}
	defer resp.Body.Close()

	log.Println("登陆成功")
}

func status(routerAddr string) {
	log.Println("开始获取状态")

	doc, _ := getData("http://" + routerAddr + "/admin/status.asp")
	uptimeDoc := htmlquery.FindOne(doc, "/html/body/blockquote/form[1]/table[1]/tbody/tr[3]/td[2]/font")
	uptimeDesc := uptimeDoc.FirstChild.Data
	uptime := parseDuration(uptimeDoc.LastChild.Data)
	cpuUsage := parsePercentValue(htmlquery.FindOne(doc, "/html/body/blockquote/form[1]/table[1]/tbody/tr[5]/td[2]/font"))
	memoryUsage := parsePercentValue(htmlquery.FindOne(doc, "/html/body/blockquote/form[1]/table[1]/tbody/tr[6]/td[2]/font"))

	log.Println("运行时间:", uptime)
	log.Println("CPU使用率:", cpuUsage)
	log.Println("内存使用率:", memoryUsage)

	uptimeMetrics.WithLabelValues().Set(uptime)
	uptimeDescMetrics.WithLabelValues(uptimeDesc).Set(uptime)
	cpuUsageMetrics.WithLabelValues().Set(cpuUsage)
	memoryUsageMetrics.WithLabelValues().Set(memoryUsage)
}

func parseDuration(durationStr string) float64 {
	durationStr = strings.ReplaceAll(durationStr, ",", "")
	durationStr = strings.ReplaceAll(durationStr, " ", "")

	if strings.Contains(durationStr, "days") {
		durationStr = strings.ReplaceAll(durationStr, "days", "d")
	}
	if strings.Contains(durationStr, "day") {
		durationStr = strings.ReplaceAll(durationStr, "day", "d")
	}
	if strings.Contains(durationStr, ":") {
		durationStr = strings.ReplaceAll(durationStr, ":", "h")
	}
	if strings.Contains(durationStr, "min") {
		durationStr = strings.ReplaceAll(durationStr, "min", "m")
	}

	if !strings.HasSuffix(durationStr, "m") {
		durationStr = durationStr + "m"
	}

	duration, err := str2duration.Str2Duration(durationStr)
	if err != nil {
		log.Println("解析时间错误: ", durationStr, err)
		return -1
	}

	return duration.Seconds()
}

func ponStatus(routerAddr string) {
	log.Println("开始获取 pon 状态")
	doc, _ := getData("http://" + routerAddr + "/status_pon.asp")

	temperature := parseFloatValue(htmlquery.FindOne(doc, "/html/body/blockquote/table[2]/tbody/tr[2]/td[2]/font"))
	voltage := parseFloatValue(htmlquery.FindOne(doc, "/html/body/blockquote/table[2]/tbody/tr[3]/td[2]/font"))
	transmitPower := parseFloatValue(htmlquery.FindOne(doc, "/html/body/blockquote/table[2]/tbody/tr[4]/td[2]/font"))
	receivePower := parseFloatValue(htmlquery.FindOne(doc, "/html/body/blockquote/table[2]/tbody/tr[5]/td[2]/font"))
	current := parseFloatValue(htmlquery.FindOne(doc, "/html/body/blockquote/table[2]/tbody/tr[6]/td[2]/font"))
	llidStatus := parseFloatValue(htmlquery.FindOne(doc, "/html/body/blockquote/table[4]/tbody/tr[3]/td[2]/font/b"))

	log.Println("温度:", temperature)
	log.Println("电压:", voltage)
	log.Println("发送功率:", transmitPower)
	log.Println("接收功率:", receivePower)
	log.Println("偏置电流:", current)
	log.Println("EPON LLID Status:", llidStatus)

	temperatureMetrics.WithLabelValues().Set(temperature)
	voltageMetrics.WithLabelValues().Set(voltage)
	transmitPowerMetrics.WithLabelValues().Set(transmitPower)
	receivePowerMetrics.WithLabelValues().Set(receivePower)
	currentMetrics.WithLabelValues().Set(current)
	llidStatusMetrics.WithLabelValues().Set(llidStatus)

}

func parsePercentValue(node *html.Node) float64 {
	value := strings.Split(node.FirstChild.Data, "%")[0]
	// 将字符串转换为浮点数
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Fatal("转换失败")
	}

	return result
}
func parseFloatValue(node *html.Node) float64 {
	value := strings.Split(node.FirstChild.Data, " ")[0]
	// 将字符串转换为浮点数
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Fatal("转换失败")
	}

	return result
}

func getData(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc, err := htmlquery.Parse(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}
	return doc, nil
}
