package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/kvvPro/metric-collector/internal/hash"
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
	// TotalMemory     float64
	// FreeMemory      float64
	// CPUutilization1 float64
}

type Client struct {
	Metrics        Metrics
	pollInterval   int
	reportInterval int
	Address        string
	contentType    string
	needToHash     bool
	hashKey        string
	queue          chan []metrics.Metric
	maxWorkerCount int
}

func NewClient(pollInterval int, reportInterval int, address string, contentType string, hashKey string, rateLimit int) (*Client, error) {
	return &Client{
		Metrics:        Metrics{},
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		Address:        address,
		contentType:    contentType,
		hashKey:        hashKey,
		needToHash:     hashKey != "",
		queue:          make(chan []metrics.Metric),
		maxWorkerCount: rateLimit,
	}, nil
}

func (cli *Client) ReadMetrics(ctx context.Context) {
	for {
		runtime.ReadMemStats(&cli.Metrics.MemStats)

		// т.к. отправляем все попытки чтений - то PollCount всегда 1
		cli.Metrics.PollCount = 1
		cli.Metrics.RandomValue = 0.1 + rand.Float64()*(1000-0.1)

		// send metrics to channel
		mslice := DeepFieldsNew(cli.Metrics)
		cli.queue <- mslice

		select {
		case <-time.After(time.Duration(cli.pollInterval) * time.Second):
		case <-ctx.Done():
			return
		}
	}
}

func (cli *Client) ReadSpecificMetrics(ctx context.Context) {
	for {
		fields := make([]metrics.Metric, 0)
		memstats, err := mem.VirtualMemory()
		cpustat, errs := cpu.PercentWithContext(ctx, 0, false)
		if err != nil {
			continue
		}
		if errs != nil {
			continue
		}
		total := float64(memstats.Total)
		newTotal := metrics.NewCommonMetric("TotalMemory", metrics.MetricTypeGauge, nil, &total)
		free := float64(memstats.Free)
		newFree := metrics.NewCommonMetric("FreeMemory", metrics.MetricTypeGauge, nil, &free)
		cpu := cpustat[0]
		newCPU := metrics.NewCommonMetric("CPUutilization1", metrics.MetricTypeGauge, nil, &cpu)

		fields = append(fields, *newTotal, *newFree, *newCPU)

		cli.queue <- fields

		select {
		case <-time.After(time.Duration(cli.pollInterval) * time.Second):
		case <-ctx.Done():
			return
		}
	}
}

func (cli *Client) PushMetricsJSON(ctx context.Context) {
	// read from channel
	for {
		// wait interval
		select {
		case <-time.After(time.Duration(cli.reportInterval) * time.Second):
		case <-ctx.Done():
			return
		}

		m, opened := <-cli.queue
		if !opened {
			// channel is closed
			return
		}

		err := retry.Do(
			func() error {
				return cli.updateBatchMetricsJSON(m)
			},
			retry.Attempts(3),
			retry.InitDelay(1000*time.Millisecond),
			retry.Step(2000*time.Millisecond),
			retry.Context(ctx),
		)
		if err != nil {
			Sugar.Infoln(err.Error())
		}
	}
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
	if cli.needToHash {
		hash := hash.GetHashSHA256(bodyBuffer.String(), cli.hashKey)
		request.Header.Set("HashSHA256", base64.URLEncoding.EncodeToString(hash))
	}
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

func (cli *Client) Run(ctx context.Context) {
	for i := 0; i < cli.maxWorkerCount; i++ {
		go cli.PushMetricsJSON(ctx)
	}
	go cli.ReadSpecificMetrics(ctx)

	cli.ReadMetrics(ctx)

}
