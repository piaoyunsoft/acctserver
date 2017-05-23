package common

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/astaxie/beego/httplib"
	"github.com/pborman/uuid"
)

var incID = 0

//GetMD5 生成32位MD5
func GetMD5(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

//GetRandomString 生成随机字符串
func GetRandomString(length int) string {
	bytes := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

//RandInRange 指定范围生成随机数
func RandInRange(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return min + r.Intn(max-min)
}

//GetGUID 生成GUID
func GetGUID() string {
	guid := uuid.NewUUID()
	return guid.String()
}

//DoZlibCompress 对数据进行Zlib压缩
func DoZlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(src)
	w.Close()
	return in.Bytes()
}

func GetID() string {
	now := time.Now().Unix()
	incID++
	return fmt.Sprintf("%d%08d", now, incID)
}

func HttpPostJson(url string, json interface{}, timeout time.Duration) ([]byte, error) {
	req := httplib.Post(url)
	req.SetTimeout(timeout, timeout)
	req.JSONBody(json)
	res, err := req.Response()
	if err != nil {
		// logger.E("请求IAP验证失败\n%s\n%s", err.Error(), receiptData)
		return nil, err
	}
	result, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	return result, nil
}
