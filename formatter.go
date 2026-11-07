package libghostty

import (
	"io"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// FormatterFormat selects the output format for a Formatter.
// C: GhosttyFormatterFormat
type FormatterFormat int

const (
	// FormatterFormatPlain emits plain text (no escape sequences).
	FormatterFormatPlain FormatterFormat = vt.GHOSTTY_FORMATTER_FORMAT_PLAIN

	// FormatterFormatVT emits VT sequences preserving colors, styles, URLs, etc.
	FormatterFormatVT FormatterFormat = vt.GHOSTTY_FORMATTER_FORMAT_VT

	// FormatterFormatHTML emits HTML with inline styles.
	FormatterFormatHTML FormatterFormat = vt.GHOSTTY_FORMATTER_FORMAT_HTML
)

// FormatterExtras selects terminal state to add to VT output. Combine values
// with |.
//
// VT output is meant to be written to another terminal to recreate the screen.
// Without extras, it reproduces the text and how each cell is styled. Each
// extra adds sequences that restore one more piece of terminal state, such as
// the cursor position or the colors in the palette. Plain text ignores extras.
//
// [FormatterFormatHTML] uses only [FormatterExtraPalette], which adds the
// <style> block defining the --vt-palette-N CSS variables. Palette colors in
// HTML output refer to those variables, so without this extra the document has
// to define them itself.
//
// A new formatter uses [FormatterExtrasDefault]. Use [FormatterExtrasAll] to
// recreate the terminal as closely as possible.
//
// The values are those of the GHOSTTY_FORMATTER_EXTRA_* macros in
// formatter.h. The translated headers in internal/vt do not carry them
// because they are expressions rather than literals.
//
// C: GhosttyFormatterExtras
type FormatterExtras uint32

const (
	// FormatterExtraPalette restores the color palette with OSC 4.
	FormatterExtraPalette FormatterExtras = 1 << 0

	// FormatterExtraModes restores terminal modes that differ from their
	// defaults with CSI h and l.
	FormatterExtraModes FormatterExtras = 1 << 1

	// FormatterExtraScrollingRegion restores the scrolling region with DECSTBM
	// and DECSLRM.
	FormatterExtraScrollingRegion FormatterExtras = 1 << 2

	// FormatterExtraTabstops restores tab stops by clearing them all and
	// setting each one again.
	FormatterExtraTabstops FormatterExtras = 1 << 3

	// FormatterExtraPwd restores the working directory with OSC 7.
	FormatterExtraPwd FormatterExtras = 1 << 4

	// FormatterExtraKeyboard restores keyboard modes such as modifyOtherKeys.
	FormatterExtraKeyboard FormatterExtras = 1 << 5

	// FormatterExtraCursor restores the cursor position with CUP.
	FormatterExtraCursor FormatterExtras = 1 << 6

	// FormatterExtraStyle restores the style the cursor writes new text with,
	// using SGR.
	FormatterExtraStyle FormatterExtras = 1 << 7

	// FormatterExtraHyperlink restores the hyperlink the cursor writes new text
	// with, using OSC 8.
	FormatterExtraHyperlink FormatterExtras = 1 << 8

	// FormatterExtraProtection restores character protection with DECSCA.
	FormatterExtraProtection FormatterExtras = 1 << 9

	// FormatterExtraKittyKeyboard restores the Kitty keyboard protocol flags.
	FormatterExtraKittyKeyboard FormatterExtras = 1 << 10

	// FormatterExtraCharsets restores the character sets, such as the line
	// drawing set.
	FormatterExtraCharsets FormatterExtras = 1 << 11

	// FormatterExtrasNone selects no extras. VT output has only the text and
	// its styles.
	FormatterExtrasNone FormatterExtras = 0

	// FormatterExtrasAll selects every extra.
	FormatterExtrasAll FormatterExtras = 0xFFF

	// FormatterExtrasDefault is what a new formatter starts with: the palette,
	// the cursor style and the cursor hyperlink. These are the extras that
	// affect how output looks.
	FormatterExtrasDefault FormatterExtras = FormatterExtraPalette | FormatterExtraStyle | FormatterExtraHyperlink
)

// FormatterOption is a functional option for configuring a Formatter. Each one
// applies a setting to the new formatter, so anything an option does can also
// be done later with the corresponding Formatter setter.
type FormatterOption func(*Formatter) error

// WithFormatterFormat sets the output format (plain, VT, or HTML).
// Defaults to FormatterFormatPlain if not specified. See [Formatter.SetEmit].
func WithFormatterFormat(format FormatterFormat) FormatterOption {
	return func(f *Formatter) error {
		return f.SetEmit(format)
	}
}

// WithFormatterUnwrap enables unwrapping of soft-wrapped lines. Defaults to
// false. See [Formatter.SetUnwrap].
func WithFormatterUnwrap(unwrap bool) FormatterOption {
	return func(f *Formatter) error {
		return f.SetUnwrap(unwrap)
	}
}

// WithFormatterTrim controls whether trailing whitespace is removed from
// lines that contain text. Defaults to true. See [Formatter.SetTrim].
func WithFormatterTrim(trim bool) FormatterOption {
	return func(f *Formatter) error {
		return f.SetTrim(trim)
	}
}

// WithFormatterSelection restricts formatter output to sel. Passing nil
// formats the entire active screen, which is the default. See
// [Formatter.SetSelection].
func WithFormatterSelection(sel *Selection) FormatterOption {
	return func(f *Formatter) error {
		return f.SetSelection(sel)
	}
}

// WithFormatterExtras sets the terminal state added to VT output. Defaults to
// [FormatterExtrasDefault]. Set this to [FormatterExtrasNone] for just the text
// and its styles, since the default adds the whole palette to VT and HTML
// output. See [Formatter.SetExtras].
func WithFormatterExtras(extras FormatterExtras) FormatterOption {
	return func(f *Formatter) error {
		return f.SetExtras(extras)
	}
}

// Formatter wraps a Ghostty formatter handle that can produce
// plain text, VT sequences, or HTML from a terminal's current state.
// The formatter stores a borrowed reference to a terminal, so the
// terminal must outlive the formatter and formatter calls must be
// serialized with all other access to that terminal.
//
// Formatter implements io.WriterTo so formatted output can be written
// directly to any io.Writer.
// C: GhosttyFormatter
type Formatter struct {
	ptr uintptr

	// writer is reused across WriteTo calls to avoid allocating handles.
	writer ghosttyWriterBridge
}

// NewFormatter creates a formatter for the given terminal's active screen.
// The terminal must outlive the formatter. The formatter captures a
// borrowed reference to the terminal and reads its current state on
// each [Formatter.Format] call, so formatter calls must be serialized
// with other access to the terminal.
//
// Without options, the formatter emits plain text with trailing whitespace
// trimmed. Settings can be changed at any time with the Formatter setters and
// stay in effect across format calls.
func NewFormatter(t *Terminal, opts ...FormatterOption) (*Formatter, error) {
	tls := getTLS()
	defer putTLS(tls)

	const n = int(unsafe.Sizeof(uintptr(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)
	if err := resultError(vt.Xghostty_formatter_terminal_new(tls, 0, out, t.ptr)); err != nil {
		return nil, err
	}

	f := &Formatter{ptr: *(*uintptr)(unsafe.Pointer(out))}
	for _, opt := range opts {
		if err := opt(f); err != nil {
			f.Close()
			return nil, err
		}
	}

	return f, nil
}

// set writes one formatter option from a value stored in C memory. A zero
// value pointer restores the option's default.
func (f *Formatter) set(option vt.GhosttyFormatterOption, value uintptr) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_formatter_set(tls, f.ptr, option, value))
}

// setFormatterScalar writes one formatter option whose value has type T.
func setFormatterScalar[T any](f *Formatter, option vt.GhosttyFormatterOption, v T) error {
	tls := getTLS()
	defer putTLS(tls)

	n := int(unsafe.Sizeof(v))
	p := tls.Alloc(n)
	defer tls.Free(n)
	*(*T)(unsafe.Pointer(p)) = v
	return resultError(vt.Xghostty_formatter_set(tls, f.ptr, option, p))
}

// getFormatterScalar reads one formatter option whose value has type T.
func getFormatterScalar[T any](f *Formatter, option vt.GhosttyFormatterOption) (T, error) {
	tls := getTLS()
	defer putTLS(tls)

	var v T
	n := int(unsafe.Sizeof(v))
	p := tls.Alloc(n)
	defer tls.Free(n)
	clear(unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
	if err := resultError(vt.Xghostty_formatter_get(tls, f.ptr, option, p)); err != nil {
		return v, err
	}
	return *(*T)(unsafe.Pointer(p)), nil
}

// SetEmit sets the output format. The change applies to the next format call.
func (f *Formatter) SetEmit(format FormatterFormat) error {
	return setFormatterScalar(f, vt.GHOSTTY_FORMATTER_OPT_EMIT, vt.GhosttyFormatterFormat(format))
}

// Emit returns the output format.
func (f *Formatter) Emit() (FormatterFormat, error) {
	v, err := getFormatterScalar[vt.GhosttyFormatterFormat](f, vt.GHOSTTY_FORMATTER_OPT_EMIT)
	if err != nil {
		return 0, err
	}
	return FormatterFormat(v), nil
}

// SetUnwrap controls whether lines that wrapped because they were too long for
// the screen are joined. When false, each row of the screen is its own line in
// the output. When true, a line the terminal wrapped onto several rows comes
// out as one line, the way the program originally wrote it.
func (f *Formatter) SetUnwrap(unwrap bool) error {
	return setFormatterScalar(f, vt.GHOSTTY_FORMATTER_OPT_UNWRAP, cBool(unwrap))
}

// Unwrap reports whether soft-wrapped lines are joined.
func (f *Formatter) Unwrap() (bool, error) {
	v, err := getFormatterScalar[uint8](f, vt.GHOSTTY_FORMATTER_OPT_UNWRAP)
	if err != nil {
		return false, err
	}
	return v != 0, nil
}

// SetTrim controls whether trailing whitespace is removed from lines that
// contain text.
//
// Blank lines at the end of the output are always removed. The one exception
// is [FormatterFormatVT] output with trim set to false, which keeps those lines
// so that the output reproduces the screen layout.
func (f *Formatter) SetTrim(trim bool) error {
	return setFormatterScalar(f, vt.GHOSTTY_FORMATTER_OPT_TRIM, cBool(trim))
}

// Trim reports whether trailing whitespace is removed.
func (f *Formatter) Trim() (bool, error) {
	v, err := getFormatterScalar[uint8](f, vt.GHOSTTY_FORMATTER_OPT_TRIM)
	if err != nil {
		return false, err
	}
	return v != 0, nil
}

// SetSelection restricts formatter output to sel. Passing nil clears it and
// formats the entire active screen again.
//
// To format the terminal's active selection, read it with [Terminal.Selection]
// and pass it here. Check for nil first: Terminal.Selection returns nil when
// nothing is selected, which here would format the whole screen. For the same
// text Ghostty copies to the clipboard, use plain output with both unwrap and
// trim enabled.
//
// The selection must come from the same terminal passed to [NewFormatter]. The
// formatter copies it, so the Selection value itself does not need to outlive
// this call. The copy still contains borrowed grid references into the
// terminal, and like any [Selection] it is only valid until the terminal next
// changes. After that, set the selection again before formatting, or clear it.
func (f *Formatter) SetSelection(sel *Selection) error {
	if sel == nil {
		return f.set(vt.GHOSTTY_FORMATTER_OPT_SELECTION, 0)
	}
	return setFormatterScalar(f, vt.GHOSTTY_FORMATTER_OPT_SELECTION, sel.toC())
}

// Selection returns the selection the formatter is restricted to, or nil when
// it formats the entire active screen. The result is subject to the same
// validity rules as the value passed to [Formatter.SetSelection].
func (f *Formatter) Selection() (*Selection, error) {
	cs, err := getFormatterScalar[vt.GhosttySelection](f, vt.GHOSTTY_FORMATTER_OPT_SELECTION)
	if err != nil {
		return noValueNilSelection(err)
	}
	sel := selectionFromC(cs)
	return &sel, nil
}

// SetExtras sets the terminal state added to VT output. To change one extra and
// keep the rest, read the current value with [Formatter.Extras] first.
func (f *Formatter) SetExtras(extras FormatterExtras) error {
	return setFormatterScalar(f, vt.GHOSTTY_FORMATTER_OPT_EXTRAS, vt.GhosttyFormatterExtras(extras))
}

// Extras returns the terminal state added to VT output.
func (f *Formatter) Extras() (FormatterExtras, error) {
	v, err := getFormatterScalar[vt.GhosttyFormatterExtras](f, vt.GHOSTTY_FORMATTER_OPT_EXTRAS)
	if err != nil {
		return 0, err
	}
	return FormatterExtras(v), nil
}

// Close frees the formatter handle. After this call, the formatter
// must not be used, except that calling Close again is a safe no-op.
func (f *Formatter) Close() {
	if f == nil || f.ptr == 0 {
		return
	}
	tls := getTLS()
	defer putTLS(tls)
	f.writer.close()
	vt.Xghostty_formatter_free(tls, f.ptr)
	f.ptr = 0
}

// Format runs the formatter and returns the output as a byte slice.
// Each call reflects the terminal's current state at the time of the
// call. Serialize Format with all other access to the underlying
// terminal. The returned buffer is allocated by libghostty and copied
// into Go memory.
func (f *Formatter) Format() ([]byte, error) {
	tls := getTLS()
	defer putTLS(tls)

	// Two out-parameters: the buffer pointer and its length.
	const n = 2 * int(unsafe.Sizeof(uintptr(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)
	outPtr := out
	outLen := out + unsafe.Sizeof(uintptr(0))
	if err := resultError(vt.Xghostty_formatter_format_alloc(tls, f.ptr, 0, outPtr, outLen)); err != nil {
		return nil, err
	}
	ptr := *(*uintptr)(unsafe.Pointer(outPtr))
	length := *(*uint64)(unsafe.Pointer(outLen))
	defer vt.Xghostty_free(tls, 0, ptr, length)

	return goBytes(ptr, int(length)), nil
}

// FormatBuf formats into buf and returns the number of bytes written. If buf
// is too small, the returned count is the required size and the error has
// result [ResultOutOfSpace]. A nil buffer can be used to query the required
// size. Output that is empty needs no space, so a query for it returns 0 and
// no error.
func (f *Formatter) FormatBuf(buf []byte) (int, error) {
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
	result := vt.Xghostty_formatter_format_buf(
		tls,
		f.ptr,
		ptr,
		uint64(len(buf)),
		written,
	)
	return int(*(*uint64)(unsafe.Pointer(written))), resultError(result)
}

// FormatString runs the formatter and returns the output as a string.
// This is a convenience wrapper around Format.
func (f *Formatter) FormatString() (string, error) {
	b, err := f.Format()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// WriteTo implements io.WriterTo by formatting the current terminal state and
// writing it to w. It does not allocate a buffer for the complete result. The
// returned count includes bytes accepted before an error.
//
// libghostty calls w synchronously for each piece of output, so WriteTo may
// block if w blocks. The writer must not call methods on f or its terminal.
// C: ghostty_formatter_format
func (f *Formatter) WriteTo(w io.Writer) (int64, error) {
	bridge := &f.writer
	if err := bridge.reset(w); err != nil {
		return 0, err
	}

	tls := getTLS()
	defer putTLS(tls)
	result := vt.Xghostty_formatter_format(tls, f.ptr, bridge.cWriter())
	written := bridge.written
	callbackErr := bridge.err
	bridge.writer = nil
	return written, resultErrorWithCallback(result, callbackErr)
}
