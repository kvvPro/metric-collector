package app

import (
	"io"
	"metric-collector/internal/storage"
	"metric-collector/internal/storage/memstorage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_UpdateHandle(t *testing.T) {
	type fields struct {
		storage storage.Storage
		Port    string
	}
	type url string
	type want struct {
		method      string
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		fields fields
		args   map[url]want
	}{
		{
			name: "1",
			fields: fields{
				storage: &memstorage.MemStorage{
					Gauges:   make(map[string]float64),
					Counters: make(map[string]int64),
				},
				Port: "8080",
			},
			args: map[url]want{
				"/update/counter/metric1/9": {
					method:      http.MethodPost,
					code:        200,
					response:    "OK!",
					contentType: "text/plain; charset=utf-8",
				},
				"/update/counter/someMetric/527": {
					method:      http.MethodPost,
					code:        200,
					response:    "OK!",
					contentType: "text/plain; charset=utf-8",
				},
				"/update/counter/": {
					method:      http.MethodPost,
					code:        404,
					response:    "Missing name of metric\n",
					contentType: "text/plain; charset=utf-8",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				storage: tt.fields.storage,
				Port:    tt.fields.Port,
			}

			for path, req := range tt.args {
				request := httptest.NewRequest(req.method, string(path), nil)
				// создаём новый Recorder
				w := httptest.NewRecorder()

				srv.UpdateHandle(w, request)

				res := w.Result()
				// проверяем код ответа
				assert.Equal(t, res.StatusCode, req.code)
				// получаем и проверяем тело запроса
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)

				require.NoError(t, err)
				assert.Equal(t, string(resBody), req.response)
				assert.Equal(t, res.Header.Get("Content-Type"), req.contentType)
			}
		})
	}
}

func TestServer_GetValueHandle(t *testing.T) {
	type fields struct {
		storage storage.Storage
		Port    string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				storage: tt.fields.storage,
				Port:    tt.fields.Port,
			}
			srv.GetValueHandle(tt.args.w, tt.args.r)
		})
	}
}

func TestServer_AllMetricsHandle(t *testing.T) {
	type fields struct {
		storage storage.Storage
		Port    string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				storage: tt.fields.storage,
				Port:    tt.fields.Port,
			}
			srv.AllMetricsHandle(tt.args.w, tt.args.r)
		})
	}
}
