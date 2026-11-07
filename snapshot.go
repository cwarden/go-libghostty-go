package libghostty

import (
	"errors"
	"io"
	"runtime"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// SnapshotDecoderOption identifies a configurable snapshot decoder option.
// C: GhosttySnapshotDecoderOption
type SnapshotDecoderOption int

const (
	// SnapshotDecoderOptionMaxContinuationBytes is the largest non-ground VT
	// continuation the decoder will accept (size_t). When continuation
	// retention is enabled, this is also the tracking limit on the returned
	// terminal.
	SnapshotDecoderOptionMaxContinuationBytes SnapshotDecoderOption = vt.GHOSTTY_SNAPSHOT_DECODER_OPT_MAX_CONTINUATION_BYTES

	// SnapshotDecoderOptionRetainContinuation controls whether the returned
	// terminal retains the decoded continuation (bool).
	SnapshotDecoderOptionRetainContinuation SnapshotDecoderOption = vt.GHOSTTY_SNAPSHOT_DECODER_OPT_RETAIN_CONTINUATION

	// SnapshotDecoderOptionCompressHistory controls whether the decoder
	// compresses scrollback as it restores it (bool). See
	// [SnapshotDecoder.SetCompressHistory].
	SnapshotDecoderOptionCompressHistory SnapshotDecoderOption = vt.GHOSTTY_SNAPSHOT_DECODER_OPT_COMPRESS_HISTORY

	// SnapshotDecoderOptionKittyImageStorageLimit is the maximum number of
	// bytes of Kitty graphics images kept for each restored screen
	// (uint64_t). See [SnapshotDecoder.SetKittyImageStorageLimit].
	SnapshotDecoderOptionKittyImageStorageLimit SnapshotDecoderOption = vt.GHOSTTY_SNAPSHOT_DECODER_OPT_KITTY_IMAGE_STORAGE_LIMIT
)

// SnapshotDecoderProgressKind is the kind of record that the most recent
// successful call to [SnapshotDecoder.Next] decoded.
// C: GhosttySnapshotDecoderProgressKind
type SnapshotDecoderProgressKind int

const (
	// SnapshotDecoderProgressKindPage is a page of scrollback history,
	// along with the Kitty graphics placements on that page. The Rows and
	// Remaining fields of [SnapshotDecoderProgress] describe it.
	SnapshotDecoderProgressKindPage SnapshotDecoderProgressKind = vt.GHOSTTY_SNAPSHOT_DECODER_PROGRESS_KIND_PAGE

	// SnapshotDecoderProgressKindImage is the pixel data of one Kitty
	// graphics image or animation frame. The ImageID, ImageFrame, and
	// ImageBytes fields of [SnapshotDecoderProgress] describe it.
	SnapshotDecoderProgressKindImage SnapshotDecoderProgressKind = vt.GHOSTTY_SNAPSHOT_DECODER_PROGRESS_KIND_IMAGE
)

// SnapshotDecoderData identifies a data field available from a
// [SnapshotDecoder].
// C: GhosttySnapshotDecoderData
type SnapshotDecoderData int

const (
	// SnapshotDecoderDataInvalid is an invalid or sentinel query.
	SnapshotDecoderDataInvalid SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_INVALID

	// SnapshotDecoderDataMaxContinuationBytes is the decoder's current
	// continuation input limit (size_t).
	SnapshotDecoderDataMaxContinuationBytes SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_MAX_CONTINUATION_BYTES

	// SnapshotDecoderDataSourceOffset is the number of source bytes consumed
	// so far (size_t).
	SnapshotDecoderDataSourceOffset SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_SOURCE_OFFSET

	// SnapshotDecoderDataHistoryRowsPrimary is the complete logical primary
	// screen history extent (uint64_t).
	SnapshotDecoderDataHistoryRowsPrimary SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_HISTORY_ROWS_PRIMARY

	// SnapshotDecoderDataHistoryRowsAlternate is the complete logical
	// alternate screen history extent (uint64_t).
	SnapshotDecoderDataHistoryRowsAlternate SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_HISTORY_ROWS_ALTERNATE

	// SnapshotDecoderDataProgressScreen is the screen that the most recently
	// decoded history page or image belongs to (GhosttyTerminalScreen).
	SnapshotDecoderDataProgressScreen SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_SCREEN

	// SnapshotDecoderDataProgressRows is the number of rows prepended by the
	// most recently decoded history page (size_t). It is only available
	// for [SnapshotDecoderProgressKindPage].
	SnapshotDecoderDataProgressRows SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_ROWS

	// SnapshotDecoderDataProgressRemaining is the number of page records
	// remaining in the current screen's history sequence (uint32_t). It is
	// only available for [SnapshotDecoderProgressKindPage].
	SnapshotDecoderDataProgressRemaining SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_REMAINING

	// SnapshotDecoderDataRetainContinuation reports whether continuation
	// tracking is retained on returned terminals (bool).
	SnapshotDecoderDataRetainContinuation SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_RETAIN_CONTINUATION

	// SnapshotDecoderDataCompressHistory reports whether the decoder
	// compresses scrollback as it restores it (bool). See
	// [SnapshotDecoder.SetCompressHistory].
	SnapshotDecoderDataCompressHistory SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_COMPRESS_HISTORY

	// SnapshotDecoderDataKittyImageStorageLimit is the maximum number of
	// bytes of Kitty graphics images kept for each restored screen
	// (uint64_t). See [SnapshotDecoder.SetKittyImageStorageLimit].
	SnapshotDecoderDataKittyImageStorageLimit SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_KITTY_IMAGE_STORAGE_LIMIT

	// SnapshotDecoderDataProgressKind is the kind of record that the most
	// recent successful call to Next decoded
	// (GhosttySnapshotDecoderProgressKind). It is available whenever
	// [SnapshotDecoderDataProgressScreen] is.
	SnapshotDecoderDataProgressKind SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_KIND

	// SnapshotDecoderDataProgressImageID is the ID of the Kitty graphics
	// image whose pixel data the most recent call to Next decoded
	// (uint32_t). It is only available for
	// [SnapshotDecoderProgressKindImage].
	SnapshotDecoderDataProgressImageID SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_IMAGE_ID

	// SnapshotDecoderDataProgressImageFrame is the frame number of that
	// pixel data (uint32_t). Frame 1 is the image itself, and frames 2 and
	// up are animation frames. It is only available for
	// [SnapshotDecoderProgressKindImage].
	SnapshotDecoderDataProgressImageFrame SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_IMAGE_FRAME

	// SnapshotDecoderDataProgressImageBytes is the number of bytes of pixel
	// data that the most recent call to Next added to its image (size_t).
	// Zero means the data was read but discarded. It is only available for
	// [SnapshotDecoderProgressKindImage].
	SnapshotDecoderDataProgressImageBytes SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_IMAGE_BYTES

	// SnapshotDecoderDataKittyImagesPrimary is the number of Kitty graphics
	// images restored on the primary screen (uint32_t). See
	// [SnapshotDecoder.KittyImageCount].
	SnapshotDecoderDataKittyImagesPrimary SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_KITTY_IMAGES_PRIMARY

	// SnapshotDecoderDataKittyImagesAlternate is the number of Kitty
	// graphics images restored on the alternate screen (uint32_t). See
	// [SnapshotDecoder.KittyImageCount].
	SnapshotDecoderDataKittyImagesAlternate SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_KITTY_IMAGES_ALTERNATE

	// SnapshotDecoderDataKittyPendingBytesPrimary is the number of bytes of
	// Kitty graphics pixel data still to come for the primary screen when
	// Ready returned (uint64_t). See
	// [SnapshotDecoder.KittyImagePendingBytes].
	SnapshotDecoderDataKittyPendingBytesPrimary SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_KITTY_PENDING_BYTES_PRIMARY

	// SnapshotDecoderDataKittyPendingBytesAlternate is the number of bytes
	// of Kitty graphics pixel data still to come for the alternate screen
	// when Ready returned (uint64_t). See
	// [SnapshotDecoder.KittyImagePendingBytes].
	SnapshotDecoderDataKittyPendingBytesAlternate SnapshotDecoderData = vt.GHOSTTY_SNAPSHOT_DECODER_DATA_KITTY_PENDING_BYTES_ALTERNATE
)

// SnapshotDecoder incrementally decodes and validates one terminal
// snapshot. It is stateful, not safe for concurrent use, and must be closed.
//
// A terminal returned by [SnapshotDecoder.Ready] must remain alive until the
// decoder reaches FINISH or is closed. Closing the decoder does not close any
// returned terminal.
// C: GhosttySnapshotDecoder
type SnapshotDecoder struct {
	ptr uintptr

	// reader is retained for the lifetime of a callback-backed decoder because
	// the C decoder stores its GhosttyReader descriptor by value.
	reader *ghosttyReaderBridge

	// sourceBytes and sourcePin keep the zero-copy input used by
	// NewSnapshotDecoderBytes alive, immovable, and therefore legal for
	// libghostty to retain across calls. The pin is released at FINISH or
	// decoder destruction.
	sourceBytes  []byte
	sourcePin    runtime.Pinner
	sourcePinned bool

	// sourcePtr is the C-owned input used by NewSnapshotDecoderBytesCopy.
	// libghostty retains this borrowed pointer until FINISH or decoder
	// destruction.
	sourcePtr unsafe.Pointer
	sourceLen uintptr
}

// snapshotReaderDecoder stores a decoder and its reader bridge in one
// allocation. The bridge remains separately addressable for use as callback
// data.
type snapshotReaderDecoder struct {
	decoder SnapshotDecoder
	reader  ghosttyReaderBridge
}

// Snapshot encodes a complete terminal snapshot and returns a Go-owned copy.
// Calls must be serialized with every other operation on t. Callers taking
// repeated snapshots can use [Terminal.SnapshotBuf] with a reusable buffer to
// avoid the result allocation and copy.
// C: ghostty_snapshot_encode_alloc
func (t *Terminal) Snapshot() ([]byte, error) {
	tls := getTLS()
	defer putTLS(tls)

	// Two out-parameters: the buffer pointer and its length.
	const n = 2 * int(unsafe.Sizeof(uintptr(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)
	outPtr := out
	outLen := out + unsafe.Sizeof(uintptr(0))
	*(*uintptr)(unsafe.Pointer(outPtr)) = 0
	*(*uint64)(unsafe.Pointer(outLen)) = 0
	if err := resultError(vt.Xghostty_snapshot_encode_alloc(tls, t.ptr, 0, outPtr, outLen)); err != nil {
		return nil, err
	}
	ptr := *(*uintptr)(unsafe.Pointer(outPtr))
	size := *(*uint64)(unsafe.Pointer(outLen))
	defer vt.Xghostty_free(tls, 0, ptr, size)

	length, valid := ghosttySizeToInt(size)
	if !valid {
		return nil, &Error{Result: ResultLimitExceeded}
	}
	if length == 0 {
		return nil, nil
	}

	outBytes := make([]byte, length)
	copy(outBytes, unsafe.Slice((*byte)(unsafe.Pointer(ptr)), length))
	return outBytes, nil
}

// SnapshotBuf encodes a complete terminal snapshot into buf and returns the
// number of bytes written. If buf is nil or too small, the returned count is
// the required capacity and the error has result [ResultOutOfSpace].
// C: ghostty_snapshot_encode_buf
func (t *Terminal) SnapshotBuf(buf []byte) (int, error) {
	tls := getTLS()
	defer putTLS(tls)

	var ptr uintptr
	if len(buf) > 0 {
		addr, unpin := pinBytes(buf)
		defer unpin()
		ptr = addr
	}

	const n = int(unsafe.Sizeof(uint64(0)))
	written := tls.Alloc(n)
	defer tls.Free(n)
	*(*uint64)(unsafe.Pointer(written)) = 0
	result := vt.Xghostty_snapshot_encode_buf(
		tls,
		t.ptr,
		ptr,
		uint64(len(buf)),
		written,
	)
	length, valid := ghosttySizeToInt(*(*uint64)(unsafe.Pointer(written)))
	if !valid {
		return 0, &Error{Result: ResultLimitExceeded}
	}
	return length, resultError(result)
}

// SnapshotWriteTo encodes a complete terminal snapshot directly to w. The
// writer is called synchronously and must not call methods on t. The returned
// count includes bytes accepted before an error.
// C: ghostty_snapshot_encode
func (t *Terminal) SnapshotWriteTo(w io.Writer) (int64, error) {
	bridge, err := newGhosttyWriter(w)
	if err != nil {
		return 0, err
	}
	defer bridge.close()

	tls := getTLS()
	defer putTLS(tls)
	result := vt.Xghostty_snapshot_encode(tls, t.ptr, bridge.cWriter())
	return bridge.written, resultErrorWithCallback(result, bridge.err)
}

// NewSnapshotDecoder creates a snapshot decoder backed by r. Reads are
// synchronous and occur only during Ready, Next, or Decode. A successful
// zero-byte read is permanent EOF; nonblocking readers must wait internally.
// C: ghostty_snapshot_decoder_new
func NewSnapshotDecoder(r io.Reader) (*SnapshotDecoder, error) {
	if r == nil {
		return nil, &Error{Result: ResultInvalidValue}
	}
	holder := &snapshotReaderDecoder{}
	holder.reader.reader = r
	holder.reader.handle = newHandle(&holder.reader)

	tls := getTLS()
	defer putTLS(tls)
	const n = int(unsafe.Sizeof(uintptr(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)
	*(*uintptr)(unsafe.Pointer(out)) = 0
	result := vt.Xghostty_snapshot_decoder_new(tls, 0, out, holder.reader.cReader())
	if err := resultError(result); err != nil {
		holder.reader.close()
		return nil, err
	}

	holder.decoder.ptr = *(*uintptr)(unsafe.Pointer(out))
	holder.decoder.reader = &holder.reader
	return &holder.decoder, nil
}

// snapshotDecoderNewBuf creates a decoder over the C-visible input at ptr.
func snapshotDecoderNewBuf(ptr uintptr, length uint64) (uintptr, error) {
	tls := getTLS()
	defer putTLS(tls)
	const n = int(unsafe.Sizeof(uintptr(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)
	*(*uintptr)(unsafe.Pointer(out)) = 0
	if err := resultError(vt.Xghostty_snapshot_decoder_new_buf(tls, 0, out, ptr, length)); err != nil {
		return 0, err
	}
	return *(*uintptr)(unsafe.Pointer(out)), nil
}

// NewSnapshotDecoderBytes creates a zero-copy snapshot decoder over data. The
// decoder pins the byte slice because libghostty retains its source pointer
// across calls. The caller must not modify data until the decoder reaches
// FINISH through [SnapshotDecoder.Decode] or [SnapshotDecoder.Next], or until
// the decoder is closed.
// C: ghostty_snapshot_decoder_new_buf
func NewSnapshotDecoderBytes(data []byte) (*SnapshotDecoder, error) {
	d := &SnapshotDecoder{sourceBytes: data}
	var source uintptr
	if len(data) > 0 {
		d.sourcePin.Pin(&data[0])
		d.sourcePinned = true
		source = uintptr(unsafe.Pointer(&data[0]))
	}

	ptr, err := snapshotDecoderNewBuf(source, uint64(len(data)))
	runtime.KeepAlive(data)
	if err != nil {
		d.releaseSource()
		return nil, err
	}

	d.ptr = ptr
	return d, nil
}

// NewSnapshotDecoderBytesCopy creates a snapshot decoder over a private copy
// of data. Unlike [NewSnapshotDecoderBytes], the caller may modify or release
// data immediately after this function returns. Prefer the zero-copy form for
// immutable snapshot input, especially for large histories.
// C: ghostty_snapshot_decoder_new_buf
func NewSnapshotDecoderBytesCopy(data []byte) (*SnapshotDecoder, error) {
	length := uintptr(len(data))
	var source unsafe.Pointer
	if length > 0 {
		source = Alloc(length)
		if source == nil {
			return nil, &Error{Result: ResultOutOfMemory}
		}
		copy(unsafe.Slice((*byte)(source), len(data)), data)
	}

	ptr, err := snapshotDecoderNewBuf(uintptr(source), uint64(length))
	if err != nil {
		Free(source, length)
		return nil, err
	}

	return &SnapshotDecoder{
		ptr:       ptr,
		sourcePtr: source,
		sourceLen: length,
	}, nil
}

// Close frees the decoder and releases any retained reader handle or copied
// source buffer. It does not close a terminal returned by Ready or Decode.
// C: ghostty_snapshot_decoder_free
func (d *SnapshotDecoder) Close() {
	if d == nil || d.ptr == 0 {
		return
	}

	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_snapshot_decoder_free(tls, d.ptr)
	d.ptr = 0
	d.reader.close()
	d.reader = nil
	d.releaseSource()
}

// releaseSource ends either form of source ownership after libghostty can no
// longer read the input. It is idempotent because FINISH releases the source
// before the decoder itself is closed.
func (d *SnapshotDecoder) releaseSource() {
	if d.sourcePinned {
		d.sourcePin.Unpin()
		d.sourcePinned = false
	}
	d.sourceBytes = nil
	if d.sourcePtr != nil {
		Free(d.sourcePtr, d.sourceLen)
		d.sourcePtr = nil
		d.sourceLen = 0
	}
}

// setSnapshotDecoderScalar writes one decoder option whose value has type T.
func setSnapshotDecoderScalar[T any](d *SnapshotDecoder, option vt.GhosttySnapshotDecoderOption, v T) error {
	tls := getTLS()
	defer putTLS(tls)

	n := int(unsafe.Sizeof(v))
	p := tls.Alloc(n)
	defer tls.Free(n)
	*(*T)(unsafe.Pointer(p)) = v
	return resultError(vt.Xghostty_snapshot_decoder_set(tls, d.ptr, option, p))
}

// SetMaxContinuationBytes sets the largest non-ground VT continuation the
// decoder will accept. When [SnapshotDecoder.SetRetainContinuation] is true,
// this also becomes the continuation tracking limit on the returned terminal.
// It may only be called before decoding begins. Zero accepts only snapshots
// whose parser and UTF-8 decoder are at ground and leaves tracking disabled.
// C: ghostty_snapshot_decoder_set
func (d *SnapshotDecoder) SetMaxContinuationBytes(limit uint) error {
	return setSnapshotDecoderScalar(d, vt.GHOSTTY_SNAPSHOT_DECODER_OPT_MAX_CONTINUATION_BYTES, uint64(limit))
}

// SetRetainContinuation controls whether the returned terminal retains the
// exact continuation decoded from the snapshot. Retention is disabled by
// default and requires a nonzero maximum continuation size to enable tracking
// on the returned terminal. It may only be called before decoding begins.
// C: ghostty_snapshot_decoder_set
func (d *SnapshotDecoder) SetRetainContinuation(retain bool) error {
	return setSnapshotDecoderScalar(d, vt.GHOSTTY_SNAPSHOT_DECODER_OPT_RETAIN_CONTINUATION, cBool(retain))
}

// SetCompressHistory controls whether the decoder compresses scrollback as
// it restores it. It is off by default and may only be called before
// decoding begins.
//
// Without this option, a restored terminal keeps all of its scrollback
// uncompressed, even if the terminal that produced the snapshot had
// compressed it. The scrollback stays that size until the application calls
// [Terminal.Compress]. For a terminal with a lot of scrollback, that can
// use many times more memory than the original terminal did.
//
// With this option, the decoder compresses each page of scrollback right
// after restoring it, so at most one uncompressed page is held at a time.
// The result is the same as decoding normally and then calling
// [Terminal.Compress] with [TerminalCompressionFull], but without the
// temporary jump in memory use. A page that is on screen when it is
// restored is left uncompressed. Compressed scrollback is decompressed
// automatically when it is needed, for example when the user scrolls or
// searches.
//
// This option only affects how the restored terminal stores its scrollback
// in memory. It works with any snapshot. On platforms that do not support
// scrollback compression, it is accepted and has no effect.
// C: ghostty_snapshot_decoder_set
func (d *SnapshotDecoder) SetCompressHistory(compress bool) error {
	return setSnapshotDecoderScalar(d, vt.GHOSTTY_SNAPSHOT_DECODER_OPT_COMPRESS_HISTORY, cBool(compress))
}

// SetKittyImageStorageLimit sets the maximum number of bytes of Kitty
// graphics images to keep for each restored screen. It must be called before
// decoding begins. The default is the limit that a new terminal uses, and
// zero turns off Kitty graphics on the restored terminal.
//
// Snapshots do not record the limit of the terminal that created them, so
// the decoder always uses this one. Images are restored oldest first. If an
// image does not fit, older images are removed to make room, just as they
// would be on a live terminal. To change the limit after decoding, call
// [Terminal.SetKittyImageStorageLimit] on the restored terminal.
// C: ghostty_snapshot_decoder_set
func (d *SnapshotDecoder) SetKittyImageStorageLimit(limit uint64) error {
	return setSnapshotDecoderScalar(d, vt.GHOSTTY_SNAPSHOT_DECODER_OPT_KITTY_IMAGE_STORAGE_LIMIT, limit)
}

// decodeTerminal runs fn, a decoder call that writes a terminal handle to
// its out-parameter, and wraps the handle.
func (d *SnapshotDecoder) decodeTerminal(fn func(tls *libc.TLS, out uintptr) vt.GhosttyResult) (*Terminal, error) {
	tls := getTLS()
	defer putTLS(tls)

	const n = int(unsafe.Sizeof(uintptr(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)
	*(*uintptr)(unsafe.Pointer(out)) = 0
	if err := d.decodeError(fn(tls, out)); err != nil {
		return nil, err
	}
	return terminalFromC(*(*uintptr)(unsafe.Pointer(out)), TerminalConfig{}), nil
}

// Ready decodes and validates the renderable snapshot prefix. The
// returned terminal belongs to the caller and is ready to use right away.
// Older history and Kitty graphics pixel data can then be restored one
// record at a time with [SnapshotDecoder.Next].
//
// Ready restores Kitty graphics images without their pixel data. You can
// place and inspect these images right away, but [KittyGraphicsImage.Pixels]
// has nothing to return until Next delivers the pixels. When the pixels
// arrive, [KittyGraphics.Generation] changes, so a renderer that watches it
// draws the image on its next frame.
// C: ghostty_snapshot_decoder_ready
func (d *SnapshotDecoder) Ready() (*Terminal, error) {
	return d.decodeTerminal(func(tls *libc.TLS, out uintptr) vt.GhosttyResult {
		return vt.Xghostty_snapshot_decoder_ready(tls, d.ptr, out)
	})
}

// Next decodes and validates one record. A record is either a page of
// history, or the pixel data of one Kitty graphics image or animation frame.
// Pixel data can arrive at any point, including between history pages.
//
// Next returns true after it decodes a record. Call [SnapshotDecoder.Progress]
// to see what kind of record it was. Next returns false with a nil error once
// the whole snapshot has been decoded, and keeps returning false after that.
//
// You can render, resize, and write live input to the terminal between
// calls. If that makes a record impossible to apply, Next still reads and
// validates it, and Progress reports zero rows or zero bytes. For example,
// this happens when live input deletes the image that pixel data belongs to.
//
// If [SnapshotDecoder.SetCompressHistory] is on, Next compresses the page
// before returning, unless the page is on screen.
// C: ghostty_snapshot_decoder_next
func (d *SnapshotDecoder) Next() (bool, error) {
	tls := getTLS()
	defer putTLS(tls)
	result := vt.Xghostty_snapshot_decoder_next(tls, d.ptr)
	switch result {
	case vt.GHOSTTY_SUCCESS:
		return true, nil
	case vt.GHOSTTY_NO_VALUE:
		d.releaseSource()
		return false, nil
	default:
		return false, d.decodeError(result)
	}
}

// Decode decodes and validates one complete snapshot in a single call.
// The returned terminal is caller-owned. Bytes following the FINISH record
// remain outside the snapshot and can be located with SourceOffset.
//
// The returned terminal has no effect callbacks. Modes are restored, so a
// snapshot taken during synchronized output yields a terminal that is
// already in a render hold; a [RenderHoldFunc] attached afterwards is not
// told about it. See [RenderHoldFunc].
// C: ghostty_snapshot_decoder_decode
func (d *SnapshotDecoder) Decode() (*Terminal, error) {
	t, err := d.decodeTerminal(func(tls *libc.TLS, out uintptr) vt.GhosttyResult {
		return vt.Xghostty_snapshot_decoder_decode(tls, d.ptr, out)
	})
	if err != nil {
		return nil, err
	}
	d.releaseSource()
	return t, nil
}

// Get reads a raw decoder data field into value. The pointer type must match
// the documented type for data; prefer the typed helpers when possible.
//
// value must point at memory that does not move: C memory, or Go memory
// that is pinned with [runtime.Pinner] for the duration of the call.
// C: ghostty_snapshot_decoder_get
func (d *SnapshotDecoder) Get(data SnapshotDecoderData, value unsafe.Pointer) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_snapshot_decoder_get(
		tls,
		d.ptr,
		vt.GhosttySnapshotDecoderData(data),
		uintptr(value),
	))
}

// getSnapshotDecoderScalar reads one decoder data field whose value has
// type T.
func getSnapshotDecoderScalar[T any](d *SnapshotDecoder, data SnapshotDecoderData) (T, error) {
	tls := getTLS()
	defer putTLS(tls)

	var v T
	n := int(unsafe.Sizeof(v))
	p := tls.Alloc(n)
	defer tls.Free(n)
	clear(unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
	if err := resultError(vt.Xghostty_snapshot_decoder_get(
		tls,
		d.ptr,
		vt.GhosttySnapshotDecoderData(data),
		p,
	)); err != nil {
		return v, err
	}
	return *(*T)(unsafe.Pointer(p)), nil
}

// GetMulti reads several raw decoder data fields in one C call. Each value
// must point to storage with the type documented by its corresponding key.
// On partial failure, written is the failing index.
// C: ghostty_snapshot_decoder_get_multi
func (d *SnapshotDecoder) GetMulti(keys []SnapshotDecoderData, values []unsafe.Pointer) (written int, err error) {
	if len(keys) != len(values) {
		return 0, errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return 0, nil
	}

	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[vt.GhosttySnapshotDecoderData](&args, keys, values)
	if err != nil {
		return 0, err
	}
	defer args.free()

	tls := getTLS()
	defer putTLS(tls)
	const n = int(unsafe.Sizeof(uint64(0)))
	cWritten := tls.Alloc(n)
	defer tls.Free(n)
	*(*uint64)(unsafe.Pointer(cWritten)) = 0
	err = resultError(vt.Xghostty_snapshot_decoder_get_multi(
		tls,
		d.ptr,
		uint64(len(keys)),
		uintptr(unsafe.Pointer(cKeys)),
		uintptr(unsafe.Pointer(args.values)),
		cWritten,
	))
	return int(*(*uint64)(unsafe.Pointer(cWritten))), err
}

// MaxContinuationBytes returns the decoder's current continuation input limit.
// When [SnapshotDecoder.RetainContinuation] is true, this is also the tracking
// limit applied to the returned terminal.
func (d *SnapshotDecoder) MaxContinuationBytes() (uint, error) {
	value, err := getSnapshotDecoderScalar[uint64](d, SnapshotDecoderDataMaxContinuationBytes)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

// CompressHistory reports whether the decoder compresses scrollback as it
// restores it. See [SnapshotDecoder.SetCompressHistory].
func (d *SnapshotDecoder) CompressHistory() (bool, error) {
	value, err := getSnapshotDecoderScalar[uint8](d, SnapshotDecoderDataCompressHistory)
	if err != nil {
		return false, err
	}
	return value != 0, nil
}

// RetainContinuation reports whether decoded continuation tracking will be
// retained on terminals returned by Ready or Decode.
func (d *SnapshotDecoder) RetainContinuation() (bool, error) {
	value, err := getSnapshotDecoderScalar[uint8](d, SnapshotDecoderDataRetainContinuation)
	if err != nil {
		return false, err
	}
	return value != 0, nil
}

// SourceOffset returns the number of snapshot source bytes consumed so far.
// At FINISH it identifies the first byte after the snapshot.
func (d *SnapshotDecoder) SourceOffset() (uint, error) {
	value, err := getSnapshotDecoderScalar[uint64](d, SnapshotDecoderDataSourceOffset)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

// HistoryRows returns the snapshot's complete logical history extent for
// screen. It becomes available after Ready validates. The alternate screen
// returns [ErrNoValue] if the snapshot has no alternate screen.
func (d *SnapshotDecoder) HistoryRows(screen TerminalScreen) (uint64, error) {
	var data SnapshotDecoderData
	switch screen {
	case ScreenPrimary:
		data = SnapshotDecoderDataHistoryRowsPrimary
	case ScreenAlternate:
		data = SnapshotDecoderDataHistoryRowsAlternate
	default:
		return 0, &Error{Result: ResultInvalidValue}
	}

	return getSnapshotDecoderScalar[uint64](d, data)
}

// KittyImageStorageLimit returns the maximum number of bytes of Kitty
// graphics images the decoder keeps for each restored screen. See
// [SnapshotDecoder.SetKittyImageStorageLimit].
func (d *SnapshotDecoder) KittyImageStorageLimit() (uint64, error) {
	return getSnapshotDecoderScalar[uint64](d, SnapshotDecoderDataKittyImageStorageLimit)
}

// KittyImageCount returns the number of Kitty graphics images that Ready
// restored on screen. Their pixel data arrives later through
// [SnapshotDecoder.Next].
//
// It is available after Ready returns. It is zero if the library was built
// without Kitty graphics support. For [ScreenAlternate], it returns
// [ErrNoValue] if the snapshot has no alternate screen.
func (d *SnapshotDecoder) KittyImageCount(screen TerminalScreen) (uint32, error) {
	var data SnapshotDecoderData
	switch screen {
	case ScreenPrimary:
		data = SnapshotDecoderDataKittyImagesPrimary
	case ScreenAlternate:
		data = SnapshotDecoderDataKittyImagesAlternate
	default:
		return 0, &Error{Result: ResultInvalidValue}
	}

	return getSnapshotDecoderScalar[uint32](d, data)
}

// KittyImagePendingBytes returns the number of bytes of Kitty graphics pixel
// data that were still to come for screen when Ready returned.
//
// Use it to show progress while the pixel data arrives. After each call to
// Next, add [SnapshotDecoderProgress.ImageBytes] to a running total for the
// screen. Discarded pixel data counts as zero bytes, so the total can end up
// below this value.
//
// It is available after Ready returns. For [ScreenAlternate], it returns
// [ErrNoValue] if the snapshot has no alternate screen.
func (d *SnapshotDecoder) KittyImagePendingBytes(screen TerminalScreen) (uint64, error) {
	var data SnapshotDecoderData
	switch screen {
	case ScreenPrimary:
		data = SnapshotDecoderDataKittyPendingBytesPrimary
	case ScreenAlternate:
		data = SnapshotDecoderDataKittyPendingBytesAlternate
	default:
		return 0, &Error{Result: ResultInvalidValue}
	}

	return getSnapshotDecoderScalar[uint64](d, data)
}

// ProgressKind returns the kind of record that the most recent successful
// call to Next decoded.
func (d *SnapshotDecoder) ProgressKind() (SnapshotDecoderProgressKind, error) {
	value, err := getSnapshotDecoderScalar[vt.GhosttySnapshotDecoderProgressKind](d, SnapshotDecoderDataProgressKind)
	if err != nil {
		return 0, err
	}
	return SnapshotDecoderProgressKind(value), nil
}

// ProgressImageID returns the ID of the Kitty graphics image whose pixel
// data the most recent call to Next decoded. It returns [ErrNoValue] if that
// record was not [SnapshotDecoderProgressKindImage].
func (d *SnapshotDecoder) ProgressImageID() (uint32, error) {
	return getSnapshotDecoderScalar[uint32](d, SnapshotDecoderDataProgressImageID)
}

// ProgressImageFrame returns the frame number of the pixel data that the
// most recent call to Next decoded. Frame 1 is the image itself, and frames
// 2 and up are animation frames. It returns [ErrNoValue] if that record was
// not [SnapshotDecoderProgressKindImage].
func (d *SnapshotDecoder) ProgressImageFrame() (uint32, error) {
	return getSnapshotDecoderScalar[uint32](d, SnapshotDecoderDataProgressImageFrame)
}

// ProgressImageBytes returns the number of bytes of pixel data that the most
// recent call to Next added to its image. It returns [ErrNoValue] if that
// record was not [SnapshotDecoderProgressKindImage].
//
// Zero means the pixel data was read and validated but then discarded. That
// happens when live input deleted the image or replaced it with a new image
// that has the same ID, when the image did not fit the storage limit, or
// when its screen no longer exists.
func (d *SnapshotDecoder) ProgressImageBytes() (uint, error) {
	value, err := getSnapshotDecoderScalar[uint64](d, SnapshotDecoderDataProgressImageBytes)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

// ProgressScreen returns the screen associated with the most recently
// decoded history page or image belongs to.
func (d *SnapshotDecoder) ProgressScreen() (TerminalScreen, error) {
	value, err := getSnapshotDecoderScalar[vt.GhosttyTerminalScreen](d, SnapshotDecoderDataProgressScreen)
	if err != nil {
		return 0, err
	}
	return TerminalScreen(value), nil
}

// ProgressRows returns the number of rows prepended by the most recently
// decoded history page. Zero means the page could not be applied safely.
// It returns [ErrNoValue] if that record was not
// [SnapshotDecoderProgressKindPage].
func (d *SnapshotDecoder) ProgressRows() (uint, error) {
	value, err := getSnapshotDecoderScalar[uint64](d, SnapshotDecoderDataProgressRows)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

// ProgressRemaining returns the number of page records remaining in the same
// screen's history sequence. It returns [ErrNoValue] if that record was not
// [SnapshotDecoderProgressKindPage].
func (d *SnapshotDecoder) ProgressRemaining() (uint32, error) {
	return getSnapshotDecoderScalar[uint32](d, SnapshotDecoderDataProgressRemaining)
}

// SnapshotDecoderProgress describes the record that the most recent
// successful call to [SnapshotDecoder.Next] decoded. Kind says which fields
// apply. Fields that don't apply are zero.
type SnapshotDecoderProgress struct {
	// Kind is the kind of record: a history page or image pixel data.
	Kind SnapshotDecoderProgressKind

	// Screen is the screen that the record belongs to.
	Screen TerminalScreen

	// Rows is the number of rows added to the top of the scrollback. Zero
	// means the page could not be applied. It is set for
	// [SnapshotDecoderProgressKindPage].
	Rows uint

	// Remaining is the number of history pages still to come for this
	// screen. It is set for [SnapshotDecoderProgressKindPage].
	Remaining uint32

	// ImageID is the ID of the Kitty graphics image that the pixel data
	// belongs to. It is set for [SnapshotDecoderProgressKindImage].
	ImageID uint32

	// ImageFrame is the frame number of the pixel data. Frame 1 is the
	// image itself, and frames 2 and up are animation frames. It is set for
	// [SnapshotDecoderProgressKindImage].
	ImageFrame uint32

	// ImageBytes is the number of bytes of pixel data added to the image.
	// Zero means the data was discarded. See
	// [SnapshotDecoder.ProgressImageBytes] for when that happens. It is set
	// for [SnapshotDecoderProgressKindImage].
	ImageBytes uint
}

// Progress returns every progress value for the record that the most recent
// successful call to Next decoded. Which values exist depends on the kind, so
// it reads the kind and screen first and then only the values that apply. It
// does not allocate, so prefer it to the individual Progress methods.
func (d *SnapshotDecoder) Progress() (SnapshotDecoderProgress, error) {
	tls := getTLS()
	defer putTLS(tls)

	// One 8-byte slot on the C stack serves every value in turn.
	const n = int(unsafe.Sizeof(uint64(0)))
	slot := tls.Alloc(n)
	defer tls.Free(n)
	get := func(data SnapshotDecoderData) error {
		*(*uint64)(unsafe.Pointer(slot)) = 0
		return resultError(vt.Xghostty_snapshot_decoder_get(
			tls, d.ptr, vt.GhosttySnapshotDecoderData(data), slot))
	}
	u32 := func() uint32 { return *(*uint32)(unsafe.Pointer(slot)) }
	u64 := func() uint64 { return *(*uint64)(unsafe.Pointer(slot)) }

	var out SnapshotDecoderProgress
	if err := get(SnapshotDecoderDataProgressKind); err != nil {
		return SnapshotDecoderProgress{}, err
	}
	out.Kind = SnapshotDecoderProgressKind(*(*vt.GhosttySnapshotDecoderProgressKind)(unsafe.Pointer(slot)))
	if err := get(SnapshotDecoderDataProgressScreen); err != nil {
		return SnapshotDecoderProgress{}, err
	}
	out.Screen = TerminalScreen(*(*vt.GhosttyTerminalScreen)(unsafe.Pointer(slot)))

	switch out.Kind {
	case SnapshotDecoderProgressKindPage:
		if err := get(SnapshotDecoderDataProgressRows); err != nil {
			return SnapshotDecoderProgress{}, err
		}
		out.Rows = uint(u64())
		if err := get(SnapshotDecoderDataProgressRemaining); err != nil {
			return SnapshotDecoderProgress{}, err
		}
		out.Remaining = u32()

	case SnapshotDecoderProgressKindImage:
		if err := get(SnapshotDecoderDataProgressImageID); err != nil {
			return SnapshotDecoderProgress{}, err
		}
		out.ImageID = u32()
		if err := get(SnapshotDecoderDataProgressImageFrame); err != nil {
			return SnapshotDecoderProgress{}, err
		}
		out.ImageFrame = u32()
		if err := get(SnapshotDecoderDataProgressImageBytes); err != nil {
			return SnapshotDecoderProgress{}, err
		}
		out.ImageBytes = uint(u64())

	default:
		// A newer library may add more progress kinds. Report the kind
		// and screen, and leave the kind-specific values zero.
	}
	return out, nil
}

// decodeError combines GHOSTTY_IO_ERROR with the original Go reader error so
// callers can inspect both through errors.Is/As.
func (d *SnapshotDecoder) decodeError(result vt.GhosttyResult) error {
	var callbackErr error
	if d.reader != nil {
		callbackErr = d.reader.err
	}
	return resultErrorWithCallback(result, callbackErr)
}
