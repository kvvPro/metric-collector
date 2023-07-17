package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/kvvPro/metric-collector/internal/metrics"
	"github.com/kvvPro/metric-collector/internal/retry"

	"go.uber.org/zap"
)

var Sugar zap.SugaredLogger

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
	Address        string
	contentType    string
}

func NewClient(pollInterval int, reportInterval int, address string, contentType string) (*Client, error) {
	return &Client{
		Metrics:        Metrics{},
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		Address:        address,
		contentType:    contentType,
	}, nil
}

func (cli *Client) ReadMetrics() {
	runtime.ReadMemStats(&cli.Metrics.MemStats)
	cli.Metrics.PollCount += 1
	cli.Metrics.RandomValue = 0.1 + rand.Float64()*(1000-0.1)
}

func (cli *Client) PushMetricsJSON() {
	mslice := DeepFieldsNew(cli.Metrics)

	err := cli.updateBatchMetricsJSON(mslice)
	if err != nil {
		err = retry.Do(
			func() error {
				return cli.updateBatchMetricsJSON(mslice)
			},
			retry.Attempts(3),
			retry.PauseBeforeFirstAttempt(true),
		)
		if err != nil {
			Sugar.Infoln(err.Error())
		}
	}
	// обнуляем PollCount
	cli.Metrics.PollCount = 0
}

func (cli *Client) updateBatchMetricsJSON(allMetrics []metrics.Metric) error {
	client := &http.Client{}
	url := "http://" + cli.Address + "/updates/"

	bodyBuffer := new(bytes.Buffer)
	gzb := gzip.NewWriter(bodyBuffer)
	json.NewEncoder(gzb).Encode(allMetrics)
	err := gzb.Close()
	if err != nil {
		Sugar.Infoln("Error encode request body: ", err.Error())
		return err
	}

	request, err := http.NewRequest(http.MethodPost, url, bodyBuffer)
	if err != nil {
		Sugar.Infoln("Error request: ", err.Error())
		return err
	}
	Sugar.Infoln("-----------NEW REQUEST---------------")
	Sugar.Infoln("client-request: ", bodyBuffer.String())

	request.Header.Set("Connection", "Keep-Alive")
	request.Header.Set("Content-Encoding", "gzip")
	response, err := client.Do(request)
	if err != nil {
		Sugar.Infoln("Error response: ", err.Error())
		return err
	}
	Sugar.Infoln("Request done")

	dataResponse, err := io.ReadAll(response.Body)
	if err != nil {
		Sugar.Infoln("Error reading response body: ", err.Error())
		return err
	}
	Sugar.Infoln("Response body was read")

	Sugar.Infoln("client-response: ", string(dataResponse))
	Sugar.Infoln(
		"uri", request.RequestURI,
		"method", request.Method,
		"status", response.Status, // получаем код статуса ответа
	)

	defer response.Body.Close()

	return nil
}

func (cli *Client) Run() {
	timefromReport := 0

	for {
		if timefromReport >= cli.reportInterval {
			timefromReport = 0
			cli.PushMetricsJSON()
		}

		cli.ReadMetrics()

		time.Sleep(time.Duration(cli.pollInterval) * time.Second)
		timefromReport += cli.pollInterval
	}
}
