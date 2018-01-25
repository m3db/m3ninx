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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"gopkg.in/validator.v2"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index"
	"github.com/m3db/m3ninx/index/segment/mem"
	xconfig "github.com/m3db/m3x/config"
	"github.com/m3db/m3x/instrument"
	xlog "github.com/m3db/m3x/log"
)

const (
	listenPort = 12345
	debugPort  = 22345
)

// Configuration is the configuration fields.
type Configuration struct {
	// Logging configuration.
	Logging xlog.Configuration `yaml:"logging"`

	// Metrics configuration.
	Metrics instrument.MetricsConfiguration `yaml:"metrics"`
}

func main() {
	runtime.SetBlockProfileRate(100)

	var (
		configFile = flag.String("config", "", "config file")
	)
	flag.Parse()

	if *configFile == "" {
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

	iopts := instrument.NewOptions().
		SetLogger(logger).
		SetMetricsScope(scope)

	memOpts := mem.NewOptions().
		SetInstrumentOptions(iopts)
	opts := index.NewOptions().
		SetInstrumentOptions(iopts).
		SetMemSegmentOptions(memOpts)

	idx, err := index.New(opts)
	if err != nil {
		log.Fatalf("Unable to create index: %v", err)
	}

	ingestHandler := ingestHandler{
		idx: idx,
	}
	mux := http.NewServeMux()
	mux.Handle("/index", ingestHandler)

	address := fmt.Sprintf("0.0.0.0:%d", listenPort)
	go func() {
		if err := http.ListenAndServe(address, mux); err != nil {
			log.Fatalf("Unable to listen, err: %v\n", err)
		}
	}()
	log.Printf("Listening on %s", address)

	debugAddress := fmt.Sprintf("0.0.0.0:%d", debugPort)
	go func() {
		if err := http.ListenAndServe(debugAddress, nil); err != nil {
			log.Fatalf("Unable to listen for debug, err: %v\n", err)
		}
	}()
	log.Printf("Debug listening on %s", debugAddress)

	select {}
}

type ingestHandler struct {
	idx index.Index
}

func (i ingestHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(response, fmt.Sprintf("unable to read input: %v", err), 400)
		return
	}

	var d document
	if err := json.Unmarshal(body, &d); err != nil {
		http.Error(response, fmt.Sprintf("unable to read input: %v", err), 400)
		return
	}

	if err := validator.Validate(d); err != nil {
		http.Error(response, fmt.Sprintf("illegal input: %v", err), 400)
		return
	}

	err = i.idx.Insert(d.toDoc())
	if err == nil {
		response.WriteHeader(http.StatusOK)
		return
	}

	// indicate present document insertion error seperate from other cases
	if err == index.ErrDocAlreadyInserted {
		http.Error(response, err.Error(), 409)
		return
	}

	http.Error(response, fmt.Sprintf("unable to insert: %v", err), 500)
}

type field struct {
	Name  string `json:"name" validate:"nonzero"`
	Value string `json:"value" validate:"nonzero"`
}

type document struct {
	ID     string  `json:"id" validate:"nonzero"`
	Fields []field `json:"fields" validate:"nonzero"`
}

func (d *document) toDoc() doc.Document {
	fields := make([]doc.Field, 0, len(d.Fields))
	for _, f := range d.Fields {
		fields = append(fields, doc.Field{
			Name:      []byte(f.Name),
			Value:     []byte(f.Value),
			ValueType: doc.StringValueType,
		})
	}
	return doc.Document{
		ID:     []byte(d.ID),
		Fields: fields,
	}
}
