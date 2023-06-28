package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

var Sugar zap.SugaredLogger

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

	io.WriteString(w, body)
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) GetValueJSONHandle(w http.ResponseWriter, r *http.Request) {
	requestedMetrics, isValid := isValidGetValueJSONParams(r, w)
	if !isValid {
		return
	}

	updatedMetrics := srv.GetRequestedValues(requestedMetrics)

	w.Header().Set("Content-Type", "application/json")

	bodyBuffer := new(bytes.Buffer)
	if len(updatedMetrics) == 1 {
		json.NewEncoder(bodyBuffer).Encode(updatedMetrics[0])
	} else {
		json.NewEncoder(bodyBuffer).Encode(updatedMetrics)
	}
	body := bodyBuffer.String()

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
	io.WriteString(w, body)
	w.WriteHeader(http.StatusOK)
}
