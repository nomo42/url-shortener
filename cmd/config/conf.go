package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
)

var Config struct {
	HostAddr     string `env:"SERVER_ADDRESS"`
	ShortcutAddr string `env:"BASE_URL"`
}

func InitFlags() {
	flag.StringVar(&Config.HostAddr, "a", "localhost:8080", "set address to run server")
	flag.StringVar(&Config.ShortcutAddr, "b", "http://localhost:8080", "set result shortcut url address")
	flag.Parse()

	err := env.Parse(&Config)
	if err != nil {
		fmt.Println(err)
		return
	}
}
