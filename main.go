package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	json2 "encoding/json"
	"fmt"
	"github.com/roseboy/httpcase/json"
	"github.com/roseboy/httpcase/requests"
	"github.com/roseboy/httpcase/util"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

//https://cloud.cn2030.com/sc/wx/HandlerSubscribe.ashx?act=auth&code=031wz10w3NbX1W2Prs2w3fkVaM1wz10c
//https://cloud.cn2030.com/sc/wx/HandlerSubscribe.ashx?act=auth&code=091vkDFa1PmkHA0WKWIa1BaIe54vkDFj

var (
	http    = requests.NewHttpSession()
	apiBase = "https://cloud.cn2030.com/sc/wx/HandlerSubscribe.ashx"
	config  Config
	args    Args
)

func main() {
	//读取配置
	read(&config, "config")
	read(&args, "_temp")

	fmt.Println("Waiting...")
	beginTime, _ := time.ParseInLocation("2006-01-02 15:04:05", config.BeginTime, time.Local)
	for time.Now().Before(beginTime) {
		time.Sleep(500 * time.Millisecond)
	}

	//查询地点
	for args.CustomerId == 0 {
		apiUrl := fmt.Sprintf("%s?act=CustomerList&city=%s&id=0&cityCode=%d&product=0",
			apiBase, url.PathEscape(config.City), config.CityCode)
		log.Println(apiUrl[len(apiBase)+5:])
		jsonObj, err := http.Get(apiUrl).Headers(header()).Timeout(2000).Send().ReadToJsonObject()
		if err != nil {
			log.Println("CustomerList error,retry...")
			continue
		}
		log.Printf("==>status:%v, msg:%v", jsonObj.Get("status"), jsonObj.Get("msg"))

		jsonObj.GetArray("list").ForEach(func(i int, object *json.Object) {
			if strings.Contains(object.Get("cname").(string), config.CustomerName) || config.CustomerName == "" {
				args.CustomerId = int(object.Get("id").(float64))
				return
			}
		})
	}
	log.Printf("CustomerId:%d\n", args.CustomerId)
	save(args, "_temp")

	//查询疫苗
	for args.CustomerProductId == 0 {
		apiUrl := fmt.Sprintf("%s?act=CustomerProduct&id=%d", apiBase, args.CustomerId)
		log.Println(apiUrl[len(apiBase)+5:])
		jsonObj, err := http.Get(apiUrl).Headers(header()).Timeout(2000).Send().ReadToJsonObject()
		if err != nil {
			log.Println("CustomerProduct error,retry...")
			continue
		}
		log.Printf("==>status:%v, msg:%v", jsonObj.Get("status"), jsonObj.Get("msg"))

		jsonObj.GetArray("list").ForEach(func(i int, object *json.Object) {
			if object.Get("text").(string) == config.CustomerProductName {
				args.CustomerProductId = int(object.Get("id").(float64))
				return
			}
		})
	}
	log.Printf("CustomerProductId:%d\n", args.CustomerProductId)
	save(args, "_temp")

	//获取验证码guid
	for args.Guid == "" || time.Now().Unix()-args.GuidTime.Unix() > 30 {
		args.Guid = GetCaptchaGuid()
		args.GuidTime = time.Now()
	}
	log.Printf("GuId:%s\n", args.Guid)
	save(args, "_temp")

	//return

	//查询可预约的日期
	for len(args.Dates) == 0 {
		apiUrl := fmt.Sprintf("%s?act=GetCustSubscribeDateAll&pid=%d&id=%d&month=%d",
			apiBase, args.CustomerProductId, args.CustomerId, config.Month)
		log.Println(apiUrl[len(apiBase)+5:])
		jsonObj, err := http.Get(apiUrl).Headers(header()).Timeout(2000).Send().ReadToJsonObject()
		if err != nil {
			log.Println("GetCustSubscribeDateAll error,retry...")
			continue
		}
		log.Printf("==>status:%v, msg:%v", jsonObj.Get("status"), jsonObj.Get("msg"))

		if jsonObj.GetArray("list").Length() == 0 {
			log.Println("预约未开始...")
			time.Sleep(500 * time.Millisecond)
			continue
		}

		jsonObj.GetArray("list").ForEach(func(i int, object *json.Object) {
			if object.Get("enable").(bool) {
				args.Dates = append(args.Dates, object.Get("date").(string))
			}
		})

		//无有效日期
		if len(args.Dates) == 0 {
			log.Println("已全部约满！")
			return
		}
	}
	log.Printf("Dates:%v\n", args.Dates)
	save(args, "_temp")

	//查询预约时间段
	dateIndex := 0
LabelGetMxId:
	for args.MxId == "" {
		args.Date = args.Dates[dateIndex%len(args.Dates)]
		apiUrl := fmt.Sprintf("%s?act=GetCustSubscribeDateDetail&pid=%d&id=%d&scdate=%s",
			apiBase, args.CustomerProductId, args.CustomerId, args.Date)
		log.Println(apiUrl[len(apiBase)+5:])
		jsonObj, err := http.Get(apiUrl).Headers(header()).Timeout(2000).Send().ReadToJsonObject()
		if err != nil {
			log.Println("GetCustSubscribeDateDetail error,retry...")
			continue
		}
		log.Printf("==>status:%v, msg:%v", jsonObj.Get("status"), jsonObj.Get("msg"))

		jsonObj.GetArray("list").ForEach(func(i int, object *json.Object) {
			if object.Get("qty").(float64) > 0 { //库存
				args.MxId = object.Get("mxid").(string)
				return
			}
		})

		if args.MxId == "" {
			log.Println("GetCustSubscribeDateDetail qty is 0,retry...")
			dateIndex++
			continue
		}
	}
	fmt.Printf("MxId:%v\n", args.MxId)
	save(args, "_temp")

LabelGetcaptcha:
	//识别验证码获取guid
	for args.Guid == "" {
		args.Guid = GetCaptchaGuid()
	}

	//提交预约
	OrderStatus := ""
	FailCount := 0
	for OrderStatus != "200" {
		apiUrl := fmt.Sprintf("%s?act=Save20&birthday=%s&tel=%s&sex=%d&cname=%s&doctype=1&idcard=%s&mxid=%s&date=%s&pid=7&Ftime=%d&guid=%s",
			apiBase, config.UserInfo.Birthday, config.UserInfo.Tel, config.UserInfo.Sex,
			url.QueryEscape(config.UserInfo.Cname), config.UserInfo.IdCard, args.MxId, args.Date, config.UserInfo.Ftime, args.Guid)
		log.Println(apiUrl[len(apiBase)+5:])
		jsonObj, err := http.Get(apiUrl).Headers(header()).Timeout(5000).Send().ReadToJsonObject()
		if err != nil {
			log.Println("Save20 error,retry...")
			time.Sleep(500 * time.Millisecond)
			continue
		}
		log.Printf("==> %s", jsonObj.ToString())

		OrderStatus = fmt.Sprintf("%v", jsonObj.Get("status"))
		if OrderStatus == "201" {
			log.Println(fmt.Sprintf("该时段预约已满,切换下个日期:%s", args.Dates[dateIndex%len(args.Dates)]))
			dateIndex++
			args.MxId = ""
			args.Guid = ""
			goto LabelGetMxId
		} else if OrderStatus != "200" {
			log.Println(fmt.Sprintf("Save20 error:%s,retry...", jsonObj.Get("msg")))
			FailCount++
			time.Sleep(1 * time.Second)
			args.Guid = ""
			goto LabelGetcaptcha
		}
	}

	time.Sleep(2 * time.Second)

	//预约状态
	apiUrl := fmt.Sprintf("%s?act=GetOrderStatus", apiBase)
	log.Printf(apiUrl[len(apiBase)+1:])
	jsonObj, err := http.Get(apiUrl).Headers(header()).Timeout(2000).Send().ReadToJsonObject()
	if err != nil {
		log.Println("GetOrderStatus error,retry...")
		time.Sleep(1 * time.Second)
		args.Guid = ""
		goto LabelGetcaptcha
	}
	log.Printf("==> %s", jsonObj.ToString())

	OrderStatus = fmt.Sprintf("%v", jsonObj.Get("status"))
	if OrderStatus == "300" { //该身份证或微信号已有预约信息.
		return
	} else if OrderStatus != "0" && OrderStatus != "200" {
		log.Println(fmt.Sprintf("GetOrderStatus:,retry..."))
		time.Sleep(1 * time.Second)
		args.Guid = ""
		goto LabelGetcaptcha
	}
	fmt.Println("Congratulations!!!")
}

