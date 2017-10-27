package bench

import (
	"context"

	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/tylertreat/bench"
)

// Request represents a request to a Prometheus remote storage endpoint. It is a sum
// type of a write request and a query request.
type Request struct {
	writeRequest *prompb.WriteRequest
	queryRequest *prompb.QueryRequest
}

// NewWriteRequest creates a new write request.
func NewWriteRequest(r *prompb.WriteRequest) Request {
	return Request{writeRequest: r}
}

// NewQueryRequest creates a new query request.
func NewQueryRequest(r *prompb.QueryRequest) Request {
	return Request{queryRequest: r}
}

// RequestGenerator generates new requests.
type RequestGenerator interface {
	// Generate generates a new request.
	Generate() Request
}

// RequesterFactory implements bench.RequesterFactory by creating a Requester
// which issues requests to Prometheus's remote storage endpoint.
type RequesterFactory struct {
	Generator RequestGenerator
	Index     int
	Config    remote.ClientConfig
}

// GetRequester returns a new bench.Requester, called for each Benchmark connection.
func (f *RequesterFactory) GetRequester(uint64) bench.Requester {
	return &requester{
		generator: f.Generator,
		index:     f.Index,
		config:    f.Config,
	}
}

// Requester implements bench.Requester by issuing a request to a Prometheus remote
// storage endpoint.
type Requester struct {
	generator RequestGenerator
	index     int
	config    remote.ClientConfig
	client    *remote.Client
}

// Setup prepares the Requester for benchmarking.
func (r *requester) Setup() error {
	client, err := remote.NewClient(p.index, &p.config)
	if err != nil {
		return err
	}
	p.client = client
	return nil
}

// Request executes a synchronous request.
func (r *requester) Request() error {
	req := r.generator.Gen()
	if req.writeRequest != nil {
		return r.client.Store(req.writeRequest)
	}
	_, err := r.client.Read(context.Background(), req.queryRequest)
	return err
}

// Teardown cleans up the Requester upon benchmark completion.
func (r *requester) Teardown() error {
	return nil
}
