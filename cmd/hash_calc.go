package cmd

import (
	"fmt"
	"hash/crc32"
)

func ShortenURL(URL []byte) string {
	key := fmt.Sprintf("%X", crc32.ChecksumIEEE(URL))
	//urlStorage := fileStorage.NewStorage()
	//if ok := urlStorage.ExistenceCheck(key); ok {
	//	return key
	//}
	//logger.Log.Info(string(URL))
	//urlStorage.WriteValue(key, string(URL))
	//fileStore := fileStorage.Get(config.Config.JSONDB)
	//err := fileStore.CreateRecord(key, string(URL))
	//if err != nil && config.Config.JSONDB != "" {
	//	logger.Log.Warn(fmt.Sprintf("fail to record hash:%s, url:%s. Error: %s", key, string(URL), err.Error()))
	//}
	return key
}
