package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
)

var Config struct {
	HostAddr     string `env:"SERVER_ADDRESS"`
	ShortcutAddr string `env:"BASE_URL"`
	LogLevel     string `env:"LOG_LVL"`
	JsonDb       string `env:"FILE_STORAGE_PATH"`
}

func InitFlags() {
	flag.StringVar(&Config.HostAddr, "a", "localhost:8080", "set address to run server")
	flag.StringVar(&Config.ShortcutAddr, "b", "http://localhost:8080", "set result shortcut url address")
	flag.StringVar(&Config.LogLevel, "l", "debug", "set log level")
	flag.StringVar(&Config.JsonDb, "f", "/tmp/short-url-db.json", "set file name for json database with URLs")
	flag.Parse()

	err := env.Parse(&Config)
	if err != nil {
		fmt.Println(err)
		return
	}
}
