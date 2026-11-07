package libghostty

import (
	"testing"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// These benchmarks measure what the bindings add on top of the translated
// library: taking a TLS from the pool for a call, registering a callback
// handle, pinning a Go slice for the duration of a call, and the cost of
// a call that does nothing but cross into the translated code.

// BenchmarkTLSPool measures getTLS/putTLS, which every call into the
// library pays once.
func BenchmarkTLSPool(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		tls := getTLS()
		putTLS(tls)
	}
}

// BenchmarkHandle measures registering and removing a callback handle,
// which a writer or reader bridge pays once per operation.
func BenchmarkHandle(b *testing.B) {
	b.ReportAllocs()
	v := new(int)
	for b.Loop() {
		h := newHandle(v)
		h.delete()
	}
}

// BenchmarkPinBytes measures pinning a slice for a call, which every
// VTWrite pays.
func BenchmarkPinBytes(b *testing.B) {
	b.ReportAllocs()
	data := make([]byte, 64)
	for b.Loop() {
		_, unpin := pinBytes(data)
		unpin()
	}
}

// BenchmarkLibraryCall measures one call into the translated library that
// does almost no work: a terminal data getter for a scalar.
func BenchmarkLibraryCall(b *testing.B) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		b.Fatal(err)
	}
	defer term.Close()

	b.ReportAllocs()
	for b.Loop() {
		if _, err := term.CursorX(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLibraryCallRaw measures the same getter called directly on the
// translated function with a TLS held across iterations, so the
// difference to BenchmarkLibraryCall is the bindings' own overhead.
func BenchmarkLibraryCallRaw(b *testing.B) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		b.Fatal(err)
	}
	defer term.Close()

	tls := getTLS()
	defer putTLS(tls)
	n := int(unsafe.Sizeof(uint16(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)

	b.ReportAllocs()
	for b.Loop() {
		if r := vt.Xghostty_terminal_get(tls, term.ptr, vt.GHOSTTY_TERMINAL_DATA_CURSOR_X, out); r != vt.GHOSTTY_SUCCESS {
			b.Fatal(r)
		}
	}
}
