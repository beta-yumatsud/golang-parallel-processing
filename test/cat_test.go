package main

import (
	"fmt"
	"os"
	"testing"
)

func seed(n int) []string {
	s := make([]string, 0, n)
	for i := 0; i < n; i++ {
		s = append(s, "a")
	}
	return s
}

func bench(b *testing.B, n int, f func(...string) string) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		f(seed(n)...)
	}
}

func BenchmarkCat3(b *testing.B) { bench(b, 3, cat) }
func BenchmarkBuf3(b *testing.B) { bench(b, 3, buf) }
func BenchmarkCat100(b *testing.B) { bench(b, 100, cat) }
func BenchmarkBuf100(b *testing.B) { bench(b, 100, buf) }
func BenchmarkCat10000(b *testing.B) { bench(b, 10000, cat) }
func BenchmarkBuf10000(b *testing.B) { bench(b, 10000, buf) }

func BenchmarkConcatenate(b *testing.B) {
	benchCases := []struct{
		name string
		n    int
		f    func(...string) string
	}{
		{"Cat", 3, cat},
		{"Buf", 3, buf},
		{"Cat", 100, cat},
		{"Buf", 100, buf},
		{"Cat", 10000, cat},
		{"Buf", 10000, buf},
	}

	for _, c := range benchCases {
		b.Run(fmt.Sprintf("%s%d", c.name, c.n),
			func(b *testing.B) {
				bench(b, c.n, c.f)
			})
	}
}

func TestMain(m *testing.M) {
	// DBの初期化とかに使えて良いみたい
	// setup()
	existCode := m.Run()
	// shutdown()
	os.Exit(existCode)
}