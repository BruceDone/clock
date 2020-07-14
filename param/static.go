package param

import (
	"github.com/sirupsen/logrus"

	"clock/config"
)

var (
	// 数据库相关
	DBConn = ""

	// 应用相关
	WebUser = ""
	WebPwd  = ""
	WebJwt  = ""
)

// 载入静态信息
func SetStatic() {
	if tmp := config.Config.GetString("login.user"); tmp == "" {
		logrus.Fatal("empty login.user")
	} else {
		logrus.Println("[param] load user")
		WebUser = tmp
	}

	if tmp := config.Config.GetString("login.pwd"); tmp == "" {
		logrus.Fatal("empty login.user")
	} else {
		WebPwd = tmp
	}

	if tmp := config.Config.GetString("login.jwt"); tmp == "" {
		logrus.Fatal("empty login.jwt")
	} else {
		WebJwt = tmp
	}

	if tmp := config.Config.GetString("storage.conn"); tmp == "" {
		logrus.Fatal("empty login.user")
	} else {
		logrus.Println("[param] load DbConn")
		DBConn = tmp
	}

}
