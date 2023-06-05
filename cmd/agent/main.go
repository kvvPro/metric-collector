package main

import (
	"runtime"
	"runtime/metrics"
)

func main() {
	metrics.All()
	runtime.ReadMemStats(&runtime.MemStats{})
}
