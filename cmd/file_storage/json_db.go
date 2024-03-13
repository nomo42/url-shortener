package file_storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"os"
	"sync"

	"github.com/nomo42/url-shortener.git/cmd/config"

	"github.com/nomo42/url-shortener.git/cmd/interfaces"

	"github.com/nomo42/url-shortener.git/cmd/logger"
)

// MaxFileDbSize is a max size of a file which will be load to the memory (200 MB default)
const MaxFileDbSize = 1024 * 1024 * 200

type FileStorage struct {
	file   *os.File
	urlMap map[string]string
}

var fileStore *FileStorage
var fMu sync.Mutex

type Result struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var urlCounter int

func Get(fName string) interfaces.Storage {
	fMu.Lock()
	defer fMu.Unlock()
	if fileStore != nil {
		return fileStore
	}

	var fileStore FileStorage
	fileStore.urlMap = make(map[string]string)

	if config.Config.JSONDB == "" {
		return &FileStorage{file: nil}
	}

	records, err := os.OpenFile(fName, os.O_CREATE|os.O_EXCL|os.O_RDWR|os.O_APPEND, 0666)
	if err == nil { // return if file is already exists
		fileStore.file = records
		return &fileStore
	}

	if !errors.Is(err, os.ErrExist) {
		logger.Log.Warn("fail to open db file", zap.String("error", err.Error()))
		return &FileStorage{file: nil}
	}

	records, err = os.OpenFile(fName, os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("fail read urlRecords: %s", err.Error()))
		return &FileStorage{file: nil}
	}

	scanner := bufio.NewScanner(io.LimitReader(records, MaxFileDbSize))
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
		fileStore.urlMap[resultingURLObj.ShortURL] = resultingURLObj.OriginalURL
	}

	fileStore.file = records
	return &fileStore
}

func (f *FileStorage) WriteValue(hash string, originalURL string) {
	if f.ExistenceCheck(hash) {
		return
	}

	f.urlMap[hash] = originalURL
	if f.file == nil {
		logger.Log.Warn("no underlying file configured")
	}

	urlCounter++
	var result Result

	result.UUID = urlCounter
	result.OriginalURL = originalURL
	result.ShortURL = hash
	record, err := json.Marshal(result)
	if err != nil {
		logger.Log.Error("INSANE. fail to marshal record")
		return
	}

	record = append(record, byte('\n'))

	_, err = f.file.Write(record)
	if err != nil {
		logger.Log.Error("fail to write new record: %s", zap.Any("error", err))
	}

	return
}

func (f *FileStorage) ReadValue(key string) (string, bool) {
	value, ok := f.urlMap[key]
	return value, ok
}

func (f *FileStorage) ExistenceCheck(key string) bool {
	if _, ok := f.urlMap[key]; ok {
		return true
	}
	return false
}

func (f *FileStorage) Clear() {
	clear(f.urlMap)
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}
