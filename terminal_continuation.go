package libghostty

import (
	"io"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// Continuation returns a Go-owned copy of the replay-safe VT byte suffix
// needed to reconstruct the terminal's unfinished parser or UTF-8 decoder
// state. It returns an empty slice when the stream is at ground.
//
// Continuation tracking must have been enabled with
// [WithContinuationMaxBytes] or [Terminal.SetContinuationMaxBytes] before the
// input that produced an unfinished state was written. When tracking is
// disabled, this returns an error with result [ResultInvalidValue], even at
// ground.
// C: ghostty_terminal_continuation_alloc
func (t *Terminal) Continuation() ([]byte, error) {
	tls := getTLS()
	defer putTLS(tls)
	// Both out-parameters live in TLS memory: the data pointer first, the
	// length after it.
	results := tls.Alloc(16)
	defer tls.Free(16)
	if err := resultError(vt.Xghostty_terminal_continuation_alloc(
		tls,
		t.ptr,
		0,
		results,
		results+8,
	)); err != nil {
		return nil, err
	}
	outPtr := *(*uintptr)(unsafe.Pointer(results))
	outLen := *(*uint64)(unsafe.Pointer(results + 8))
	defer vt.Xghostty_free(tls, 0, outPtr, outLen)

	length, valid := ghosttySizeToInt(outLen)
	if !valid {
		return nil, &Error{Result: ResultLimitExceeded}
	}
	if length == 0 {
		return nil, nil
	}

	out := make([]byte, length)
	copy(out, unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), length))
	return out, nil
}

// ContinuationBuf copies the replay-safe VT continuation into buf and returns
// the number of bytes written. If buf is nil or too small, the returned count
// is the required size and the error has result [ResultOutOfSpace]. With
// tracking enabled, a stream at ground has an empty continuation that needs no
// space, so this returns 0 and no error, even for a nil buf. When tracking is
// disabled, this returns an error with result [ResultInvalidValue].
// C: ghostty_terminal_continuation_buf
func (t *Terminal) ContinuationBuf(buf []byte) (int, error) {
	tls := getTLS()
	defer putTLS(tls)
	var ptr uintptr
	if len(buf) > 0 {
		var unpin func()
		ptr, unpin = pinBytes(buf)
		defer unpin()
	}

	out := tls.Alloc(8)
	defer tls.Free(8)
	result := vt.Xghostty_terminal_continuation_buf(
		tls,
		t.ptr,
		ptr,
		uint64(len(buf)),
		out,
	)
	length, valid := ghosttySizeToInt(*(*uint64)(unsafe.Pointer(out)))
	if !valid {
		return 0, &Error{Result: ResultLimitExceeded}
	}
	return length, resultError(result)
}

// ContinuationWriteTo writes the replay-safe VT continuation directly to w.
// The writer is called synchronously and must not call methods on t. The
// returned count includes bytes accepted before an error.
// C: ghostty_terminal_continuation_write
func (t *Terminal) ContinuationWriteTo(w io.Writer) (int64, error) {
	bridge, err := newGhosttyWriter(w)
	if err != nil {
		return 0, err
	}
	defer bridge.close()

	tls := getTLS()
	defer putTLS(tls)
	result := vt.Xghostty_terminal_continuation_write(tls, t.ptr, bridge.cWriter())
	return bridge.written, resultErrorWithCallback(result, bridge.err)
}
