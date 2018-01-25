// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/m3db/m3ninx/index/segment/bench"
	"github.com/uber-go/tally"

	xconfig "github.com/m3db/m3x/config"
	"github.com/m3db/m3x/instrument"
	xlog "github.com/m3db/m3x/log"
)

// Configuration is the configuration fields.
type Configuration struct {
	// Logging configuration.
	Logging xlog.Configuration `yaml:"logging"`

	// Metrics configuration.
	Metrics instrument.MetricsConfiguration `yaml:"metrics"`
}

func main() {
	var (
		configFile       = flag.String("config", "", "config file")
		serverAddress    = flag.String("server-address", "http://localhost:12345", "idxnode server address")
		duration         = flag.Duration("duration", time.Minute, "load test duration")
		numIngestWorkers = flag.Int("num-workers", 20, "number of go-routines performing ingestion")
		ingestQPS        = flag.Int("ingest-qps", 100, "ingest-qps")
	)
	flag.Parse()

	if *configFile == "" ||
		*duration < time.Second ||
		*numIngestWorkers < 0 ||
		*ingestQPS < 0 {
		flag.Usage()
		os.Exit(1)
	}
	var cfg Configuration
	if err := xconfig.LoadFile(&cfg, *configFile); err != nil {
		fmt.Fprintf(os.Stderr, "unable to load %s: %v", *configFile, err)
	}

	logger, err := cfg.Logging.BuildLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create logger: %v", err)
	}

	scope, closer, err := cfg.Metrics.NewRootScope()
	if err != nil {
		logger.Fatalf("could not connect to metrics: %v", err)
	}
	defer closer.Close()

	ih, err := newIngestHarness(scope, *serverAddress, *duration, *numIngestWorkers, *ingestQPS)
	if err != nil {
		logger.Fatalf("unable to create ingest harness: %v", err)
	}

	logger.Infof("starting load generation")
	ih.run()
	logger.Infof("finished load generation")
}

type ingestHarness struct {
	ingestAddress    string
	loadDuration     time.Duration
	numIngestWorkers int
	ingestQPS        int

	httpClient         *http.Client
	genSpec            *bench.GeneratorSpec
	workerWg           sync.WaitGroup
	workerTickDuration time.Duration

	metrics *ingestMetrics
}

func (h *ingestHarness) run() {
	h.workerWg.Add(h.numIngestWorkers)
	for i := 0; i < h.numIngestWorkers; i++ {
		go h.workerFn()
	}
	h.workerWg.Wait()
}

func (h *ingestHarness) workerFn() {
	workerStartTime := time.Now()
	ticker := time.NewTicker(h.workerTickDuration)
	defer func() {
		h.workerWg.Done()
		ticker.Stop()
	}()
	for {
		if time.Since(workerStartTime) > h.loadDuration {
			return
		}
		<-ticker.C
		startTime := time.Now()

		d := h.genSpec.Generate()
		jsonBytes, err := json.Marshal(d)
		if err != nil {
			h.metrics.generate.ReportError(time.Since(startTime))
			continue
		}

		req, err := http.NewRequest("POST", h.ingestAddress, bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := h.httpClient.Do(req)

		// release any resources held by response.Body to ensure underlying connection
		// can be reused
		responseCode := 200
		if resp != nil {
			responseCode = resp.StatusCode
		}
		h.metrics.addResponseCode(responseCode)

		if resp != nil && resp.Body != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}

		if err != nil && rand.Float64() < 0.05 {
			println(fmt.Sprintf("error: %v", err))
		}

		h.metrics.index.ReportSuccessOrError(err, time.Since(startTime))
	}
}

func newIngestHarness(
	scope tally.Scope,
	serverAddress string,
	loadDuration time.Duration,
	numIngestWorkers int,
	ingestQPS int,
) (*ingestHarness, error) {
	gs, err := bench.NewGeneratorSpec(time.Now().Unix())
	if err != nil {
		return nil, err
	}
	td := time.Second / time.Duration(ingestQPS)

	return &ingestHarness{
		ingestAddress:    fmt.Sprintf("%s/index", serverAddress),
		loadDuration:     loadDuration,
		numIngestWorkers: numIngestWorkers,
		ingestQPS:        ingestQPS,

		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        10000,
				MaxIdleConnsPerHost: 10000,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		genSpec:            gs,
		workerTickDuration: td,

		metrics: newIngestMetrics(scope),
	}, nil
}

type ingestMetrics struct {
	sync.RWMutex

	scope              tally.Scope
	generate           instrument.MethodMetrics
	index              instrument.MethodMetrics
	statusCodeCounters map[int]tally.Counter
}

func newIngestMetrics(s tally.Scope) *ingestMetrics {
	scope := s.SubScope("client")
	return &ingestMetrics{
		scope:              scope,
		index:              instrument.NewMethodMetrics(scope, "index", 1.0),
		generate:           instrument.NewMethodMetrics(scope, "generate", 1.0),
		statusCodeCounters: make(map[int]tally.Counter, 10),
	}
}

func (im *ingestMetrics) addResponseCode(c int) {
	im.RLock()
	count, ok := im.statusCodeCounters[c]
	im.RUnlock()
	if ok {
		count.Inc(1)
		return
	}

	im.Lock()
	count, ok = im.statusCodeCounters[c]
	if ok {
		im.Unlock()
		count.Inc(1)
		return
	}

	count = im.scope.Tagged(map[string]string{
		"code": fmt.Sprintf("%d", c),
	}).Counter("response")
	im.statusCodeCounters[c] = count
	im.Unlock()

	count.Inc(1)
}
