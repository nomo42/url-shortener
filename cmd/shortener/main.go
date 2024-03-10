package main

import (
	"encoding/json"
	"fmt"
	"github.com/nomo42/url-shortener.git/cmd/gzencode"
	"go.uber.org/zap"
	"io"

	"strings"

	"net/http"

	"hash/crc32"

	"github.com/go-chi/chi/v5"

	"github.com/nomo42/url-shortener.git/cmd/config"

	"github.com/nomo42/url-shortener.git/cmd/storage"

	"github.com/nomo42/url-shortener.git/cmd/logger"
)

// Сделать не глобальной эту шляпу
var urlStorage = storage.NewStorage()

func main() {
	config.InitFlags()
	if err := logger.Initialize(config.Config.LogLevel); err != nil {
		fmt.Printf("Ошибка %v\n", err)
	}
	err := http.ListenAndServe(config.Config.HostAddr, logger.LogMware(gzencode.GzipWriteMware(newMuxer())))
	if err != nil {
		fmt.Printf("Ошибка %v\n", err)
	}
}

func shortenURL(URL []byte) string {
	key := fmt.Sprintf("%X", crc32.ChecksumIEEE(URL))
	if ok := urlStorage.ExistenceCheck(key); ok {
		return key
	}
	urlStorage.WriteValue(key, string(URL))
	return key
}

func createShortcutHandler(w http.ResponseWriter, r *http.Request) {
	//encode := r.Header.Values("Content-Type")
	//for _, v := range encode {
	//	if v == "application/json" {
	//		buf, _ := io.ReadAll(r.Body)
	//		logger.Log.Info(fmt.Sprintf("%v", string(buf)))
	//		_, _ = w.Write(buf)
	//		return
	//	}
	//}

	//проверяем наличие в поле Content-Type строки text/plain
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-gzip") {
		http.Error(w, "Invalid request method"+fmt.Sprintf(": %s", r.Header.Get("Content-Type")), http.StatusBadRequest)
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
	url, ok := urlStorage.ReadValue(hash)
	if ok != nil {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func createShortcutJSONHandler(w http.ResponseWriter, r *http.Request) {
	type URL struct {
		URL string `json:"url"`
	}

	type resultURL struct {
		Result string `json:"result"`
	}

	givenURL := &URL{}
	shortURL := &resultURL{}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-gzip") {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Fail read request body", http.StatusInternalServerError)
		return
	}

	if err = json.Unmarshal(buf, givenURL); err != nil {
		logger.Log.Error("Fail unmarshal json", zap.String("body", string(buf)))
		http.Error(w, "Fail unmarshal json", http.StatusInternalServerError)
	}
	byteURL := []byte(givenURL.URL)
	shortURL.Result = fmt.Sprintf("%s/%v", config.Config.ShortcutAddr, shortenURL(byteURL))

	buf, err = json.Marshal(shortURL)
	if err != nil {
		http.Error(w, "Fail marshaling result", http.StatusInternalServerError)
	}
	logger.Log.Info(string(buf))
	for _, v := range r.Header.Values("Accept-Encoding") {
		if strings.Contains(v, "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
		}
	}
	w.WriteHeader(http.StatusCreated)
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
