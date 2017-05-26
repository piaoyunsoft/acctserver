package controllers

import (
	"errors"
	"fmt"
	"moton/acctserver/models"
	"moton/logger"
	"net/url"
	"sort"
	"strconv"

	"io/ioutil"

	"encoding/pem"

	"crypto/sha1"
	"crypto/x509"

	"crypto/rsa"

	"crypto"

	"encoding/base64"

	"io"

	"github.com/astaxie/beego"
)

type AlipayController struct {
	beego.Controller
}

func (c *AlipayController) rsaVerify(token string, base64Sign string, pubKey *rsa.PublicKey) error {
	hash := sha1.New()
	if _, err := io.WriteString(hash, token); err != nil {
		return err
	}

	decodeSign, err := base64.StdEncoding.DecodeString(base64Sign)
	if err != nil {
		return err
	}

	logger.I("base64解码后的sign：%s", string(decodeSign))

	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA1, hash.Sum(nil), decodeSign); err != nil {
		return err
	}
	return nil
}

func (c *AlipayController) verifySign(values url.Values) error {
	var str string
	var ampersand string
	var sortedKeys []string

	sign, ok := values["sign"]
	if !ok {
		return errors.New(fmt.Sprintf("支付参数中没有签名信息，参数=%v", values))
	}

	// tradeStatus, ok := values["trade_state"]

	for k := range values {
		if k != "sign" && k != "sign_type" {
			sortedKeys = append(sortedKeys, k)
		}
	}

	sort.Strings(sortedKeys)
	for _, k := range sortedKeys {
		str += ampersand + k + "=" + values[k][0]
		if len(ampersand) == 0 {
			ampersand = "&"
		}
	}

	logger.I("拼接后的字符串：%s", str)

	pubKeyPath := beego.AppConfig.String("alipay::alipaypublic")
	data, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return errors.New(fmt.Sprintf("无法打开公钥文件，path=%s err=%s", pubKeyPath, err.Error()))
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return errors.New("pem解码失败!")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return errors.New(fmt.Sprintf("生成公钥失败，err=%s", err.Error()))
	}

	if err = c.rsaVerify(str, sign[0], pubKey.(*rsa.PublicKey)); err != nil {
		return errors.New(fmt.Sprintf("rsa验签失败，err=%s\nvalues=%v", err.Error(), values))
	}
	return nil
}

func (c *AlipayController) checkRequestParams(values url.Values) error {
	// sellerEmail, ok := values["seller_email"]
	// if !ok {
	// 	return errors.New(fmt.Sprintf("支付参数中没有seller_email字段，values=%v", values))
	// }

	// if sellerEmail[0] != beego.AppConfig.String("alipay::selleremail") {
	// 	return errors.New(fmt.Sprintf("验证商户邮件失败，期望=%s，收到=%s", sellerEmail[0], beego.AppConfig.String("alipay::selleremail")))
	// }

	return c.verifySign(values)
}

func (c *AlipayController) PayResult() {
	resp := "failure"
	defer func() {
		c.Ctx.WriteString(resp)
	}()

	values := c.Input()
	err := c.checkRequestParams(values)
	if err != nil {
		logger.E("支付宝支付请求验证失败: %s", err.Error())
		return
	}

	tradeStatus, ok := values["trade_status"]
	if !ok {
		logger.E("支付宝支付请求参数中没有交易状态trade_staus参数，values=%v", values)
		return
	}

	if tradeStatus[0] == "WAIT_BUYER_PAY" || tradeStatus[0] == "TRADE_CLOSED" {
		resp = "success"
		return
	}

	if tradeStatus[0] != "TRADE_SUCCESS" && tradeStatus[0] != "TRADE_FINISH" {
		logger.E("支付宝支付请求参数中交易状态不正确，tradeStatus=%s", tradeStatus[0])
		return
	}

	totalFee, ok := values["total_fee"]
	if !ok {
		logger.E("支付宝支付请求参数中没有支付金额total_fee参数，values=%v", values)
		return
	}

	transId, ok := values["trade_no"]
	if !ok {
		logger.E("支付宝支付请求参数中没有支付平台订单trade_no参数，values=%v", values)
		return
	}

	outTradeNo, ok := values["out_trade_no"]
	if !ok {
		logger.E("支付宝支付请求参数中没有支付平台订单out_trade_no参数，values=%v", values)
		return
	}

	orderId, err := strconv.ParseInt(outTradeNo[0], 10, 64)
	if err != nil {
		logger.E("支付宝支付转换订单信息失败, orderId=%s", orderId)
		return
	}

	price, err := strconv.ParseFloat(totalFee[0], 64)
	if err != nil {
		logger.E("支付宝支付转换支付金额失败, totalFee=%s", totalFee[0])
		return
	}

	if !models.DeliveryProduct(orderId, price, transId[0], true) {
		logger.E("支付宝支付发送商品失败!\n[sdk_order_id=%s]\n%v", transId[0], values)
		return
	}

	resp = "success"
}
