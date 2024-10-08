package main

import (
	"runtime"

	"github.com/alphadose/haxmap"
)

// Term interface (already implemented by BlankNode, Resource, and Literal in your code)
type Term interface {
	String() string
}

// weakTerm wraps a Term with a finalizer to remove it from the map when it's no longer used.
type weakTerm struct {
	term  Term
	final func()
}

// InternPool using haxmap
type InternPool struct {
	pool *haxmap.Map[string, *weakTerm]
}

// NewInternPool creates a new instance of the InternPool
func NewInternPool() *InternPool {
	return &InternPool{
		pool: haxmap.New[string, *weakTerm](),
	}
}

// Intern adds a Term to the pool if it's not already present, and returns the interned Term.
func (p *InternPool) Intern(t Term) Term {
	key := t.String()

	// Check if it already exists
	if val, ok := p.pool.Get(key); ok {
		return val.term
	}

	// Otherwise, add it with a finalizer
	wTerm := &weakTerm{
		term: t,
		final: func() {
			// Cleanup on finalization
			p.pool.Del(key)
		},
	}
	runtime.SetFinalizer(wTerm, func(wt *weakTerm) {
		wt.final()
	})
	p.pool.Set(key, wTerm)

	return t
}

func main() {
	// Create the intern pool
	pool := NewInternPool()

	// Assume BlankNode, Resource, and Literal structs are implemented elsewhere.
	// Intern a few terms (for demonstration purposes)
	b1 := pool.Intern(&BlankNode{ID: "1"})
	b2 := pool.Intern(&BlankNode{ID: "1"})

	// Output true because b1 and b2 are interned and should be the same instance
	println(b1 == b2)
}
