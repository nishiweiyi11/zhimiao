package main

import (
	"fmt"
	"github.com/roseboy/httpcase/requests"
)

const auth = `Bearer eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiI3NzU1MTA1IiwiaWF0IjoxNjE2NDgzNTE3LCJzdWIiOiJ7XCJzZXNzaW9uSWRcIjpcIjc3NTUxMDVcIixcInVzZXJJZFwiOjc3NTUxMDUsXCJjb2RlXCI6XCIyMDIxMDIwMTAyOTA5NzEwNDAwXCIsXCJjZWxscGhvbmVcIjpcIjE2NioqKio5ODAwXCIsXCJpZGVudGl0eUF1dGhlbnRpY2F0aW9uU3RhdHVzXCI6MSxcInJlZ2lzdGVyQ2hhbm5lbFwiOlwiMDJcIixcInBvc3BhbFVzZXJJZFwiOjEyMDE4NjYxMTQ0Mzc2MzEzMSxcIm5hbWVcIjpcIi5LXCIsXCJhY2NvdW50VHlwZVwiOlwiQ1VTVE9NRVJfVVNFUlwiLFwiaGFzaENwXCI6XCIyYmtQQTc5NVZCdjNnMUpNXCJ9IiwiZXhwIjoxNjE5MDc1NTE3fQ._CEMTuhL-TZGghmg5ANUZWgXzPrAL3I1ZqsYCL8E5m4`

func main() {
	res, _ := requests.Post("https://app.relxtech.com/api/crm/coupon/draw").
		Body(`{"couponId":380,"activityId":64}`).Headers(headers()).Send().ReadToText()
	fmt.Println(res)

	res, _ = requests.Post("https://app.relxtech.com/api/crm/coupon/draw").
		Body(`{"couponId":381,"activityId":64}`).Headers(headers()).Send().ReadToText()
	fmt.Println(res)

	res, _ = requests.Post("https://app.relxtech.com/api/crm/coupon/draw").
		Body(`{"couponId":382,"activityId":64}`).Headers(headers()).Send().ReadToText()
	fmt.Println(res)

	res, _ = requests.Post("https://app.relxtech.com/api/crm/coupon/draw").
		Body(`{"couponId":383,"activityId":64}`).Headers(headers()).Send().ReadToText()
	fmt.Println(res)

}

func headers() map[string]string {
	h := make(map[string]string)
	h["Host"] = "app.relxtech.com"
	h["Origin"] = "https://app.relxtech.com"
	h["Referer"] = "https://app.relxtech.com/mcrm/activity/64"
	h["Accept-Language"] = "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7"
	h["X-Requested-With"] = "com.tencent.mm"
	h["XAccept"] = "application/json"
	h["Content-Type"] = "application/json;charset=UTF-8"
	h["User-Agent"] = "Mozilla/5.0 (Linux; Android 10; DT1901A Build/QKQ1.191222.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/77.0.3865.120 MQQBrowser/6.2 TBS/045521 Mobile Safari/537.36 MMWEBID/6454 MicroMessenger/8.0.1.1841(0x2800015D) Process/tools WeChat/arm64 Weixin NetType/WIFI Language/zh_CN ABI/arm64"
	h["Authorization"] = auth
	return h
}
