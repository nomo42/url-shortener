package main

import (
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"slices"
)

var URLmap = make(map[string]string)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const keylenth = 6

func main() {
	http.HandleFunc("/", PostHandler)
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		fmt.Printf("Ошибка %v\n", err)
	}
	//ShortURL("/api/dev/govno")
	//fmt.Println("Hello")

}

func ShortURL(URL []byte) string {
	key := "/" + fmt.Sprintf("%X", crc32.ChecksumIEEE(URL))
	if _, ok := URLmap[key]; ok {
		return key
	}
	URLmap[key] = string(URL)
	return key
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	//если спрашивают уже созданный URL передаем управление GeHandler'у
	if r.Method == http.MethodGet && slices.Contains(r.Header["Content-Type"], "text/plain") {
		GetHandler(w, r)
		return
	}

	//проверяем метод и наличие в поле Content-Type строки tesxt/plain
	if r.Method != http.MethodPost || !slices.Contains(r.Header["Content-Type"], "text/plain") {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	//раз прошли проверку заранее пишем статус 201 в хедер
	w.WriteHeader(http.StatusCreated)
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	//далее пишем в ответ сокращенный url
	_, err = w.Write([]byte(fmt.Sprintf("http://localhost:8080%v\n", ShortURL(buf))))
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "text/plain")
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := URLmap[r.URL.String()]; !ok {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", URLmap[r.URL.String()])
	w.WriteHeader(http.StatusTemporaryRedirect)
}
