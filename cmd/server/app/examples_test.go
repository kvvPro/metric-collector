package app

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kvvPro/metric-collector/internal/metrics"
)

func ExampleServer_PingHandle() {
	client := &http.Client{}
	address := "localhost:8080"
	url := "http://" + address + "/ping"

	bodyBuffer := new(bytes.Buffer)
	request, err := http.NewRequest(http.MethodGet, url, bodyBuffer)
	if err != nil {
		fmt.Printf("Error request: %v", err.Error())
		return
	}

	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("Error response: %v", err.Error())
		return
	}
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Error response")
		return
	}
}

func ExampleServer_UpdateBatchJSONHandle() {
	client := &http.Client{}
	address := "localhost:8080"
	url := "http://" + address + "/updates/"

	mval := 34.5
	mem := &metrics.Metric{
		ID:    "virtual_memory",
		MType: "gauge",
		Value: &mval,
	}
	cval := int64(12)
	cpu := &metrics.Metric{
		ID:    "cpu",
		MType: "counter",
		Delta: &cval,
	}
	arr := make([]metrics.Metric, 0)
	arr = append(arr, *mem)
	arr = append(arr, *cpu)

	bodyBuffer := new(bytes.Buffer)
	gzb := gzip.NewWriter(bodyBuffer)
	json.NewEncoder(gzb).Encode(arr)
	err := gzb.Close()
	if err != nil {
		fmt.Printf("Error encode request body: %v", err.Error())
		return
	}

	request, err := http.NewRequest(http.MethodPost, url, bodyBuffer)
	if err != nil {
		fmt.Printf("Error request: %v", err.Error())
		return
	}

	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("Error response: %v", err.Error())
		return
	}
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Error response")
		return
	}
}

func ExampleServer_AllMetricsHandle() {
	client := &http.Client{}
	address := "localhost:8080"
	url := "http://" + address + "/"

	bodyBuffer := new(bytes.Buffer)
	request, err := http.NewRequest(http.MethodGet, url, bodyBuffer)
	if err != nil {
		fmt.Printf("Error request: %v", err.Error())
		return
	}

	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("Error response: %v", err.Error())
		return
	}

	dataResponse, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v", err.Error())
		return
	}

	var allmetrics []metrics.Metric
	reader := io.NopCloser(bytes.NewReader(dataResponse))
	if err := json.NewDecoder(reader).Decode(&allmetrics); err != nil {
		fmt.Print("Error to parse response body")
		return
	}

	if response.StatusCode != http.StatusOK {
		fmt.Printf("Error response")
		return
	}
	fmt.Printf("all metrics: %v", allmetrics)
}
