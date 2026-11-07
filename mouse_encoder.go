package libghostty

// Mouse encoder — encodes mouse events into terminal escape sequences.
// Wraps the C APIs from mouse/encoder.h.

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// MouseEncoder encodes mouse events into terminal escape sequences,
// supporting X10, UTF-8, SGR, URxvt, and SGR-Pixels mouse protocols.
// It maintains mutable encoder state and is not safe for concurrent
// use.
//
// Basic usage:
//  1. Create an encoder with NewMouseEncoder.
//  2. Configure options with the Set* methods or SetFromTerminal.
//  3. Create mouse events, encode them with Encode, and free them.
//  4. Free the encoder with Close when done.
//
// C: GhosttyMouseEncoder
type MouseEncoder struct {
	ptr uintptr
}

// MouseTrackingMode selects which mouse events the terminal tracks.
//
// C: GhosttyMouseTrackingMode
type MouseTrackingMode int

const (
	// MouseTrackingNone disables mouse reporting.
	MouseTrackingNone MouseTrackingMode = vt.GHOSTTY_MOUSE_TRACKING_NONE

	// MouseTrackingX10 enables X10 mouse mode.
	MouseTrackingX10 MouseTrackingMode = vt.GHOSTTY_MOUSE_TRACKING_X10

	// MouseTrackingNormal enables normal mouse mode (button press/release only).
	MouseTrackingNormal MouseTrackingMode = vt.GHOSTTY_MOUSE_TRACKING_NORMAL

	// MouseTrackingButton enables button-event tracking mode.
	MouseTrackingButton MouseTrackingMode = vt.GHOSTTY_MOUSE_TRACKING_BUTTON

	// MouseTrackingAny enables any-event tracking mode.
	MouseTrackingAny MouseTrackingMode = vt.GHOSTTY_MOUSE_TRACKING_ANY
)

// MouseFormat selects the wire format for mouse escape sequences.
//
// C: GhosttyMouseFormat
type MouseFormat int

const (
	MouseFormatX10       MouseFormat = vt.GHOSTTY_MOUSE_FORMAT_X10
	MouseFormatUTF8      MouseFormat = vt.GHOSTTY_MOUSE_FORMAT_UTF8
	MouseFormatSGR       MouseFormat = vt.GHOSTTY_MOUSE_FORMAT_SGR
	MouseFormatURxvt     MouseFormat = vt.GHOSTTY_MOUSE_FORMAT_URXVT
	MouseFormatSGRPixels MouseFormat = vt.GHOSTTY_MOUSE_FORMAT_SGR_PIXELS
)

// MouseEncoderSize describes the rendered terminal geometry used to
// convert surface-space positions into encoded coordinates.
//
// C: GhosttyMouseEncoderSize
type MouseEncoderSize struct {
	// ScreenWidth is the full screen width in pixels.
	ScreenWidth uint32

	// ScreenHeight is the full screen height in pixels.
	ScreenHeight uint32

	// CellWidth is the cell width in pixels. Must be non-zero.
	CellWidth uint32

	// CellHeight is the cell height in pixels. Must be non-zero.
	CellHeight uint32

	// PaddingTop is the top padding in pixels.
	PaddingTop uint32

	// PaddingBottom is the bottom padding in pixels.
	PaddingBottom uint32

	// PaddingRight is the right padding in pixels.
	PaddingRight uint32

	// PaddingLeft is the left padding in pixels.
	PaddingLeft uint32
}

// MouseEncoderOption identifies a mouse encoder configuration option.
// Each one has its own setter on [MouseEncoder].
//
// C: GhosttyMouseEncoderOption
type MouseEncoderOption int

const (
	// MouseEncoderOptEvent sets the mouse tracking mode
	// (value: MouseTrackingMode).
	MouseEncoderOptEvent MouseEncoderOption = vt.GHOSTTY_MOUSE_ENCODER_OPT_EVENT

	// MouseEncoderOptFormat sets the mouse output format
	// (value: MouseFormat).
	MouseEncoderOptFormat MouseEncoderOption = vt.GHOSTTY_MOUSE_ENCODER_OPT_FORMAT

	// MouseEncoderOptSize sets the renderer size context
	// (value: MouseEncoderSize).
	MouseEncoderOptSize MouseEncoderOption = vt.GHOSTTY_MOUSE_ENCODER_OPT_SIZE

	// MouseEncoderOptAnyButtonPressed sets whether any mouse button
	// is currently pressed (value: bool).
	MouseEncoderOptAnyButtonPressed MouseEncoderOption = vt.GHOSTTY_MOUSE_ENCODER_OPT_ANY_BUTTON_PRESSED

	// MouseEncoderOptTrackLastCell enables motion deduplication by
	// last cell (value: bool).
	MouseEncoderOptTrackLastCell MouseEncoderOption = vt.GHOSTTY_MOUSE_ENCODER_OPT_TRACK_LAST_CELL
)

