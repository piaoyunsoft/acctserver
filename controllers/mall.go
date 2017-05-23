package controllers

import (
	"encoding/json"
	"fmt"
	"moton/acctserver/common"
	"moton/acctserver/models"
	"moton/logger"
	"sort"
	"time"

	"net/url"

	"strconv"

	"github.com/astaxie/beego"
)

type MallController struct {
	beego.Controller
}

type mallArgs struct {
	Type      int    `json:"type"`
	Category  int    `json:"category"`
	ProductID int    `json:"product_id"`
	OrderID   int64  `json:"order_id"`
	CharDBID  int64  `json:"char_dbid"`
	GameID    string `json:"game_id"`
	ServerID  int    `json:"server_id"`
	Channel   string `json:"channel"`
}

type mallOpcode struct {
	Opcode string   `json:"opcode"`
	Args   mallArgs `json:"arg,omitempty"`
}

type mallResult struct {
	Opcode      string           `json:"opcode"`
	ErrorCode   int              `json:"error_code"`
	ErrorMsg    string           `json:"error_msg"`
	ProductList []models.Product `json:"products,omitempty"`
	OrderID     *int64           `json:"order_id,omitempty"`
	ProductID   *int             `json:"product_id,omitempty"`
	Price       *float64         `json:"price,omitempty"`
	ProductName string           `json:"product_name,omitempty"`
}

type starscloudOrder struct {
	Paytype      string `form:"type"`         //通知类型，pay表示支付通知	否
	ProductName  string `form:"productName"`  //商品名称	否
	ProductID    string `form:"productId"`    //商品ID	否
	Amount       string `form:"amount"`       //支付金额（分）	是
	ChannOrderID string `form:"channOrderId"` //渠道分配的订单号	是
	ChannType    string `form:"channType"`    //渠道类型	是
	PmOrderID    string `form:"pmOrderId"`    //支付中间件服务分配的订单号	是
	UID          string `form:"uid"`          //渠道分配的终端用户ID	是
	PmAppID      string `form:"pmAppId"`      //支付中间件服务分配的应用标识	是
	PackName     string `form:"packName"`     //应用程序包名	否
	ExtraInfo    string `form:"extraInfo"`    //终端用户扩展信息	否
	Sign         string `form:"sign"`         //md5签名，签名内容如下	否
}

/*
order_id	string	订单号，AnySDK 产生的订单号
product_count	string	要购买商品数量（暂不提供具体数量）
下列渠道请不要使用此金额字段作为发放道具的依据，而应该使用 product_id 作为发放道具的依据：
    GooglePlay
    Apple appstore
    心动（非越狱）
amount	string	支付金额，单位元 值根据不同渠道的要求可能为浮点类型
pay_status	string	支付状态，1 为成功，非1则为其他异常状态，游服请在成功的状态下发货
pay_time	string	支付时间，YYYY-mm-dd HH:ii:ss 格式
user_id	string	用户 ID，用户系统的用户 ID
order_type	string	支付方式 详见 支付渠道标识表
game_user_id	string	游戏内用户 ID，支付时传入的 Role_Id 参数
server_id	string	服务器 ID，支付时传入的 Server_Id 参数
product_name	string	商品名称，支付时传入的 Product_Name 参数
product_id	string	商品 ID，支付时传入的 Product_Id 参数
下列渠道请使用此 product_id 作为发放道具的依据，而不要使用金额amount字段作为发放道具的依据：
    GooglePlay
    Apple appstore
    心动（非越狱）
channel_product_id	string	product_id 字段的值对应的渠道商品 ID，对应关系可以在 dev 后台 -> 游戏列表 -> 管理商品 页面进行配置。
private_data	string	自定义参数，调用客户端支付函数时传入的EXT参数，透传给游戏服务器
channel_number	string	渠道编号 渠道列表
sign	string	通用签名串，通用验签参考签名算法
source	string	渠道服务器通知 AnySDK 时请求的参数
enhanced_sign	string	增强签名串，验签参考签名算法（有增强密钥的游戏有效）
channel_order_id	string	渠道订单号，如果渠道通知过来的参数没有渠道订单号则为空。
game_id	string	游戏 ID，AnySDK 服务端为游戏分配的唯一标识
plugin_id	string	插件 ID，AnySDK 插件数字唯一标识
*/
type anysdkOrder struct {
	OrderID      string `form:"order_id"`      //订单号，AnySDK 产生的订单号
	ProductCount string `form:"product_count"` /*要购买商品数量（暂不提供具体数量）
	下列渠道请不要使用此金额字段作为发放道具的依据，而应该使用 product_id 作为发放道具的依据：
		GooglePlay
		Apple appstore
		心动（非越狱）*/
	Amount      string `form:"amount"`       //支付金额，单位元 值根据不同渠道的要求可能为浮点类型
	PayStatus   string `form:"pay_status"`   //支付状态，1 为成功，非1则为其他异常状态，游服请在成功的状态下发货
	PayTime     string `form:"pay_time"`     //支付时间，YYYY-mm-dd HH:ii:ss 格式
	UserID      string `form:"user_id"`      //用户 ID，用户系统的用户 ID
	OrderType   string `form:"order_type"`   //支付方式 详见 支付渠道标识表
	GameUserID  string `form:"game_user_id"` //游戏内用户 ID，支付时传入的 Role_Id 参数
	ServerID    string `form:"server_id"`    //服务器 ID，支付时传入的 Server_Id 参数
	ProductName string `form:"product_name"` //商品名称，支付时传入的 Product_Name 参数
	ProductID   string `form:"product_id"`   /*商品 ID，支付时传入的 Product_Id 参数
	下列渠道请使用此 product_id 作为发放道具的依据，而不要使用金额amount字段作为发放道具的依据：
		GooglePlay
		Apple appstore
		心动（非越狱）*/
	ChannelProductID string `form:"channel_product_id"` //product_id 字段的值对应的渠道商品 ID，对应关系可以在 dev 后台 -> 游戏列表 -> 管理商品 页面进行配置。
	PrivateData      string `form:"private_data"`       //自定义参数，调用客户端支付函数时传入的EXT参数，透传给游戏服务器
	ChannelNumber    string `form:"channel_number"`     //渠道编号 渠道列表
	Sign             string `form:"sign"`               //通用签名串，通用验签参考签名算法
	Source           string `form:"source"`             //渠道服务器通知 AnySDK 时请求的参数
	EnhancedSign     string `form:"enhanced_sign"`      //增强签名串，验签参考签名算法（有增强密钥的游戏有效）
	ChannelOrderID   string `form:"channel_order_id"`   //渠道订单号，如果渠道通知过来的参数没有渠道订单号则为空。
	GameID           string `form:"game_id"`            //游戏 ID，AnySDK 服务端为游戏分配的唯一标识
	PluginID         string `form:"plugin_id"`          //插件 ID，AnySDK 插件数字唯一标识
}

