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

func BenchmarkTermsDict(b *testing.B) {
	benchmarks := []struct {
		name string
		fn   func(docs []doc.Document, b *testing.B)
	}{
		{
			name: "benchmark Insert with simple terms dictionary",
			fn:   benchmarkInsertSimpleTermsDict,
		},
		{
			name: "benchmark Insert with trigram terms dictionary",
			fn:   benchmarkInsertTrigramTermsDict,
		},
		{
			name: "benchmark MatchExact with simple terms dictionary",
			fn:   benchmarkMatchExactSimpleTermsDict,
		},
		{
			name: "benchmark MatchExact with trigram terms dictionary",
			fn:   benchmarkMatchExactTrigramTermsDict,
		},
		{
			name: "benchmark MatchRegex with simple terms dictionary",
			fn:   benchmarkMatchRegexSimpleTermsDict,
		},
		{
			name: "benchmark MatchRegex with trigram terms dictionary",
			fn:   benchmarkMatchRegexTrigramTermsDict,
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

func benchmarkInsertSimpleTermsDict(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		dict := newSimpleTermsDict(NewOptions())
		b.StartTimer()

		for i, d := range docs {
			for _, f := range d.Fields {
				dict.Insert(f, postings.ID(i))
			}
		}
	}
}

func benchmarkInsertTrigramTermsDict(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		dict := newTrigramTermsDict(NewOptions())
		b.StartTimer()

		for i, d := range docs {
			for _, f := range d.Fields {
				dict.Insert(f, postings.ID(i))
			}
		}
	}
}

func benchmarkMatchExactSimpleTermsDict(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	dict := newSimpleTermsDict(NewOptions())
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

func benchmarkMatchExactTrigramTermsDict(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	dict := newTrigramTermsDict(NewOptions())
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

func benchmarkMatchRegexSimpleTermsDict(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	dict := newSimpleTermsDict(NewOptions())
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

func benchmarkMatchRegexTrigramTermsDict(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	dict := newTrigramTermsDict(NewOptions())
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
