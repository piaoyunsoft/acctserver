package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"moton/acctserver/common"
	"moton/acctserver/models"
	"moton/logger"

	"fmt"

	"github.com/astaxie/beego"
)

type AccountController struct {
	beego.Controller
}

type accountArgs struct {
	UserName   string `json:"username"`
	Password   string `json:"password,omitempty"`
	Channel    string `json:"channel"`
	GameID     string `json:"game_id"`
	ApiKey     string `json:"api_key"`
	Phone      string `json:"phone"`
	SMSCaptcha string `json:"smscaptcha"`
	Token      string `json:"token"`
	Timestamp  string `json:"timestamp"`
	Sign       string `json:"sign"`
}

type accountOpcode struct {
	Opcode string      `json:"opcode"`
	Args   accountArgs `json:"arg"`
}

type accountResult struct {
	Opcode    string `json:"opcode"`
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	UserName  string `json:"username,omitempty"`
}

type anysdkLoginCommon struct {
	Channel  string `json:"channel"`
	UserSdk  string `json:"user_sdk"`
	Uid      string `json:"uid"`
	ServerID string `json:"server_id"`
	PluginID string `json:"plugin_id"`
}

type anysdkLoginResult struct {
	Status string            `json:"status"`
	Common anysdkLoginCommon `json:"common"`
	SN     string            `json:"sn"`
	Ext    string            `json:"ext"`
}

type anysdkVerifyData struct {
	Captcha string `json:"captchar"`
	GUID    string `json:"guid"`
}

const (
	anysdkLoginTimeout = 60
)

func (c *AccountController) decodeAccountArgs() *accountOpcode {
	opcode := &accountOpcode{}
	reqData := c.GetString("req_data")
	// logger.I("XXXXXX:%s", reqData)
	decoder := json.NewDecoder(strings.NewReader(reqData))
	err := decoder.Decode(&opcode)
	if err != nil {
		logger.E(err.Error())
		return nil
	}
	return opcode
}

func (c *AccountController) tryRequest(url string, params []byte, header string) ([]byte, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, time.Second*10)
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(time.Second * 10))
				return conn, nil
			},
			ResponseHeaderTimeout: time.Second * 10,
		},
	}

	body := bytes.NewBuffer(params)
	res, err := client.Post(url, header, body)
	if err != nil {
		return nil, err
	}

	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *AccountController) tryRgister(opcode *accountOpcode) int {
	//用户名检查
	if opcode.Args.UserName == "" || len(opcode.Args.UserName) > 48 {
		return common.USERNAME_IS_INVALID
	}

	//密码检查
	if opcode.Args.Password == "" || len(opcode.Args.Password) > 48 {
		return common.PASSWORD_IS_INVALID
	}

	//账号名转小写
	opcode.Args.UserName = strings.ToLower(opcode.Args.UserName)

	//生成盐
	//18字节，包含大小写字母
	salt := common.GetRandomString(18)
	pwdMD5 := common.GetMD5(opcode.Args.Password + salt)
	guid := strings.ToUpper(common.GetGUID())

	accountID, outResult := models.AddAccount(guid, opcode.Args.UserName, pwdMD5, salt, c.Ctx.Request.RemoteAddr)
	if outResult == -1 {
		return common.SE_DB_HANDLER_ERROR
	}

	if outResult == 1 {
		return common.USERNAME_ALREADY_EXIST
	}

	if outResult == 2 {
		logger.E("Guid already exist %s", guid)
		return common.USERNAME_ALREADY_EXIST
	}

	//设置当前用户信息
	user := &models.User{
		AccountID: accountID,
		GUID:      guid,
		Channel:   opcode.Args.Channel,
		GameID:    opcode.Args.GameID,
		Salt:      salt,
	}

	c.SetSession("user", user)
	return common.SUCCEED
}

