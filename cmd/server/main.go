package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMemStorage() MemStorage {
	return MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

func (s MemStorage) Update(t string, n string, v string) error {
	if t == "gauge" {
		if fval, err := strconv.ParseFloat(v, 32); err == nil {
			s.Gauges[n] = fval
		}
	} else if t == "counter" {
		if ival, err := strconv.ParseInt(v, 10, 64); err == nil {
			s.Counters[n] += ival
		}
	} else {
		return errors.New("uknown metric type")
	}
	return nil
}

func isValidURL(url string) bool {
	re := regexp.MustCompile(`^/update/(counter|gauge)/\w+/\d+(?:\.\d+){0,1}$`)
	return re.MatchString(url)
}

func isNameMissing(url string) bool {
	re := regexp.MustCompile(`^/update/(counter|gauge)/\d+(?:\.\d+){0,1}$`)
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

func isValidParams(p string, w http.ResponseWriter) ([]string, bool) {
	if !isValidURL(p) {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return nil, false
	}
	if isNameMissing(p) {
		http.Error(w, "Missing name of metric", http.StatusNotFound)
		return nil, false
	}

	params := strings.Split(p, "/")

	metric_type := params[2]
	metric_value := params[4]

	if !isValidType(metric_type) || !isValidValue(metric_value) {
		http.Error(w, "Invalid type or value", http.StatusBadRequest)
		return nil, false
	}

	return params, true
}

// func validateQuery(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		checkParams(r.URL.Path, w)
// 		next.ServeHTTP(w, r)
// 	})
// }

func mainHandle(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Start handling\r\n")

	params, isValid := isValidParams(r.URL.Path, w)

	if !isValid {
		return
	}

	metric_type := params[2]
	metric_name := params[3]
	metric_value := params[4]

	body := fmt.Sprintf("Method: %s\r\n", r.Method)
	body += "Params ===============\r\n"
	body += fmt.Sprintf("%s: %v\r\n", "metric_type", metric_type)
	body += fmt.Sprintf("%s: %v\r\n", "metric_name", metric_name)
	body += fmt.Sprintf("%s: %v\r\n", "metric_value", metric_value)

	io.WriteString(w, body)

	err := Storage.Update(metric_type, metric_name, metric_value)
	if err != nil {
		panic(err)
	}

	io.WriteString(w, "Finish handling\r\n")

	w.WriteHeader(http.StatusOK)
}

var Storage = NewMemStorage()

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(mainHandle))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
