package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/nomo42/url-shortener.git/cmd/config"
	"hash/crc32"
	"io"
	"net/http"
	"strings"
)

var URLmap = make(map[string]string)

func main() {
	config.InitFlags()
	err := http.ListenAndServe(config.Config.HostAddr, newMuxer())
	if err != nil {
		fmt.Printf("Ошибка %v\n", err)
	}
}

func shortURL(URL []byte) string {
	key := fmt.Sprintf("%X", crc32.ChecksumIEEE(URL))
	if _, ok := URLmap[key]; ok {
		return key
	}
	URLmap[key] = string(URL)
	return key
}

func createShortcutHandler(w http.ResponseWriter, r *http.Request) {
	//проверяем наличие в поле Content-Type строки text/plain
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	//раз прошли проверку заранее пишем статус 201 в хедер
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Fail read request body", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	//далее пишем в ответ сокращенный url
	_, err = w.Write([]byte(fmt.Sprintf("%s%v", config.Config.ShortcutAddr, shortURL(buf))))
	if err != nil {
		return
	}

}

func resolveShortcutHandler(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	url, ok := URLmap[hash]
	if !ok {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
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

func newMuxer() chi.Router {
	mux := chi.NewRouter()
	mux.Get("/{hash}", resolveShortcutHandler)
	mux.Post("/", createShortcutHandler)
	return mux
}