func (c *AccountController) tryVeriySMSCaptcha(opcode *accountOpcode) int {
	if len(opcode.Args.SMSCaptcha) != 4 {
		return common.CAPTCHA_INVALID
	}

	memcache, err := common.GetMemoryCache()
	if err != nil {
		logger.E(err.Error())
		return common.SE_GET_CACHE_ERROR
	}

	cachekey := "captcha" + opcode.Args.Phone
	// logger.I("tryVeriySMSCaptcha:%s", cachekey)
	if !memcache.IsExist(cachekey) {
		return common.CAPTCHA_INVALID
	}

	// logger.I("tryVeriySMSCaptcha:%v %v", opcode.Args.SMSCaptcha, memcache.Get(cachekey))
	if opcode.Args.SMSCaptcha != memcache.Get(cachekey).(string) {
		return common.CAPTCHA_INVALID
	}

	memcache.Delete(cachekey)

	return common.SUCCEED
}

func (c *AccountController) tryFindChannelAccount(userid string, channel string) (*models.AccountBase, int) {
	acct := models.FindAccountByChannel(channel, userid)
	if acct == nil {
		//生成盐
		//18字节，包含大小写字母
		salt := common.GetRandomString(18)
		pwdMD5 := common.GetMD5(userid + salt)
		guid := strings.ToUpper(common.GetGUID())
		username := channel + userid
		outResult := models.AddAccountByChannel(guid, channel, userid, username, pwdMD5, salt, c.Ctx.Request.RemoteAddr)
		if outResult == -1 {
			return nil, common.SE_DB_HANDLER_ERROR
		}

		if outResult == 1 {
			return nil, common.USERNAME_ALREADY_EXIST
		}

		if outResult == 2 {
			logger.E("Guid already exist %s", guid)
			return nil, common.USERNAME_ALREADY_EXIST
		}

		if outResult == 3 {
			logger.E("Channel id already exist %s", userid)
			return nil, common.USERNAME_ALREADY_EXIST
		}

		acct = models.FindAccountByChannel(channel, userid)
		if acct == nil {
			return nil, common.NOT_FIND_USERNAME
		}
	}

	return acct, common.SUCCEED
}

//Login 处理账号登录流程
func (c *AccountController) Login() {
	result := accountResult{"login", common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), ""}
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	sessionObj := c.GetSession("user")

	// 已登录
	if sessionObj != nil {
		result.ErrorCode = common.ALREADY_LOGIN
		return
	}

	opcode := c.decodeAccountArgs()

	// //游戏检查
	// gameID := beego.AppConfig.String("common::gameid")
	// if opcode.Args.GameID != gameID {
	// 	result.ErrorCode = common.UNKNOW_GAME_ID
	// 	return
	// }

	fmt.Println(opcode)
	//用户名检查
	if opcode.Args.UserName == "" || len(opcode.Args.UserName) > 48 {
		result.ErrorCode = common.USERNAME_IS_INVALID
		return
	}

	//密码检查
	if opcode.Args.Password == "" || len(opcode.Args.Password) > 48 {
		result.ErrorCode = common.PASSWORD_IS_INVALID
		return
	}

	//用户名转小写
	opcode.Args.UserName = strings.ToLower(opcode.Args.UserName)

	//获得账号
	acct := models.GetAccount(opcode.Args.UserName)
	if acct == nil {
		result.ErrorCode = common.NOT_FIND_USERNAME
		return
	}

	pwdMD5 := common.GetMD5(opcode.Args.Password + acct.Salt)
	if pwdMD5 != acct.Password {
		result.ErrorCode = common.USERNAME_OR_PASSWORD_ERROR
		return
	}

	//设置当前用户信息
	user := &models.User{
		AccountID: acct.AccountId,
		GUID:      acct.AccountGuid,
		Salt:      acct.Salt,
		Channel:   opcode.Args.Channel,
		GameID:    opcode.Args.GameID,
	}

	c.SetSession("user", user)
	result.ErrorCode = common.SUCCEED
	return
}

