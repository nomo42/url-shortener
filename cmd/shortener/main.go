package main

import (
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"strings"
)

var URLmap = make(map[string]string)

func main() {
	http.HandleFunc("/", handle)
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		fmt.Printf("Ошибка %v\n", err)
	}
	//shortURL("/api/dev/govno")
	//fmt.Println("Hello")

}

func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		resolveShortcutHandler(w, r)
	case http.MethodPost:
		createShortcutHandler(w, r)
	default:
		http.Error(w, "Invalid request", http.StatusBadRequest)
	}
}

func shortURL(URL []byte) string {
	key := "/" + fmt.Sprintf("%X", crc32.ChecksumIEEE(URL))
	if _, ok := URLmap[key]; ok {
		return key
	}
	URLmap[key] = string(URL)
	return key
}

func createShortcutHandler(w http.ResponseWriter, r *http.Request) {
	//проверяем наличие в поле Content-Type строки tesxt/plain
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
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
	_, err = w.Write([]byte(fmt.Sprintf("http://localhost:8080%v", shortURL(buf))))
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "text/plain")
}

func resolveShortcutHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := URLmap[r.URL.String()]; !ok {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", URLmap[r.URL.String()])
	w.WriteHeader(http.StatusTemporaryRedirect)
}