// NewMouseEncoder creates a new mouse encoder with default options.
// The encoder must be freed with Close when no longer needed.
func NewMouseEncoder() (*MouseEncoder, error) {
	tls := getTLS()
	defer putTLS(tls)

	const n = int(unsafe.Sizeof(uintptr(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)
	if err := resultError(vt.Xghostty_mouse_encoder_new(tls, 0, out)); err != nil {
		return nil, err
	}
	return &MouseEncoder{ptr: *(*uintptr)(unsafe.Pointer(out))}, nil
}

// Close frees the underlying mouse encoder handle. After this call,
// the encoder must not be used, except that calling Close again is a
// safe no-op.
func (enc *MouseEncoder) Close() {
	if enc == nil || enc.ptr == 0 {
		return
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_mouse_encoder_free(tls, enc.ptr)
	enc.ptr = 0
}

// setMouseEncoderScalar writes one encoder option whose value has type T.
func setMouseEncoderScalar[T any](enc *MouseEncoder, opt vt.GhosttyMouseEncoderOption, v T) error {
	tls := getTLS()
	defer putTLS(tls)

	n := int(unsafe.Sizeof(v))
	p := tls.Alloc(n)
	defer tls.Free(n)
	*(*T)(unsafe.Pointer(p)) = v
	return resultError(vt.Xghostty_mouse_encoder_set(tls, enc.ptr, opt, p))
}

// SetTrackingMode sets the mouse tracking mode on the encoder.
func (enc *MouseEncoder) SetTrackingMode(mode MouseTrackingMode) error {
	return setMouseEncoderScalar(enc, vt.GHOSTTY_MOUSE_ENCODER_OPT_EVENT, vt.GhosttyMouseTrackingMode(mode))
}

// SetFormat sets the mouse output format on the encoder.
func (enc *MouseEncoder) SetFormat(format MouseFormat) error {
	return setMouseEncoderScalar(enc, vt.GHOSTTY_MOUSE_ENCODER_OPT_FORMAT, vt.GhosttyMouseFormat(format))
}

// SetSize sets the renderer size context on the encoder.
func (enc *MouseEncoder) SetSize(s MouseEncoderSize) error {
	// GhosttyMouseEncoderSize is a sized struct: size records the struct
	// size this binding was built with.
	cs := vt.GhosttyMouseEncoderSize{
		Fsize:           uint64(unsafe.Sizeof(vt.GhosttyMouseEncoderSize{})),
		Fscreen_width:   s.ScreenWidth,
		Fscreen_height:  s.ScreenHeight,
		Fcell_width:     s.CellWidth,
		Fcell_height:    s.CellHeight,
		Fpadding_top:    s.PaddingTop,
		Fpadding_bottom: s.PaddingBottom,
		Fpadding_right:  s.PaddingRight,
		Fpadding_left:   s.PaddingLeft,
	}
	return setMouseEncoderScalar(enc, vt.GHOSTTY_MOUSE_ENCODER_OPT_SIZE, cs)
}

// SetAnyButtonPressed sets whether any mouse button is currently
// pressed.
func (enc *MouseEncoder) SetAnyButtonPressed(pressed bool) error {
	return setMouseEncoderScalar(enc, vt.GHOSTTY_MOUSE_ENCODER_OPT_ANY_BUTTON_PRESSED, cBool(pressed))
}

// SetTrackLastCell enables or disables motion deduplication by
// last cell.
func (enc *MouseEncoder) SetTrackLastCell(track bool) error {
	return setMouseEncoderScalar(enc, vt.GHOSTTY_MOUSE_ENCODER_OPT_TRACK_LAST_CELL, cBool(track))
}

// SetFromTerminal reads the terminal's current mouse tracking mode
// and output format and applies them to the encoder. It does not
// modify size or any-button state. It does clear the last tracked cell,
// so motion deduplication starts over and calling this before every
// event defeats [MouseEncoder.SetTrackLastCell]. The caller must
// serialize access to both the encoder and the terminal during this
// call.
func (enc *MouseEncoder) SetFromTerminal(t *Terminal) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_mouse_encoder_set_from_terminal(tls, enc.ptr, t.ptr))
}

// Reset clears internal encoder state such as motion deduplication
// (last tracked cell).
func (enc *MouseEncoder) Reset() error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_mouse_encoder_reset(tls, enc.ptr))
}

// Encode encodes a mouse event into a terminal escape sequence and
// returns the result as a byte slice. Not all mouse events produce
// output; in that case, a nil slice and nil error are returned.
func (enc *MouseEncoder) Encode(event *MouseEvent) ([]byte, error) {
	tls := getTLS()
	defer putTLS(tls)

	// Most mouse escape sequences fit in 128 bytes. Encode into a buffer
	// on the C stack first; fall back to a larger allocation if needed.
	const bufLen = 128
	const n = bufLen + int(unsafe.Sizeof(uint64(0)))
	buf := tls.Alloc(n)
	defer tls.Free(n)
	outLen := buf + bufLen
	*(*uint64)(unsafe.Pointer(outLen)) = 0
	result := vt.Xghostty_mouse_encoder_encode(
		tls,
		enc.ptr,
		event.ptr,
		buf,
		bufLen,
		outLen,
	)
	length := *(*uint64)(unsafe.Pointer(outLen))

	if result == vt.GHOSTTY_SUCCESS {
		if length == 0 {
			return nil, nil
		}
		out := make([]byte, length)
		copy(out, unsafe.Slice((*byte)(unsafe.Pointer(buf)), length))
		return out, nil
	}

	if result == vt.GHOSTTY_OUT_OF_SPACE {
		// length is the required buffer size.
		dynBuf := make([]byte, length)
		addr, unpin := pinBytes(dynBuf)
		defer unpin()
		*(*uint64)(unsafe.Pointer(outLen)) = 0
		if err := resultError(vt.Xghostty_mouse_encoder_encode(
			tls,
			enc.ptr,
			event.ptr,
			addr,
			length,
			outLen,
		)); err != nil {
			return nil, err
		}
		written := *(*uint64)(unsafe.Pointer(outLen))
		if written == 0 {
			return nil, nil
		}
		return dynBuf[:written], nil
	}

	return nil, &Error{Result: Result(result)}
}