//ChannelLogin 渠道登录
func (c *AccountController) ChannelLogin() {
	result := accountResult{"channel_login", common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), ""}
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	sessionObj := c.GetSession("user")

	// 已登录
	if sessionObj != nil {
		result.ErrorCode = common.ALREADY_LOGIN
		return
	}

	opcode := c.decodeAccountArgs()

	// //游戏检查
	// gameID := beego.AppConfig.String("common::gameid")
	// if opcode.Args.GameID != gameID {
	// 	result.ErrorCode = common.UNKNOW_GAME_ID
	// 	return
	// }

	//渠道ID检查
	if opcode.Args.UserName == "" {
		result.ErrorCode = common.USERNAME_IS_INVALID
		return
	}

	// UserName == sid
	ucAcct := models.GetUCAccount(opcode.Args.UserName)
	if ucAcct == nil {
		result.ErrorCode = common.NOT_FIND_USERNAME
		return
	}

	// ucAcct := &models.UCAccount{
	// 	AccountID: opcode.Args.UserName,
	// }

	acct := models.FindAccountByChannel(opcode.Args.Channel, ucAcct.AccountID)
	if acct == nil {
		//生成盐
		//18字节，包含大小写字母
		salt := common.GetRandomString(18)
		pwdMD5 := common.GetMD5(ucAcct.AccountID + salt)
		guid := strings.ToUpper(common.GetGUID())
		outResult := models.AddAccountByChannel(guid, opcode.Args.Channel, ucAcct.AccountID, ucAcct.AccountID, pwdMD5, salt, c.Ctx.Request.RemoteAddr)
		if outResult == -1 {
			result.ErrorCode = common.SE_DB_HANDLER_ERROR
			return
		}

		if outResult == 1 {
			result.ErrorCode = common.USERNAME_ALREADY_EXIST
			return
		}

		if outResult == 2 {
			result.ErrorCode = common.USERNAME_ALREADY_EXIST
			logger.E("Guid already exist %s", guid)
			return
		}

		if outResult == 3 {
			result.ErrorCode = common.USERNAME_ALREADY_EXIST
			logger.E("Channel id already exist %s", ucAcct.AccountID)
			return
		}

		acct := models.FindAccountByChannel(opcode.Args.Channel, ucAcct.AccountID)
		if acct == nil {
			result.ErrorCode = common.NOT_FIND_USERNAME
			return
		}
	}

	//设置当前用户信息
	user := &models.User{
		AccountID: acct.AccountId,
		GUID:      acct.AccountGuid,
		Salt:      acct.Salt,
		Channel:   opcode.Args.Channel,
		GameID:    opcode.Args.GameID,
	}

	c.SetSession("user", user)
	result.ErrorCode = common.SUCCEED

	return
}

//Register 处理账号注册流程
func (c *AccountController) Register() {
	result := accountResult{"register", common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), ""}
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	sessionObj := c.GetSession("user")

	// 已登录，不能注册
	if sessionObj != nil {
		result.ErrorCode = common.ALREADY_LOGIN_NOT_REGISTER
		return
	}

	opcode := c.decodeAccountArgs()

	// //游戏检查
	// gameID := beego.AppConfig.String("common::gameid")
	// if opcode.Args.GameID != gameID {
	// 	result.ErrorCode = common.UNKNOW_GAME_ID
	// 	return
	// }

	result.ErrorCode = c.tryRgister(opcode)
	return
}

//Tourists 游客登录
func (c *AccountController) Tourists() {
	result := accountResult{"tourists", common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), ""}
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	sessionObj := c.GetSession("user")

	// 已登录，不能注册
	if sessionObj != nil {
		result.ErrorCode = common.ALREADY_LOGIN_NOT_REGISTER
	}

	opcode := c.decodeAccountArgs()

	// //游戏检查
	// gameID := beego.AppConfig.String("common::gameid")
	// if opcode.Args.GameID != gameID {
	// 	result.ErrorCode = common.UNKNOW_GAME_ID
	// 	return
	// }

	//用户名检查
	if len(opcode.Args.UserName) > 36 {
		result.ErrorCode = common.USERNAME_IS_INVALID
		return
	}

	var user *models.User
	var username string
	//获得账号
	acct := models.FindAccountByGUID(opcode.Args.UserName)
	if acct == nil {
		//生成新账号
		salt := common.GetRandomString(18)
		pwdMD5 := common.GetMD5(salt)
		guid := strings.ToUpper(common.GetGUID())
		username = guid
		accountID, outResult := models.AddAccount(guid, username, pwdMD5, salt, c.Ctx.Request.RemoteAddr)
		if outResult == -1 {
			result.ErrorCode = common.SE_DB_HANDLER_ERROR
			return
		}

		if outResult == 1 {
			result.ErrorCode = common.USERNAME_ALREADY_EXIST
			return
		}

		if outResult == 2 {
			result.ErrorCode = common.USERNAME_ALREADY_EXIST
			logger.E("Guid already exist %s", guid)
			return
		}

		//设置当前用户信息
		user = &models.User{
			AccountID: accountID,
			GUID:      guid,
			Channel:   opcode.Args.Channel,
			GameID:    opcode.Args.GameID,
			Salt:      salt,
		}
	} else {
		username = acct.Username
		//设置当前用户信息
		user = &models.User{
			AccountID: acct.AccountId,
			GUID:      acct.AccountGuid,
			Salt:      acct.Salt,
			Channel:   opcode.Args.Channel,
			GameID:    opcode.Args.GameID,
		}
	}

	c.SetSession("user", user)
	result.ErrorCode = common.SUCCEED
	result.UserName = username
	return
}

