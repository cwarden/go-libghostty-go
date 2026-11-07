package libghostty

// Color scheme report encoding from color_scheme.h.

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// encodeInto calls an encoder that writes into buf and stores the number
// of bytes written or needed in a size_t out-parameter. It returns the
// result and that count. encode receives the TLS, the buffer address and
// its length, and the address of the count.
func encodeInto(buf []byte, encode func(tls *libc.TLS, buf uintptr, bufLen uint64, written uintptr) vt.GhosttyResult) (vt.GhosttyResult, int) {
	tls := getTLS()
	defer putTLS(tls)

	// The count lives in TLS memory; the buffer is a Go slice, pinned so
	// that it does not move during the call.
	const n = int(unsafe.Sizeof(uint64(0)))
	written := tls.Alloc(n)
	defer tls.Free(n)
	*(*uint64)(unsafe.Pointer(written)) = 0

	var addr uintptr
	if len(buf) > 0 {
		var unpin func()
		addr, unpin = pinBytes(buf)
		defer unpin()
	}
	result := encode(tls, addr, uint64(len(buf)), written)
	return result, int(*(*uint64)(unsafe.Pointer(written)))
}

// ColorSchemeReportEncode encodes a color scheme report escape sequence.
func ColorSchemeReportEncode(scheme ColorScheme) ([]byte, error) {
	encode := func(tls *libc.TLS, buf uintptr, bufLen uint64, written uintptr) vt.GhosttyResult {
		return vt.Xghostty_color_scheme_report_encode(
			tls,
			vt.GhosttyColorScheme(scheme),
			buf,
			bufLen,
			written,
		)
	}

	buf := make([]byte, 32)
	result, written := encodeInto(buf, encode)
	if result == vt.GHOSTTY_SUCCESS {
		return append([]byte(nil), buf[:written]...), nil
	}
	if result != vt.GHOSTTY_OUT_OF_SPACE {
		return nil, &Error{Result: Result(result)}
	}

	out := make([]byte, written)
	result, outWritten := encodeInto(out, encode)
	if err := resultError(result); err != nil {
		return nil, err
	}
	return out[:outWritten], nil
}
