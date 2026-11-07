package libshim

import (
	"sync"
	"unsafe"

	"modernc.org/libc"
)

// Zig's libc allocator calls posix_memalign for alignments above what
// malloc guarantees, and then free, realloc and malloc_usable_size on the
// result. modernc.org/libc has no posix_memalign, so the generator renames
// these four calls to the ccgo_ functions below. An over-aligned block is
// carved out of a larger malloc block and remembered in aligned, so that
// the other three functions can tell it from a plain malloc block.

// alignedBlock records an over-aligned allocation: the malloc block it
// was carved out of and the size that was requested.
type alignedBlock struct {
	raw  uintptr
	size uint64
}

var (
	alignedMu sync.Mutex
	// aligned maps the aligned pointer handed to the caller to its block.
	aligned = map[uintptr]alignedBlock{}
)

// mallocAlign is the alignment modernc.org/libc's malloc guarantees.
const mallocAlign = 2 * uint64(unsafe.Sizeof(uintptr(0)))

// Xccgo_posix_memalign implements posix_memalign(memptr, alignment, size).
// alignment is a power of two; sizes are uint64 because the Zig code
// declares them as uintptr_t.
func Xccgo_posix_memalign(tls *libc.TLS, memptr uintptr, alignment, size uint64) int32 {
	if alignment <= mallocAlign {
		p := libc.Xmalloc(tls, libc.Tsize_t(size))
		if p == 0 {
			return libc.ENOMEM
		}
		*(*uintptr)(unsafe.Pointer(memptr)) = p
		return 0
	}
	raw := libc.Xmalloc(tls, libc.Tsize_t(size+alignment))
	if raw == 0 {
		return libc.ENOMEM
	}
	p := (raw + uintptr(alignment)) &^ uintptr(alignment-1)
	alignedMu.Lock()
	aligned[p] = alignedBlock{raw: raw, size: size}
	alignedMu.Unlock()
	*(*uintptr)(unsafe.Pointer(memptr)) = p
	return 0
}

func lookupAligned(p uintptr) (alignedBlock, bool) {
	alignedMu.Lock()
	b, ok := aligned[p]
	alignedMu.Unlock()
	return b, ok
}

// Xccgo_free implements free for blocks from malloc or
// Xccgo_posix_memalign.
func Xccgo_free(tls *libc.TLS, p uintptr) {
	if p == 0 {
		return
	}
	alignedMu.Lock()
	b, ok := aligned[p]
	if ok {
		delete(aligned, p)
	}
	alignedMu.Unlock()
	if ok {
		libc.Xfree(tls, b.raw)
		return
	}
	libc.Xfree(tls, p)
}

// Xccgo_malloc_usable_size implements malloc_usable_size. For an
// over-aligned block it reports the requested size; the Zig allocator only
// needs the result to be at least that.
func Xccgo_malloc_usable_size(tls *libc.TLS, p uintptr) uint64 {
	if b, ok := lookupAligned(p); ok {
		return b.size
	}
	return uint64(libc.Xmalloc_usable_size(tls, p))
}

// Xccgo_realloc implements realloc. An over-aligned block is copied to a
// plain malloc block; the Zig allocator only calls realloc for blocks whose
// alignment malloc guarantees.
func Xccgo_realloc(tls *libc.TLS, p uintptr, size uint64) uintptr {
	b, ok := lookupAligned(p)
	if !ok {
		return libc.Xrealloc(tls, p, libc.Tsize_t(size))
	}
	q := libc.Xmalloc(tls, libc.Tsize_t(size))
	if q == 0 {
		return 0
	}
	n := b.size
	if size < n {
		n = size
	}
	copy(unsafe.Slice((*byte)(unsafe.Pointer(q)), n), unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
	Xccgo_free(tls, p)
	return q
}
