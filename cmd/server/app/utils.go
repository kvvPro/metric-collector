package app

import (
	"net/http"
	"regexp"
	"strings"
)

func isValidURL(url string) bool {
	re := regexp.MustCompile(`^/update/(counter|gauge)/\w+/\d+(?:\.\d+){0,1}$`)
	return re.MatchString(url)
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

func isValidParams(r *http.Request, w *http.ResponseWriter) ([]string, bool) {
	p := r.URL.Path

	if r.Method != http.MethodPost {
		http.Error(*w, "Invalid method", http.StatusMethodNotAllowed)
		return nil, false
	}
	if isNameMissing(p) {
		http.Error(*w, "Missing name of metric", http.StatusNotFound)
		return nil, false
	}

	params := strings.Split(p, "/")

	metricType := params[2]
	metricValue := params[4]

	if !isValidType(metricType) || !isValidValue(metricValue) {
		http.Error(*w, "Invalid type or value", http.StatusBadRequest)
		return nil, false
	}
	// full regexp for check all path
	if !isValidURL(p) {
		http.Error(*w, "Invalid query", http.StatusBadRequest)
		return nil, false
	}

	return params, true
}
