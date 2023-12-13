package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_createShortcutHandler(t *testing.T) {
	//после выполнения теста очищаем мапу с URL'ами
	defer func() {
		clear(URLmap)
	}()
	type want struct {
		code        int
		response    string
		contentType string
		isMap       bool
	}

	type request struct {
		body        string
		method      string
		contentType string
	}

	tests := []struct {
		name    string
		want    want
		request request
	}{
		{
			name: "wiki test",
			want: want{
				code:        http.StatusCreated,
				response:    "http://localhost:8080/D63CDBB3",
				contentType: "text/plain",
				isMap:       true},
			request: request{body: "https://wikipedia.org", method: http.MethodPost, contentType: "text/plain"},
		},
		{
			name: "google test",
			want: want{code: http.StatusCreated,
				response:    "http://localhost:8080/5B1A2675",
				contentType: "text/plain",
				isMap:       true},
			request: request{body: "https://google.com", method: http.MethodPost, contentType: "text/plain"},
		},
		{
			name: "wrong content-type test",
			want: want{code: http.StatusBadRequest,
				response:    "Invalid request method\n",
				contentType: "",
				isMap:       false},
			request: request{body: "https://dontcare.ru", method: http.MethodPost, contentType: "wrong"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, "/api", strings.NewReader(test.request.body))
			request.Header.Set("Content-Type", test.request.contentType)
			recorder := httptest.NewRecorder()

			createShortcutHandler(recorder, request)
			res := recorder.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
			if res.StatusCode != http.StatusCreated {
				return
			}
			//отрезаем от ответа вида http//localhost:8080/<hash_string> префикс http//localhost:8080 чтобы получить ключ мапы
			key, _ := strings.CutPrefix(test.want.response, "http://localhost:8080/")
			address, ok := URLmap[key]
			//проверяем наличие элемента в мапе
			assert.Equal(t, test.want.isMap, ok)
			//проверяем что значение по этому ключу является нужным нужным адресом
			assert.Equal(t, test.request.body, address)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}

func Test_resolveShortcutHandler(t *testing.T) {
	defer func() {
		clear(URLmap)
	}()
	URLmap["D63CDBB3"] = "https://wikipedia.org"
	URLmap["5B1A2675"] = "https://google.com"
	type want struct {
		code     int
		body     string
		location string
	}

	type request struct {
		url    string
		method string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "google.com test",
			request: request{
				url:    "/5B1A2675",
				method: http.MethodGet,
			},
			want: want{
				code:     http.StatusTemporaryRedirect,
				body:     "",
				location: "https://google.com",
			},
		},
		{
			name: "non-existent element of map test",
			request: request{
				url:    "/5B134275",
				method: http.MethodGet,
			},
			want: want{
				code:     http.StatusBadRequest,
				body:     "Invalid request method\n",
				location: "empty",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)
			recorder := httptest.NewRecorder()
			//resolveShortcutHandler(recorder, request)
			newMuxer().ServeHTTP(recorder, request)
			res := recorder.Result()
			defer res.Body.Close()
			assert.Equal(t, test.want.code, res.StatusCode)
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.body, string(body))
			//если вопрос невалидный, то location не задаётся
			if res.StatusCode != http.StatusTemporaryRedirect {
				return
			}
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
		})
	}
}
