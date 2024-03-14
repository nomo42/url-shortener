package api

import (
	"github.com/nomo42/url-shortener.git/cmd/config"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"testing"

	"github.com/nomo42/url-shortener.git/cmd/filestorage"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

//func TestStuff(t *testing.T) {
//	//config.InitFlags()
//	//f, err := os.OpenFile(config.Config.JSONDB, os.O_APPEND|os.O_RDWR, 0666)
//	//require.NoError(t, err, "must open file")
//	//var r1, r2 filestorage.Result
//	//r1.UUID = 1
//	//r1.ShortURL = "EAC67D10"
//	//r1.OriginalURL = "kolivan.org"
//	//r2.UUID = 2
//	//r2.ShortURL = "EBCF7D12"
//	//r2.OriginalURL = "govno.com"
//	//p, err := json.Marshal(r1)
//	//require.NoError(t, err, "must marshal")
//	//_, err = f.Write(p)
//	//_, err = f.Write([]byte("\n"))
//	//require.NoError(t, err, "must write file")
//	//p, err = json.Marshal(r2)
//	//require.NoError(t, err, "must marshal")
//	//_, err = f.Write(p)
//	//require.NoError(t, err, "must write file")
//	f, err := os.Open(config.Config.JSONDB)
//	require.NoError(t, err, "must open file")
//	scanner := bufio.NewScanner(f)
//	for scanner.Scan() {
//		url := scanner.Bytes()
//		logger.Log.Info(fmt.Sprintf("%s", url))
//		//require.NoError(t, err, fmt.Sprintf("must marshal: %s", url))
//		//require.JSONEq(t, `{"uuid":1,"short_url":"EAC67D10","original_url":"kolivan.org"}`, string(bytesUrl))
//		var resultingUrlObj filestorage.Result
//		err := json.Unmarshal(url, &resultingUrlObj)
//		require.NoErrorf(t, err, "must unmarshal")
//		t.Log(resultingUrlObj)
//	}
//
//}

func Test_createShortcutHandler(t *testing.T) {

	config.InitFlags()
	storage = filestorage.Get("/tmp/test-storage.json")

	//после выполнения теста очищаем сторедж с URL'ами
	defer func() {
		storage.Close()
		os.Remove("/tmp/test-storage.json")
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
				response:    "Invalid request method: wrong\n",
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
			//отрезаем от ответа вида http//localhost:8080/<hash_string> префикс http//localhost:8080/ чтобы получить ключ мапы
			key, _ := strings.CutPrefix(test.want.response, "http://localhost:8080/")
			address, ok := storage.ReadValue(key)
			//проверяем наличие элемента в мапе
			assert.Equal(t, test.want.isMap, ok)
			//проверяем что значение по этому ключу является нужным нужным адресом
			assert.Equal(t, test.request.body, address)

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}

func Test_resolveShortcutHandler(t *testing.T) {
	storage = filestorage.Get("/tmp/test-storage.json")

	//после выполнения теста очищаем сторедж с URL'ами
	defer func() {
		storage.Close()
		os.Remove("/tmp/test-storage.json")
	}()

	storage.WriteValue("D63CDBB3", "https://wikipedia.org")
	storage.WriteValue("5B1A2675", "https://google.com")
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

func Test_createShortcutJSONHandler(t *testing.T) {
	storage = filestorage.Get("/tmp/test-storage.json")

	//после выполнения теста очищаем сторедж с URL'ами
	defer func() {
		storage.Close()
		os.Remove("/tmp/test-storage.json")
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
				response:    "{\"result\":\"http://localhost:8080/D63CDBB3\"}",
				contentType: "application/json",
				isMap:       true},
			request: request{body: "{\"url\": \"https://wikipedia.org\"}", method: http.MethodPost, contentType: "application/json"},
		},
		{
			name: "google test",
			want: want{code: http.StatusCreated,
				response:    "{\"result\":\"http://localhost:8080/5B1A2675\"}",
				contentType: "application/json",
				isMap:       true},
			request: request{body: "{\"url\": \"https://google.com\"}", method: http.MethodPost, contentType: "application/json"},
		},
		{
			name: "wrong content-type test",
			want: want{code: http.StatusBadRequest,
				response:    "Invalid request method\n",
				contentType: "",
				isMap:       false},
			request: request{body: "\"url\": \"https://google.com\"", method: http.MethodPost, contentType: "wrong"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, "/api/shorten", strings.NewReader(test.request.body))
			request.Header.Set("Content-Type", test.request.contentType)
			recorder := httptest.NewRecorder()

			createShortcutJSONHandler(recorder, request)
			res := recorder.Result()

			assert.Equal(t, test.want.code, res.StatusCode)

			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.response, string(resBody))

			if res.StatusCode != http.StatusCreated {
				return
			}
			//отрезаем от ответа вида http//localhost:8080/<hash_string> префикс http//localhost:8080/ чтобы получить ключ мапы
			key, _ := strings.CutPrefix(test.want.response, "{\"result\":\"http://localhost:8080/")
			key, _ = strings.CutSuffix(key, "\"}")
			address, ok := storage.ReadValue(key)
			//проверяем наличие элемента в мапе
			assert.Equal(t, test.want.isMap, ok)
			//проверяем что значение по этому ключу является нужным нужным адресом
			value, _ := strings.CutPrefix(test.request.body, "{\"url\": \"")
			value, _ = strings.CutSuffix(value, "\"}")
			assert.Equal(t, value, address)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}
