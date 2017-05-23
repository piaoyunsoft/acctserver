package controllers

import (
	"encoding/json"
	"moton/acctserver/common"
	"moton/acctserver/models"
	"strconv"

	"moton/logger"

	"github.com/astaxie/beego"
)

type beecloudPayresult struct {
	Sign            string                 `json:"sign"`             //服务器端通过计算 App ID + App Secret + timestamp 的MD5生成的签名(32字符十六进制),请在接受数据时自行按照此方式验证sign的正确性，不正确不返回success即可
	Timestamp       int64                  `json:"timestamp"`        //服务端的时间（毫秒），用以验证sign, MD5计算请参考sign的解释
	ChannelType     string                 `json:"channel_type"`     //WX/ALI/UN/KUAIQIAN/JD/BD/YEE/PAYPAL 分别代表微信/支付宝/银联/快钱/京东/百度/易宝/PAYPAL
	SubChannelType  string                 `json:"sub_channel_type"` //代表以上各个渠道的子渠道，参看字段说明
	TransactionType string                 `json:"transaction_type"` //PAY/REFUND 分别代表支付和退款的结果确认
	TransactionID   string                 `json:"transaction_id"`   //交易单号，对应支付请求的bill_no或者退款请求的refund_no,对于秒支付button为传入的out_trade_no
	TransactionFee  int                    `json:"transaction_fee"`  //交易金额，是以分为单位的整数，对应支付请求的total_fee或者退款请求的refund_fee
	TradeSuccess    bool                   `json:"trade_success"`    //交易是否成功，目前收到的消息都是交易成功的消息
	MessageDetail   map[string]interface{} `json:"message_detail"`   //{orderId:xxx…..} 从支付渠道方获得的详细结果信息，例如支付的订单号，金额， 商品信息等，详见附录
	Optional        map[string]interface{} `json:"optional"`         //附加参数，为一个JSON格式的Map，客户在发起购买或者退款操作时添加的附加信息
}

type BeecloudController struct {
	beego.Controller
}

func (c *BeecloudController) PayResult() {
	returnStr := "fail"
	defer func() {
		logger.I(string(c.Ctx.Input.RequestBody))
		c.Ctx.WriteString(returnStr)
	}()

	requestBody := c.Ctx.Input.RequestBody
	payResult := &beecloudPayresult{}
	err := json.Unmarshal(requestBody, payResult)
	if err != nil {
		logger.E("解析json失败\n%s\n%s", err.Error(), string(requestBody))
		return
	}

	appID := beego.AppConfig.String("beecloud::appid")
	appSecret := beego.AppConfig.String("beecloud::appsecret")
	timeStamp := strconv.FormatInt(payResult.Timestamp, 10)
	sign := common.GetMD5(appID + appSecret + timeStamp)
	if sign != payResult.Sign {
		logger.E("签名验证失败\n%v", payResult)
		return
	}

	returnStr = "success"
	param, ok := payResult.Optional["orderId"]
	if !ok {
		logger.E("无效订单\n%v", payResult)
		return
	}

	strOrderID := param.(string)
	orderID, err := strconv.ParseInt(strOrderID, 10, 64)
	if err != nil {
		logger.E("无效订单\n%v", payResult)
		return
	}

	price := float64(payResult.TransactionFee) / 100.0

	if !models.DeliveryProduct(orderID, price, payResult.TransactionID, true) {
		logger.E("发送商品失败!\n[order_id=%d]\n%v", orderID, payResult)
		return
	}
}
