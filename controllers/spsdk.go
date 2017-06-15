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

type SpsdkVerifyRequestParams struct {
	SessionId string `json:"session_id"`
	GameId    string `json:"game_id"`
}

type SpsdkVerifyResult struct {
	ErrorCode int    `json:"errcode"`
	ErrorMsg  string `json:"errmsg"`
	UserId    string `json:"user_id"`
}

type SpsdkPayResultRequestParams struct {
	OrderId   string `json:"order_id"`
	Username  string `json:"username"`
	ServerId  string `json:"server_id"`
	Server    string `json:"_server"`
	GameCoin  string `json:"game_coin"`
	Money     string `json:"money"`
	ExtraInfo string `json:"extra_info"`
	Time      string `json:"time"`
	Sign      string `json:"sign"`
}

type SpsdkPayResultResult struct {
	State int    `json:"state"`
	Msg   string `json:"msg"`
}

type SpsdkController struct {
	beego.Controller
}

func (c *SpsdkController) checkSign(params *SpsdkPayResultRequestParams) error {
	str := params.OrderId
	str += params.Username
	str += params.ServerId
	str += params.GameCoin
	str += params.ExtraInfo
	str += params.Time
	str += beego.AppConfig.String("spsdk::paykey")
	logger.I("MD5前：%s", str)
	if common.GetMD5(str) != params.Sign {
		return errors.New(fmt.Sprintf("签名对比失败，签名计算前：%s\n计算后：%s\n期望值：%s", str, common.GetMD5(str), params.Sign))
	}

	return nil
}

func (c *SpsdkController) Verify() {
	result := &SpsdkVerifyResult{
		ErrorCode: common.SYSTEM_ERROR,
	}
	c.Data["json"] = result
	defer func() {
		logger.I(string(c.Ctx.Input.RequestBody))
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	reqParams := &SpsdkVerifyRequestParams{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, reqParams)
	if err != nil {
		logger.E("思璞sdk登录验证请求解析参数失败:\nparams=%s\nerr=%s", c.Ctx.Input.RequestBody, err.Error())
		result.ErrorCode = common.DATA_ILLEGAL
		return
	}

	postResult, err := common.HttpPostJson(beego.AppConfig.String("spsdk::verifyurl"), []byte(reqParams.SessionId), 10*time.Second)
	if err != nil {
		logger.E("思璞sdk登录验证请求验证失败:\nparams=%v\nerr=%s", reqParams, err.Error())
		result.ErrorCode = common.CHANNEL_LOGIN_VERIFY_FAILED
		return
	}

	accountId := string(postResult)
	acct, errcode := models.ChannelLogin(beego.AppConfig.String("spsdk::channel"), accountId, c.Ctx.Request.RemoteAddr)
	if errcode != common.SUCCEED {
		logger.E("思璞sdk登录验证请求游戏登录失败:\naccountId=%s\nparams=%v\nerrcode=%s", accountId, reqParams, errcode)
		result.ErrorCode = errcode
		return
	}

	user := &models.User{
		AccountID: acct.AccountId,
		GUID:      acct.AccountGuid,
		Salt:      acct.Salt,
		Channel:   beego.AppConfig.String("spsdk::channel"),
		GameID:    "",
	}

	logger.I(accountId)
	c.SetSession("user", user)
	result.UserId = accountId
	result.ErrorCode = common.SUCCEED
}

func (c *SpsdkController) PayResult() {
	result := &SpsdkPayResultResult{
		State: 0,
		Msg:   "未知错误",
	}
	c.Data["json"] = result
	defer func() {
		c.ServeJSON()
	}()

	reqParams := &SpsdkPayResultRequestParams{}
	logger.I("%q\n", reqParams)
	err := c.ParseForm(reqParams)
	// err := json.Unmarshal(c.Ctx.Input.RequestBody, reqParams)
	if err != nil {
		logger.E("思璞sdk支付请求解析参数失败:\nerr=%s", err.Error())
		result.Msg = "解析请求参数失败"
		return
	}

	err = c.checkSign(reqParams)
	if err != nil {
		logger.E("思璞sdk支付请求验证签名失败:\nerr=%s", err.Error())
		result.Msg = "签名验证失败"
		return
	}

	orderId, err := strconv.ParseInt(reqParams.ExtraInfo, 10, 64)
	if err != nil {
		logger.E("思璞sdk支付转换订单信息失败:\nextra_info=%s\nerr=%s", reqParams.ExtraInfo, err.Error())
		result.Msg = "获取订单号失败"
		return
	}

	price, err := strconv.ParseFloat(reqParams.Money, 64)
	if err != nil {
		logger.E("思璞sdk支付转换支付金额失败:\nmoney=%s\nerr=%s", reqParams.Money, err.Error())
		result.Msg = "购买价格数据非法"
		return
	}

	if !models.DeliveryProduct(orderId, price, reqParams.OrderId, true) {
		logger.E("思璞sdk支付发送商品失败!\n[sdk_order_id=%s]\n%v", reqParams.OrderId, reqParams)
		result.Msg = "发送商品失败"
		return
	}

	result.State = 1
	result.Msg = "成功"
}
