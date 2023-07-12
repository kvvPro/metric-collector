package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var Sugar zap.SugaredLogger
var ContentTypesForCompress = "application/json; text/html"

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		duration := time.Since(start)

		Sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
		)
	}
	return http.HandlerFunc(logFn)
}

func GzipMiddleware(h http.Handler) http.Handler {
	compressFunc := func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		// enableCompress := strings.Contains(ContentTypesForCompress, w.Header().Get("Content-Type"))
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(w)
			cw.Header().Set("Content-Encoding", "gzip")
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.EqualFold(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(compressFunc)
}

func (srv *Server) PingHandle(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbpool, err := pgxpool.New(ctx, srv.DBConnection)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer dbpool.Close()

	err = dbpool.Ping(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body := "OK!"
	io.WriteString(w, body)
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) UpdateHandle(w http.ResponseWriter, r *http.Request) {
	params, isValid := isValidUpdateParams(r, w)
	if !isValid {
		return
	}

	metricType := params[2]
	metricName := params[3]
	metricValue := params[4]
	err := srv.AddMetric(metricType, metricName, metricValue)
	if err != nil {
		panic(err)
	}

	body := "OK!"
	io.WriteString(w, body)
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) UpdateJSONHandle(w http.ResponseWriter, r *http.Request) {
	requestedMetrics, isValid := isValidUpdateJSONParams(r, w)
	if !isValid {
		return
	}

	for _, m := range requestedMetrics {
		err := srv.AddMetricNew(m)
		if err != nil {
			panic(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")

	updatedMetrics := srv.GetRequestedValues(requestedMetrics)
	bodyBuffer := new(bytes.Buffer)
	if len(updatedMetrics) == 1 {
		json.NewEncoder(bodyBuffer).Encode(updatedMetrics[0])
	} else {
		json.NewEncoder(bodyBuffer).Encode(updatedMetrics)
	}
	body := bodyBuffer.String()

	Sugar.Infoln("body-response: ", body)

	io.WriteString(w, body)
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) GetValueJSONHandle(w http.ResponseWriter, r *http.Request) {
	requestedMetrics, isValid := isValidGetValueJSONParams(r, w)
	if !isValid {
		return
	}

	// // check if ID+Type of requested metric not found in our storage
	// for _, el := range requestedMetrics {
	// 	if _, err := srv.GetMetricValue(el.MType, el.ID); err != nil {
	// 		http.Error(w, err.Error(), http.StatusBadRequest)
	// 		return
	// 	}
	// }

	updatedMetrics := srv.GetRequestedValues(requestedMetrics)
	allmetrics := srv.GetAllMetricsNew()

	// error if one or more requested metrics weren't found in our store
	if len(requestedMetrics) != len(updatedMetrics) {
		Sugar.Infoln("!!! Missing name of metric")
		Sugar.Infoln("all-metrics: ", allmetrics)
		Sugar.Infoln("requsted-metrics: ", requestedMetrics)
		Sugar.Infoln("updated-metrics: ", updatedMetrics)
		http.Error(w, "Missing name of metric", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	bodyBuffer := new(bytes.Buffer)
	if len(updatedMetrics) == 1 {
		json.NewEncoder(bodyBuffer).Encode(updatedMetrics[0])
	} else {
		json.NewEncoder(bodyBuffer).Encode(updatedMetrics)
	}
	body := bodyBuffer.String()

	Sugar.Infoln("body-response: ", body)

	io.WriteString(w, body)
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) GetValueHandle(w http.ResponseWriter, r *http.Request) {
	params, isValid := isValidGetValueParams(r, w)
	if !isValid {
		return
	}
	metricType := params[2]
	metricName := params[3]

	val, err := srv.GetMetricValue(metricType, metricName)
	if val == nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	io.WriteString(w, fmt.Sprintf("%v", val))
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) AllMetricsHandle(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if !isValidURL(p) {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	metrics := srv.GetAllMetrics()
	body := `<html>
				<head>
				<title></title>
				</head>
				<body>
					<table border="1" cellpadding="1" cellspacing="1" style="width: 500px">
						<thead>
							<tr>
								<th scope="col">Metric name</th>
								<th scope="col">Value</th>
							</tr>
						</thead>
						<tbody>
							%rows
						</tbody>
					</table>
				</body>
			</html>`
	rows := ""
	for _, el := range metrics {
		rows += fmt.Sprintf("<tr><th>%v</th><th>%v</th></tr>", el.GetName(), el.GetValue())
	}

	body = strings.ReplaceAll(body, "%rows", rows)
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, body)
	w.WriteHeader(http.StatusOK)
}
