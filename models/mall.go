package models

import (
	"moton/logger"

	"github.com/astaxie/beego/orm"
)

type Product struct {
	Id            int     `orm:"pk" json:"id"`
	Name          string  `json:"name"`
	Type          int     `json:"type"`
	Category      int     `json:"category"`
	GameId        string  `json:"-"`
	ServerList    string  `json:"-"`
	BeginTime     int64   `json:"begintime"`
	EndTime       int64   `json:"endtime"`
	Price         float64 `json:"price"`
	DiscountPrice float64 `json:"discount_price"`
	Dscount       string  `json:"discount"`
	Items         string  `json:"items"`
	BonusItems    string  `json:"bonus_items"`
	BtnName       string  `json:"btn_name"`
	UiIcon        string  `json:"ui_icon"`
	Icon          string  `json:"icon"`
	Pos           int     `json:"pos"`
	Times         int     `json:"times"`
	Stock         int     `json:"stock"`
	Proplling     int     `json:"proplling"`
}

type WebTask struct {
	Opcode  string   `json:"opcode"`
	DBID    int64    `json:"link_dbid"`
	OrderID int64    `json:"order_id"`
	Data    *Product `json:"data"`
}

func GetProduct(id interface{}) *Product {
	defer ShowAutoCommit()

	product := &Product{}
	o := orm.NewOrm()
	err := o.Raw("select * from mall where id=? limit 1", id).QueryRow(product)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return nil
	}
	return product
}

func GetProductList(category int) []Product {
	defer ShowAutoCommit()

	var productList []Product
	o := orm.NewOrm()
	_, err := o.Raw("select * from mall where category=?", category).QueryRows(&productList)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return nil
	}
	return productList
}