//PhoneRegister 手机注册
func (c *AccountController) PhoneRegister() {
	result := accountResult{"register", common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), ""}
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	sessionObj := c.GetSession("user")

	// 已登录，不能注册
	if sessionObj != nil {
		result.ErrorCode = common.ALREADY_LOGIN_NOT_REGISTER
		return
	}

	opcode := c.decodeAccountArgs()

	//游戏检查
	// gameID := beego.AppConfig.String("common::gameid")
	// if opcode.Args.GameID != gameID {
	// 	result.ErrorCode = common.UNKNOW_GAME_ID
	// 	return
	// }

	ret := c.tryVeriySMSCaptcha(opcode)
	if ret != common.SUCCEED {
		result.ErrorCode = ret
		return
	}

	opcode.Args.UserName = opcode.Args.Phone
	result.ErrorCode = c.tryRgister(opcode)
	return
}

func (c *AccountController) SMSCaptcha() {
	result := accountResult{"smscaptcha", common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), ""}
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	opcode := c.decodeAccountArgs()

	// //游戏检查
	// gameID := beego.AppConfig.String("common::gameid")
	// if opcode.Args.GameID != gameID {
	// 	result.ErrorCode = common.UNKNOW_GAME_ID
	// 	return
	// }

	memcache, err := common.GetMemoryCache()
	if err != nil {
		result.ErrorCode = common.SE_GET_CACHE_ERROR
		logger.E(err.Error())
		return
	}

	cachekey := "captcha" + opcode.Args.Phone
	if memcache.IsExist(cachekey) {
		result.ErrorCode = common.CAPTCHA_NOT_EXPIRED
		return
	}

	expiredtime, err := beego.AppConfig.Int64("sendsms::expiredtime")
	if err != nil {
		expiredtime = 180
	}

	code := strconv.Itoa(common.RandInRange(1000, 9999))
	memcache.Put(cachekey, code, time.Duration(expiredtime)*time.Second)

	logger.I("SMSCaptcha:[%s], [%s], [%d]", cachekey, code, expiredtime)

	// if memcache.IsExist(cachekey) {
	// 	logger.I("SMSCaptcha:%v", memcache.Get(cachekey))
	// }
	//发送短信
	models.SendSMSCaptcha(opcode.Args.Phone, code, beego.AppConfig.String("sendsms::templateid"))
	result.ErrorCode = common.SUCCEED
	return
}

