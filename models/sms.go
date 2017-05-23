package models

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"moton/acctserver/common"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
)

type Response struct {
	StatusCode    string //请求状态码，取值000000（成功）
	SmsMessageSid string //短信唯一标识符
	DateCreated   string //短信的创建时间，格式：年-月-日 时:分:秒（如2013-02-01 15:38:09）
	StatusMsg     string //"【账号】请求包头Authorization参数解码后格式有误"
}

type RequestPacket struct {
	To         string   `json:"to"`         //短信接收端手机号码集合，用英文逗号分开，每批发送的手机号数量不得超过100个
	AppId      string   `json:"appId"`      //应用Id
	TemplateId string   `json:"templateId"` //模板Id
	Datas      []string `json:"datas"`      //内容数据外层节点
	//Data       string   //内容数据，用于替换模板中{序号}
}

func SendSMSCaptcha(phone, captcha, templateID string) {
	restUrl := beego.AppConfig.String("sendsms::url")

	accountsid := beego.AppConfig.String("sendsms::accountsid")
	authtoken := beego.AppConfig.String("sendsms::authtoken")
	appId := beego.AppConfig.String("sendsms::appid")

	t := time.Now()
	timeStamp := t.Format("20060102150405")
	sigParameter := common.GetMD5(accountsid + authtoken + timeStamp)
	authorization := base64.StdEncoding.EncodeToString([]byte(accountsid + ":" + timeStamp))
	url := restUrl + "/2013-12-26/Accounts/" + accountsid + "/SMS/TemplateSMS?sig=" + sigParameter

	var response Response
	requestPacket := RequestPacket{
		To:         phone,
		AppId:      appId,
		TemplateId: templateID,
		Datas:      []string{captcha, "30"},
	}

	requestbody, _ := json.Marshal(requestPacket)
	req := httplib.Post(url)
	req.SetTimeout(10*time.Second, 30*time.Second)
	req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	req.Header("Accept", "application/json")
	req.Header("Content-Type", "application/json;charset=utf-8")
	req.Header("Authorization", authorization)
	req.Body(requestbody)
	req.ToJSON(&response)
}
