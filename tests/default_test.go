package test

import (
	"moton/acctserver/controllers"
	_ "moton/acctserver/routers"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/astaxie/beego"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	_, file, _, _ := runtime.Caller(1)
	apppath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".."+string(filepath.Separator))))
	beego.TestBeegoInit(apppath)
}

// TestMain is a sample to run an endpoint test
// func TestMain(t *testing.T) {
// 	r, _ := http.NewRequest("GET", "/", nil)
// 	w := httptest.NewRecorder()
// 	beego.BeeApp.Handlers.ServeHTTP(w, r)

// 	beego.Trace("testing", "TestMain", "Code[%d]\n%s", w.Code, w.Body.String())

// 	Convey("Subject: Test Station Endpoint\n", t, func() {
// 		Convey("Status Code Should Be 200", func() {
// 			So(w.Code, ShouldEqual, 200)
// 		})
// 		Convey("The Result Should Not Be Empty", func() {
// 			So(w.Body.Len(), ShouldBeGreaterThan, 0)
// 		})
// 	})
// }

func TestAccount(t *testing.T) {
	v := url.Values{}
	v.Set("req_data", "{\"opcode\":\"login\", \"arg\":{\"username\":\"d02\", \"password\":\"c4ca4238a0b923820dcc509a6f75849b\", \"channel\":\"default_self\", \"game_id\":\"_self_game\"}}")
	r, _ := http.NewRequest("POST", "/account/login", strings.NewReader(v.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.Add(beego.AppConfig.String("route::Login"), &controllers.AccountController{}, "post:Login")
	beego.BeeApp.Handlers.Add(beego.AppConfig.String("route::ChannelLogin"), &controllers.AccountController{}, "post:ChannelLogin")
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestMain", "Code[%d]\n%s", w.Code, w.Body.String())
	Convey("Subject: Test Station Endpoint\n", t, func() {
		Convey("Status Code Should Be 200", func() {
			So(w.Code, ShouldEqual, 200)
		})
		Convey("The Result Should Not Be Empty", func() {
			So(w.Body.Len(), ShouldBeGreaterThan, 0)
		})
	})
}
