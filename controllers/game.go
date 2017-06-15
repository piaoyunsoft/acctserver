package controllers

import (
	"encoding/json"
	"moton/acctserver/common"
	"moton/acctserver/models"
	"moton/logger"
	"strings"

	"github.com/astaxie/beego"
)

type gameServerInfo struct {
	New       bool   `json:"new"`
	Recommend bool   `json:"recommend"`
	State     int    `json:"state"`
	Info      string `json:"info"`
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Port      string `json:"port"`
}

type gameArgs struct {
	ServerID string `json:"server_id"`
	RandomA  string `json:"random_a"`
}

type rechargeArgs struct {
	Rmb      string `json:"rmb"`
	ServerID string `json:"server_id"`
	CharDBID string `json:"char_dbid"`
}

type buyitemArgs struct {
	Id string `json:"id"`
}

type getServerArgs struct {
	Version string `json:"version"`
	GameId  string `json:"game_id"`
}

type gameOpcode struct {
	Opcode string      `json:"opcode"`
	Args   interface{} `json:"arg"`
}

type gameResult struct {
	Opcode     string                 `json:"opcode"`
	ErrorCode  int                    `json:"error_code"`
	ErrorMsg   string                 `json:"error_msg"`
	ServerList map[int]gameServerInfo `json:"server_list,omitempty"`
	AccountID  int64                  `json:"account_id,omitempty"`
	MStr       string                 `json:"M,omitempty"`
}

type GameController struct {
	beego.Controller
}

func (c *GameController) checkUser(user *models.User) int {
	if user == nil {
		return common.NEED_LOGIN_FIRST
	}

	return common.SUCCEED
}

func (c *GameController) ServerList() {
	result := gameResult{"server_list", common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), make(map[int]gameServerInfo), 0, ""}
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		// c.ServeJSON()

		data, err := json.Marshal(&result)
		if err != nil {
			return
		}
		compressed := common.DoZlibCompress(data)
		c.Ctx.Output.Header("Content-Type", "application/octet-streamn; charset=utf-8")
		c.Ctx.Output.Body(compressed)
	}()

	//验证玩家
	sessionObj := c.GetSession("user")

	if sessionObj == nil {
		result.ErrorCode = common.NEED_LOGIN_FIRST
		return
	}

	user := sessionObj.(*models.User)

	//验证用户
	errcode := c.checkUser(user)
	if errcode != common.SUCCEED {
		result.ErrorCode = errcode
		return
	}

	//解析参数
	var args getServerArgs
	err := json.Unmarshal([]byte(c.GetString("req_data")), &args)
	if err != nil {
		result.ErrorCode = common.DATA_ILLEGAL
		logger.E(err.Error())
		return
	}

	realAddr := c.Ctx.Request.Header.Get("X-Forwarded-For")
	if realAddr == "" {
		fullAddr := c.Ctx.Request.RemoteAddr
		realAddr = fullAddr[:strings.LastIndex(fullAddr, ":")]
	}
	// logger.D("[%s], [%s], [%s], [%s]", c.Ctx.Request.RemoteAddr, realAddr, c.Ctx.Request.Header.Get("X-Forwarded-For"), c.Ctx.Request.Header.Get("Remote_addr"))
	gameId := args.GameId
	if gameId == "" {
		gameId = user.GameID
	}
	logger.D("GameID:%s, GameVersion:%s", gameId, args.Version)
	serverList := models.GetServerList(gameId, realAddr, args.Version)
	for _, v := range serverList {
		result.ServerList[v.Id] = gameServerInfo{
			v.Isnew,
			v.Isrecmd,
			1,
			v.ClientInformation,
			v.Name,
			v.Bindip,
			v.Bindport,
		}
	}

	result.ErrorCode = common.SUCCEED
	return
}

