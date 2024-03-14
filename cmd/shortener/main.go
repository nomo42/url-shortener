package main

import (
	"fmt"
	"github.com/nomo42/url-shortener.git/cmd/api"
	"github.com/nomo42/url-shortener.git/cmd/config"

	"github.com/nomo42/url-shortener.git/cmd/fileStorage"

	"github.com/nomo42/url-shortener.git/cmd/logger"
)

// Сделать не глобальной эту шляпу

func main() {
	config.InitFlags()

	if err := logger.Initialize(config.Config.LogLevel); err != nil {
		fmt.Printf("Ошибка %v\n", err)
	}

	fileStore := fileStorage.Get(config.Config.JSONDB)
	defer fileStore.Close()

	api.LaunchServer(fileStore)
}
