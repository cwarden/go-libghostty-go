package libghostty

// Synchronous Go I/O adapters for GhosttyReader and GhosttyWriter.
//
// libghostty invokes both callbacks only while the originating C API call is
// active. A handle (see ccgo.go) carries the Go reader or writer through C
// userdata without storing a Go pointer in C memory.

import (
	"errors"
	"fmt"
	"io"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// MIMEReaderFn streams the requested MIME representation to writer. The MIME
// string and writer are valid only for the duration of the callback. Return a
// non-nil error when the representation cannot be read or writer rejects a
// write.
//
// C: GhosttyMimeReaderFn
type MIMEReaderFn func(mime string, writer io.Writer) error

// ghosttyMIMEReaderBridge owns the handle used by a synchronous
// GhosttyMimeReader. It also preserves the original Go callback error so the
// outer operation can expose it through errors.Is and errors.As.
type ghosttyMIMEReaderBridge struct {
	reader MIMEReaderFn
	handle handle
	err    error
}

// cMIMEReader returns the C descriptor that routes reads to b.
func (b *ghosttyMIMEReaderBridge) cMIMEReader() vt.GhosttyMimeReader {
	return vt.GhosttyMimeReader{
		Fread:     vt.Fp(goGhosttyMIMEReaderTrampoline),
		Fuserdata: uintptr(b.handle),
	}
}

// close releases the MIME reader's handle after C can no longer invoke
// the callback.
func (b *ghosttyMIMEReaderBridge) close() {
	if b == nil || b.handle == 0 {
		return
	}
	b.handle.delete()
	b.handle = 0
	b.reader = nil
}

// ghosttyMIMEWriter exposes a borrowed GhosttyWriter as an io.Writer for the
// duration of one MIME reader callback. The writer is called with the TLS of
// the callback in progress.
type ghosttyMIMEWriter struct {
	tls    *libc.TLS
	writer vt.GhosttyWriter
	err    error
}

// Write passes all of p to the borrowed C writer. GhosttyWriter has an
// all-or-error contract, so a refusal is reported as a zero-byte write and an
// error even though the underlying destination may already hold a prefix.
func (w *ghosttyMIMEWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if w.err != nil {
		return 0, w.err
	}

	data, unpin := pinBytes(p)
	defer unpin()
	write := *(*func(*libc.TLS, uintptr, uintptr, uint64) uint8)(unsafe.Pointer(&struct{ uintptr }{w.writer.Fwrite}))
	if write(w.tls, w.writer.Fuserdata, data, uint64(len(p))) == 0 {
		w.err = errors.New("ghostty: MIME writer refused write")
		return 0, w.err
	}
	return len(p), nil
}

// ghosttyReaderBridge owns the handle installed in a GhosttyReader. The
// decoder retains that C descriptor, so this bridge must live until the
// decoder is closed even though each individual read is synchronous.
type ghosttyReaderBridge struct {
	reader     io.Reader
	handle     handle
	err        error
	pendingErr error
	eof        bool
}

// cReader returns the C descriptor that routes reads to b.
func (b *ghosttyReaderBridge) cReader() vt.GhosttyReader {
	return vt.GhosttyReader{
		Fread:     vt.Fp(goGhosttyReaderTrampoline),
		Fuserdata: uintptr(b.handle),
	}
}

// close releases the reader's handle. It must be called only after C can
// no longer invoke the callback.
func (b *ghosttyReaderBridge) close() {
	if b == nil || b.handle == 0 {
		return
	}
	b.handle.delete()
	b.handle = 0
	b.reader = nil
}

// ghosttyWriterBridge adapts an io.Writer to GhosttyWriter. It records the
// accepted byte count and the original Go error, if any.
type ghosttyWriterBridge struct {
	writer  io.Writer
	handle  handle
	written int64
	err     error
}

// newGhosttyWriter builds a writer bridge backed by w. The caller passes the
// descriptor from cWriter to libghostty, and must close the bridge after the
// synchronous libghostty operation returns.
func newGhosttyWriter(w io.Writer) (*ghosttyWriterBridge, error) {
	b := &ghosttyWriterBridge{}
	if err := b.reset(w); err != nil {
		return nil, err
	}
	return b, nil
}

// reset prepares b to write to w. It reuses the existing handle when
// possible. Calls using the same bridge must be serialized.
func (b *ghosttyWriterBridge) reset(w io.Writer) error {
	if w == nil {
		return &Error{Result: ResultInvalidValue}
	}

	b.writer = w
	b.written = 0
	b.err = nil
	if b.handle == 0 {
		b.handle = newHandle(b)
	}
	return nil
}

// cWriter returns the C descriptor that routes writes to b.
func (b *ghosttyWriterBridge) cWriter() vt.GhosttyWriter {
	return vt.GhosttyWriter{
		Fwrite:    vt.Fp(goGhosttyWriterTrampoline),
		Fuserdata: uintptr(b.handle),
	}
}

// close releases the writer's handle.
func (b *ghosttyWriterBridge) close() {
	if b == nil || b.handle == 0 {
		return
	}
	b.handle.delete()
	b.handle = 0
	b.writer = nil
}

// resultErrorWithCallback preserves libghostty's I/O result while also
// exposing the original Go reader or writer error through errors.Is/As.
func resultErrorWithCallback(result vt.GhosttyResult, callbackErr error) error {
	err := resultError(result)
	if result != vt.GHOSTTY_IO_ERROR || callbackErr == nil {
		return err
	}
	return errors.Join(err, callbackErr)
}

// goGhosttyMIMEReaderTrampoline is the GhosttyMimeReaderFn installed by
// cMIMEReader.
func goGhosttyMIMEReaderTrampoline(
	tls *libc.TLS,
	userdata uintptr,
	mime vt.GhosttyString,
	writer vt.GhosttyWriter,
) (ok uint8) {
	b := handle(userdata).value().(*ghosttyMIMEReaderBridge)

	// A panic must never unwind into libghostty, which cannot recover from
	// it. Preserve it as the callback error and make the originating
	// operation return IO_ERROR.
	defer func() {
		if recovered := recover(); recovered != nil {
			b.err = fmt.Errorf("ghostty MIME reader panic: %v", recovered)
			ok = cFalse
		}
	}()

	mimeBytes, valid := copyGhosttyString(mime)
	if !valid {
		b.err = &Error{Result: ResultInvalidValue}
		return cFalse
	}

	w := &ghosttyMIMEWriter{tls: tls, writer: writer}
	if err := b.reader(string(mimeBytes), w); err != nil {
		b.err = err
		return cFalse
	}
	if w.err != nil {
		// Treat a writer error as fatal even if a callback accidentally ignored
		// the error returned by io.Writer.Write.
		b.err = w.err
		return cFalse
	}

	return cTrue
}

// goGhosttyReaderTrampoline is the GhosttyReaderFn installed by cReader.
func goGhosttyReaderTrampoline(
	_ *libc.TLS,
	userdata uintptr,
	buffer uintptr,
	capacity uint64,
	outRead uintptr,
) (ok uint8) {
	b := handle(userdata).value().(*ghosttyReaderBridge)
	*(*uint64)(unsafe.Pointer(outRead)) = 0

	// A panic must never unwind into libghostty. Convert it to the same
	// fatal read failure that an ordinary Go reader error would produce.
	defer func() {
		if recovered := recover(); recovered != nil {
			b.err = fmt.Errorf("ghostty reader panic: %v", recovered)
			ok = cFalse
		}
	}()

	if b.pendingErr != nil {
		b.err = b.pendingErr
		b.pendingErr = nil
		return cFalse
	}
	if b.eof {
		return cTrue
	}

	length, valid := ghosttySizeToInt(capacity)
	if !valid || length <= 0 || buffer == 0 {
		b.err = &Error{Result: ResultLimitExceeded}
		return cFalse
	}

	n, err := b.reader.Read(unsafe.Slice((*byte)(unsafe.Pointer(buffer)), length))
	if n < 0 || n > length {
		b.err = fmt.Errorf("ghostty reader returned invalid byte count %d for capacity %d", n, length)
		return cFalse
	}
	*(*uint64)(unsafe.Pointer(outRead)) = uint64(n)

	if err == nil {
		// GhosttyReader defines a successful zero-byte read as permanent EOF.
		b.eof = n == 0
		return cTrue
	}
	if errors.Is(err, io.EOF) {
		// Deliver bytes returned with io.EOF first, then report definitive EOF
		// without consulting the Go reader again.
		b.eof = true
		return cTrue
	}
	if n > 0 {
		// Go permits data and an error in the same read. Deliver the data now
		// and report the fatal error on libghostty's next callback.
		b.pendingErr = err
		return cTrue
	}

	b.err = err
	return cFalse
}

// goGhosttyWriterTrampoline is the GhosttyWriterFn installed by cWriter.
func goGhosttyWriterTrampoline(
	_ *libc.TLS,
	userdata uintptr,
	data uintptr,
	length uint64,
) (ok uint8) {
	b := handle(userdata).value().(*ghosttyWriterBridge)

	// A panic must never unwind into libghostty. Preserve it as a Go error
	// and make libghostty return GHOSTTY_IO_ERROR.
	defer func() {
		if recovered := recover(); recovered != nil {
			b.err = fmt.Errorf("ghostty writer panic: %v", recovered)
			ok = cFalse
		}
	}()

	count, valid := ghosttySizeToInt(length)
	if !valid || count <= 0 || data == 0 {
		b.err = &Error{Result: ResultLimitExceeded}
		return cFalse
	}
	remaining := unsafe.Slice((*byte)(unsafe.Pointer(data)), count)

	for len(remaining) > 0 {
		n, err := b.writer.Write(remaining)
		if n < 0 || n > len(remaining) {
			b.err = fmt.Errorf("ghostty writer returned invalid byte count %d for length %d", n, len(remaining))
			return cFalse
		}

		const maxInt64 = int64(^uint64(0) >> 1)
		if b.written > maxInt64-int64(n) {
			b.err = &Error{Result: ResultLimitExceeded}
			return cFalse
		}
		b.written += int64(n)
		remaining = remaining[n:]

		if err != nil {
			b.err = err
			return cFalse
		}
		if n == 0 {
			b.err = io.ErrShortWrite
			return cFalse
		}
	}

	return cTrue
}
