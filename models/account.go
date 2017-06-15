package models

import (
	"moton/acctserver/common"
	"moton/logger"
	"strings"

	"github.com/astaxie/beego/orm"
)

type AccountBase struct {
	AccountId    int64 `orm:"pk"`
	AccountGuid  string
	Username     string
	Password     string
	Salt         string
	RegisterTime int
	RegisterIp   string
	State        int
}

type ChannelAccount struct {
	AccountGuid      string
	UserChannel      string
	ChannelAccountId string
}

func init() {
	orm.RegisterModel(new(AccountBase))
}

func (m *AccountBase) TableName() string {
	return "account_base"
}

//AddAccount 添加账号到数据库
func AddAccount(guid, username, password, salt, registerip string) (int64, int) {
	defer ShowAutoCommit()

	var outAccountID int64
	var outResult int
	o := orm.NewOrm()

	err := o.Raw("call create_new_account(?, ?, ?, ?, ?, @out_account_id, @result)", guid, username, password, salt, registerip).QueryRow(&outAccountID, &outResult)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return -1, -1
	}

	return outAccountID, outResult
}

//GetAccount 找到账号
func GetAccount(username string) *AccountBase {
	defer ShowAutoCommit()
	o := orm.NewOrm()

	acct := &AccountBase{}

	err := o.Raw("select * from account_base where username = ? limit 1", username).QueryRow(acct)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return nil
	}

	return acct
}

//
func BindAccount(oldusername, newusername, password string) bool {
	defer ShowAutoCommit()
	o := orm.NewOrm()

	r, err := o.Raw("update account_base set username = ?, password = ? where account_guid = ?", newusername, password, oldusername).Exec()
	if err != nil {
		logger.E(err.Error())
		return false
	}

	num, err := r.RowsAffected()
	if err != nil {
		logger.E(err.Error())
		return false
	}

	if num != 1 {
		return false
	}

	return true
}

//FindAccountByChannel 通过渠道信息找到账号
func FindAccountByChannel(channel, channelAccountID string) *AccountBase {
	defer ShowAutoCommit()
	o := orm.NewOrm()

	// var count int
	acct := &AccountBase{}
	err := o.Raw("select * from account_base, channel_account_mapping where channel_account_mapping.user_channel = ? and channel_account_mapping.channel_account_id = ? and channel_account_mapping.account_guid = account_base.account_guid limit 1", channel, channelAccountID).QueryRow(acct)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return nil
	}

	return acct
}

//FindAccountByGUID 通过guid查找账号
func FindAccountByGUID(guid string) *AccountBase {
	defer ShowAutoCommit()
	o := orm.NewOrm()

	acct := &AccountBase{}

	err := o.Raw("select * from account_base where account_guid = ? limit 1", guid).QueryRow(acct)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return nil
	}

	return acct
}

func AddAccountByChannel(guid, channel, channelAccountId, username, password, salt, registerip string) int {
	defer ShowAutoCommit()
	o := orm.NewOrm()

	var outAccountID int64
	var result int
	var errno int
	err := o.Raw("call create_new_channel_account(?, ?, ?, ?, ?, ?, ?, @out_account_id, @result)", guid, channel, channelAccountId, username, password, salt, registerip).QueryRow(&outAccountID, &result, &errno)
	if err != nil {
		if err != orm.ErrNoRows {
			logger.E(err.Error())
		}
		return -1
	}

	return result
}

func ChannelLogin(channel, channelAccountId, ip string) (*AccountBase, int) {
	acct := FindAccountByChannel(channel, channelAccountId)
	if acct == nil {
		salt := common.GetRandomString(18)
		pwdMD5 := common.GetMD5(channelAccountId + salt)
		guid := strings.ToUpper(common.GetGUID())
		username := channel + channelAccountId
		outResult := AddAccountByChannel(guid, channel, channelAccountId, username, pwdMD5, salt, ip)
		if outResult == -1 {
			return nil, common.SE_DB_HANDLER_ERROR
		}

		if outResult == 1 {
			return nil, common.USERNAME_ALREADY_EXIST
		}

		if outResult == 2 {
			logger.E("Guid already exist %s", guid)
			return nil, common.USERNAME_ALREADY_EXIST
		}

		if outResult == 3 {
			logger.E("Channel id already exist %s", channelAccountId)
			return nil, common.USERNAME_ALREADY_EXIST
		}
	}

	acct = FindAccountByChannel(channel, channelAccountId)
	if acct == nil {
		return nil, common.NOT_FIND_USERNAME
	}

	return acct, common.SUCCEED
}
