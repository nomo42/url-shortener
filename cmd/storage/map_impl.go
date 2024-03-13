package storage

import (
	"fmt"
	"github.com/nomo42/url-shortener.git/cmd/interfaces"
)

type mapStorage map[string]string

func (m mapStorage) WriteValue(key, value string) {
	m[key] = value
}

func (m mapStorage) ReadValue(key string) (string, error) {
	value, ok := m[key]
	if !ok {
		return "", fmt.Errorf("no value")
	}
	return value, nil
}

func (m mapStorage) ExistenceCheck(key string) bool {
	if _, ok := m[key]; ok {
		return true
	}
	return false
}

func (m mapStorage) Clear() {
	clear(m)
}

var urlMap mapStorage = make(map[string]string)

// Пока других реализаций хранения URL нету, так что NewStorage возвращает именно мапу. Далее с помощью флагов буду
// определять какая реализация нужна и NewStorage будет определён в файле mem_storage_interfaces.go
func NewStorage() interfaces.Storage {
	return urlMap
}
