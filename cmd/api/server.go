package api

import (
	"fmt"
	"github.com/nomo42/url-shortener.git/cmd/api/gzencode"
	"github.com/nomo42/url-shortener.git/cmd/config"
	"github.com/nomo42/url-shortener.git/cmd/interfaces"
	"github.com/nomo42/url-shortener.git/cmd/logger"
	"net/http"
)

var storage interfaces.Storage

func LaunchServer(_storage interfaces.Storage) {
	storage = _storage

	err := http.ListenAndServe(config.Config.HostAddr, logger.LogMware(gzencode.GzipWriteMware(newMuxer())))
	if err != nil {
		fmt.Printf("Ошибка %v\n", err)
	}
}
