package main

import (
	"bytes"
	"github.com/antchfx/htmlquery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	go func() {
		intervalValue := 10
		intervalStr := os.Getenv("SCRAP_INTERVAL")
		if intervalStr != "" {
			interval, err := strconv.Atoi(intervalStr)
			if err != nil {
				log.Fatal("无法将环境变量转换为整数:", err)
			}
			intervalValue = interval
		}
		log.Println("SCRAP_INTERVAL: ", intervalValue)
		for range time.Tick(time.Duration(intervalValue) * time.Second) {
			log.Println("开始获取 Pon Stick 状态")
			login()
			ponStatus()
			status()
		}
	}()
	log.Println("服务启动完成")
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9001", nil))
}

func login() {
	log.Println("开始登陆")
	url := "http://192.168.1.1/boaform/admin/formLogin"
	payload := "challenge=&username=admin&password=admin&save=%E7%99%BB%E5%BD%95&submit-url=%2Fadmin%2Flogin.asp"

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

func status() {
	log.Println("开始获取状态")

	doc, _ := getData("http://192.168.1.1/admin/status.asp")
	uptimeDoc := htmlquery.FindOne(doc, "/html/body/blockquote/form[1]/table[1]/tbody/tr[3]/td[2]/font")
	uptimeDesc := uptimeDoc.FirstChild.Data
	uptime := parseUptime(uptimeDoc)
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

func parseUptime(uptimeDoc *html.Node) float64 {
	uptimeString := uptimeDoc.LastChild.Data
	uptimeString = strings.ReplaceAll(uptimeString, " ", "")
	log.Println("运行时间: ", uptimeString)
	// 24h以内: 20:13
	if strings.Contains(uptimeString, ":") {
		values := strings.Split(uptimeString, ":")
		hours, err := strconv.Atoi(values[0])
		if err != nil {
			log.Println("解析时间-小时错误: ", uptimeString)
			return -1
		}
		minutes, err := strconv.Atoi(values[1])
		if err != nil {
			log.Println("解析时间-分钟错误: ", uptimeString)
			return -1
		}

		uptimeSecond := (hours*60 + minutes) * 60
		return float64(uptimeSecond)
	} else if strings.Contains(uptimeString, "day") {
		// 大于24h: 1 day, 0 min
		values := strings.Split(uptimeString, "day,")
		days, err := strconv.Atoi(values[0])
		if err != nil {
			log.Println("解析时间-天错误: ", uptimeString)
			return -1
		}
		values = strings.Split(values[1], "min")
		minutes, err := strconv.Atoi(values[0])
		if err != nil {
			log.Println("解析时间-分钟错误: ", uptimeString)
			return -1
		}

		uptimeSecond := (days*24*60 + minutes) * 60
		return float64(uptimeSecond)
	} else {
		// 小于1h: 10
		minutes, err := strconv.Atoi(uptimeString)
		if err != nil {
			log.Println("解析时间-天错误: ", uptimeString)
			return -1
		}
		uptimeSecond := minutes * 60
		return float64(uptimeSecond)
	}
	//log.Println("未知时间类型: ", uptimeString)
	//return -2
}

func ponStatus() {
	log.Println("开始获取 pon 状态")
	doc, _ := getData("http://192.168.1.1/status_pon.asp")

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
