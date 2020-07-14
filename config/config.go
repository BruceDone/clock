package config

import (
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Config *viper.Viper

func SetConfig(filePath string) {
	log.Infof("[config] run the env with:%s", filePath)

	Config = viper.New()
	Config.SetConfigFile(filePath)
	if err := Config.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	// set log by default
	setLog()
}

// set log level
func setLog() {
	l := Config.Get("log.level")

	if l == "" {
		log.SetLevel(log.InfoLevel)
	} else if l == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if l == "error" {
		log.SetLevel(log.ErrorLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

}

func relativePath(basedir string, path *string) {
	p := *path
	if len(p) > 0 && p[0] != '/' {
		*path = filepath.Join(basedir, p)
	}
}
