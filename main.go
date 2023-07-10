package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/antchfx/htmlquery"
)

func main() {
	ponStatus()
	status()
}

func status() {
	url := "http://192.168.1.1/admin/status.asp"

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
		panic(err)
	}
	uptime := htmlquery.FindOne(doc, "/html/body/blockquote/form[1]/table[1]/tbody/tr[3]/td[2]/font")
	cpuUsage := htmlquery.FindOne(doc, "/html/body/blockquote/form[1]/table[1]/tbody/tr[5]/td[2]/font")
	memoryUsage := htmlquery.FindOne(doc, "/html/body/blockquote/form[1]/table[1]/tbody/tr[6]/td[2]/font")

	fmt.Println("运行时间:", uptime.FirstChild.Data)
	fmt.Println("CPU使用率:", cpuUsage.FirstChild.Data)
	fmt.Println("内存使用率:", memoryUsage.FirstChild.Data)
}

func ponStatus() {
	url := "http://192.168.1.1/status_pon.asp"

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
		panic(err)
	}
	temperature := htmlquery.FindOne(doc, "/html/body/blockquote/table[2]/tbody/tr[2]/td[2]/font")
	voltage := htmlquery.FindOne(doc, "/html/body/blockquote/table[2]/tbody/tr[3]/td[2]/font")
	transmitPower := htmlquery.FindOne(doc, "/html/body/blockquote/table[2]/tbody/tr[4]/td[2]/font")
	receivePower := htmlquery.FindOne(doc, "/html/body/blockquote/table[2]/tbody/tr[5]/td[2]/font")
	current := htmlquery.FindOne(doc, "/html/body/blockquote/table[2]/tbody/tr[6]/td[2]/font")
	llidStatus := htmlquery.FindOne(doc, "/html/body/blockquote/table[4]/tbody/tr[3]/td[2]/font/b")

	fmt.Println("温度:", temperature.FirstChild.Data)
	fmt.Println("电压:", voltage.FirstChild.Data)
	fmt.Println("发送功率:", transmitPower.FirstChild.Data)
	fmt.Println("接收功率:", receivePower.FirstChild.Data)
	fmt.Println("偏置电流:", current.FirstChild.Data)
	fmt.Println("EPON LLID Status:", llidStatus.FirstChild.Data)
}
