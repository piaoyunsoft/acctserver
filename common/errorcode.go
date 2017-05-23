package common

const (
	SYSTEM_ERROR                     = -1 //系统错误
	SYSTEM_GET_PLATFORM_DB_ERROR     = -2 //系统错误，获取平台数据库失败
	SE_DB_HANDLER_ERROR              = -3 //系统错误，数据库操作失败
	SE_LOAD_GAME_SERVER_LIST_ERROR   = -4 //系统错误，加载服务器列表失败
	SE_CONNECT_GAME_SERVER_DB_ERROR  = -5 //系统错误，连接游戏服务器数据库失败
	SE_CONNECT_GAME_GIFT_DB_ERROR    = -6 //系统错误，连接游戏礼包数据库失败
	SE_LOAD_GAME_RECHARGE_LIST_ERROR = -7 //系统错误，加载游戏充值列表失败
	SE_GET_CACHE_ERROR               = -8 //系统错误，获取缓存失败

	SUCCEED              = 0 //成功
	PARAM_LESS           = 1 //缺少参数
	DATA_ILLEGAL         = 2 //数据非法
	INVALID_REQUEST_TYPE = 3 //无效的请求类型
	REQUEST_PARAM_LESS   = 4 //请求类型的参数不足

	NEED_LOGIN_FIRST                        = 101 //需要先登录
	ALREADY_LOGIN                           = 102 //已登录
	ALREADY_LOGIN_NOT_REGISTER              = 103 //已登录，不能注册
	USERNAME_ALREADY_EXIST                  = 104 //用户名已存在
	USERNAME_IS_INVALID                     = 105 //无效的用户名
	PASSWORD_IS_INVALID                     = 106 //无效的密码
	CHANNEL_INVALID                         = 107 //无效的渠道
	UNKNOW_GAME_ID                          = 108 //未知的游戏
	NOT_FIND_USERNAME                       = 109 //用户不存在
	USERNAME_OR_PASSWORD_ERROR              = 110 //用户名或密码错误
	PASSWORD_NEW_AND_OLD_CANNOT_BE_SAME     = 111 //新密码和老密码不能相同
	USERNAME_NEW_AND_OLD_CANNOT_BE_SAME     = 112 //新用户名和老用户名不能相同
	SERVER_ID_IS_INVALID                    = 113 //无效的服务器id
	GIFT_CODE_IS_INVALID                    = 114 //无效的礼包码
	GIFT_CODE_ALREADY_USED                  = 115 //礼包码已被使用过
	GIFT_CODE_TYPE_CANT_NOT_GET             = 116 //不能重复领取该类型礼包
	THE_GAME_DO_NOT_USE_USER_RECHARGE       = 117 //当前游戏不能使用用户充值，请进行付费充值
	USER_RECHARGE_RMB_VALUE_INVALID         = 118 //用户充值时，无效的金额
	USER_RECHARGE_UNKNOW_ERROR              = 119 //用户充值时，未知的错误
	USER_NOT_IS_GM                          = 120 //不是gm账号
	GM_USER_LOGIN_IP_NOT_ALLOW              = 121 //gm登录时，ip不被允许
	NOT_FIND_ACCOUNTID_BY_CHAR_DBID         = 122 //找不到角色dbid对应的账号id
	PAY_ORDERID_ALREADY_EXIST               = 123 //支付时，订单号已存在
	PAY_UNKNOW_ERROR                        = 124 //支付时，未知的错误
	PAY_RMB_VALUE_INVALID                   = 125 //支付时，无效的金额(找不到金额对应的充值配置)
	PAY_PRIVATE_DATA_ERROR                  = 126 //支付方传递的游戏私有信息错误
	PAY_PRIVATE_DATA_DECODE_FROM_JSON_ERROR = 127 //支付方传递的私有信息从json解析为对象失败
	PAY_APPID_NOT_EQ_GAME_ID_FOR_APPID      = 128 //支付方传递的appid和游戏标识所查到的appid不相等
	PAY_IP_NOT_IN_WHITE_LIST                = 129 //支付方当前ip不在白名单中
	PAY_SIGN_CHECK_FAILED                   = 130 //支付方支付签名验证失败
	CAPTCHA_NOT_EXPIRED                     = 131 //验证码还没过期
	CAPTCHA_INVALID                         = 132 //验证码无效
	PHONE_IS_INVALID                        = 133 //手机号无效
	PHONE_USERNAME_ALREADY_EXIST            = 134 //手机号已被注册
	TOURISTS_BIND_FAILED                    = 135 //绑定账号失败
	IS_NOT_TOURIST                          = 136 //非游客账号不能绑定
	CAN_NOT_FIND_PRODUCT                    = 137 //无法找到商品
	PRODCUT_TO_JSON_FAILED                  = 138 //商品数据转换失败
	SEND_PRODUCT_FAILED                     = 139 //发送商品到游戏服失败
	LOGIN_SIGN_CHECK_FAILED                 = 140 //登录签名验证失败
	NEED_CHANNEL_LOGIN                      = 141 //未通过渠道登录
	HAVE_NO_PRODUCTS_IN_STOCK               = 142 //已达到购买上限
	CREATE_ORDER_ID_FAILED                  = 143 //生成订单号失败
	INSERT_ORDER_FAILED                     = 144 //创建订单失败
	ORDER_NOT_FINISHED                      = 145 //有未完成的订单
)

