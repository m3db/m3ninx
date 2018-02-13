package mem

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/m3db/m3ninx/index/segment"

	"github.com/m3db/m3ninx/doc"
)

var (
	benchDocs []doc.Document
)

func init() {
	var err error
	if benchDocs, err = readDocuments("../../../testdata/node_exporter.json", 2000); err != nil {
		panic(fmt.Sprintf("unable to read documents for benchmarks: %v", err))
	}
}

func BenchmarkInsert_SimpleTermsDictionary(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		dict := newSimpleTermsDictionary(NewOptions())
		b.StartTimer()

		for i, d := range benchDocs {
			for _, f := range d.Fields {
				dict.Insert(f, segment.DocID(i))
			}
		}
	}
}

func BenchmarkInsert_TrigramTermsDictionary(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		dict := newTrigramTermsDictionary(NewOptions())
		b.StartTimer()

		for i, d := range benchDocs {
			for _, f := range d.Fields {
				dict.Insert(f, segment.DocID(i))
			}
		}
	}
}

func BenchmarkFetch_SimpleTermsDictionary(b *testing.B) {
	dict := newSimpleTermsDictionary(NewOptions())
	for i, d := range benchDocs {
		for _, f := range d.Fields {
			dict.Insert(f, segment.DocID(i))
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, d := range benchDocs {
			for _, f := range d.Fields {
				dict.Fetch(f.Name, f.Value, termFetchOptions{false})
			}
		}
	}
}

func BenchmarkFetch_TrigramTermsDictionary(b *testing.B) {
	dict := newTrigramTermsDictionary(NewOptions())
	for i, d := range benchDocs {
		for _, f := range d.Fields {
			dict.Insert(f, segment.DocID(i))
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, d := range benchDocs {
			for _, f := range d.Fields {
				// The trigram terms dictionary can return false postives so we may want to
				// consider verifying the results returned are matches to provide a more
				// fair comparison with the simple terms dictionary.
				dict.Fetch(f.Name, f.Value, termFetchOptions{false})
			}
		}
	}
}

func BenchmarkFetchRegex_SimpleTermsDictionary(b *testing.B) {
	dict := newSimpleTermsDictionary(NewOptions())
	for i, d := range benchDocs {
		for _, f := range d.Fields {
			dict.Insert(f, segment.DocID(i))
		}
	}

	var (
		name   = []byte("__name__")
		filter = []byte("node_netstat_Tcp_.*")
		opts   = termFetchOptions{true}
	)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dict.Fetch(name, filter, opts)
	}
}

func BenchmarkFetchRegex_TrigramTermsDictionary(b *testing.B) {
	dict := newTrigramTermsDictionary(NewOptions())
	for i, d := range benchDocs {
		for _, f := range d.Fields {
			dict.Insert(f, segment.DocID(i))
		}
	}

	var (
		name   = []byte("__name__")
		filter = []byte("node_netstat_Tcp_.*")
		opts   = termFetchOptions{true}
	)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// The trigram terms dictionary can return false postives so we may want to
		// consider verifying the results returned are matches to provide a more
		// fair comparison with the simple terms dictionary.
		dict.Fetch(name, filter, opts)
	}
}

func readDocuments(fn string, n int) ([]doc.Document, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		docs    []doc.Document
		scanner = bufio.NewScanner(f)
	)
	for scanner.Scan() && len(docs) < n {
		var fieldsMap map[string]string
		if err := json.Unmarshal(scanner.Bytes(), &fieldsMap); err != nil {
			return nil, err
		}

		fields := make([]doc.Field, 0, len(fieldsMap))
		for k, v := range fieldsMap {
			fields = append(fields, doc.Field{
				Name:  []byte(k),
				Value: doc.Value(v),
			})
		}
		docs = append(docs, doc.Document{
			Fields: fields,
		})
	}

	if len(docs) != n {
		return nil, fmt.Errorf("requested %d metrics but found %d", n, len(docs))
	}

	return docs, nil
}