func (c *AccountController) TouristsBind() {
	result := accountResult{"touristsbind", common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), ""}
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	opcode := c.decodeAccountArgs()

	// //游戏检查
	// gameID := beego.AppConfig.String("common::gameid")
	// if opcode.Args.GameID != gameID {
	// 	result.ErrorCode = common.UNKNOW_GAME_ID
	// 	return
	// }

	// logger.I("TouristsBind:[%v]", opcode.Args)
	ret := c.tryVeriySMSCaptcha(opcode)
	if ret != common.SUCCEED {
		result.ErrorCode = ret
		return
	}

	//用户名检查
	if opcode.Args.UserName == "" || len(opcode.Args.UserName) > 48 {
		result.ErrorCode = common.USERNAME_IS_INVALID
		return
	}

	//手机号检查
	if opcode.Args.Phone == "" || len(opcode.Args.Phone) > 48 {
		result.ErrorCode = common.PHONE_IS_INVALID
		return
	}

	//密码检查
	if opcode.Args.Password == "" || len(opcode.Args.Password) > 48 {
		result.ErrorCode = common.PASSWORD_IS_INVALID
		return
	}

	// 检查手机用户名是否重名
	acct := models.GetAccount(opcode.Args.Phone)
	if acct != nil {
		// logger.I("TouristsBind: 111 [%v]", acct)
		result.ErrorCode = common.PHONE_USERNAME_ALREADY_EXIST
		return
	}

	// 找到游客账号数据
	acct = models.GetAccount(opcode.Args.UserName)
	if acct == nil {
		result.ErrorCode = common.NOT_FIND_USERNAME
		return
	}

	if acct.AccountGuid != acct.Username {
		result.ErrorCode = common.IS_NOT_TOURIST
		return
	}

	// logger.I("TouristsBind: 222 [%v]", acct)
	//生成盐
	//18字节，包含大小写字母
	pwdMD5 := common.GetMD5(opcode.Args.Password + acct.Salt)

	// 更新账号和密码
	if !models.BindAccount(opcode.Args.UserName, opcode.Args.Phone, pwdMD5) {
		result.ErrorCode = common.TOURISTS_BIND_FAILED
		return
	}

	result.ErrorCode = common.SUCCEED
	return
}

func (c *AccountController) StarscloudLogin() {
	result := accountResult{"starscloud_login", common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), ""}
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	sessionObj := c.GetSession("user")

	// 已登录
	if sessionObj != nil {
		result.ErrorCode = common.ALREADY_LOGIN
		return
	}

	opcode := c.decodeAccountArgs()

	// //游戏检查
	// gameID := beego.AppConfig.String("common::gameid")
	// if opcode.Args.GameID != gameID {
	// 	result.ErrorCode = common.UNKNOW_GAME_ID
	// 	return
	// }

	//渠道ID检查
	if opcode.Args.UserName == "" || opcode.Args.Channel == "" || opcode.Args.Token == "" || opcode.Args.Timestamp == "" {
		result.ErrorCode = common.PARAM_LESS
		return
	}

	appid := beego.AppConfig.String("starscloud::appid")
	pmsecret := beego.AppConfig.String("starscloud::pmsecret")
	if appid == "" || pmsecret == "" {
		result.ErrorCode = common.SYSTEM_ERROR
		logger.E("appid和pmsecret不能为空！")
		return
	}

	sign := common.GetMD5(appid + opcode.Args.Channel + opcode.Args.UserName + opcode.Args.Token + opcode.Args.Timestamp + pmsecret)
	if sign != opcode.Args.Sign {
		logger.E("签名验证失败：\r\n[%s]\r\n%v", sign, opcode.Args)
		result.ErrorCode = common.LOGIN_SIGN_CHECK_FAILED
		return
	}

	acct := models.FindAccountByChannel(opcode.Args.Channel, opcode.Args.UserName)
	if acct == nil {
		//生成盐
		//18字节，包含大小写字母
		salt := common.GetRandomString(18)
		pwdMD5 := common.GetMD5(opcode.Args.UserName + salt)
		guid := strings.ToUpper(common.GetGUID())
		outResult := models.AddAccountByChannel(guid, opcode.Args.Channel, opcode.Args.UserName, opcode.Args.UserName, pwdMD5, salt, c.Ctx.Request.RemoteAddr)
		if outResult == -1 {
			result.ErrorCode = common.SE_DB_HANDLER_ERROR
			return
		}

		if outResult == 1 {
			result.ErrorCode = common.USERNAME_ALREADY_EXIST
			return
		}

		if outResult == 2 {
			result.ErrorCode = common.USERNAME_ALREADY_EXIST
			logger.E("Guid already exist %s", guid)
			return
		}

		if outResult == 3 {
			result.ErrorCode = common.USERNAME_ALREADY_EXIST
			logger.E("Channel id already exist %s", opcode.Args.UserName)
			return
		}

		acct := models.FindAccountByChannel(opcode.Args.Channel, opcode.Args.UserName)
		if acct == nil {
			result.ErrorCode = common.NOT_FIND_USERNAME
			return
		}
	}

	//设置当前用户信息
	user := &models.User{
		AccountID: acct.AccountId,
		GUID:      acct.AccountGuid,
		Salt:      acct.Salt,
		Channel:   opcode.Args.Channel,
		GameID:    opcode.Args.GameID,
	}

	c.SetSession("user", user)
	result.ErrorCode = common.SUCCEED

	return
}

