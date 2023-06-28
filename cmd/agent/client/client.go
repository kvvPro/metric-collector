package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/kvvPro/metric-collector/internal/metrics"
)

type Agent interface {
	ReadMetrics()
	PushMetrics()
}

type IMetric interface {
	GetName() string
	GetType() string
	GetValue() any
	GetTypeForQuery() string
}

type Metrics struct {
	runtime.MemStats
	PollCount   int64
	RandomValue float64
}

type Client struct {
	Metrics        Metrics
	pollInterval   int
	reportInterval int
	host           string
	port           string
	contentType    string
}

func NewClient(pollInterval int, reportInterval int, host string, port string, contentType string) (*Client, error) {
	return &Client{
		Metrics:        Metrics{},
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		host:           host,
		port:           port,
		contentType:    contentType,
	}, nil
}

func (cli *Client) ReadMetrics() {
	runtime.ReadMemStats(&cli.Metrics.MemStats)
	cli.Metrics.PollCount += 1
}

func (cli *Client) PushMetrics() {
	mslice := DeepFields(cli.Metrics)

	for _, m := range mslice {
		cli.updateMetric(m)
	}
}

func (cli *Client) PushMetricsJSON() {
	mslice := DeepFieldsNew(cli.Metrics)

	cli.updateMetricsJSON(mslice)
}

func (cli *Client) updateMetricsJSON(allMetrics []metrics.Metric) error {
	client := &http.Client{}
	url := "http://" + cli.host + ":" + cli.port + "/update/"

	bodyBuffer := new(bytes.Buffer)
	json.NewEncoder(bodyBuffer).Encode(allMetrics)

	request, err := http.NewRequest(http.MethodPost, url, bodyBuffer)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	_, serr := io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if serr != nil {
		panic(serr)
	}
	return nil
}

func (cli *Client) updateMetric(metric IMetric) error {
	client := &http.Client{}
	// metric := m.(Metric)
	metricType := metric.GetTypeForQuery()
	metricName := metric.GetName()
	metricValue := metric.GetValue()
	url := "http://" + cli.host + ":" + cli.port + "/update/" +
		metricType + "/" + metricName + "/" + fmt.Sprintf("%v", metricValue)

	var body []byte
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", cli.contentType)
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	_, serr := io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if serr != nil {
		panic(serr)
	}
	return nil
}

func (cli *Client) Run() {
	timefromReport := 0

	for {
		if timefromReport >= cli.reportInterval {
			timefromReport = 0
			// cli.PushMetrics()
			cli.PushMetricsJSON()
		}

		cli.ReadMetrics()

		time.Sleep(time.Duration(cli.pollInterval) * time.Second)
		timefromReport += cli.pollInterval
	}
}
