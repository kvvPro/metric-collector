package app

import (
	"io"
	"net/http"
)

func (srv *Server) MainHandle(w http.ResponseWriter, r *http.Request) {

	params, isValid := isValidParams(r, w)

	if !isValid {
		return
	}

	metricType := params[2]
	metricName := params[3]
	metricValue := params[4]

	// body := fmt.Sprintf("Method: %s\r\n", r.Method)
	// body += "Params ===============\r\n"
	// body += fmt.Sprintf("%s: %v\r\n", "metricType", metricType)
	// body += fmt.Sprintf("%s: %v\r\n", "metricName", metricName)
	// body += fmt.Sprintf("%s: %v\r\n", "metricValue", metricValue)

	err := srv.AddMetric(metricType, metricName, metricValue)
	if err != nil {
		panic(err)
	}

	body := "OK!"

	io.WriteString(w, body)

	// io.WriteString(w, "Finish handling\r\n")
	// w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(http.StatusOK)
}
