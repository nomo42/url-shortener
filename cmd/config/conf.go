package config

import (
	"flag"
)

var Config struct {
	HostAddr     string
	ShortcutAddr string
}

func InitFlags() {
	flag.StringVar(&Config.HostAddr, "a", "localhost:8080", "set address to run server")
	flag.StringVar(&Config.ShortcutAddr, "b", "http://localhost:8080/", "set result shortcut url address")
	flag.Parse()
}