func (c *MallController) createMallResult(opcode string) *mallResult {
	return &mallResult{opcode, common.SYSTEM_ERROR, common.GetErrorMsg(common.SYSTEM_ERROR), nil, nil, nil, nil, ""}
}

func (c *MallController) checkAnySdkSign(params string, sign string, tail string) bool {
	md5 := common.GetMD5(params)
	md5 = common.GetMD5(md5 + tail)
	if md5 != sign {
		return false
	}
	return true
}

func (c *MallController) checkAnySdkOrder(values url.Values) bool {
	var params string
	var sortList []string
	for k := range values {
		if (k != "sign") && (k != "enhanced_sign") {
			sortList = append(sortList, k)
		}
	}

	sort.Strings(sortList)
	for _, v := range sortList {
		params += values[v][0]
	}

	if !c.checkAnySdkSign(params, values["enhanced_sign"][0], beego.AppConfig.String("anysdk::enhancedkey")) {
		return false
	}

	return true
}

func (c *MallController) createOrderID(charDBID int64) (int64, error) {
	now := time.Now()
	nowStr := strconv.FormatInt(now.Unix(), 10)
	strCharDBID := strconv.FormatInt(charDBID, 10)
	strOrderID := fmt.Sprintf("%s%s%s", now.Format("060102"), nowStr[len(nowStr)-8:], strCharDBID[len(strCharDBID)-4:])

	orderID, err := strconv.ParseInt(strOrderID, 10, 64)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

func (c *MallController) getOrderIDFromSession() int64 {
	sessionObj := c.GetSession("order")
	if sessionObj == nil {
		return 0
	}

	return sessionObj.(int64)
}

func (c *MallController) ProductList() {
	result := c.createMallResult("productlist")
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	opcode := mallOpcode{}
	err := json.Unmarshal([]byte(c.GetString("req_data")), &opcode)
	if err != nil {
		logger.E(err.Error())
		result.ErrorCode = common.DATA_ILLEGAL
		return
	}

	productList := models.GetProductList(opcode.Args.Category)
	// productList := models.GetProductList("_self_game")
	if productList == nil {
		result.ErrorCode = common.SE_DB_HANDLER_ERROR
		return
	}

	nowUnix := time.Now().Unix()
	for _, product := range productList {
		if product.BeginTime > 0 && nowUnix < product.BeginTime {
			continue
		}

		if product.EndTime > 0 && nowUnix > product.EndTime {
			continue
		}

		var serverList []int
		err := json.Unmarshal([]byte(product.ServerList), &serverList)
		if err != nil {
			logger.E("解析服务器列表到json失败: %s\n%s", product.ServerList, err.Error())
			continue
		}
		for _, serverID := range serverList {
			if serverID == opcode.Args.ServerID {
				if product.Times > 0 {
					count, err := models.GetOrderFinishedCount(product.Id, serverID, opcode.Args.CharDBID)
					if err != nil {
						logger.E(err.Error())
						break
					}
					product.Stock = product.Times - count
					if product.Stock <= 0 {
						break
					}
				}

				result.ProductList = append(result.ProductList, product)
				break
			}
		}
	}

	result.ErrorCode = common.SUCCEED
	return
}

func (c *MallController) Order() {
	result := c.createMallResult("order")
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	opcode := mallOpcode{}
	err := json.Unmarshal([]byte(c.GetString("req_data")), &opcode)
	if err != nil {
		result.ErrorCode = common.DATA_ILLEGAL
		return
	}

	logger.I("[Order] Opcode: %v", opcode)

	if opcode.Args.Channel == "" {
		result.ErrorCode = common.CHANNEL_INVALID
		return
	}

	// now := time.Now()

	gameServer := models.GetServer(opcode.Args.ServerID)
	if gameServer == nil {
		logger.E("找不到服务器!\n%v", opcode)
		return
	}

	if !gameServer.CharExists(opcode.Args.CharDBID) {
		logger.E("找不到指定的角色!\n%v", opcode)
		return
	}

	// unfinishedOrder, err := models.GetOrderUnfinished(opcode.Args.ServerID, opcode.Args.CharDBID)
	// if err != nil {
	// 	logger.E(err.Error())
	// 	return
	// }

	// if unfinishedOrder != nil {
	// timeout, err := beego.AppConfig.Int64("ordertimeout")
	// if err != nil {
	// 	timeout = common.OrderTimeoutTime
	// }

	// if unfinishedOrder.OrderTime+timeout > now.Unix() {
	// 	result.ErrorCode = common.ORDER_NOT_FINISHED
	// 	return
	// }

	// unfinishedOrder.DeliveryTime = now.Unix()
	// unfinishedOrder.State = common.OrderStateTimeout
	// err = models.UpdateOrder(unfinishedOrder)
	// if err != nil {
	// 	logger.E(err.Error())
	// 	return
	// }
	// }

	product := models.GetProduct(opcode.Args.ProductID)
	if product == nil {
		result.ErrorCode = common.CAN_NOT_FIND_PRODUCT
		return
	}

	if product.Times != 0 {
		count, err := models.GetOrderFinishedCount(opcode.Args.ProductID, opcode.Args.ServerID, opcode.Args.CharDBID)
		if err != nil {
			logger.E(err.Error())
			return
		}
		if product.Times-count < 0 {
			result.ErrorCode = common.HAVE_NO_PRODUCTS_IN_STOCK
			return
		}
	}

	orderID, err := c.createOrderID(opcode.Args.CharDBID)
	if err != nil {
		result.ErrorCode = common.CREATE_ORDER_ID_FAILED
		return
	}

	price := 0.0
	if product.DiscountPrice > 0 {
		price = product.DiscountPrice
	} else {
		price = product.Price
	}

	order := &models.Order{
		OrderID:      orderID,
		CharDBID:     opcode.Args.CharDBID,
		ProductID:    opcode.Args.ProductID,
		Price:        price,
		ServerID:     opcode.Args.ServerID,
		Channel:      opcode.Args.Channel,
		OrderTime:    time.Now().Unix(),
		DeliveryTime: 0,
		State:        0,
	}

	logger.I("%v", order)

	err = models.InsertOrder(order)
	if err != nil {
		logger.E(err.Error())
		result.ErrorCode = common.INSERT_ORDER_FAILED
		return
	}

	result.Price = &price
	result.OrderID = &orderID
	result.ProductID = &opcode.Args.ProductID
	result.ProductName = product.Name
	result.ErrorCode = common.SUCCEED
	return
}

func (c *MallController) CancelOrder() {
	result := c.createMallResult("cancel_order")
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	opcode := mallOpcode{}
	err := json.Unmarshal([]byte(c.GetString("req_data")), &opcode)
	if err != nil {
		result.ErrorCode = common.DATA_ILLEGAL
		return
	}

	// logger.D("%v\n%s", opcode, c.GetString("req_data"))

	if !models.CancelOrder(opcode.Args.OrderID) {
		result.ErrorCode = common.SUCCEED
		return
	}
	result.ErrorCode = common.SUCCEED
}

func (c *MallController) BuyProduct() {
	result := c.createMallResult("buyproduct")
	c.Data["json"] = &result
	defer func() {
		result.ErrorMsg = common.GetErrorMsg(result.ErrorCode)
		c.ServeJSON()
	}()

	opcode := mallOpcode{}
	err := json.Unmarshal([]byte(c.GetString("req_data")), &opcode)
	if err != nil {
		result.ErrorCode = common.DATA_ILLEGAL
		return
	}

	// gameServer := models.GetServer("6")
	gameServer := models.GetServer(strconv.Itoa(opcode.Args.ServerID))
	if gameServer == nil {
		result.ErrorCode = common.SERVER_ID_IS_INVALID
		return
	}

	// product := models.GetProduct(opcode.Args.ProductID, "_self_game")
	product := models.GetProduct(opcode.Args.ProductID)
	if product == nil {
		result.ErrorCode = common.CAN_NOT_FIND_PRODUCT
		return
	}

	task := &models.WebTask{
		"buyproduct",
		opcode.Args.CharDBID,
		0,
		product,
	}

	data, err := json.Marshal(task)
	if err != nil {
		logger.E(err.Error())
		result.ErrorCode = common.PRODCUT_TO_JSON_FAILED
		return
	}

	succeed, err := gameServer.WebAddTask(string(data))
	if err != nil || !succeed {
		logger.E(err.Error())
		result.ErrorCode = common.SEND_PRODUCT_FAILED
		return
	}

	result.ErrorCode = common.SUCCEED
	return
}

func (c *MallController) StarscloudBuyProduct() {
	result := "fail"
	defer func() {
		c.Ctx.WriteString(result)
	}()

	order := starscloudOrder{}
	if err := c.ParseForm(&order); err != nil {
		logger.E("参数错误！")
		return
	}

	appid := beego.AppConfig.String("starscloud::appid")
	pmsecret := beego.AppConfig.String("starscloud::pmsecret")
	if appid == "" || pmsecret == "" {
		logger.E("appid和pmsecret不能为空！")
		return
	}

	sign := fmt.Sprintf("amount=%s&channOrderId=%s&channType=%s&pmOrderId=%s&uid=%s&pmAppId=%s&pmSecret=%s",
		order.Amount, order.ChannOrderID, order.ChannType, order.PmOrderID, order.UID, appid, pmsecret)
	sign = common.GetMD5(sign)
	if sign != order.Sign {
		logger.E("订单签名错误！\n%v\n%s", order, sign)
		return
	}

	args := mallArgs{}
	err := json.Unmarshal([]byte(order.ExtraInfo), &args)
	if err != nil {
		logger.E("解析订单额外信息失败!\n%v", order)
		return
	}

	// gameServer := models.GetServer("6")
	gameServer := models.GetServer(strconv.Itoa(args.ServerID))
	if gameServer == nil {
		logger.E("找不到服务器!\n%v", args)
		return
	}

	// product := models.GetProduct(opcode.Args.ProductID, "_self_game")
	product := models.GetProduct(args.ProductID)
	if product == nil {
		logger.E("找不到指定商品!\n%v", args)
		return
	}

	if !gameServer.CharExists(args.CharDBID) {
		logger.E("找不到指定的角色!\n%v", args)
		return
	}

	task := &models.WebTask{
		"buyproduct",
		args.CharDBID,
		0,
		product,
	}

	data, err := json.Marshal(task)
	if err != nil {
		logger.E("生成商品任务信息失败!\n%v", task)
		return
	}

	succeed, err := gameServer.WebAddTask(string(data))
	if err != nil || !succeed {
		logger.E("添加商品任务信息到游戏服失败!\n%v", string(data))
		return
	}

	result = "ok"
}

func (c *MallController) AnySdkPayResult() {
	// c.Ctx.Output.Header("Access-Control-Allow-Origin", "http://dev.anysdk.com")
	returnStr := "fail"
	defer func() {
		logger.I("PayResult: %s", string(c.Ctx.Input.RequestBody))
		c.Ctx.WriteString(returnStr)
	}()

	if !c.checkAnySdkOrder(c.Input()) {
		return
	}

	returnStr = "ok"
	sdkOrder := anysdkOrder{}
	if err := c.ParseForm(&sdkOrder); err != nil {
		logger.E("解析参数错误！\n%s", string(c.Ctx.Input.RequestBody))
		return
	}

	orderID, err := strconv.ParseInt(sdkOrder.PrivateData, 10, 64)
	if err != nil {
		logger.E("无效的订单！\n%v", sdkOrder)
		return
	}

	price, err := strconv.ParseFloat(sdkOrder.Amount, 64)
	if err != nil {
		logger.E("无效的价格！\n%v", sdkOrder)
		return
	}

	if !models.DeliveryProduct(orderID, price, sdkOrder.OrderID, false) {
		logger.E("发送商品失败!\n[order_id=%d]\n%v", orderID, sdkOrder)
		return
	}
}
