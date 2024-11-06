package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	host        = "192.168.200.2:801"
	referer     = "http://192.168.200.2/"
	requestUrl  = "http://192.168.200.2:801/eportal/"
	c           = "Portal"
	a           = "login"
	callback    = "dr1003"
	loginMethod = "1"
	jsVersion   = "3.3.3"
)

type NetworkInterface struct {
	Name string
	Ipv4 string
	Ipv6 string
	Mac  string
}

type AuthResult struct {
	Result  string `json:"result"`
	Msg     string `json:"msg"`
	RetCode int    `json:"ret_code"`
}

var (
	uaMap = map[string]string{
		"phone":   "Mozilla/5.0 (Linux; Android 12; Pixel 6 Build/SD1A.210817.023; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/94.0.4606.71 Mobile Safari/537.36",
		"pad":     "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15 Chrome/117.0.5938.62",
		"desktop": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.5938.62 Safari/537.36",
	}
	ispMap = map[string]struct{}{
		"telecom": {},
		"cmcc":    {},
		"unicom":  {},
		"xyw":     {},
	}
	ua       string
	isp      string
	username string
	password string
)

func init() {
	flag.StringVar(&ua, "ua", "desktop", "auth ua, eg: -ua phone")
	flag.StringVar(&isp, "isp", "telecom", "auth isp, eg: -isp telecom")
	flag.StringVar(&username, "username", "", "auth login username, eg: -username 16xxxxx")
	flag.StringVar(&password, "password", "", "auth login password, eg: -password *******")
}

func main() {
	flag.Parse()
	if "" == username || "" == password {
		log.Fatalf("认证用户名和密码不能为空，username: %s, password: %s \n", username, password)
	}
	userAgent, ok := uaMap[ua]
	if !ok {
		log.Fatalf("认证请求UA不存在，UA: %s \n", ua)
	}
	_, ok = ispMap[isp]
	if !ok {
		log.Fatalf("认证类型不存在，type: %s \n", isp)
	}

	networkInterfaces, err := GetNetworkInterfaces()
	if err != nil {
		log.Fatalf("获取网络接口失败，err: %s \n", err.Error())
	}
	activeInterface := SelectActiveNetworkInterface(networkInterfaces)
	if activeInterface == nil {
		log.Fatalln("没有网络接口连接到校园网，请连接后重试！")
	}
	u, err := url.Parse(requestUrl)
	if err != nil {
		log.Fatalf("解析请求Url失败，err: %s \n", err.Error())
	}
	GenerateQueryParams(u, activeInterface)
	request, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	request.Header.Add("Host", host)
	request.Header.Add("Referer", referer)
	request.Header.Add("User-Agent", userAgent)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("请求服务器失败，err: %s \n", err.Error())
	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("解析请求响应体失败，err：%s \n", err.Error())
	}
	resultString := string(bytes)
	result := &AuthResult{}
	err = json.Unmarshal([]byte(resultString[7:len(resultString)-1]), &result)
	if err != nil {
		log.Printf("解析请求响应体失败，err：%s \n", err.Error())
		log.Fatalln(resultString)
	}
	if result.Result == "0" && result.RetCode == 2 {
		log.Println("当前设备已认证")
		return
	}
	log.Printf("response code: %s, message: %s ret_code: %d \n", result.Result, result.Msg, result.RetCode)
}

func GenerateQueryParams(u *url.URL, activeInterface *NetworkInterface) {
	query := u.Query()
	query.Add("c", c)
	query.Add("a", a)
	query.Add("callback", callback)
	query.Add("login_method", loginMethod)
	var device uint8 = 1
	if ua == "desktop" {
		device = 0
	}
	query.Add("user_account", fmt.Sprintf(",%d,%s@%s", device, username, isp))
	query.Add("user_password", password)
	query.Add("wlan_user_ip", activeInterface.Ipv4)
	query.Add("wlan_user_mac", strings.ReplaceAll(activeInterface.Mac, ":", ""))
	query.Add("jsVersion", jsVersion)
	u.RawQuery = query.Encode()
}

// SelectActiveNetworkInterface 获取校园网可用的网络接口
func SelectActiveNetworkInterface(networks []*NetworkInterface) *NetworkInterface {
	actives := make([]*NetworkInterface, 0)
	for _, item := range networks {
		if strings.HasPrefix(item.Ipv4, "10.") {
			actives = append(actives, item)
		}
	}
	length := len(actives)
	if length == 0 {
		return nil
	}
	if length == 1 {
		return actives[0]
	}
	fmt.Println("当前存在多个可能的网络IP，请选择连接校园网的IP和MAC地址：")
	for i, item := range actives {
		fmt.Printf("(%d) Name: %s, Ipv4: %s, Ipv6: %s, MAC: %s \n", i, item.Name, item.Ipv4, item.Ipv6, item.Mac)
	}
	fmt.Printf("请输入网络接口下标（default 0）: ")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		log.Println("读取输入错误，已自动使用`默认值")
		return actives[0]
	}
	index, err := strconv.Atoi(input)
	if err != nil || index < 0 || index > (length-1) {
		fmt.Printf("输入值格式错误，input: %s, 已自动使用默认值", input)
		return actives[0]
	}
	return actives[index]
}

// GetNetworkInterfaces 获取本机的所有网络接口
func GetNetworkInterfaces() ([]*NetworkInterface, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	length := len(netInterfaces)
	if length == 0 {
		return nil, fmt.Errorf("device not network interface")
	}
	result := make([]*NetworkInterface, 0)
	for _, netInterface := range netInterfaces {
		addrs, addrErr := netInterface.Addrs()
		if addrErr != nil {
			log.Printf("get network interface addrs error, msg: %s", addrErr.Error())
			continue
		}
		var ipv4, ipv6 string
		for i := 0; i < len(addrs); i++ {
			ipNet, ok := addrs[i].(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
			} else if ipNet.IP.To16() != nil {
				ipv6 = ipNet.IP.String()
			}
		}
		if ipv4 == "" {
			log.Printf("interface %s not active ipv4 address, break", netInterface.Name)
			continue
		}
		result = append(result, &NetworkInterface{
			Name: netInterface.Name,
			Ipv4: ipv4,
			Ipv6: ipv6,
			Mac:  netInterface.HardwareAddr.String(),
		})
	}
	return result, nil
}
