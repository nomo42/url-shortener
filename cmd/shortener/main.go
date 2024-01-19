package main

import (
	"encoding/json"
	"fmt"
	"io"

	"strings"

	"net/http"

	"hash/crc32"

	"github.com/go-chi/chi/v5"

	"github.com/nomo42/url-shortener.git/cmd/config"

	"github.com/nomo42/url-shortener.git/cmd/storage"

	"github.com/nomo42/url-shortener.git/cmd/logger"
)

var URLStorage = storage.NewStorage()

func main() {
	config.InitFlags()
	if err := logger.Initialize(config.Config.LogLevel); err != nil {
		fmt.Printf("Ошибка %v\n", err)
	}
	err := http.ListenAndServe(config.Config.HostAddr, logger.RequestResponseLogger(newMuxer()))
	if err != nil {
		fmt.Printf("Ошибка %v\n", err)
	}
}

func shortenURL(URL []byte) string {
	key := fmt.Sprintf("%X", crc32.ChecksumIEEE(URL))
	if ok := URLStorage.ExistenceCheck(key); ok {
		return key
	}
	URLStorage.WriteValue(key, string(URL))
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
	_, err = w.Write([]byte(fmt.Sprintf("%s/%v", config.Config.ShortcutAddr, shortenURL(buf))))
	if err != nil {
		return
	}

}

func resolveShortcutHandler(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	url, ok := URLStorage.ReadValue(hash)
	if ok != nil {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func createShortcutJSONHandler(w http.ResponseWriter, r *http.Request) {
	type URL struct {
		Url string `json:"url"`
	}

	type shortURL struct {
		Result string `json:"result"`
	}

	givenUrl := &URL{}
	shortUrl := &shortURL{}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Fail read request body", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err = json.Unmarshal(buf, givenUrl); err != nil {
		http.Error(w, "Fail unmarshal json", http.StatusInternalServerError)
	}
	byteUrl := []byte(givenUrl.Url)
	shortUrl.Result = fmt.Sprintf("%s/%v", config.Config.ShortcutAddr, shortenURL(byteUrl))

	buf, err = json.Marshal(shortUrl)
	if err != nil {
		http.Error(w, "Fail marshaling result", http.StatusInternalServerError)
	}

	_, err = w.Write(buf)
	if err != nil {
		return
	}
}

func newMuxer() chi.Router {
	mux := chi.NewRouter()
	mux.Get("/{hash}", resolveShortcutHandler)
	mux.Post("/", createShortcutHandler)
	mux.Post("/api/shorten", createShortcutJSONHandler)
	return mux
}
