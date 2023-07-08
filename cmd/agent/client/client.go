package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/kvvPro/metric-collector/internal/metrics"

	"go.uber.org/zap"
)

var Sugar zap.SugaredLogger

type Agent interface {
	ReadMetrics()
	PushMetrics()
}

type Metric interface {
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
	cli.Metrics.RandomValue = 0.1 + rand.Float64()*(1000-0.1)
}

func (cli *Client) PushMetrics() {
	mslice := DeepFields(cli.Metrics)

	for _, m := range mslice {
		cli.updateMetric(m)
	}
}

func (cli *Client) PushMetricsJSON() {
	mslice := DeepFieldsNew(cli.Metrics)

	err := cli.updateMetricsJSON(mslice)
	if err != nil {
		panic(err)
	}
	// обнуляем PollCount
	cli.Metrics.PollCount = 0
}

func (cli *Client) updateMetricsJSON(allMetrics []metrics.Metric) error {
	client := &http.Client{}
	url := "http://" + cli.host + ":" + cli.port + "/update/"

	for _, m := range allMetrics {
		bodyBuffer := new(bytes.Buffer)
		gzb := gzip.NewWriter(bodyBuffer)
		json.NewEncoder(gzb).Encode(m)
		err := gzb.Close()
		if err != nil {
			panic(err)
		}

		request, err := http.NewRequest(http.MethodPost, url, bodyBuffer)
		if err != nil {
			panic(err)
		}
		Sugar.Infoln("-----------NEW REQUEST---------------")
		Sugar.Infoln("client-request: ", bodyBuffer.String())

		request.Header.Set("Connection", "Keep-Alive")
		request.Header.Set("Content-Encoding", "gzip")
		response, err := client.Do(request)
		if err != nil {
			Sugar.Infoln("Error response: ", err.Error())
			continue
		}
		Sugar.Infoln("Request done")

		dataResponse, err := io.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}
		Sugar.Infoln("Response body was read")

		Sugar.Infoln("client-response: ", string(dataResponse))
		Sugar.Infoln(
			"uri", request.RequestURI,
			"method", request.Method,
			"status", response.Status, // получаем код статуса ответа
		)

		defer response.Body.Close()
	}

	return nil
}

func (cli *Client) updateMetric(metric Metric) error {
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

	request.Header.Set("Content-Encoding", "gzip")
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
