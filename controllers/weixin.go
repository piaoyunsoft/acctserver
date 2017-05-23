package controllers

import (
	"encoding/xml"
	"errors"
	"moton/acctserver/models"
	"moton/gjrobot/common"
	"moton/logger"
	"sort"

	"net/url"

	"fmt"

	"strconv"

	"strings"

	"github.com/astaxie/beego"
)

/*
参数文档
https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3

应用ID	appid	是	String(32)	wx8888888888888888	微信开放平台审核通过的应用APPID
商户号	mch_id	是	String(32)	1900000109	微信支付分配的商户号
设备号	device_info	否	String(32)	013467007045764	微信支付分配的终端设备号，
随机字符串	nonce_str	是	String(32)	5K8264ILTKCH16CQ2502SI8ZNMTM67VS	随机字符串，不长于32位
签名	sign	是	String(32)	C380BEC2BFD727A4B6845133519F3AD6	签名，详见签名算法
业务结果	result_code	是	String(16)	SUCCESS	SUCCESS/FAIL
错误代码	err_code	否	String(32)	SYSTEMERROR	错误返回的信息描述
错误代码描述	err_code_des	否	String(128)	系统错误	错误返回的信息描述
用户标识	openid	是	String(128)	wxd930ea5d5a258f4f	用户在商户appid下的唯一标识
是否关注公众账号	is_subscribe	否	String(1)	Y	用户是否关注公众账号，Y-关注，N-未关注，仅在公众账号类型支付有效
交易类型	trade_type	是	String(16)	APP	APP
付款银行	bank_type	是	String(16)	CMC	银行类型，采用字符串类型的银行标识，银行类型见银行列表
总金额	total_fee	是	Int	100	订单总金额，单位为分
货币种类	fee_type	否	String(8)	CNY	货币类型，符合ISO4217标准的三位字母代码，默认人民币：CNY，其他值列表详见货币类型
现金支付金额	cash_fee	是	Int	100	现金支付金额订单现金支付金额，详见支付金额
现金支付货币类型	cash_fee_type	否	String(16)	CNY	货币类型，符合ISO4217标准的三位字母代码，默认人民币：CNY，其他值列表详见货币类型
代金券金额	coupon_fee	否	Int	10	代金券或立减优惠金额<=订单总金额，订单总金额-代金券或立减优惠金额=现金支付金额，详见支付金额
代金券使用数量	coupon_count	否	Int	1	代金券或立减优惠使用数量
代金券ID	coupon_id_$n	否	String(20)	10000	代金券或立减优惠ID,$n为下标，从0开始编号
单个代金券支付金额	coupon_fee_$n	否	Int	100	单个代金券或立减优惠支付金额,$n为下标，从0开始编号
微信支付订单号	transaction_id	是	String(32)	1217752501201407033233368018	微信支付订单号
商户订单号	out_trade_no	是	String(32)	1212321211201407033568112322	商户系统内部订单号，要求32个字符内，只能是数字、大小写字母_-|*@ ，且在同一个商户号下唯一。
商家数据包	attach	否	String(128)	123456	商家数据包，原样返回
支付完成时间	time_end	是	String(14)	20141030133525	支付完成时间，格式为yyyyMMddHHmmss，如2009年12月25日9点10分10秒表示为20091225091010。其他详见时间规则
*/

type WeixinPayResultRequest struct {
	XMLName       xml.Name `xml:"xml"`
	AppId         string   `xml:"appid"`
	MchId         string   `xml:"mch_id"`
	NonceStr      string   `xml:"nonce_str"`
	Sign          string   `xml:"sign"`
	ResultCode    string   `xml:"result_code"`
	ErrCode       string   `xml:"err_code"`
	ErrCodeDes    string   `xml:"err_code_des"`
	OpenId        string   `xml:"openid"`
	TradeType     string   `xml:"trade_type"`
	BankType      string   `xml:"bank_type"`
	TotalFee      int      `xml:"total_fee"`
	FeeType       string   `xml:"fee_type"`
	CashFee       int      `xml:"cash_fee"`
	TransactionId string   `xml:"transaction_id"`
	OutiTradeNo   string   `xml:"out_trade_no"`
	Attach        string   `xml:"attach"`
	TimeEnd       string   `xml:"time_end"`
}

