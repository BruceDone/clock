package main

import (
	"flag"

	"github.com/sirupsen/logrus"

	"clock/config"
	"clock/param"
	"clock/server"
	"clock/storage"
)

var (
	configFilePath string
)

func main() {
	flag.StringVar(&configFilePath, "c", "config.yaml", "config file path")
	flag.Parse()

	config.SetConfig(configFilePath)
	param.SetStatic()
	storage.SetDb()

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
