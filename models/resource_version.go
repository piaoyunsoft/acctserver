package models

import (
	"moton/logger"

	"time"

	"github.com/astaxie/beego/orm"
)

type ResoureVersion struct {
	ID             int    `orm:"pk,column(Id)"`
	ProductVersion int    `orm:"column(productVersion)"`
	ChannelID      int    `orm:"column(channelId)"`
	Version        int    `orm:"column(version)"`
	VersionEnd     int    `orm:"column(version_end)"`
	ReleaseTime    int    `orm:"column(releaseTime)"`
	AllowIPs       string `orm:"column(allowIps)"`
	FileURL        string `orm:"column(fileURL)"`
	Upload         int    `orm:"column(upload)"`
}

func GetResourceVersions(productID, programVersion, resVersion, sdkID, channelID string) []ResoureVersion {
	o := orm.NewOrm()
	o.Using("patch")

	var record []ResoureVersion
	sql := `select  resource_version.fileURL, resource_version.version, resource_version.upload, resource_version.allowIps from resource_version, product_version
			where resource_version.productVersion = product_version.id and product_version.productId = ? 
			and product_version.version = ? and product_version.sdkId = ? 
			and resource_version.version_end = 0 
			and resource_version.channelId = ? 
			and resource_version.version > ? and resource_version.releaseTime <= ? order by resource_version.version`
	now := time.Now().Unix()
	_, err := o.Raw(sql, productID, programVersion, sdkID, channelID, resVersion, now).QueryRows(&record)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return nil
	}
	return record
}

func GetBatchResourceVersions(productID, programVersion, resVersion, sdkID, channelID string) []ResoureVersion {

	o := orm.NewOrm()
	o.Using("patch")

	var records []ResoureVersion
	sql := `select resource_version.fileURL, resource_version.version, resource_version.version_end, resource_version.upload, resource_version.allowIps from resource_version, product_version
			where resource_version.productVersion = product_version.id and product_version.productId = ? 
			and product_version.version = ? and product_version.sdkId = ? 
			and resource_version.version_end > 0
			and resource_version.channelId = ? 
			and resource_version.version = ? and resource_version.releaseTime <= ? limit 1`
	now := time.Now().Unix()
	_, err := o.Raw(sql, productID, programVersion, sdkID, channelID, resVersion, now).QueryRows(&records)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return nil
	}
	return records
}
