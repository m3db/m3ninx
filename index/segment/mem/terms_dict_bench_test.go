package mem

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/postings"
)

var (
	benchMatchRegexName    = []byte("__name__")
	benchMatchRegexPattern = []byte("node_netstat_Tcp_.*")
	benchMatchRegexRE      = regexp.MustCompile(string(benchMatchRegexPattern))
)

func BenchmarkTermsDictionary(b *testing.B) {
	benchmarks := []struct {
		name string
		fn   func(docs []doc.Document, b *testing.B)
	}{
		{
			name: "benchmark Insert with simple terms dictionary",
			fn:   benchmarkInsertSimpleTermsDictionary,
		},
		{
			name: "benchmark Insert with trigram terms dictionary",
			fn:   benchmarkInsertTrigramTermsDictionary,
		},
		{
			name: "benchmark MatchExact with simple terms dictionary",
			fn:   benchmarkMatchExactSimpleTermsDictionary,
		},
		{
			name: "benchmark MatchExact with trigram terms dictionary",
			fn:   benchmarkMatchExactTrigramTermsDictionary,
		},
		{
			name: "benchmark MatchRegex with simple terms dictionary",
			fn:   benchmarkMatchRegexSimpleTermsDictionary,
		},
		{
			name: "benchmark MatchRegex with trigram terms dictionary",
			fn:   benchmarkMatchRegexTrigramTermsDictionary,
		},
	}

	docs, err := readDocuments("../../../testdata/node_exporter.json", 2000)
	if err != nil {
		b.Fatalf("unable to read documents for benchmarks: %v", err)
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			bm.fn(docs, b)
		})
	}
}

func benchmarkInsertSimpleTermsDictionary(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		dict := newSimpleTermsDictionary(NewOptions())
		b.StartTimer()

		for i, d := range docs {
			for _, f := range d.Fields {
				dict.Insert(f, postings.ID(i))
			}
		}
	}
}

func benchmarkInsertTrigramTermsDictionary(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		dict := newTrigramTermsDictionary(NewOptions())
		b.StartTimer()

		for i, d := range docs {
			for _, f := range d.Fields {
				dict.Insert(f, postings.ID(i))
			}
		}
	}
}

func benchmarkMatchExactSimpleTermsDictionary(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	dict := newSimpleTermsDictionary(NewOptions())
	for i, d := range docs {
		for _, f := range d.Fields {
			dict.Insert(f, postings.ID(i))
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, d := range docs {
			for _, f := range d.Fields {
				dict.MatchExact(f.Name, f.Value)
			}
		}
	}
}

func benchmarkMatchExactTrigramTermsDictionary(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	dict := newTrigramTermsDictionary(NewOptions())
	for i, d := range docs {
		for _, f := range d.Fields {
			dict.Insert(f, postings.ID(i))
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, d := range docs {
			for _, f := range d.Fields {
				// The trigram terms dictionary can return false postives so we may want to
				// consider verifying the results returned are matches to provide a more
				// fair comparison with the simple terms dictionary.
				dict.MatchExact(f.Name, f.Value)
			}
		}
	}
}

func benchmarkMatchRegexSimpleTermsDictionary(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	dict := newSimpleTermsDictionary(NewOptions())
	for i, d := range docs {
		for _, f := range d.Fields {
			dict.Insert(f, postings.ID(i))
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dict.MatchRegex(benchMatchRegexName, benchMatchRegexPattern, benchMatchRegexRE)
	}
}

func benchmarkMatchRegexTrigramTermsDictionary(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	dict := newTrigramTermsDictionary(NewOptions())
	for i, d := range docs {
		for _, f := range d.Fields {
			dict.Insert(f, postings.ID(i))
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// The trigram terms dictionary can return false postives so we may want to
		// consider verifying the results returned are matches to provide a more
		// fair comparison with the simple terms dictionary.
		dict.MatchRegex(benchMatchRegexName, benchMatchRegexPattern, benchMatchRegexRE)
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
				Value: []byte(v),
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