func GetCaptchaGuid() string {
	//获取验证吗
	apiUrl := fmt.Sprintf("%s?act=GetCaptcha", apiBase)
	log.Println(apiUrl[len(apiBase)+5:])
	jsonObj, err := http.Get(apiUrl).Headers(header()).Send().ReadToJsonObject()
	if err != nil {
		log.Println("GetCaptcha error,retry...")
		return ""
	}
	log.Printf("==>status:%v, msg:%v", jsonObj.Get("status"), jsonObj.Get("msg"))

	if fmt.Sprintf("%v", jsonObj.Get("status")) != "0" {
		log.Println(fmt.Sprintf("CaptchaVerify GetCaptcha:%s,retry...", jsonObj.Get("msg")))
		time.Sleep(1 * time.Second)
		return ""
	}

	//识别验证码
	apiUrl = "http://127.0.0.1:8080/captcha"
	log.Printf(apiUrl)
	jsonObj, err = http.Post(apiUrl).Body(jsonObj.ToString()).Send().ReadToJsonObject()
	if err != nil {
		log.Println("IdentifyVerify error,retry...")
		return ""
	}
	x := jsonObj.Get("x")
	log.Printf("==> %s", jsonObj.ToString())

	//提交验证码
	apiUrl = fmt.Sprintf("%s?act=CaptchaVerify&token=&x=%v&y=%d", apiBase, x, 5)
	log.Println(apiUrl[len(apiBase)+5:])
	jsonObj, err = http.Get(apiUrl).Headers(header()).Timeout(2000).Send().ReadToJsonObject()
	if err != nil {
		log.Println("CaptchaVerify error,retry...")
		time.Sleep(1 * time.Second)
		return ""
	}
	log.Printf("==> %s", jsonObj.ToString())

	if fmt.Sprintf("%v", jsonObj.Get("status")) == "408" {
		log.Println("Cookie 失效...")
		os.Exit(0)
	} else if fmt.Sprintf("%v", jsonObj.Get("status")) != "200" {
		log.Println(fmt.Sprintf("CaptchaVerify error:(%v)%s,retry...", jsonObj.Get("status"), jsonObj.Get("msg")))
		time.Sleep(2 * time.Second)
		return ""
	}
	return jsonObj.Get("guid").(string)
}

