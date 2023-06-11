package app

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

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
