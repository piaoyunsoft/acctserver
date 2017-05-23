package models

import (
	"fmt"
	"moton/logger"

	"github.com/astaxie/beego/orm"
)

type GameServer struct {
	Id                int `orm:"pk"`
	Name              string
	Gameid            string
	Isnew             bool
	Isrecmd           bool
	State             int
	Ip                string
	Port              string
	Login             string
	Password          string
	Db                string
	Bindip            string
	Bindport          string
	ClientInformation string
}

type RechargeConfig struct {
	Id                 int `orm:"pk"`
	Gameid             string
	Name               string
	RechargeAmount     int
	ExchangeAmount     int
	FirstRechargeBonus int
}

func init() {
	orm.RegisterModel(new(GameServer))
}

func (m GameServer) TableName() string {
	return "server_list"
}

func (m GameServer) WebLogin(accountId int64, strM, ip string) (int, error) {
	_, err := orm.GetDB(m.Name)
	if err != nil {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&autocommit=true", m.Login, m.Password, m.Ip, m.Port, m.Db)
		logger.I("dsn:%s", dsn)
		orm.RegisterDataBase(m.Name, "mysql", dsn)
	}

	o := orm.NewOrm()
	o.Using(m.Name)
	var newUser int
	err = o.Raw("call web_login(?, ?, ?, @new_user)", accountId, strM, ip).QueryRow(&newUser)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		logger.D("%s", err.Error())
		return 0, err
	}

	return newUser, nil
}

func (m GameServer) WebAddTask(data string) (bool, error) {
	_, err := orm.GetDB(m.Name)
	if err != nil {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&autocommit=true", m.Login, m.Password, m.Ip, m.Port, m.Db)
		orm.RegisterDataBase(m.Name, "mysql", dsn)
	}

	o := orm.NewOrm()
	o.Using(m.Name)

	var hasError int
	err = o.Raw("call web_add_web_task(?)", data).QueryRow(&hasError)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return false, err
	}

	if hasError == 1 {
		return false, nil
	}
	return true, nil
}

func (m GameServer) IsFirstRecharge(accountID int64, amount int) (bool, error) {
	_, err := orm.GetDB(m.Name)
	if err != nil {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&autocommit=true", m.Login, m.Password, m.Ip, m.Port, m.Db)
		orm.RegisterDataBase(m.Name, "mysql", dsn)
	}

	o := orm.NewOrm()
	o.Using(m.Name)
	var outRechargeCount int
	err = o.Raw("call web_do_recharge_rmb(?, ?, @already_num)", accountID, amount).QueryRow(&outRechargeCount)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return false, err
	}

	if outRechargeCount != 0 {
		return false, nil
	}
	return true, nil
}

func (m GameServer) WebRecharge(orderId string, channel string, accountId int64, exchangeAmount int, bonusAmount int) error {
	_, err := orm.GetDB(m.Name)
	if err != nil {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&autocommit=true", m.Login, m.Password, m.Ip, m.Port, m.Db)
		orm.RegisterDataBase(m.Name, "mysql", dsn)
	}

	o := orm.NewOrm()
	o.Using(m.Name)
	var outResult int
	var outError int
	err = o.Raw("call web_recharge(?, ?, ?, ?, ?, @out_result, @out_error)", orderId, channel, accountId, exchangeAmount, bonusAmount).QueryRow(&outResult, &outError)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return err
	}
	return nil
}

func (m GameServer) CharExists(dbid int64) bool {
	_, err := orm.GetDB(m.Name)
	if err != nil {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&autocommit=true", m.Login, m.Password, m.Ip, m.Port, m.Db)
		orm.RegisterDataBase(m.Name, "mysql", dsn)
	}

	o := orm.NewOrm()
	o.Using(m.Name)
	var outCount int
	err = o.Raw("select count(*) from game_chardata where dbid=?", dbid).QueryRow(&outCount)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return false
	}

	if outCount <= 0 {
		return false
	}

	return true
}

func GetServerList(game_id string, ip string, version string) []GameServer {
	var serverList []GameServer
	o := orm.NewOrm()
	count, err := o.Raw("select * from server where `group` in (select pattern.server_group from pattern where pattern.name=?) and allow_ver=?", game_id, version).QueryRows(&serverList)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return nil
	}
	logger.D("Count:%d", count)
	return serverList
}

func GetServer(serverID interface{}) *GameServer {
	gameServer := &GameServer{}
	o := orm.NewOrm()
	err := o.Raw("select * from server where id=? limit 1", serverID).QueryRow(gameServer)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return nil
	}
	return gameServer
}

func GetRechargeCfg(gameId string, amount int) *RechargeConfig {
	rechargeCfg := &RechargeConfig{}
	o := orm.NewOrm()
	err := o.Raw("select * from recharge_list where gameid=? and recharge_amount=?", gameId, amount).QueryRow(rechargeCfg)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return nil
	}
	return rechargeCfg
}
