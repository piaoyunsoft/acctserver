package controllers

import (
	"fmt"
	"strconv"

	"moton/acctserver/models"
	"moton/logger"

	"github.com/astaxie/beego"
)

type PatchController struct {
	beego.Controller
}

type patchResInfo struct {
	ProductID      string `form:"nProductId"`
	ProgramVersion string `form:"nProgramVer"`
	ResVersion     string `form:"nResVer"`
	SdkID          string `form:"sdkId"`
	ChannelID      string `form:"channelId"`
}

func (p *patchResInfo) Check() bool {
	if len(p.ProductID) == 0 || len(p.ProgramVersion) == 0 || len(p.ResVersion) == 0 || len(p.SdkID) == 0 || len(p.ChannelID) == 0 {
		return false
	}

	return true
}

func (c *PatchController) PatchList() {
	info := patchResInfo{}
	if err := c.ParseForm(&info); err != nil {
		logger.E("Parse patchResInfo failed!")
		return
	}

	if !info.Check() {
		logger.E("patchResInfo invalid! %v", info)
		return
	}

	downloadInfo := models.GetDownloadInfo()
	if downloadInfo == nil {
		logger.E("GetDownloadInfo failed!")
		return
	}

	batchVersions := models.GetBatchResourceVersions(info.ProductID, info.ProgramVersion, info.ResVersion, info.SdkID, info.ChannelID)
	// if batchVersions == nil {
	// 	logger.E("GetBatchResourceVersions failed!")
	// 	return
	// }

	batchVersionsCount := len(batchVersions)
	if batchVersionsCount == 1 {
		info.ResVersion = strconv.Itoa(batchVersions[0].VersionEnd)
	}

	versions := models.GetResourceVersions(info.ProductID, info.ProgramVersion, info.ResVersion, info.SdkID, info.ChannelID)
	if versions == nil {
		logger.E("GetResourceVersions failed!")
		return
	}

	versionsCount := len(versions)
	if batchVersionsCount == 1 {
		versionsCount++
	}

	result := strconv.Itoa(versionsCount)
	if batchVersionsCount == 1 {
		result = fmt.Sprintf("%s|%d|%s", result, batchVersions[0].VersionEnd, batchVersions[0].FileURL)
	}

	for _, v := range versions {
		result = fmt.Sprintf("%s|%d|%s", result, v.VersionEnd, v.FileURL)
	}

	c.Ctx.WriteString(result)
}
