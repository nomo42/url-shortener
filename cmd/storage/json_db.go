package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/nomo42/url-shortener.git/cmd/config"

	"github.com/nomo42/url-shortener.git/cmd/interfaces"

	"github.com/nomo42/url-shortener.git/cmd/logger"
)

type FileStorage struct {
	file *os.File
}

var fileStore *FileStorage
var fMu sync.Mutex

type Result struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var urlCounter int

func GetFileStorage(store interfaces.Storage) interfaces.FileStorage {
	fMu.Lock()
	defer fMu.Unlock()
	if fileStore != nil {
		return fileStore
	}

	if config.Config.JSONDB == "" {
		return &FileStorage{file: nil}
	}

	records, err := os.OpenFile(config.Config.JSONDB, os.O_CREATE|os.O_EXCL|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {

		file, err := os.Open(config.Config.JSONDB)
		file, err = os.OpenFile(config.Config.JSONDB, os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			logger.Log.Warn(fmt.Sprintf("fail read urlRecords: %s", err.Error()))
			return &FileStorage{file: nil}
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			url := scanner.Bytes()
			var resultingURLObj Result
			err = json.Unmarshal(url, &resultingURLObj)
			if err != nil {
				logger.Log.Warn(fmt.Sprintf("fail to unmarshal url json: %s", err.Error()))
			}
			if urlCounter < resultingURLObj.UUID {
				urlCounter = resultingURLObj.UUID
			}
			store.WriteValue(resultingURLObj.ShortURL, resultingURLObj.OriginalURL)

		}
		records = file
	}
	var file FileStorage
	file.file = records
	fileStore = &file
	return fileStore
}

func (f *FileStorage) CreateRecord(hash string, originalURL string) error {
	if config.Config.JSONDB == "" {
		return fmt.Errorf("no file")
	}

	urlCounter++
	var result Result

	result.UUID = urlCounter
	result.OriginalURL = originalURL
	result.ShortURL = hash
	record, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if urlCounter == 1 {
		_, err := f.file.Write(record)
		if err != nil {
			return fmt.Errorf("fail to write new record: %s", err.Error())
		}
		return nil
	}

	_, err = f.file.Write([]byte("\n"))
	if err != nil {
		return err
	}

	_, err = f.file.Write(record)
	if err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}
