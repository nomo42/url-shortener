package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/nomo42/url-shortener.git/cmd/config"
	"os"
)

type Result struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var urlCounter int

func InitJSONDB(store Storage) error {
	if config.Config.JsonDb == "" {
		return nil
	}

	records, err := os.OpenFile(config.Config.JsonDb, os.O_CREATE|os.O_EXCL|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {

		file, err := os.Open(config.Config.JsonDb)
		if err != nil {
			return fmt.Errorf("fail read urlRecords: %s", err.Error())
		}
		defer func() {
			_ = file.Close()
		}()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			url := scanner.Bytes()
			var resultingURLObj Result
			err = json.Unmarshal(url, &resultingURLObj)
			if err != nil {
				return fmt.Errorf("fail to unmarshal url json: %s", err.Error())
			}
			//is this a necessary? maybe delete
			if urlCounter < resultingURLObj.UUID {
				urlCounter = resultingURLObj.UUID
			}
			store.WriteValue(resultingURLObj.ShortURL, resultingURLObj.OriginalURL)

			//parse results and init urlCounter
		}
		if err != nil {
			return fmt.Errorf("fail to parse json: %s", err.Error())
		}
	}
	defer func() {
		_ = records.Close()
	}()
	return nil
}

func CreateRecord(hash string, originalURL string) error {
	urlRecords, err := os.OpenFile(config.Config.JsonDb, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("fail to open file: %s", err.Error())
	}
	defer func() {
		_ = urlRecords.Close()
	}()
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
		_, err := urlRecords.Write(record)
		if err != nil {
			return fmt.Errorf("fail to write new record: %s", err.Error())
		}
		return nil
	}

	_, err = urlRecords.Write([]byte("\n"))
	if err != nil {
		return err
	}

	_, err = urlRecords.Write(record)
	if err != nil {
		return err
	}

	return nil
}
