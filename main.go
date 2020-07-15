package main

import (
	"clock/config"
	"clock/param"
	"clock/server"
	"clock/storage"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

var (
	filePath string // 配置文件路径
	help     bool   // 帮助
)

func usage() {
	fmt.Fprintf(os.Stdout, `clock - simlpe scheduler
Usage: clock [-h help] [-c ./config.yaml]
Options:
`)
	flag.PrintDefaults()
}

func main() {
	flag.StringVar(&filePath, "c", "./config.yaml", "配置文件所在")
	flag.BoolVar(&help, "h", false, "帮助")
	flag.Usage = usage
	flag.Parse()
	if help {
		flag.PrintDefaults()
		return
	}

	// 设置配置文件和静态变量
	config.SetConfig(filePath)
	param.SetStatic()
	storage.SetDb()

	defer storage.Db.Close()
	address := config.Config.GetString("server.host")
	if address == "" {
		logrus.Fatal("can not find any server host config")
	}

	engine, err := server.CreateEngine()
	if err != nil {
		logrus.Fatal(err)
	}

	if e := engine.Start(address); e != nil {
		logrus.Fatal(e)
	}

}
