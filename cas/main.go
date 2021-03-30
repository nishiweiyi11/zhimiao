package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/roseboy/httpcase/requests"
	"strings"
)

var http = requests.NewHttpSession()

func main() {
	var (
		url      = "https://ssodev.yuexiu.com/cas/login?service=https%3A%2F%2Fpbm-uat.yuexiu.com%2Fedj-home-frontend%2F%23%2FMain%2FHome"
		username = "peng.shaokai"
		password = "Password1"
		pubKey   = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCVloufcFiGVeT5N9a3njcLi3wDQ3iqIVD35AqPZuar2STQYSRSkMuwXIpJDu4XyLZmqWS0beRcCfGTpvPsCJ8DuGLn28Lj0w0u6/LtDV49D+r7+knYHXxT/y3v0FTeTIBo/HdBw/7EwJmZdvR1eyay0QKo63vCQkTrwzB4LSY9rwIDAQAB
-----END PUBLIC KEY-----`
	)

	pwdByte := RsaEncrypt([]byte(password), []byte(pubKey))
	password = base64.StdEncoding.EncodeToString(pwdByte)
	fmt.Println(password)

	res := http.Get(url).Send()
	fmt.Println(res.Response.Headers)
	html, _ := res.ReadToText()
	ltIndex := strings.Index(html, `name="lt" value="`) + 17
	lt := html[ltIndex : ltIndex+44]
	executionIndex := strings.Index(html, `name="execution" value="`) + 24
	execution := html[executionIndex : executionIndex+4]

	params := make(map[string]string)
	params["lt"] = lt
	params["_eventId"] = "submit"
	params["execution"] = execution
	params["username"] = username
	params["password"] = "ENCODEDPASSWORD" + password
	fmt.Println(params)
	res = http.Post(url).Headers(header()).AllowRedirect(false).Params(params).Send()
	html, _ = res.ReadToText()

	fmt.Println(res.Response.Status)
	fmt.Println(res.Response.Headers)
}

// 公钥加密
func RsaEncrypt(data, keyBytes []byte) []byte {
	//解密pem格式的公钥
	block, _ := pem.Decode(keyBytes)

	// 解析公钥
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)
	//加密
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		panic(err)
	}
	return ciphertext
}

func header() map[string]string {
	h := make(map[string]string)
	h[`Accept`] = `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`
	h[`Accept-Encoding`] = `gzip, deflate, br`
	h[`Accept-Language`] = `zh-CN,zh;q=0.9`
	h[`Cache-Control`] = `no-cache`
	h[`Connection`] = `keep-alive`
	h[`Content-Type`] = `application/x-www-form-urlencoded`
	h[`Host`] = `ssodev.yuexiu.com`
	h[`Origin`] = `https://ssodev.yuexiu.com`
	h[`Pragma`] = `no-cache`
	h[`Referer`] = `https://ssodev.yuexiu.com/cas/login?service=https%3A%2F%2Fpbm-uat.yuexiu.com%2Fedj-home-frontend%2F%23%2FMain%2FHome`
	h[`sec-ch-ua`] = `"Google Chrome";v="89", "Chromium";v="89", ";Not A Brand";v="99"`
	h[`sec-ch-ua-mobile`] = `?0`
	h[`Sec-Fetch-Dest`] = `document`
	h[`Sec-Fetch-Mode`] = `navigate`
	h[`Sec-Fetch-Site`] = `same-origin`
	h[`Sec-Fetch-User`] = `?1`
	h[`Upgrade-Insecure-Requests`] = `1`
	h[`User-Agent`] = `Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36`
	return h
}
