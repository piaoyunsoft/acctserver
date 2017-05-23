package models

import (
	"moton/logger"

	"github.com/astaxie/beego/orm"
)

type DownloadInfo struct {
	Id   int    `orm:"pk,column(id)"`
	Path string `orm:"column(path)"`
}

func GetDownloadInfo() *DownloadInfo {
	o := orm.NewOrm()
	o.Using("patch")

	record := &DownloadInfo{}
	sql := `select * from downloadinfo limit 1`
	err := o.Raw(sql).QueryRow(&record)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return nil
	}
	return record
}