var errorMsgMap map[int]string

func init() {
	errorMsgMap = make(map[int]string)
	errorMsgMap[SYSTEM_ERROR] = "系统错误"
	errorMsgMap[SYSTEM_GET_PLATFORM_DB_ERROR] = "系统错误，获取平台数据库失败"
	errorMsgMap[SE_DB_HANDLER_ERROR] = "系统错误，数据库操作失败"
	errorMsgMap[SE_LOAD_GAME_SERVER_LIST_ERROR] = "系统错误，加载服务器列表失败"
	errorMsgMap[SE_CONNECT_GAME_SERVER_DB_ERROR] = "系统错误，连接游戏服务器数据库失败"
	errorMsgMap[SE_CONNECT_GAME_GIFT_DB_ERROR] = "系统错误，连接游戏礼包数据库失败"
	errorMsgMap[SE_LOAD_GAME_RECHARGE_LIST_ERROR] = "系统错误，加载游戏充值列表失败"
	errorMsgMap[SE_GET_CACHE_ERROR] = "系统错误，获取缓存失败"

	errorMsgMap[SUCCEED] = "成功"
	errorMsgMap[PARAM_LESS] = "缺少参数"
	errorMsgMap[DATA_ILLEGAL] = "数据非法"
	errorMsgMap[INVALID_REQUEST_TYPE] = "无效的请求类型"
	errorMsgMap[REQUEST_PARAM_LESS] = "请求类型的参数不足"

	errorMsgMap[NEED_LOGIN_FIRST] = "需要先登录"
	errorMsgMap[ALREADY_LOGIN] = "已登录"
	errorMsgMap[ALREADY_LOGIN_NOT_REGISTER] = "已登录，不能注册"
	errorMsgMap[USERNAME_ALREADY_EXIST] = "用户名已存在"
	errorMsgMap[USERNAME_IS_INVALID] = "无效的用户名"
	errorMsgMap[PASSWORD_IS_INVALID] = "无效的密码"
	errorMsgMap[CHANNEL_INVALID] = "无效的渠道"
	errorMsgMap[UNKNOW_GAME_ID] = "未知的游戏"
	errorMsgMap[NOT_FIND_USERNAME] = "用户不存在"
	errorMsgMap[USERNAME_OR_PASSWORD_ERROR] = "用户名或密码错误"
	errorMsgMap[PASSWORD_NEW_AND_OLD_CANNOT_BE_SAME] = "新密码和老密码不能相同"
	errorMsgMap[USERNAME_NEW_AND_OLD_CANNOT_BE_SAME] = "新用户名和老用户名不能相同"
	errorMsgMap[SERVER_ID_IS_INVALID] = "无效的服务器id"
	errorMsgMap[GIFT_CODE_IS_INVALID] = "无效的礼包码"
	errorMsgMap[GIFT_CODE_ALREADY_USED] = "礼包码已被使用过"
	errorMsgMap[GIFT_CODE_TYPE_CANT_NOT_GET] = "不能重复领取该类型礼包"
	errorMsgMap[THE_GAME_DO_NOT_USE_USER_RECHARGE] = "当前游戏不能使用用户充值，请进行付费充值"
	errorMsgMap[USER_RECHARGE_RMB_VALUE_INVALID] = "用户充值时，无效的金额"
	errorMsgMap[USER_RECHARGE_UNKNOW_ERROR] = "用户充值时，未知的错误"
	errorMsgMap[USER_NOT_IS_GM] = "不是gm账号"
	errorMsgMap[GM_USER_LOGIN_IP_NOT_ALLOW] = "gm登录时，ip不被允许"
	errorMsgMap[NOT_FIND_ACCOUNTID_BY_CHAR_DBID] = "找不到角色dbid对应的账号id"
	errorMsgMap[PAY_ORDERID_ALREADY_EXIST] = "支付时，订单号已存在"
	errorMsgMap[PAY_UNKNOW_ERROR] = "支付时，未知的错误"
	errorMsgMap[PAY_RMB_VALUE_INVALID] = "支付时，无效的金额(找不到金额对应的充值配置)"
	errorMsgMap[PAY_PRIVATE_DATA_ERROR] = "支付方传递的游戏私有信息错误"
	errorMsgMap[PAY_PRIVATE_DATA_DECODE_FROM_JSON_ERROR] = "支付方传递的私有信息从json解析为对象失败"
	errorMsgMap[PAY_APPID_NOT_EQ_GAME_ID_FOR_APPID] = "支付方传递的appid和游戏标识所查到的appid不相等"
	errorMsgMap[PAY_IP_NOT_IN_WHITE_LIST] = "支付方当前ip不在白名单中"
	errorMsgMap[PAY_SIGN_CHECK_FAILED] = "支付方支付签名验证失败"
	errorMsgMap[CAPTCHA_NOT_EXPIRED] = "验证码过期前不能重复发送"
	errorMsgMap[CAPTCHA_INVALID] = "验证码无效"
	errorMsgMap[PHONE_IS_INVALID] = "手机号无效"
	errorMsgMap[PHONE_USERNAME_ALREADY_EXIST] = "手机号已被注册"
	errorMsgMap[TOURISTS_BIND_FAILED] = "绑定账号失败"
	errorMsgMap[IS_NOT_TOURIST] = "非游客账号不能绑定"
	errorMsgMap[CAN_NOT_FIND_PRODUCT] = "无法找到商品"
	errorMsgMap[PRODCUT_TO_JSON_FAILED] = "商品数据转换失败"
	errorMsgMap[SEND_PRODUCT_FAILED] = "发送商品到游戏服失败"
	errorMsgMap[LOGIN_SIGN_CHECK_FAILED] = "登录签名验证失败"
	errorMsgMap[NEED_CHANNEL_LOGIN] = "未通过渠道登录"
	errorMsgMap[HAVE_NO_PRODUCTS_IN_STOCK] = "已达到购买上限"
	errorMsgMap[CREATE_ORDER_ID_FAILED] = "生成订单号失败"
	errorMsgMap[INSERT_ORDER_FAILED] = "创建订单失败"
	errorMsgMap[ORDER_NOT_FINISHED] = "有未完成的订单"
}

//GetErrorMsg 根据错误值返回错误消息
func GetErrorMsg(code int) string {
	msg, ok := errorMsgMap[code]
	if ok {
		return msg
	}
	return ""
}
