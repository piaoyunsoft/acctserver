package controllers

import (
	"moton/logger"

	"net/url"

	"github.com/astaxie/beego"
)

type ThirdPartyController struct {
	beego.Controller
}

func (c *ThirdPartyController) createAlipayLink(values url.Values) string {
	var link string
	return link
}

func (c *ThirdPartyController) createWeixinLink(values url.Values) string {
	var link string
	return link
}

func (c *ThirdPartyController) Create3rdPartyPayLink() {
	resp := ""
	defer func() {
		c.Ctx.WriteString(resp)
	}()

	values := c.Input()
	payType, ok := values["payType"]
	if !ok {
		logger.E("向第三方请求支付失败，缺少payType参数，values=%v", values)
		return
	}

	if payType[0] == "ali" {
		resp = c.createAlipayLink(values)
	} else if payType[0] == "wx" {
		resp = c.createWeixinLink(values)
	} else {
		logger.E("向第三方请求支付失败，payType参数值无效，payType=%s", payType[0])
		return
	}
}
