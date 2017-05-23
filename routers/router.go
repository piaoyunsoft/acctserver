package routers

import (
	"moton/acctserver/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router(beego.AppConfig.String("router::login"), &controllers.AccountController{}, "post:Login")
	beego.Router(beego.AppConfig.String("router::channellogin"), &controllers.AccountController{}, "post:ChannelLogin")
	beego.Router(beego.AppConfig.String("router::register"), &controllers.AccountController{}, "post:Register")
	beego.Router(beego.AppConfig.String("router::phoneregister"), &controllers.AccountController{}, "post:PhoneRegister")
	beego.Router(beego.AppConfig.String("router::tourists"), &controllers.AccountController{}, "post:Tourists")
	beego.Router(beego.AppConfig.String("router::smscaptcha"), &controllers.AccountController{}, "post:SMSCaptcha")
	beego.Router(beego.AppConfig.String("router::touristsbind"), &controllers.AccountController{}, "post:TouristsBind")

	beego.Router(beego.AppConfig.String("router::serverlist"), &controllers.GameController{}, "post:ServerList")
	beego.Router(beego.AppConfig.String("router::entergame"), &controllers.GameController{}, "post:EnterGame")
	beego.Router(beego.AppConfig.String("router::recharge"), &controllers.GameController{}, "post:Recharge")

	beego.Router(beego.AppConfig.String("router::patchlist"), &controllers.PatchController{}, "post:PatchList")

	beego.Router(beego.AppConfig.String("router::sendfeedback"), &controllers.FeedbackController{}, "post:SendFeedback")
	beego.Router(beego.AppConfig.String("router::getfeedback"), &controllers.FeedbackController{}, "post:GetFeedback")

	beego.Router(beego.AppConfig.String("router::cdkey"), &controllers.CDKeyController{}, "post:GetCDKey")

	beego.Router(beego.AppConfig.String("router::productlist"), &controllers.MallController{}, "post:ProductList")
	beego.Router(beego.AppConfig.String("router::order"), &controllers.MallController{}, "post:Order")
	beego.Router(beego.AppConfig.String("router::cancelorder"), &controllers.MallController{}, "post:CancelOrder")

	release := beego.AppConfig.String("common::release")
	if release != "true" {
		beego.Router(beego.AppConfig.String("router::buyproduct"), &controllers.MallController{}, "post:BuyProduct")
	}

	beego.Router(beego.AppConfig.String("router::sclogin"), &controllers.AccountController{}, "post:StarscloudLogin")
	beego.Router(beego.AppConfig.String("router::scbuyproduct"), &controllers.MallController{}, "post:StarscloudBuyProduct")

	beego.Router(beego.AppConfig.String("router::anysdkveriy"), &controllers.AccountController{}, "post:AnySdkVerify")
	beego.Router(beego.AppConfig.String("router::anysdklogin"), &controllers.AccountController{}, "post:AnySdkLogin")
	beego.Router(beego.AppConfig.String("router::anysdkpayresult"), &controllers.MallController{}, "post:AnySdkPayResult")

	beego.Router(beego.AppConfig.String("router::beecloudpayresult"), &controllers.BeecloudController{}, "post:PayResult")

	beego.Router(beego.AppConfig.String("router::alipaypayresult"), &controllers.AlipayController{}, "post:PayResult")

	beego.Router(beego.AppConfig.String("router::weixinpayreault"), &controllers.WeixinController{}, "post:PayResult")

	beego.Router(beego.AppConfig.String("router::appleverifyreceipt"), &controllers.AppleController{}, "post:IAPVerifyReceipt")
}
