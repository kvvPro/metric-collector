package app

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kvvPro/metric-collector/internal/encrypt"
	"github.com/kvvPro/metric-collector/internal/hash"
	mc "github.com/kvvPro/metric-collector/internal/metrics"
	ip "github.com/kvvPro/metric-collector/internal/net"
	"go.uber.org/zap"
)

// Main logger
var Sugar zap.SugaredLogger

// type needed compress
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

	// определяем еще один тип, чтобы переопределить только метод Write
	hashResponseWriter struct {
		http.ResponseWriter
		SetHash bool
		HashKey string
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

func (r *hashResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	if r.SetHash {
		hash := hash.GetHashSHA256(string(b), r.HashKey)
		r.ResponseWriter.Header().Set("HashSHA256", base64.URLEncoding.EncodeToString(hash))
	}
	return r.ResponseWriter.Write(b)
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

func (srv *Server) ValidateIP(h http.Handler) http.Handler {
	validateIPFunc := func(w http.ResponseWriter, r *http.Request) {

		if srv.TrustedSubnet != "" {
			clientIP := r.Header.Get("X-Real-IP")
			if clientIP == "" {
				http.Error(w, errors.New("not found client IP").Error(), http.StatusBadRequest)
				return
			}
			trusted, err := ip.CheckIPInSubnet(clientIP, srv.TrustedSubnet)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !trusted {
				http.Error(w, errors.New("client IP not in trusted subnet").Error(), http.StatusBadRequest)
				return
			}
		}

		// передаём управление хендлеру
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(validateIPFunc)
}

func (srv *Server) DecryptMiddleware(h http.Handler) http.Handler {
	decryptFunc := func(w http.ResponseWriter, r *http.Request) {
		ow := w

		if srv.UseEncryption {
			// decrypt request body
			data, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// decrypt only non-empty data
			if len(data) > 0 {
				decryptBody, err := encrypt.Decrypt(srv.PrivateKeyPath, string(data))
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				// возвращаем тело запроса
				r.Body = io.NopCloser(bytes.NewReader([]byte(decryptBody)))
			}
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(decryptFunc)
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

func (srv *Server) CheckHashMiddleware(h http.Handler) http.Handler {
	checkHashFunc := func(w http.ResponseWriter, r *http.Request) {
		requestHash := r.Header.Get("HashSHA256")
		if requestHash != "" {
			// проверяем хэш
			data, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			originalHash := hash.GetHashSHA256(string(data), srv.HashKey)
			decodeHash, err := base64.URLEncoding.DecodeString(requestHash)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !hmac.Equal(originalHash, decodeHash) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// возвращаем тело запроса
			r.Body = io.NopCloser(bytes.NewReader(data))
		}

		// подменяем на наш writer
		hw := hashResponseWriter{
			ResponseWriter: w,
			SetHash:        srv.CheckHash,
			HashKey:        srv.HashKey,
		}

		// передаём управление хендлеру
		h.ServeHTTP(&hw, r)
	}
	return http.HandlerFunc(checkHashFunc)
}

// PingHandle godoc
// @Tags checks
// @Summary Checking db connection
// @Description Checking db connection
// @ID checksPing
// @Accept  plain
// @Produce plain
// @Success 200 {string} string "OK"
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router /ping [get]
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

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	body := "OK!"
	io.WriteString(w, body)
}

// UpdateHandle godoc
// @Tags update
// @Summary Update existed metric or add new metric
// @Description Update existed metric or add new metric
// @ID update
// @Accept  plain
// @Produce plain
// @Param type path string true "Metric type"
// @Param name path string true "Metric name"
// @Param value path number true "Metric value"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Invalid type or value"
// @Failure 404 {string} string "Missing name of metric"
// @Failure 405 {string} string "Invalid request type"
// @Failure 500 {string} string "Internal error"
// @Router /update/{type}/{name}/{value} [post]
func (srv *Server) UpdateHandle(w http.ResponseWriter, r *http.Request) {
	params, isValid := isValidUpdateParams(r, w)
	if !isValid {
		return
	}

	metricType := params[2]
	metricName := params[3]
	metricValue := params[4]
	err := srv.AddMetric(r.Context(), metricType, metricName, metricValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body := "OK!"
	io.WriteString(w, body)
	w.WriteHeader(http.StatusOK)
}

// UpdateJSONHandle godoc
// @Tags update
// @Summary Update metric, data in JSON
// @Description Update existed metric or add new metric from JSON data
// @ID updateJSON
// @Accept  json
// @Produce json
// @Param metric body []metrics.Metric true "Array of metrics"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Invalid type or value"
// @Failure 404 {string} string "Missing name of metric"
// @Failure 405 {string} string "Invalid request type"
// @Failure 500 {string} string "Internal error"
// @Router /update/ [post]
func (srv *Server) UpdateJSONHandle(w http.ResponseWriter, r *http.Request) {
	requestedMetrics, isValid := isValidUpdateJSONParams(r, w)
	if !isValid {
		return
	}

	for _, m := range requestedMetrics {
		err := srv.AddMetricNew(r.Context(), m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	updatedMetrics, err := srv.GetRequestedValues(r.Context(), requestedMetrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

// UpdateBatchJSONHandle godoc
// @Tags update
// @Summary Batch update array of metrics, data in JSON
// @Description Update existed metric or add new metric from JSON array of metrics
// @ID updateBatch
// @Accept  json
// @Produce json
// @Param metrics body []metrics.Metric true "Array of metrics"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Invalid type or value"
// @Failure 404 {string} string "Missing name of metric"
// @Failure 405 {string} string "Invalid request type"
// @Failure 500 {string} string "Internal error"
// @Router /updates/ [post]
func (srv *Server) UpdateBatchJSONHandle(w http.ResponseWriter, r *http.Request) {
	requestedMetrics, isValid := isValidUpdateJSONParams(r, w)
	if !isValid {
		return
	}

	err := srv.AddMetricsBatch(r.Context(), requestedMetrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	testbody := "OK!"
	io.WriteString(w, testbody)
	w.WriteHeader(http.StatusOK)
}

// GetValueJSONHandle godoc
// @Tags getvalue
// @Summary Get metric value
// @Description Get metric value
// @ID getvalueJSON
// @Accept  json
// @Produce json
// @Param metric body metrics.Metric true "metric"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Invalid type or value"
// @Failure 404 {string} string "Missing name of metric"
// @Failure 405 {string} string "Invalid request type"
// @Failure 500 {string} string "Internal error"
// @Router /value/ [get]
func (srv *Server) GetValueJSONHandle(w http.ResponseWriter, r *http.Request) {
	requestedMetrics, isValid := isValidGetValueJSONParams(r, w)
	if !isValid {
		return
	}

	updatedMetrics, err := srv.GetRequestedValues(r.Context(), requestedMetrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	allmetrics, err := srv.GetAllMetricsNew(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

// GetValueHandle godoc
// @Tags getvalue
// @Summary Get value of existed metric
// @Description Get value of existed metric
// @ID getvalue
// @Accept  plain
// @Produce plain
// @Param type path string true "Metric type"
// @Param name path string true "Metric name"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Invalid type"
// @Failure 404 {string} string "Missing name of metric"
// @Failure 405 {string} string "Invalid request type"
// @Failure 500 {string} string "Internal error"
// @Router /value/{type}/{name} [get]
func (srv *Server) GetValueHandle(w http.ResponseWriter, r *http.Request) {
	params, isValid := isValidGetValueParams(r, w)
	if !isValid {
		return
	}
	metricType := params[2]
	metricName := params[3]

	val, err := srv.GetMetricValue(r.Context(), metricType, metricName)
	if val == nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	io.WriteString(w, fmt.Sprintf("%v", val))
	w.WriteHeader(http.StatusOK)
}

// AllMetricsHandle godoc
// @Tags getvalue
// @Summary Get all metrics
// @Description Get all metrics with current values
// @ID getvalue
// @Accept  plain
// @Produce json
// @Success 200 {string} string "OK"
// @Failure 405 {string} string "Invalid request type"
// @Failure 500 {string} string "Internal error"
// @Router /value/{type}/{name} [get]
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

	metrics, err := srv.GetAllMetricsNew(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
		if el.MType == mc.MetricTypeCounter {
			rows += fmt.Sprintf("<tr><th>%v</th><th>%v</th></tr>", el.ID, *(el.Delta))
		} else {
			rows += fmt.Sprintf("<tr><th>%v</th><th>%v</th></tr>", el.ID, *(el.Value))
		}
	}

	body = strings.ReplaceAll(body, "%rows", rows)
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, body)
	w.WriteHeader(http.StatusOK)
}