func (c *AccountController) AnySdkVerify() {
	anysdkResult := &anysdkLoginResult{}
	c.Data["json"] = anysdkResult
	defer func() {
		c.ServeJSON()
	}()

	resdata, err := c.tryRequest(beego.AppConfig.String("anysdk::loginurl"), c.Ctx.Input.RequestBody, "application/x-www-form-urlencoded;charset=utf-8")
	if err != nil {
		logger.E(err.Error())
		return
	}

	err = json.Unmarshal(resdata, &anysdkResult)
	if err != nil {
		logger.E(err.Error())
		return
	}

	if anysdkResult.Status != "ok" {
		return
	}

	acct, errcode := c.tryFindChannelAccount(anysdkResult.Common.Uid, anysdkResult.Common.UserSdk)
	if errcode != common.SUCCEED {
		return
	}

	memcache, err := common.GetMemoryCache()
	if err != nil {
		// result.ErrorCode = common.SE_GET_CACHE_ERROR
		logger.E(err.Error())
		return
	}

	expiredtime, err := beego.AppConfig.Int64("anysdk::logintimeout")
	if err != nil {
		expiredtime = anysdkLoginTimeout
	}

	captcha := common.GetRandomString(8)
	key := "anysdk_" + anysdkResult.Common.Channel + "_" + anysdkResult.Common.Uid
	logger.I("AnySdkVerify:%s", key)
	data := &anysdkVerifyData{
		captcha,
		acct.AccountGuid,
	}

	memcache.Put(key, data, time.Duration(expiredtime)*time.Second)
	anysdkResult.Ext = captcha
	// logger.D("%v", data)
}

func (c *AccountController) AnySdkLogin() {
	result := accountResult{"anysdk_login", common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), ""}
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	opcode := c.decodeAccountArgs()
	if opcode == nil {
		result.ErrorCode = common.DATA_ILLEGAL
		return
	}

	// logger.D("xxx:%v", opcode)

	if opcode.Args.UserName == "" {
		result.ErrorCode = common.USERNAME_IS_INVALID
		return
	}

	if opcode.Args.Channel == "" {
		result.ErrorCode = common.CHANNEL_INVALID
		return
	}

	memcache, err := common.GetMemoryCache()
	if err != nil {
		logger.E(err.Error())
		return
	}

	cachekey := "anysdk_" + opcode.Args.Channel + "_" + opcode.Args.UserName
	if !memcache.IsExist(cachekey) {
		logger.E("Cache [%s] 不存在", cachekey)
		return
	}

	data := memcache.Get(cachekey).(*anysdkVerifyData)
	if data == nil {
		logger.E("Cache [%s:%v] 数据转换失败", cachekey, memcache.Get(cachekey))
		return
	}

	if opcode.Args.Password != data.Captcha {
		logger.E("验证失败[%s, %s]", opcode.Args.Password, data.Captcha)
		return
	}

	// memcache.Delete(cachekey)

	acct := models.FindAccountByGUID(data.GUID)
	if acct == nil {
		logger.E("%v", opcode)
		result.ErrorCode = common.NOT_FIND_USERNAME
		return
	}
	//设置当前用户信息
	user := &models.User{
		AccountID: acct.AccountId,
		GUID:      acct.AccountGuid,
		Salt:      acct.Salt,
		Channel:   "anysdk" + opcode.Args.Channel,
		GameID:    opcode.Args.GameID,
	}

	c.SetSession("user", user)
	result.ErrorCode = common.SUCCEED
}
