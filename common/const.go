package common

const (
	//orderTimeoutTime 订单超时时间
	OrderTimeoutTime = 60 * 5

	//orderStateUnfinished 订单还未完成
	OrderStateUnfinished = 0
	//orderStateFinished 订单已完成
	OrderStateFinished = 1

	//orderStateFailed 订单已失败
	OrderStateFailed = 10
	//orderStateTimeout 订单超时
	OrderStateTimeout = 11
	//orderStateCancelled 订单已取消
	OrderStateCancelled = 12
)
