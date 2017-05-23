package main

import (
	_ "moton/acctserver/routers"
	"moton/helper"
	"moton/logger"
	"path"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	dsnAcct := beego.AppConfig.String("dsn::acct")
	dsnPatch := beego.AppConfig.String("dsn::patch")
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", dsnAcct)
	orm.RegisterDataBase("patch", "mysql", dsnPatch)
	// orm.RegisterDataBase("game", "mysql", "root:root@tcp(127.0.0.1:3306)/gamedb?charset=utf8&loc=Local")
	// orm.RegisterModel(&entity.JobInfo{}, &entity.JobInfoHistory{}, &entity.JobSnapshot{})

}

func main() {
	appdir := helper.GetCurrPath()
	logPath := path.Join(appdir, "log", "acct")
	logger.Instance().Init(logPath, 0, 0)

	beego.SetLogger("file", `{"filename":"log/beego.log"}`)
	beego.BeeLogger.DelLogger("console")

	staticPath := beego.AppConfig.String("common::staticpath")
	beego.SetStaticPath("/"+staticPath, staticPath)
	beego.Run()

	logger.Instance().Close()
}
