package api

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/nomo42/url-shortener.git/cmd"
	"github.com/nomo42/url-shortener.git/cmd/config"
	"github.com/nomo42/url-shortener.git/cmd/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

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
	var ok bool
	var buf []byte
	for _, v := range r.Header.Values("Content-Encoding") {
		if strings.Contains(v, "gzip") {
			ok = true
			break
		}
	}
	if ok {
		gzReader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "Fail read gzipped body", http.StatusInternalServerError)
			return
		}
		buf, err = io.ReadAll(gzReader)
		if err != nil {
			http.Error(w, "Fail read gzipped body", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(fmt.Sprintf("%s/%v", config.Config.ShortcutAddr, cmd.ShortenURL(buf))))
		if err != nil {
			return
		}
		return
	}
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Fail read request body", http.StatusInternalServerError)
		return
	}
	//раз прошли проверку заранее пишем статус 201 в хедер
	w.WriteHeader(http.StatusCreated)
	//далее пишем в ответ сокращенный url
	urlHash := cmd.ShortenURL(buf)
	_, err = w.Write([]byte(fmt.Sprintf("%s/%v", config.Config.ShortcutAddr, urlHash)))
	if err != nil {
		logger.Log.Error("fail to write response", zap.String("error", err.Error()))
	}

	storage.WriteValue(urlHash, string(buf))
}

func resolveShortcutHandler(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	urlStorage := storage
	url, ok := urlStorage.ReadValue(hash)
	if !ok {
		// не лучше ли возвращать статус 404?
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
	var ok bool
	var buf []byte
	for _, v := range r.Header.Values("Content-Encoding") {
		if strings.Contains(v, "gzip") {
			ok = true
			break
		}
	}
	if ok {
		gzReader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "Fail read gzipped body", http.StatusInternalServerError)
			return
		}
		buf, err = io.ReadAll(gzReader)
		if err != nil {
			http.Error(w, "Fail read gzipped body", http.StatusInternalServerError)
			return
		}

		if err = json.Unmarshal(buf, givenURL); err != nil {
			logger.Log.Error("Fail unmarshal json", zap.String("body", string(buf)))
			http.Error(w, "Fail unmarshal json", http.StatusInternalServerError)
		}
		byteURL := []byte(givenURL.URL)
		shortURL.Result = fmt.Sprintf("%s/%v", config.Config.ShortcutAddr, cmd.ShortenURL(byteURL))

		buf, err = json.Marshal(shortURL)
		if err != nil {
			http.Error(w, "Fail marshaling result", http.StatusInternalServerError)
			return
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
			logger.Log.Error("fail to write response")
			return
		}
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
		return
	}

	urlHash := cmd.ShortenURL([]byte(givenURL.URL))
	shortURL.Result = fmt.Sprintf("%s/%v", config.Config.ShortcutAddr, urlHash)

	storage.WriteValue(urlHash, givenURL.URL)

	buf, err = json.Marshal(shortURL)
	if err != nil {
		http.Error(w, "Fail marshaling result", http.StatusInternalServerError)
		return
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
		logger.Log.Error("fail to write response")
		return
	}
}
