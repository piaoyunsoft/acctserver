package models

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"moton/acctserver/common"
	"moton/logger"
	"net"
	"net/http"
	"time"

	"github.com/astaxie/beego"
)

type UCData struct {
	SID string `json:"sid"`
}

type UCGame struct {
	GameID int `json:"gameId,int"`
}

type UCRequest struct {
	ID   int64  `json:"id,int"`
	Data UCData `json:"data"`
	Game UCGame `json:"game"`
	Sign string `json:"sign"`
}

type UCResponseState struct {
	Code int    `json:"code,int"`
	Msg  string `json:"msg"`
}

type UCAccount struct {
	AccountID string `json:"accountId"`
	Creator   string `json:"creator"`
	Nickname  string `json:"nickName"`
}

type UCResponse struct {
	ID    int64           `json:"id,int"`
	State UCResponseState `json:"state"`
	Data  UCAccount       `json:"data"`
}

func GetUCAccount(sid string) *UCAccount {
	gameId, err := beego.AppConfig.Int("uc::gameid")
	if err != nil {
		logger.E(err.Error())
		return nil
	}
	ucreq := &UCRequest{
		ID: time.Now().UnixNano() / int64(time.Millisecond),
		Data: UCData{
			sid,
		},
		Game: UCGame{
			gameId,
		},
	}

	ucreq.Sign = common.GetMD5("sid=" + sid + beego.AppConfig.String("uc::apikey"))

	data, err := json.Marshal(ucreq)
	if err != nil {
		logger.E("json err:", err)
	}

	body := bytes.NewBuffer(data)
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, time.Second*10)
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(time.Second * 10))
				return conn, nil
			},
			ResponseHeaderTimeout: time.Second * 10,
		},
	}

	res, err := client.Post(beego.AppConfig.String("uc::url"), "application/json;charset=utf-8", body)
	if err != nil {
		logger.E(err.Error())
		return nil
	}
	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		logger.E(err.Error())
		return nil
	}

	ucres := &UCResponse{}
	json.Unmarshal(result, &ucres)
	return &ucres.Data
}
