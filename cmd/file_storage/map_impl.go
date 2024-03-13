package file_storage

import (
	"fmt"
)

type mapStorage0 map[string]string

func (m mapStorage0) WriteValue(key, value string) {
	m[key] = value
}

func (m mapStorage0) ReadValue(key string) (string, error) {
	value, ok := m[key]
	if !ok {
		return "", fmt.Errorf("no value")
	}
	return value, nil
}

func (m mapStorage0) ExistenceCheck(key string) bool {
	if _, ok := m[key]; ok {
		return true
	}
	return false
}

func (m mapStorage0) Clear() {
	clear(m)
}

var urlMap mapStorage0 = make(map[string]string)

// Пока других реализаций хранения URL нету, так что NewStorage возвращает именно мапу. Далее с помощью флагов буду
// определять какая реализация нужна и NewStorage будет определён в файле mem_storage_interfaces.go
//func NewStorage() interfaces.Storage {
//
//	return urlMap
//}
