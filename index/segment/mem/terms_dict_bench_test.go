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
	benchField    = []byte("__name__")
	benchRegexp   = []byte("node_netstat_Tcp_.*")
	benchCompiled = regexp.MustCompile(string(benchRegexp))
)

func BenchmarkTermsDict(b *testing.B) {
	benchmarks := []struct {
		name string
		fn   func(docs []doc.Document, b *testing.B)
	}{
		{
			name: "benchmark Insert",
			fn:   benchmarkTermsDictInsert,
		},
		{
			name: "benchmark MatchTerm",
			fn:   benchmarkTermsDictMatchTerm,
		},
		{
			name: "benchmark MatchRegex",
			fn:   benchmarkTermsDictMatchRegex,
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

func benchmarkTermsDictInsert(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		dict := newTermsDict(NewOptions())
		b.StartTimer()

		for i, d := range docs {
			for _, f := range d.Fields {
				dict.Insert(f, postings.ID(i))
			}
		}
	}
}

func benchmarkTermsDictMatchTerm(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	dict := newTermsDict(NewOptions())
	for i, d := range docs {
		for _, f := range d.Fields {
			dict.Insert(f, postings.ID(i))
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, d := range docs {
			for _, f := range d.Fields {
				dict.MatchTerm(f.Name, f.Value)
			}
		}
	}
}

func benchmarkTermsDictMatchRegex(docs []doc.Document, b *testing.B) {
	b.ReportAllocs()

	dict := newTermsDict(NewOptions())
	for i, d := range docs {
		for _, f := range d.Fields {
			dict.Insert(f, postings.ID(i))
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dict.MatchRegexp(benchField, benchRegexp, benchCompiled)
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