//*************************************************************************

type Config struct {
	BeginTime string
	//cookie，抓包获取（必填）
	Cookie string
	//省市（必填）
	City string //`["广东省","清远市",""]`
	//该地区身份证号前6位（必填）
	CityCode int
	//医院名称关键字，为空取第一个
	CustomerName string
	//疫苗关键字（必填）
	CustomerProductName string
	//年月（必填）
	Month int
	//用户信息
	UserInfo UserInfo
}

type UserInfo struct {
	Birthday string //生日（必填）
	Tel      string //手机（必填）
	Sex      int    //性别（必填）1男 2女
	Cname    string //姓名（必填）
	Ftime    int    //针（必填）默认1针
	IdCard   string //身份证号（必填）
}

type Args struct {
	// 医院id
	CustomerId int
	//疫苗id
	CustomerProductId int
	//可预约日期
	Dates []string
	Date  string
	//预约id
	MxId     string
	Guid     string
	GuidTime time.Time
}

//*************************************************************************

func save(args interface{}, file string) {
	var str bytes.Buffer
	data, _ := json.Marshal(args)
	_ = json2.Indent(&str, data, "", "  ")
	_ = util.WriteText(str.String(), fmt.Sprintf("%s.json", file))
}

func read(args interface{}, file string) {
	txt, err := util.ReadText(fmt.Sprintf("%s.json", file))
	if err == nil {
		_ = json.Unmarshal([]byte(txt), &args)
	}
}

func header() map[string]string {
	headers := make(map[string]string)
	headers["Host"] = "cloud.cn2030.com"
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "*/*"
	headers["Connection"] = "keep-alive"
	headers["Cookie"] = config.Cookie
	headers["User-Agent"] = "Mozilla/5.0 (iPhone; CPU iPhone OS 12_1_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/16D57 MicroMessenger/7.0.3(0x17000321) NetType/WIFI Language/zh_CN"
	headers["Referer"] = "https://servicewechat.com/wx2c7f0f3c30d99445/72/page-frame.html"
	headers["zftsl"] = zftsl()
	headers["Accept-Language"] = "zh-cn"
	headers["Accept-Encoding"] = "gzip,deflate,br"
	return headers
}

func zftsl() string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("zfsw_%d", time.Now().Unix()/10)))
	return hex.EncodeToString(h.Sum(nil))
}
