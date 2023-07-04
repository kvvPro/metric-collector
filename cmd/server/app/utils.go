package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/kvvPro/metric-collector/internal/metrics"
)

type IMetric interface {
	GetName() string
	GetType() string
	GetValue() any
	GetTypeForQuery() string
}

func isValidURL(url string) bool {
	// update
	re := regexp.MustCompile(`^/update/(counter|gauge)/\w+/\d+(?:\.\d+){0,1}$`)
	// get value
	reget := regexp.MustCompile(`^/value/(counter|gauge)/\w+$`)
	// get all metrics
	reall := regexp.MustCompile(`^/$`)
	return re.MatchString(url) || reget.MatchString(url) || reall.MatchString(url)
}

func isValidURLJSON(url string) bool {
	// update
	re := regexp.MustCompile(`^/update/$`)
	// get value
	reget := regexp.MustCompile(`^/value/$`)
	return re.MatchString(url) || reget.MatchString(url)
}

func isNameMissing(url string) bool {
	re := regexp.MustCompile(`^/update/(counter|gauge)/(\d+(?:\.\d+){0,1}){0,1}$`)
	return re.MatchString(url)
}

func isValidType(t string) bool {
	re := regexp.MustCompile(`^(counter|gauge)$`)
	return re.MatchString(t)
}

func isValidValue(v string) bool {
	re := regexp.MustCompile(`^\d+(?:\.\d+){0,1}$`)
	return re.MatchString(v)
}

func isValidUpdateJSONParams(r *http.Request, w http.ResponseWriter) ([]metrics.Metric, bool) {
	p := r.URL.Path

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return nil, false
	}
	// read body
	var oneMetric metrics.Metric
	var body []metrics.Metric

	data, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	Sugar.Infoln("body-request: ", string(data[:]))

	reader := io.NopCloser(bytes.NewReader(data))
	reader2 := io.NopCloser(bytes.NewReader(data))

	// 1 - try parse to array
	if err := json.NewDecoder(reader).Decode(&body); err != nil {
		// 2 - try parse to  1 Metric
		if err := json.NewDecoder(reader2).Decode(&oneMetric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, false
		}
		// it's one Metric
		body = append(body, oneMetric)
	}

	for _, m := range body {
		if m.ID == "" {
			http.Error(w, "Missing name of metric", http.StatusNotFound)
			return nil, false
		}

		if !isValidType(m.MType) || m.Delta == nil && m.Value == nil {
			http.Error(w, "Invalid type or value", http.StatusBadRequest)
			return nil, false
		}
	}

	// full regexp for check all path
	if !isValidURLJSON(p) {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return nil, false
	}

	return body, true
}

func isValidGetValueJSONParams(r *http.Request, w http.ResponseWriter) ([]metrics.Metric, bool) {
	p := r.URL.Path

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return nil, false
	}
	// read body
	var oneMetric metrics.Metric
	var body []metrics.Metric

	data, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	Sugar.Infoln("body-request: ", string(data[:]))

	reader := io.NopCloser(bytes.NewReader(data))
	reader2 := io.NopCloser(bytes.NewReader(data))

	// 1 - try parse to array
	if err := json.NewDecoder(reader).Decode(&body); err != nil {
		// 2 - try parse to  1 Metric
		if err := json.NewDecoder(reader2).Decode(&oneMetric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, false
		}
		// it's one Metric
		body = append(body, oneMetric)
	}

	// for _, m := range body {
	// 	if m.ID == "" {
	// 		http.Error(w, "Missing name of metric", http.StatusNotFound)
	// 		return nil, false
	// 	}

	// 	if !isValidType(m.MType) {
	// 		http.Error(w, "Invalid type", http.StatusBadRequest)
	// 		return nil, false
	// 	}
	// }

	// full regexp for check all path
	if !isValidURLJSON(p) {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return nil, false
	}

	return body, true
}

func isValidUpdateParams(r *http.Request, w http.ResponseWriter) ([]string, bool) {
	p := r.URL.Path

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return nil, false
	}
	if isNameMissing(p) {
		http.Error(w, "Missing name of metric", http.StatusNotFound)
		return nil, false
	}

	params := strings.Split(p, "/")

	metricType := params[2]
	metricValue := params[4]

	if !isValidType(metricType) || !isValidValue(metricValue) {
		http.Error(w, "Invalid type or value", http.StatusBadRequest)
		return nil, false
	}
	// full regexp for check all path
	if !isValidURL(p) {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return nil, false
	}

	return params, true
}

func isValidGetValueParams(r *http.Request, w http.ResponseWriter) ([]string, bool) {
	p := r.URL.Path

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return nil, false
	}
	// full regexp for check all path
	if !isValidURL(p) {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return nil, false
	}

	params := strings.Split(p, "/")

	metricType := params[2]

	if !isValidType(metricType) {
		http.Error(w, "Invalid type", http.StatusBadRequest)
		return nil, false
	}

	return params, true
}