type WeixinPayResultResponse struct {
	XMLName    xml.Name  `xml:"xml"`
	ReturnCode CDATAText `xml:"return_code"`
}

type CDATAText struct {
	Text string `xml:",cdata"`
}

type WeixinPayResultAttach struct {
	OrderId int64 `json:"orderId"`
}

type WeixinController struct {
	beego.Controller
}

func (c *WeixinController) makeSign(values url.Values) string {
	var sign string
	var ampersand string
	var sortedKeys []string

	for k := range values {
		if k != "sign" {
			sortedKeys = append(sortedKeys, k)
		}
	}

	sort.Strings(sortedKeys)
	for _, k := range sortedKeys {
		sign += k + "=" + values[k][0] + ampersand
	}

	sign += "key=" + beego.AppConfig.String("weixin::apikey")

	return common.GetMD5(sign)
}

func (c *WeixinController) checkPayResult(values url.Values) error {
	appid, ok := values["appid"]
	if !ok {
		return errors.New(fmt.Sprintf("支付参数中没有appid信息，参数=%v", values))
	}

	if appid[0] != beego.AppConfig.String("weixin::appid") {
		return errors.New(fmt.Sprintf("appid不匹配，期望=%s，收到=%s", beego.AppConfig.String("weixin::appid"), appid[0]))
	}

	sign, ok := values["sign"]
	if !ok {
		return errors.New(fmt.Sprintf("支付参数中没有签名信息，参数=%v", values))
	}

	selfSign := c.makeSign(values)
	if selfSign != sign[0] {
		return errors.New(fmt.Sprintf("签名信息不匹配，期望=%s，收到=%s", selfSign, sign[0]))
	}

	return nil
}

func (c *WeixinController) PayResult() {
	resp := &WeixinPayResultResponse{
		ReturnCode: CDATAText{"FAIL"},
	}

	c.Data["xml"] = &resp
	defer func() {
		c.ServeXML()
	}()

	logger.I("body: %s\n", string(c.Ctx.Input.RequestBody))
	values := c.Input()
	err := c.checkPayResult(values)
	if err != nil {
		logger.E("微信支付请求验证失败: %s", err.Error())
		return
	}

	attach, ok := values["attach"]
	if !ok {
		logger.E("微信支付请求参数中没有附加的attach信息, values=%v", values)
		return
	}

	totalFee, ok := values["total_fee"]
	if !ok {
		logger.E("微信支付请求参数中没有支付金额参数，values=%v", values)
		return
	}

	transId, ok := values["transaction_id"]
	if !ok {
		logger.E("微信支付请求参数中没有支付平台订单id参数，values=%v", values)
		return
	}

	attachValues := strings.Split(attach[0], ",")
	if len(attachValues) == 0 || len(attachValues[0]) == 0 {
		logger.E("微信支付附加信息格式错误, attach=%s", attach[0])
		return
	}

	orderId, err := strconv.ParseInt(attachValues[0], 10, 64)
	if err != nil {
		logger.E("微信支付转换订单信息失败, orderId=%s", attachValues[0])
		return
	}

	price, err := strconv.ParseFloat(totalFee[0], 64)
	if err != nil {
		logger.E("微信支付转换支付金额失败, totalFee=%s", totalFee[0])
		return
	}

	price /= 100.0

	if !models.DeliveryProduct(orderId, price, transId[0], true) {
		logger.E("微信支付发送商品失败!\n[sdk_order_id=%s]\n%v", transId[0], values)
		return
	}

	resp.ReturnCode.Text = "SUCCESS"
}
