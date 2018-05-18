package main

import (
	"bytes"
	"crypto/md5"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	TestAPI = "https://apitest.mch.weixin.qq.com/sandboxnew/pay/getsignkey"
)

type PostData struct {
	XMLName  xml.Name `xml:"xml"`
	MchID    string   `xml:"mch_id"`
	NonceStr string   `xml:"nonce_str"`
	Sign     string   `xml:"sign"`
}

func main() {
	var (
		mchID     string
		appSecret string
	)

	flag.StringVar(&mchID, "m", "", "微信商户号")
	flag.StringVar(&mchID, "--mchId", "", "微信商户号")
	flag.StringVar(&appSecret, "k", "", "appSecret")
	flag.StringVar(&appSecret, "--appSecret", "", "appSecret")

	var usageStr = `

Usage: 微信根证书验证工具 [options]

Server Options:
    -m, --mchId <mchId>          微信商户号
	-k, --appSecret <appSecret>  加密密钥
`

	flag.Usage = func() {
		fmt.Printf("%s\n", usageStr)
		os.Exit(0)
	}

	if !flag.Parsed() {
		flag.Parse()
	}

	if mchID == "" {
		log.Fatalln("微信商户号不能为空")
	}

	if appSecret == "" {
		log.Fatalln("密钥不能为空")
	}

	nonceStr := getNonceStr(8)

	data := make(map[string]interface{})
	data["mch_id"] = mchID
	data["nonce_str"] = nonceStr

	pd := PostData{
		MchID:    mchID,
		NonceStr: nonceStr,
		Sign:     makeSign(data, appSecret),
	}

	xmlBody, err := xml.Marshal(pd)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(TestAPI, "application/xml", bytes.NewReader(xmlBody))
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	rb := string(respBody)

	if strings.Contains(rb, "SUCCESS") {
		log.Println("支持")
	} else {
		log.Println("不支持")
		log.Println(rb)
	}
}

// 产生随机字符串
func getNonceStr(n int) string {
	chars := []byte("abcdefghijklmnopqrstuvwxyz0123456789")
	value := []byte{}
	m := len(chars)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < n; i++ {
		value = append(value, chars[r.Intn(m)])
	}

	return string(value)
}

// 生成sign
func makeSign(params map[string]interface{}, key string) string {
	var keys []string
	var sorted []string

	for k, v := range params {
		if k != "sign" && v != "" {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)
	for _, k := range keys {
		sorted = append(sorted, fmt.Sprintf("%s=%v", k, params[k]))
	}

	s := strings.Join(sorted, "&")
	s += "&key=" + key

	return fmt.Sprintf("%X", md5.Sum([]byte(s)))
}
