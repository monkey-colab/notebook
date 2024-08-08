package rdf2go

import (
	"fmt"
	"testing"
)

// BenchmarkGraphAdd benchmarks the Add function of the Graph
func BenchmarkGraphAdd(b *testing.B) {
	graph := NewGraph("http://example.org")

	for i := 0; i < b.N; i++ {
		s := NewResource(fmt.Sprintf("http://example.org/subject%d", i))
		p := NewResource(fmt.Sprintf("http://example.org/predicate%d", i))
		o := NewLiteral(fmt.Sprintf("object%d", i))

		graph.AddTriple(s, p, o)
	}
}

// BenchmarkGraphOne benchmarks the One function of the Graph
func BenchmarkGraphOne(b *testing.B) {
	graph := NewGraph("http://example.org")

	s := NewResource("http://example.org/subject")
	p := NewResource("http://example.org/predicate")
	for i := 0; i < 100; i++ {
		graph.AddTriple(s, p, NewLiteral(fmt.Sprintf("object%d", i)))
	}

	o := NewLiteral("object55")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		graph.One(s, p, o)
	}
}

// BenchmarkGraphAll benchmarks the All function of the Graph
func BenchmarkGraphAll(b *testing.B) {
	graph := NewGraph("http://example.org")

	s := NewResource("http://example.org/subject")
	p := NewResource("http://example.org/predicate")
	for i := 0; i < 100; i++ {
		graph.AddTriple(s, p, NewLiteral(fmt.Sprintf("object%d", i)))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		graph.All(s, p, nil)
	}
}

// BenchmarkGraphRemove benchmarks the Remove function of the Graph
func BenchmarkGraphRemove(b *testing.B) {
	graph := NewGraph("http://example.org")

	s := NewResource("http://example.org/subject")
	p := NewResource("http://example.org/predicate")
	o := NewLiteral("object")
	triple := NewTriple(s, p, o)
	graph.Add(triple)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		graph.Remove(triple)
		graph.Add(triple)
	}
}