func (c *GameController) EnterGame() {
	result := gameResult{"enter_game", common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), nil, 0, ""}
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	sessionObj := c.GetSession("user")

	if sessionObj == nil {
		result.ErrorCode = common.NEED_LOGIN_FIRST
		return
	}

	user := sessionObj.(*models.User)

	//验证用户
	errcode := c.checkUser(user)
	if errcode != common.SUCCEED {
		result.ErrorCode = errcode
		return
	}

	//解析参数
	args := &gameArgs{}
	opcode := gameOpcode{"", args}
	err := json.Unmarshal([]byte(c.GetString("req_data")), &opcode)
	if err != nil {
		result.ErrorCode = common.DATA_ILLEGAL
		logger.E(err.Error())
		return
	}

	//获取服务器信息
	gameServer := models.GetServer(args.ServerID)
	if gameServer == nil {
		result.ErrorCode = common.SERVER_ID_IS_INVALID
		return
	}

	//验证随机因子
	if len(args.RandomA) != 16 {
		result.ErrorCode = common.DATA_ILLEGAL
		return
	}

	//随机32字节的字符串
	strB := common.GetRandomString(32)
	strM := common.GetMD5(args.RandomA + user.GUID + strB)

	//插入认证信息到服务器
	_, err = gameServer.WebLogin(user.AccountID, strM, c.Ctx.Request.RemoteAddr)
	if err != nil {
		result.ErrorCode = common.SE_DB_HANDLER_ERROR
		return
	}

	//保存到用户对象
	user.ServerID = args.ServerID

	//回馈账号id、认证信息M到客户端
	result.ErrorCode = common.SUCCEED
	result.AccountID = user.AccountID
	result.MStr = strM
	return
}

// func (c *GameController) Recharge() {
// 	result := gameResult{"do_recharge", common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), nil, 0, ""}
// 	c.Data["json"] = &result
// 	defer func() {
// 		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
// 		c.ServeJSON()
// 	}()

// 	sessionObj := c.GetSession("user")

// 	if sessionObj == nil {
// 		result.ErrorCode = common.NEED_LOGIN_FIRST
// 		return
// 	}

// 	user := sessionObj.(*models.User)

// 	//验证用户
// 	errcode := c.checkUser(user)
// 	if errcode != common.SUCCEED {
// 		result.ErrorCode = errcode
// 		return
// 	}

// 	//解析参数
// 	args := &rechargeArgs{}
// 	opcode := gameOpcode{"", args}
// 	err := json.Unmarshal([]byte(c.GetString("req_data")), &opcode)
// 	if err != nil {
// 		result.ErrorCode = common.DATA_ILLEGAL
// 		logger.E(err.Error())
// 		return
// 	}

// 	rechargeAmount, err := strconv.Atoi(args.Rmb)
// 	if err != nil {
// 		result.ErrorCode = common.DATA_ILLEGAL
// 		return
// 	}

// 	//获取服务器信息
// 	gameServer := models.GetServer(args.ServerID)
// 	if gameServer == nil {
// 		result.ErrorCode = common.SERVER_ID_IS_INVALID
// 		return
// 	}

// 	//获取充值配置信息
// 	rechargeCfg := models.GetRechargeCfg(user.GameID, rechargeAmount)
// 	if rechargeCfg == nil {
// 		result.ErrorCode = common.USER_RECHARGE_RMB_VALUE_INVALID
// 		return
// 	}

// 	//检查是否是首次充值
// 	bonusAmount := 0
// 	ret, err := gameServer.IsFirstRecharge(user.AccountID, rechargeAmount)
// 	if err != nil {
// 		result.ErrorCode = common.USER_RECHARGE_UNKNOW_ERROR
// 		return
// 	}

// 	if ret {
// 		bonusAmount = rechargeCfg.FirstRechargeBonus
// 	}

// 	//执行充值操作
// 	guid := strings.ToUpper(common.GetGUID())
// 	err = gameServer.WebRecharge(guid, user.GameID, user.AccountID, rechargeCfg.ExchangeAmount, bonusAmount)
// 	if err != nil {
// 		result.ErrorCode = common.USER_RECHARGE_UNKNOW_ERROR
// 		return
// 	}

// 	result.ErrorCode = common.SUCCEED
// 	return
// }
