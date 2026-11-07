package libghostty

// Unicode display-width utilities from unicode.h.

import (
	"runtime"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// UnicodeCodepointWidth returns the terminal display width of cp in cells.
// It uses the same width table as libghostty's terminal text layout.
func UnicodeCodepointWidth(cp uint32) uint8 {
	tls := getTLS()
	defer putTLS(tls)
	return vt.Xghostty_unicode_codepoint_width(tls, cp)
}

// UnicodeGraphemeWidth measures the first grapheme cluster in codepoints.
// It returns the number of codepoints consumed and that cluster's display
// width in cells. Empty input returns zero for both values.
func UnicodeGraphemeWidth(codepoints []uint32) (consumed int, width uint8) {
	if len(codepoints) == 0 {
		return 0, 0
	}
	tls := getTLS()
	defer putTLS(tls)

	// The C API is read-only for this call, so the Go-owned slice is
	// passed without a copy, pinned so that it does not move.
	var pin runtime.Pinner
	pin.Pin(&codepoints[0])
	defer pin.Unpin()

	out := tls.Alloc(1)
	defer tls.Free(1)
	n := vt.Xghostty_unicode_grapheme_width(
		tls,
		uintptr(unsafe.Pointer(&codepoints[0])),
		uint64(len(codepoints)),
		out,
	)
	return int(n), *(*uint8)(unsafe.Pointer(out))
}
