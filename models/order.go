package models

import (
	"encoding/json"
	"moton/acctserver/common"
	"moton/logger"
	"time"

	"github.com/astaxie/beego/orm"
)

type Order struct {
	OrderID        int64   `orm:"column(order_id)"`
	AccountID      int64   `orm:"column(account_id)"`
	CharDBID       int64   `orm:"column(char_dbid)"`
	ServerID       int     `orm:"column(server_id)"`
	ProductID      int     `orm:"column(product_id)"`
	Price          float64 `orm:"column(price)"`
	ChannelOrderID string  `orm:"column(channel_order_id)"`
	Channel        string  `orm:"column(channel)"`
	OrderTime      int64   `orm:"column(order_time)"`
	DeliveryTime   int64   `orm:"column(delivery_time)"`
	State          int     `orm:"column(state)"`
}

func ShowAutoCommit() {
	var key = ""
	var value = ""
	if err := orm.NewOrm().Raw("show variables like '%autocommit%'").QueryRow(&key, &value); err != nil {
		logger.E(err.Error())
	} else {
		logger.I("%s=%s", key, value)
	}
}

// func FindMallOrder(int serverId) {
// 	gameServer := GetServerById(serverId)
// 	o := orm.NewOrm()
// 	err := o.Raw("select * from server where id=? limit 1", serverID).QueryRow(gameServer)
// 	if err != nil {
// 		if err != orm.ErrNoRows {
// 			logger.E(err.Error())
// 		}
// 		return nil
// 	}
// 	return gameServer
// }

func InsertOrder(order *Order) error {
	o := orm.NewOrm()
	_, err := o.Raw("insert into mall_order values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		order.OrderID, order.AccountID, order.CharDBID, order.ServerID, order.ProductID, order.Price, order.ChannelOrderID, order.Channel, order.OrderTime, order.DeliveryTime, order.State).Exec()
	if err != nil {
		return err
	}
	return nil
}

func GetOrder(orderID int64) *Order {
	order := &Order{}
	o := orm.NewOrm()
	err := o.Raw("select * from mall_order where order_id=? limit 1", orderID).QueryRow(order)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return nil
	}
	return order
}

func GetOrderUnfinished(serverID int, charDBID int64) (*Order, error) {
	order := &Order{}
	o := orm.NewOrm()
	err := o.Raw("select * from mall_order where server_id=? and char_dbid=? and state=0", serverID, charDBID).QueryRow(order)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
			return nil, err
		}
		return nil, nil
	}
	return order, nil
}

func GetOrderFinishedCount(productID int, serverID int, chardDBID int64) (int, error) {
	var count int
	o := orm.NewOrm()
	err := o.Raw("select count(order_id) from mall_order where product_id=? and server_id=? and char_dbid=? and state=1", productID, serverID, chardDBID).QueryRow(&count)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return 0, err
	}
	return count, nil
}

func UpdateOrder(order *Order) error {
	o := orm.NewOrm()
	_, err := o.Raw("update mall_order set channel_order_id=?, delivery_time=?, state=? where order_id=?", order.ChannelOrderID, order.DeliveryTime, order.State, order.OrderID).Exec()
	if err != nil {
		return err
	}
	return nil
}

func CancelOrder(orderID int64) bool {
	order := GetOrder(orderID)
	if order == nil {
		logger.E("找不到订单！")
		return false
	}

	if order.State != common.OrderStateUnfinished {
		return false
	}
	order.DeliveryTime = time.Now().Unix()
	order.State = common.OrderStateCancelled
	err := UpdateOrder(order)
	if err != nil {
		logger.E(err.Error())
	}
	return true
}

func DeliveryProduct(orderID int64, price float64, channelOrderID string, checkPrice bool) bool {
	defer ShowAutoCommit()

	order := GetOrder(orderID)
	if order == nil {
		logger.E("找不到订单！")
		return false
	}

	if order.State != common.OrderStateUnfinished {
		logger.E("该订单已完成无法操作!\n")
		return false
	}

	if checkPrice && price != order.Price {
		logger.E("该商品需要%f元，玩家只支付了%f元\n", order.Price, price)
		return false
	}

	// logger.D("\n%v\n", order)
	order.ChannelOrderID = channelOrderID
	order.DeliveryTime = time.Now().Unix()
	order.State = common.OrderStateFailed
	defer func() {
		err := UpdateOrder(order)
		if err != nil {
			logger.E(err.Error())
		}
	}()

	product := GetProduct(order.ProductID)
	if product == nil {
		logger.E("找不到指定商品!\n%v", order)
		return false
	}

	gameServer := GetServer(order.ServerID)
	if gameServer == nil {
		logger.E("找不到服务器!\n%v", order)
		return false
	}

	if !gameServer.CharExists(order.CharDBID) {
		logger.E("找不到指定的角色!\n%v", order)
		return false
	}

	task := &WebTask{
		"buyproduct",
		order.CharDBID,
		order.OrderID,
		product,
	}

	data, err := json.Marshal(task)
	if err != nil {
		logger.E("生成商品任务信息失败!\n%v", task)
		return false
	}

	succeed, err := gameServer.WebAddTask(string(data))
	if err != nil || !succeed {
		logger.E("添加商品任务信息到游戏服失败!\n%v", string(data))
		return false
	}

	order.State = common.OrderStateFinished
	return true
}
