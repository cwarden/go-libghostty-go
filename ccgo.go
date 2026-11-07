package libghostty

// Support for calling the translated libghostty-vt in internal/vt.
//
// Every translated function takes a *libc.TLS, which holds the C stack of
// the call. One TLS must not be used by two calls at the same time, so the
// bindings take one from a pool for the duration of each call. Callbacks
// that libghostty makes into Go receive the TLS of the call in progress
// and pass it on when they call back into libghostty.
//
// libghostty stores the userdata of a callback in C memory, which the
// garbage collector does not scan, so a Go value is never stored there.
// The bindings register the value in a table and store the integer key.

import (
	"runtime"
	"sync"
	"unsafe"

	"modernc.org/libc"
)

// C bool values as the translated code represents them.
const (
	cFalse uint8 = 0
	cTrue  uint8 = 1
)

func cBool(b bool) uint8 {
	if b {
		return cTrue
	}
	return cFalse
}

// goBytes copies n bytes of C memory at p into a new Go slice. It takes
// the place of C.GoBytes; libc.GoBytes returns a view of the C memory,
// which is wrong for memory the library frees after the call.
func goBytes(p uintptr, n int) []byte {
	if n == 0 {
		return []byte{}
	}
	return append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(p)), n)...)
}

// pinBytes returns the address of the first byte of p for passing to the
// translated code, and a function that releases it. p must not be empty.
//
// The translated code runs on the goroutine's stack, and the runtime can
// move that stack while it runs; a uintptr to a Go stack variable would
// then point at the old copy. Pinning forces p's array onto the heap,
// where the collector does not move it, and keeps it there until unpin
// is called.
func pinBytes(p []byte) (addr uintptr, unpin func()) {
	var pin runtime.Pinner
	pin.Pin(&p[0])
	return uintptr(unsafe.Pointer(&p[0])), pin.Unpin
}

var tlsPool = sync.Pool{New: func() any { return libc.NewTLS() }}

// getTLS returns a TLS for one call into libghostty. Return it with putTLS
// when the call has returned.
func getTLS() *libc.TLS {
	return tlsPool.Get().(*libc.TLS)
}

func putTLS(tls *libc.TLS) {
	tlsPool.Put(tls)
}

// handle is the key of a Go value registered for use as C userdata.
type handle uintptr

var handles struct {
	sync.Mutex
	next   handle
	values map[handle]any
}

// newHandle registers v and returns its key. Keys start at 1, so a zero
// handle means none.
func newHandle(v any) handle {
	handles.Lock()
	defer handles.Unlock()
	if handles.values == nil {
		handles.values = map[handle]any{}
	}
	handles.next++
	handles.values[handles.next] = v
	return handles.next
}

// value returns the registered value, and panics for a handle that is not
// registered, as a callback with stale userdata is a bug in the bindings.
func (h handle) value() any {
	handles.Lock()
	defer handles.Unlock()
	v, ok := handles.values[h]
	if !ok {
		panic("libghostty: unknown callback handle")
	}
	return v
}

// delete unregisters the handle. It must be called only after libghostty
// can no longer invoke a callback with it.
func (h handle) delete() {
	handles.Lock()
	defer handles.Unlock()
	delete(handles.values, h)
}
