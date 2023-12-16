package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	rpprof "runtime/pprof"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/kvvPro/metric-collector/cmd/agent/config"
	"github.com/kvvPro/metric-collector/internal/encrypt"
	"github.com/kvvPro/metric-collector/internal/hash"
	"github.com/kvvPro/metric-collector/internal/metrics"
	ip "github.com/kvvPro/metric-collector/internal/net"
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
	// All runtime memory stats
	runtime.MemStats
	// counts of reading stats attempts
	PollCount int64
	// random value
	RandomValue float64
}

type Client struct {
	// array of metrics
	Metrics Metrics
	// timeout to read stats
	pollInterval int
	// timeout to send metrics to server
	reportInterval int
	// server address
	Address string
	// true if client will encrypt request body
	needToHash bool
	// key to encrypt request
	hashKey string
	// channel for exchange between goroutines - read and push threads
	queue chan []metrics.Metric
	// max count of parallel workers to push metrics to server
	maxWorkerCount int
	// True if agent encrypts all messages
	needToEncrypt bool
	// Path to public key RSA
	publicKey string
	// Path to file where mem stats will be saved
	MemProfile string
	// Exchange mode
	ExchangeMode string
	// wait group for sync
	wg *sync.WaitGroup
	// func to cancel context
	cancelFunc context.CancelFunc
}

// NewClient creates instance of client
func NewClient(settings *config.ClientFlags) (*Client, error) {
	return &Client{
		Metrics:        Metrics{},
		pollInterval:   settings.PollInterval,
		reportInterval: settings.ReportInterval,
		Address:        settings.Address,
		hashKey:        settings.HashKey,
		needToHash:     settings.HashKey != "",
		queue:          make(chan []metrics.Metric, settings.RateLimit),
		maxWorkerCount: settings.RateLimit,
		publicKey:      settings.CryptoKey,
		needToEncrypt:  settings.CryptoKey != "",
		MemProfile:     settings.MemProfile,
		ExchangeMode:   settings.ExchangeMode,
	}, nil
}

// ReadMetrics reads memory stats and send it to queue
func (cli *Client) ReadMetrics(ctx context.Context) {
	for {
		runtime.ReadMemStats(&cli.Metrics.MemStats)

		// т.к. отправляем все попытки чтений - то PollCount всегда 1
		cli.Metrics.PollCount = 1
		cli.Metrics.RandomValue = 0.1 + rand.Float64()*(1000-0.1)

		// send metrics to channel
		mslice := DeepFieldsNew(cli.Metrics)

		Sugar.Infoln("Read metrics - 1")
		cli.queue <- mslice

		select {
		case <-time.After(time.Duration(cli.pollInterval) * time.Second):
		case <-ctx.Done():
			return
		}
	}
}

// ReadSpecificMetrics reads special stats and send it to queue
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
		fields = append(fields, *newTotal, *newFree)

		for i := 0; i < len(cpustat); i++ {
			cpu := cpustat[i]
			newCPU := metrics.NewCommonMetric("CPUutilization"+fmt.Sprint(i), metrics.MetricTypeGauge, nil, &cpu)
			fields = append(fields, *newCPU)
		}

		Sugar.Infoln("Read specific metrics - 1")

		cli.queue <- fields

		select {
		case <-time.After(time.Duration(cli.pollInterval) * time.Second):
		case <-ctx.Done():
			return
		}
	}
}

// PushMetricsJSON push metrics to server
func (cli *Client) PushMetricsJSON(ctx context.Context) {
	// read from channel
	for {
		// wait interval
		select {
		case <-time.After(time.Duration(cli.reportInterval) * time.Second):
		case <-ctx.Done():
			return
		}

		Sugar.Infoln("Push metrics - 1")

		m, opened := <-cli.queue
		if !opened {
			// channel is closed
			return
		}

		err := retry.Do(
			func() error {
				if cli.ExchangeMode == "http" {
					return cli.updateBatchMetricsJSON(m)
				} else if cli.ExchangeMode == "grpc" {
					return cli.updateMetrics(ctx, m)
				} else {
					return fmt.Errorf("uknown exchange type - %v", cli.ExchangeMode)
				}
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

	if cli.needToEncrypt {
		encryptedBody, err := encrypt.Encrypt(cli.publicKey, bodyBuffer.String())
		if err != nil {
			Sugar.Infoln("Error encode request body: ", err.Error())
			return err
		}
		bodyBuffer = new(bytes.Buffer)
		_, err = bodyBuffer.WriteString(encryptedBody)
		if err != nil {
			Sugar.Infoln("Error to write request body: ", err.Error())
			return err
		}
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
	localIP := ip.GetOutboundIP(cli.Address)
	request.Header.Set("X-Real-IP", localIP.String())

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

// Run start client - read and push threads
func (cli *Client) Run(ctx context.Context) {
	asyncCtx, cancelAgent := context.WithCancel(ctx)
	cli.cancelFunc = cancelAgent
	cli.wg = &sync.WaitGroup{}
	for i := 0; i < cli.maxWorkerCount; i++ {
		cli.wg.Add(1)
		go func() {
			defer cli.wg.Done()
			cli.PushMetricsJSON(asyncCtx)
		}()
	}

	cli.wg.Add(1)
	go func() {
		defer cli.wg.Done()
		cli.ReadSpecificMetrics(asyncCtx)
	}()

	cli.wg.Add(1)
	go func() {
		defer cli.wg.Done()
		cli.ReadMetrics(asyncCtx)
	}()
}

func (cli *Client) Stop() {
	// создаём файл журнала профилирования памяти
	fmem, err := os.Create(cli.MemProfile)
	if err != nil {
		Sugar.Fatalw(err.Error())
	}
	defer fmem.Close()
	runtime.GC() // получаем статистику по использованию памяти
	if err := rpprof.WriteHeapProfile(fmem); err != nil {
		Sugar.Fatalw(err.Error())
	}

	cli.cancelFunc()
	cli.wg.Wait()
}
