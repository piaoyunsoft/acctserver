package controllers

import (
	"encoding/json"
	"errors"
	"moton/acctserver/common"
	"moton/acctserver/models"
	"moton/logger"
	"strconv"

	"time"

	"fmt"

	"github.com/astaxie/beego"
)

/*
	        $appleJSON = $postData['transactionReceipt'];
            $sdkOrderId = $postData['orderId'];

            $orderId = $postData['gameOrder']['orderId'];
            $price = $postData['gameOrder']['price'];
            $goodsId = $postData['gameOrder']['goodsId'];
            $this->product_id = $goodsId;
            $userId = $postData['gameOrder']['userId'];
            $sdkId = $postData['gameOrder']['sdkId'];
            $channelId = $postData['gameOrder']['channelId'];
            $accountId = $postData['gameOrder']['accountId'];
            $serverId = $postData['gameOrder']['serverId'];

            $gameOrder = $postData['gameOrder'];
*/

type IAPVerifyReceiptGameOrderParams struct {
	OrderId   string `json:"orderId"`
	Price     string `json:"price"`
	GoodsId   string `json:"goodsId"`
	UserId    string `json:"userId"`
	SdkId     string `json:"sdkId"`
	ChannelId string `json:"channelId"`
	AccountId string `json:"accountId"`
	ServerId  string `json:"serverId"`
}

type IAPVerifyReceiptParams struct {
	TransactionReceipt string                          `json:"transactionReceipt"`
	SdkOrderId         string                          `json:"orderId"`
	GameOrder          IAPVerifyReceiptGameOrderParams `json:"gameOrder"`
}

type IAPVerifyReceiptData struct {
	ReceiptData string `json:"receipt-data"`
}

/*
"original_purchase_date_pst":"2017-04-09 05:02:22 America\/Los_Angeles",
"unique_identifier":"6a213a90681f1c74af375126e60580855e1f31d1",i
"original_transaction_id":"60000314575006",
"bvrs":"1.4.5",
"app_item_id":"1135840912",
"transaction_id":"60000314575006",
"quantity":"1",
"unique_vendor_identifier":"904EEDF1-22D7-4724-AFB7-2DEE1B376942",
"product_id":"zswz006",
"item_id":"1141211768",
"version_external_identifier":"821190053",
"bid":"com.qjdl.arbic",
"purchase_date_ms":"1491739342415",
"purchase_date":"2017-04-09 12:02:22 Etc\/GMT",
"purchase_date_pst":"2017-04-09 05:02:22 America\/Los_Angeles",
"original_purchase_date":"2017-04-09 12:02:22 Etc\/GMT",
"original_purchase_date_ms":"1491739342415"
*/
type IAPReceipt struct {
	OriginalPurchaseDatePst   string `json:"original_purchase_date_pst"`
	UniqueIdentifier          string `json:"unique_identifier"`
	OriginalTransactionId     string `json:"original_transaction_id"`
	Bvrs                      string `json:"bvrs"`
	AppItemId                 string `json:"app_item_id"`
	TransactionId             string `json:"transaction_id"`
	Quantity                  string `json:"quantity"`
	UniqueVendorIdentifier    string `json:"unique_vendor_identifier"`
	ProductId                 string `json:"product_id"`
	ItemId                    string `json:"item_id"`
	VersionExternalIdentifier string `json:"version_external_identifier"`
	BundleId                  string `json:"bid"`
	PurchaseDateMs            string `json:"purchase_date_ms"`
	PurchaseDate              string `json:"purchase_date"`
	PurchaseDatePst           string `json:"purchase_date_pst"`
	OriginalPurchaseDate      string `json:"original_purchase_date"`
	OriginalPurchaseDateMs    string `json:"original_purchase_date_ms"`
}

type IAPVerifyReceiptResult struct {
	Status  int        `json:"status"`
	Receipt IAPReceipt `json:"receipt"`
}

type IAPVerifyReceiptResponse struct {
	Status  int    `json:"status"`
	OrderId string `json:"orderId"`
}

type AppleController struct {
	beego.Controller
}

func (c *AppleController) doIAPVerifyReceipt(url string, receiptData *IAPVerifyReceiptData, usesandbox bool) error {
	// 验证IAP
	jsonReceiptResult, err := common.HttpPostJson(url, receiptData, 10*time.Second)
	if err != nil {
		return errors.New(fmt.Sprintf("请求IAP验证失败\nerror:%s\nurl:%s\nreceiptData:%s", err.Error(), url, receiptData))
	}

	// 解析验证结果
	receiptResult := &IAPVerifyReceiptResult{}
	err = json.Unmarshal(jsonReceiptResult, receiptResult)
	if err != nil {
		return errors.New(fmt.Sprintf("解析IAP返回结果失败\n%s\n%s", err.Error(), jsonReceiptResult))
	}

	// 验证返回结果的状态
	if receiptResult.Status != 0 {
		if receiptResult.Status == 21007 && usesandbox {
			return c.doIAPVerifyReceipt(url, receiptData, !usesandbox)
		}
		return errors.New(fmt.Sprintf("验证失败\nStatus:%d", receiptResult.Status))
	}

	return nil
}

func (c *AppleController) IAPVerifyReceipt() {
	resp := &IAPVerifyReceiptResponse{
		Status: 1,
	}
	c.Data["json"] = &resp
	requestBody := c.Ctx.Input.RequestBody
	defer func() {
		c.ServeJSON()
		// logger.I("IAPVerifyReceipt: %s", string(requestBody))
	}()

	// 解析请求参数
	params := &IAPVerifyReceiptParams{}
	// err := c.ParseForm(params)
	err := json.Unmarshal(requestBody, params)
	if err != nil {
		logger.E("解析IAP请求验证json失败\n%s\n%s", err.Error(), string(requestBody))
		return
	}
	resp.OrderId = params.SdkOrderId
	// fmt.Printf("params:%v\n", params)

	// 通过配置获得沙盒或者正式地址
	var url = ""
	usesandbox := beego.AppConfig.DefaultBool("apple::usesandbox", false)
	if usesandbox {
		url = beego.AppConfig.String("apple::sandbox")
	} else {
		url = beego.AppConfig.String("apple::verifyreceipt")
	}

	// 尝试验证回执
	receiptData := &IAPVerifyReceiptData{
		ReceiptData: params.TransactionReceipt,
	}

	err = c.doIAPVerifyReceipt(url, receiptData, usesandbox)
	if err != nil {
		logger.E(err.Error())
		return
	}

	// 获取订单id
	strOrderID := params.GameOrder.OrderId
	logger.I("IAP验证成功，订单号:%s", strOrderID)
	orderID, err := strconv.ParseInt(strOrderID, 10, 64)
	if err != nil {
		logger.E("无效订单\n%v", params.GameOrder)
		return
	}

	// 获取价格
	price, err := strconv.ParseFloat(params.GameOrder.Price, 64)
	if err != nil {
		logger.E("无效的价格！\n%v", params.GameOrder)
		return
	}

	// 发货
	if !models.DeliveryProduct(orderID, price, params.SdkOrderId, true) {
		logger.E("发送商品失败!\n[sdk_order_id=%s]\n%v", params.SdkOrderId, params.GameOrder)
		return
	}

	resp.Status = 0
}
