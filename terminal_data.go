package libghostty

// Terminal data getters wrapping ghostty_terminal_get().
// Functions are ordered alphabetically.

import (
	"errors"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// TerminalData identifies a data field for terminal queries.
// C: GhosttyTerminalData
type TerminalData int

const (
	// TerminalDataInvalid is an invalid / sentinel value.
	TerminalDataInvalid TerminalData = vt.GHOSTTY_TERMINAL_DATA_INVALID

	// TerminalDataCols is the terminal width in cells (uint16_t).
	TerminalDataCols TerminalData = vt.GHOSTTY_TERMINAL_DATA_COLS

	// TerminalDataRows is the terminal height in cells (uint16_t).
	TerminalDataRows TerminalData = vt.GHOSTTY_TERMINAL_DATA_ROWS

	// TerminalDataCursorX is the cursor column position, 0-indexed (uint16_t).
	TerminalDataCursorX TerminalData = vt.GHOSTTY_TERMINAL_DATA_CURSOR_X

	// TerminalDataCursorY is the cursor row position within the active area,
	// 0-indexed (uint16_t).
	TerminalDataCursorY TerminalData = vt.GHOSTTY_TERMINAL_DATA_CURSOR_Y

	// TerminalDataCursorPendingWrap indicates whether the cursor has a
	// pending wrap (bool).
	TerminalDataCursorPendingWrap TerminalData = vt.GHOSTTY_TERMINAL_DATA_CURSOR_PENDING_WRAP

	// TerminalDataActiveScreen is the currently active screen
	// (GhosttyTerminalScreen).
	TerminalDataActiveScreen TerminalData = vt.GHOSTTY_TERMINAL_DATA_ACTIVE_SCREEN

	// TerminalDataCursorVisible indicates whether the cursor is visible,
	// DEC mode 25 (bool).
	TerminalDataCursorVisible TerminalData = vt.GHOSTTY_TERMINAL_DATA_CURSOR_VISIBLE

	// TerminalDataKittyKeyboardFlags is the current Kitty keyboard protocol
	// flags (uint8_t).
	TerminalDataKittyKeyboardFlags TerminalData = vt.GHOSTTY_TERMINAL_DATA_KITTY_KEYBOARD_FLAGS

	// TerminalDataScrollbar is the scrollbar state for the terminal viewport
	// (GhosttyTerminalScrollbar).
	TerminalDataScrollbar TerminalData = vt.GHOSTTY_TERMINAL_DATA_SCROLLBAR

	// TerminalDataCursorStyle is the current SGR style of the cursor
	// (GhosttyStyle).
	TerminalDataCursorStyle TerminalData = vt.GHOSTTY_TERMINAL_DATA_CURSOR_STYLE

	// TerminalDataMouseTracking indicates whether any mouse tracking mode
	// is active (bool).
	TerminalDataMouseTracking TerminalData = vt.GHOSTTY_TERMINAL_DATA_MOUSE_TRACKING

	// TerminalDataTitle is the terminal title as set by escape sequences
	// (GhosttyString).
	TerminalDataTitle TerminalData = vt.GHOSTTY_TERMINAL_DATA_TITLE

	// TerminalDataPwd is the terminal's current working directory as set
	// by escape sequences (GhosttyString).
	TerminalDataPwd TerminalData = vt.GHOSTTY_TERMINAL_DATA_PWD

	// TerminalDataTotalRows is the total number of rows in the active screen
	// including scrollback (size_t).
	TerminalDataTotalRows TerminalData = vt.GHOSTTY_TERMINAL_DATA_TOTAL_ROWS

	// TerminalDataScrollbackRows is the number of scrollback rows (size_t).
	TerminalDataScrollbackRows TerminalData = vt.GHOSTTY_TERMINAL_DATA_SCROLLBACK_ROWS

	// TerminalDataWidthPx is the total terminal width in pixels (uint32_t).
	TerminalDataWidthPx TerminalData = vt.GHOSTTY_TERMINAL_DATA_WIDTH_PX

	// TerminalDataHeightPx is the total terminal height in pixels (uint32_t).
	TerminalDataHeightPx TerminalData = vt.GHOSTTY_TERMINAL_DATA_HEIGHT_PX

	// TerminalDataColorForeground is the effective foreground color
	// (GhosttyColorRgb).
	TerminalDataColorForeground TerminalData = vt.GHOSTTY_TERMINAL_DATA_COLOR_FOREGROUND

	// TerminalDataColorBackground is the effective background color
	// (GhosttyColorRgb).
	TerminalDataColorBackground TerminalData = vt.GHOSTTY_TERMINAL_DATA_COLOR_BACKGROUND

	// TerminalDataColorCursor is the effective cursor color
	// (GhosttyColorRgb).
	TerminalDataColorCursor TerminalData = vt.GHOSTTY_TERMINAL_DATA_COLOR_CURSOR

	// TerminalDataColorPalette is the current 256-color palette
	// (GhosttyColorRgb[256]).
	TerminalDataColorPalette TerminalData = vt.GHOSTTY_TERMINAL_DATA_COLOR_PALETTE

	// TerminalDataColorForegroundDefault is the default foreground color,
	// ignoring OSC overrides (GhosttyColorRgb).
	TerminalDataColorForegroundDefault TerminalData = vt.GHOSTTY_TERMINAL_DATA_COLOR_FOREGROUND_DEFAULT

	// TerminalDataColorBackgroundDefault is the default background color,
	// ignoring OSC overrides (GhosttyColorRgb).
	TerminalDataColorBackgroundDefault TerminalData = vt.GHOSTTY_TERMINAL_DATA_COLOR_BACKGROUND_DEFAULT

	// TerminalDataColorCursorDefault is the default cursor color,
	// ignoring OSC overrides (GhosttyColorRgb).
	TerminalDataColorCursorDefault TerminalData = vt.GHOSTTY_TERMINAL_DATA_COLOR_CURSOR_DEFAULT

	// TerminalDataColorPaletteDefault is the default 256-color palette,
	// ignoring OSC overrides (GhosttyColorRgb[256]).
	TerminalDataColorPaletteDefault TerminalData = vt.GHOSTTY_TERMINAL_DATA_COLOR_PALETTE_DEFAULT

	// TerminalDataKittyImageStorageLimit is the Kitty image storage limit
	// in bytes for the active screen (uint64_t).
	TerminalDataKittyImageStorageLimit TerminalData = vt.GHOSTTY_TERMINAL_DATA_KITTY_IMAGE_STORAGE_LIMIT

	// TerminalDataKittyImageMediumFile indicates whether the file medium
	// is enabled for Kitty image loading (bool).
	TerminalDataKittyImageMediumFile TerminalData = vt.GHOSTTY_TERMINAL_DATA_KITTY_IMAGE_MEDIUM_FILE

	// TerminalDataKittyImageMediumTempFile is the directory allowed for
	// temporary-file image loading, or an empty string when disabled
	// (GhosttyString).
	TerminalDataKittyImageMediumTempFile TerminalData = vt.GHOSTTY_TERMINAL_DATA_KITTY_IMAGE_MEDIUM_TEMP_FILE

	// TerminalDataKittyImageMediumSharedMem indicates whether the shared
	// memory medium is enabled for Kitty image loading (bool).
	TerminalDataKittyImageMediumSharedMem TerminalData = vt.GHOSTTY_TERMINAL_DATA_KITTY_IMAGE_MEDIUM_SHARED_MEM

	// TerminalDataKittyGraphics is the Kitty graphics image storage for
	// the active screen (GhosttyKittyGraphics).
	TerminalDataKittyGraphics TerminalData = vt.GHOSTTY_TERMINAL_DATA_KITTY_GRAPHICS

	// TerminalDataSelection is the active screen's current selection
	// (GhosttySelection).
	TerminalDataSelection TerminalData = vt.GHOSTTY_TERMINAL_DATA_SELECTION

	// TerminalDataViewportActive indicates whether the viewport is pinned
	// to the active terminal area rather than scrolled into history (bool).
	TerminalDataViewportActive TerminalData = vt.GHOSTTY_TERMINAL_DATA_VIEWPORT_ACTIVE

	// TerminalDataVTProcessingError indicates whether VT processing ever
	// encountered a non-gracefully handled semantic update failure (bool).
	TerminalDataVTProcessingError TerminalData = vt.GHOSTTY_TERMINAL_DATA_VT_PROCESSING_ERROR

	// TerminalDataScrollbackMaxBytes is the configured approximate byte
	// limit (size_t), or GHOSTTY_NO_VALUE when unlimited.
	TerminalDataScrollbackMaxBytes TerminalData = vt.GHOSTTY_TERMINAL_DATA_SCROLLBACK_MAX_BYTES

	// TerminalDataScrollbackMaxLines is the configured approximate physical
	// line limit (size_t), or GHOSTTY_NO_VALUE when unlimited.
	TerminalDataScrollbackMaxLines TerminalData = vt.GHOSTTY_TERMINAL_DATA_SCROLLBACK_MAX_LINES

	// TerminalDataContinuationMaxBytes is the configured maximum retained VT
	// continuation size in bytes (size_t). Zero means tracking is disabled.
	TerminalDataContinuationMaxBytes TerminalData = vt.GHOSTTY_TERMINAL_DATA_CONTINUATION_MAX_BYTES

	// TerminalDataMode gets the current value of a terminal mode. The caller
	// initializes the mode field of a GhosttyTerminalModeConfig and the query
	// writes its value field (GhosttyTerminalModeConfig).
	TerminalDataMode TerminalData = vt.GHOSTTY_TERMINAL_DATA_MODE

	// TerminalDataVTGround indicates whether VT processing is between UTF-8
	// codepoints and terminal sequences (bool).
	TerminalDataVTGround TerminalData = vt.GHOSTTY_TERMINAL_DATA_VT_GROUND

	// TerminalDataCursorAtPrompt indicates whether the cursor is currently at
	// a semantic shell prompt or input area (bool).
	TerminalDataCursorAtPrompt TerminalData = vt.GHOSTTY_TERMINAL_DATA_CURSOR_AT_PROMPT

	// TerminalDataKittyClipboardWriteMaxBytes is the configured maximum number
	// of decoded bytes in a single Kitty clipboard protocol write (size_t).
	TerminalDataKittyClipboardWriteMaxBytes TerminalData = vt.GHOSTTY_TERMINAL_DATA_CLIPBOARD_WRITE_MAX_BYTES

	// TerminalDataMouseShape is the mouse pointer shape requested through
	// OSC 22 (GhosttyMouseShape). See [Terminal.MouseShape].
	TerminalDataMouseShape TerminalData = vt.GHOSTTY_TERMINAL_DATA_MOUSE_SHAPE

	// TerminalDataMemoryUsage is the memory held by the terminal
	// (GhosttyTerminalMemoryUsage). See [Terminal.MemoryUsage].
	TerminalDataMemoryUsage TerminalData = vt.GHOSTTY_TERMINAL_DATA_MEMORY_USAGE
)

// terminalGet reads one terminal data field of C type T. The value is read
// into TLS stack memory, because the translated code must not write through
// a pointer to a Go variable.
func terminalGet[T any](t *Terminal, data vt.GhosttyTerminalData) (T, error) {
	tls := getTLS()
	defer putTLS(tls)
	var zero T
	n := int(unsafe.Sizeof(zero))
	p := tls.Alloc(n)
	defer tls.Free(n)
	clear(unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
	if err := resultError(vt.Xghostty_terminal_get(tls, t.ptr, data, p)); err != nil {
		return zero, err
	}
	return *(*T)(unsafe.Pointer(p)), nil
}

// terminalGetBool reads a bool terminal data field.
func terminalGetBool(t *Terminal, data vt.GhosttyTerminalData) (bool, error) {
	v, err := terminalGet[uint8](t, data)
	return v != 0, err
}

// terminalGetString reads a GhosttyString terminal data field and copies
// the string into Go memory.
func terminalGetString(t *Terminal, data vt.GhosttyTerminalData) (string, error) {
	s, err := terminalGet[vt.GhosttyString](t, data)
	if err != nil {
		return "", err
	}
	return string(goBytes(s.Fptr, int(s.Flen1))), nil
}

// ActiveScreen returns which screen buffer is currently active.
func (t *Terminal) ActiveScreen() (TerminalScreen, error) {
	v, err := terminalGet[vt.GhosttyTerminalScreen](t, vt.GHOSTTY_TERMINAL_DATA_ACTIVE_SCREEN)
	return TerminalScreen(v), err
}

// Cols returns the terminal width in cells.
func (t *Terminal) Cols() (uint16, error) {
	return terminalGet[uint16](t, vt.GHOSTTY_TERMINAL_DATA_COLS)
}

// ColorBackground returns the effective background color (OSC override
// or default). Returns nil if no background color is set.
func (t *Terminal) ColorBackground() (*ColorRGB, error) {
	return t.getColorRGB(vt.GHOSTTY_TERMINAL_DATA_COLOR_BACKGROUND)
}

// ColorBackgroundDefault returns the default background color, ignoring
// any OSC override. Returns nil if no default is set.
func (t *Terminal) ColorBackgroundDefault() (*ColorRGB, error) {
	return t.getColorRGB(vt.GHOSTTY_TERMINAL_DATA_COLOR_BACKGROUND_DEFAULT)
}

// ColorCursor returns the effective cursor color (OSC override or
// default). Returns nil if no cursor color is set.
func (t *Terminal) ColorCursor() (*ColorRGB, error) {
	return t.getColorRGB(vt.GHOSTTY_TERMINAL_DATA_COLOR_CURSOR)
}

// ColorCursorDefault returns the default cursor color, ignoring any
// OSC override. Returns nil if no default is set.
func (t *Terminal) ColorCursorDefault() (*ColorRGB, error) {
	return t.getColorRGB(vt.GHOSTTY_TERMINAL_DATA_COLOR_CURSOR_DEFAULT)
}

// ColorForeground returns the effective foreground color (OSC override
// or default). Returns nil if no foreground color is set.
func (t *Terminal) ColorForeground() (*ColorRGB, error) {
	return t.getColorRGB(vt.GHOSTTY_TERMINAL_DATA_COLOR_FOREGROUND)
}

// ColorForegroundDefault returns the default foreground color, ignoring
// any OSC override. Returns nil if no default is set.
func (t *Terminal) ColorForegroundDefault() (*ColorRGB, error) {
	return t.getColorRGB(vt.GHOSTTY_TERMINAL_DATA_COLOR_FOREGROUND_DEFAULT)
}

// ColorPalette returns the current 256-color palette (with any OSC
// overrides applied).
func (t *Terminal) ColorPalette() (*Palette, error) {
	return t.getPalette(vt.GHOSTTY_TERMINAL_DATA_COLOR_PALETTE)
}

// ColorPaletteDefault returns the default 256-color palette, ignoring
// any OSC overrides.
func (t *Terminal) ColorPaletteDefault() (*Palette, error) {
	return t.getPalette(vt.GHOSTTY_TERMINAL_DATA_COLOR_PALETTE_DEFAULT)
}

// GetMulti queries multiple terminal data fields in a single library call.
// This is a low-level function. Prefer the typed getters (Cols, Rows,
// CursorX, etc.) for normal use. GetMulti is useful when you need many
// fields at once and want to avoid per-field call overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/terminal.h, GhosttyTerminalData enum).
// Use a Go type with the same size as the C type, such as uint32
// for uint32_t, bool for bool, and int32 for an enum.
//
// GetMulti returns an error if keys and values have different lengths
// or if a key cannot be read. In the second case, values for the keys
// before it may already have been written.
//
// Example:
//
//	var cols, rows uint16
//	err := t.GetMulti(
//		[]TerminalData{TerminalDataCols, TerminalDataRows},
//		[]unsafe.Pointer{unsafe.Pointer(&cols), unsafe.Pointer(&rows)},
//	)
//
// C: ghostty_terminal_get_multi
func (t *Terminal) GetMulti(keys []TerminalData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[vt.GhosttyTerminalData](&args, keys, values)
	if err != nil {
		return err
	}
	defer args.free()
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_terminal_get_multi(
		tls,
		t.ptr,
		uint64(len(keys)),
		uintptr(unsafe.Pointer(cKeys)),
		uintptr(unsafe.Pointer(args.values)),
		0,
	))
}

// CursorAtPrompt reports whether the cursor is currently at a semantic shell
// prompt or input area. It returns false when semantic prompt information is
// unavailable or the alternate screen is active.
func (t *Terminal) CursorAtPrompt() (bool, error) {
	return terminalGetBool(t, vt.GHOSTTY_TERMINAL_DATA_CURSOR_AT_PROMPT)
}

// CursorPendingWrap reports whether the cursor has a pending wrap
// (the next printed character will soft-wrap to the next line).
func (t *Terminal) CursorPendingWrap() (bool, error) {
	return terminalGetBool(t, vt.GHOSTTY_TERMINAL_DATA_CURSOR_PENDING_WRAP)
}

// CursorStyle returns the current SGR style of the cursor. This is
// the style that will be applied to newly printed characters.
func (t *Terminal) CursorStyle() (*Style, error) {
	cs, err := terminalGet[vt.GhosttyStyle](t, vt.GHOSTTY_TERMINAL_DATA_CURSOR_STYLE)
	if err != nil {
		return nil, err
	}
	return &Style{c: cs}, nil
}

// CursorVisible reports whether the cursor is visible (DEC mode 25).
func (t *Terminal) CursorVisible() (bool, error) {
	return terminalGetBool(t, vt.GHOSTTY_TERMINAL_DATA_CURSOR_VISIBLE)
}

// CursorX returns the cursor column position (0-indexed).
func (t *Terminal) CursorX() (uint16, error) {
	return terminalGet[uint16](t, vt.GHOSTTY_TERMINAL_DATA_CURSOR_X)
}

// CursorY returns the cursor row position within the active area
// (0-indexed).
func (t *Terminal) CursorY() (uint16, error) {
	return terminalGet[uint16](t, vt.GHOSTTY_TERMINAL_DATA_CURSOR_Y)
}

// ContinuationMaxBytes returns the configured maximum retained VT
// continuation size. Zero means continuation tracking is disabled.
func (t *Terminal) ContinuationMaxBytes() (uint, error) {
	v, err := terminalGet[uint64](t, vt.GHOSTTY_TERMINAL_DATA_CONTINUATION_MAX_BYTES)
	return uint(v), err
}

// HeightPx returns the total terminal height in pixels
// (rows * cell_height_px as set by Resize).
func (t *Terminal) HeightPx() (uint32, error) {
	return terminalGet[uint32](t, vt.GHOSTTY_TERMINAL_DATA_HEIGHT_PX)
}

// KittyClipboardWriteMaxBytes returns the maximum number of decoded bytes in a
// single Kitty clipboard protocol write. The limit applies to OSC 5522 and not
// OSC 52.
func (t *Terminal) KittyClipboardWriteMaxBytes() (uint, error) {
	v, err := terminalGet[uint64](t, vt.GHOSTTY_TERMINAL_DATA_CLIPBOARD_WRITE_MAX_BYTES)
	return uint(v), err
}

// KittyKeyboardFlags returns the current Kitty keyboard protocol flags.
func (t *Terminal) KittyKeyboardFlags() (KittyKeyFlags, error) {
	v, err := terminalGet[uint8](t, vt.GHOSTTY_TERMINAL_DATA_KITTY_KEYBOARD_FLAGS)
	return KittyKeyFlags(v), err
}

// KittyImageMediumFile reports whether Kitty image loading via arbitrary
// files is enabled for the active screen.
func (t *Terminal) KittyImageMediumFile() (bool, error) {
	return terminalGetBool(t, vt.GHOSTTY_TERMINAL_DATA_KITTY_IMAGE_MEDIUM_FILE)
}

// KittyImageMediumSharedMem reports whether Kitty image loading via shared
// memory is enabled for the active screen.
func (t *Terminal) KittyImageMediumSharedMem() (bool, error) {
	return terminalGetBool(t, vt.GHOSTTY_TERMINAL_DATA_KITTY_IMAGE_MEDIUM_SHARED_MEM)
}

// KittyImageMediumTempFile returns the directory allowed for Kitty
// temporary-file image loading. An empty string means the medium is disabled.
func (t *Terminal) KittyImageMediumTempFile() (string, error) {
	return terminalGetString(t, vt.GHOSTTY_TERMINAL_DATA_KITTY_IMAGE_MEDIUM_TEMP_FILE)
}

// KittyImageStorageLimit returns the active screen's Kitty image storage
// limit in bytes. Zero means the protocol is disabled.
func (t *Terminal) KittyImageStorageLimit() (uint64, error) {
	return terminalGet[uint64](t, vt.GHOSTTY_TERMINAL_DATA_KITTY_IMAGE_STORAGE_LIMIT)
}

// MemoryUsage reports how much memory the terminal holds. See
// [MemoryUsage] for what each figure means.
//
// MemoryUsage never decompresses scrollback, but it does visit every page,
// so its cost grows with the amount of scrollback. Call it periodically,
// for example on a timer, rather than after every write.
func (t *Terminal) MemoryUsage() (MemoryUsage, error) {
	tls := getTLS()
	defer putTLS(tls)
	n := int(unsafe.Sizeof(vt.GhosttyTerminalMemoryUsage{}))
	p := tls.Alloc(n)
	defer tls.Free(n)
	clear(unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
	// A sized struct: the caller sets its size field so the library knows
	// which fields exist (GHOSTTY_INIT_SIZED).
	v := (*vt.GhosttyTerminalMemoryUsage)(unsafe.Pointer(p))
	v.Fsize = uint64(n)
	if err := resultError(vt.Xghostty_terminal_get(tls, t.ptr, vt.GHOSTTY_TERMINAL_DATA_MEMORY_USAGE, p)); err != nil {
		return MemoryUsage{}, err
	}
	return MemoryUsage{
		CompressionSupported: v.Fcompression_supported != 0,
		Primary: ScreenMemoryUsage{
			Pages:           v.Fprimary_pages,
			VirtualBytes:    v.Fprimary_virtual_bytes,
			ResidentBytes:   v.Fprimary_resident_bytes,
			CompressedPages: v.Fprimary_compressed_pages,
			CompressedBytes: v.Fprimary_compressed_bytes,
			ImageBytes:      v.Fprimary_image_bytes,
		},
		Alternate: ScreenMemoryUsage{
			Pages:           v.Falternate_pages,
			VirtualBytes:    v.Falternate_virtual_bytes,
			ResidentBytes:   v.Falternate_resident_bytes,
			CompressedPages: v.Falternate_compressed_pages,
			CompressedBytes: v.Falternate_compressed_bytes,
			ImageBytes:      v.Falternate_image_bytes,
		},
	}, nil
}

// Mode returns the current value of a terminal mode.
func (t *Terminal) Mode(mode Mode) (bool, error) {
	tls := getTLS()
	defer putTLS(tls)
	n := int(unsafe.Sizeof(vt.GhosttyTerminalModeConfig{}))
	p := tls.Alloc(n)
	defer tls.Free(n)
	config := (*vt.GhosttyTerminalModeConfig)(unsafe.Pointer(p))
	*config = vt.GhosttyTerminalModeConfig{Fmode: vt.GhosttyMode(mode)}
	if err := resultError(vt.Xghostty_terminal_get(
		tls,
		t.ptr,
		vt.GHOSTTY_TERMINAL_DATA_MODE,
		p,
	)); err != nil {
		return false, err
	}
	return config.Fvalue != 0, nil
}

// MouseTracking reports whether any mouse tracking mode is active.
func (t *Terminal) MouseTracking() (bool, error) {
	return terminalGetBool(t, vt.GHOSTTY_TERMINAL_DATA_MOUSE_TRACKING)
}

// MouseShape returns the mouse pointer shape requested through OSC 22 by the
// program running in the terminal. The initial shape is [MouseShapeText].
// This value does not include changes made by the host application, such as
// showing a hand when the pointer is over a link.
func (t *Terminal) MouseShape() (MouseShape, error) {
	v, err := terminalGet[vt.GhosttyMouseShape](t, vt.GHOSTTY_TERMINAL_DATA_MOUSE_SHAPE)
	return MouseShape(v), err
}

// Pwd returns the terminal's current working directory as set by
// escape sequences (e.g. OSC 7). Returns an empty string if unset.
// The returned string is copied; it remains valid after subsequent
// calls to VTWrite or Reset.
func (t *Terminal) Pwd() (string, error) {
	return terminalGetString(t, vt.GHOSTTY_TERMINAL_DATA_PWD)
}

// Rows returns the terminal height in cells.
func (t *Terminal) Rows() (uint16, error) {
	return terminalGet[uint16](t, vt.GHOSTTY_TERMINAL_DATA_ROWS)
}

// Selection returns the active screen selection. It returns nil when
// there is no active selection.
//
// The returned selection contains borrowed grid references into terminal
// internals. Any later mutating terminal operation may invalidate them, so
// callers should read and cache needed data immediately.
func (t *Terminal) Selection() (*Selection, error) {
	cs, err := terminalGet[vt.GhosttySelection](t, vt.GHOSTTY_TERMINAL_DATA_SELECTION)
	if err != nil {
		var ge *Error
		if errors.As(err, &ge) && ge.Result == ResultNoValue {
			return nil, nil
		}
		return nil, err
	}

	sel := selectionFromC(cs)
	return &sel, nil
}

// Scrollbar returns the scrollbar state for the terminal viewport.
// This may be expensive to calculate depending on the viewport position;
// call only as needed.
func (t *Terminal) Scrollbar() (Scrollbar, error) {
	v, err := terminalGet[vt.GhosttyTerminalScrollbar](t, vt.GHOSTTY_TERMINAL_DATA_SCROLLBAR)
	if err != nil {
		return Scrollbar{}, err
	}
	return Scrollbar{
		Total:  v.Ftotal,
		Offset: v.Foffset,
		Len:    v.Flen1,
	}, nil
}

// ScrollbackRows returns the number of scrollback rows (total rows
// minus viewport rows).
func (t *Terminal) ScrollbackRows() (uint, error) {
	v, err := terminalGet[uint64](t, vt.GHOSTTY_TERMINAL_DATA_SCROLLBACK_ROWS)
	return uint(v), err
}

// ScrollbackMaxBytes returns the configured approximate scrollback byte
// limit. Nil means the byte limit is unlimited.
func (t *Terminal) ScrollbackMaxBytes() (*uint, error) {
	return t.getOptionalSize(vt.GHOSTTY_TERMINAL_DATA_SCROLLBACK_MAX_BYTES)
}

// ScrollbackMaxLines returns the configured approximate physical line
// limit. Nil means the line limit is unlimited.
func (t *Terminal) ScrollbackMaxLines() (*uint, error) {
	return t.getOptionalSize(vt.GHOSTTY_TERMINAL_DATA_SCROLLBACK_MAX_LINES)
}

// Title returns the terminal title as set by escape sequences
// (e.g. OSC 0/2). Returns an empty string if unset. The returned
// string is copied; it remains valid after subsequent calls to
// VTWrite or Reset.
func (t *Terminal) Title() (string, error) {
	return terminalGetString(t, vt.GHOSTTY_TERMINAL_DATA_TITLE)
}

// TotalRows returns the total number of rows in the active screen
// including scrollback.
func (t *Terminal) TotalRows() (uint, error) {
	v, err := terminalGet[uint64](t, vt.GHOSTTY_TERMINAL_DATA_TOTAL_ROWS)
	return uint(v), err
}

// VTGround reports whether VT processing is at the stateless point between
// UTF-8 codepoints and terminal sequences. Out-of-band VT data can be safely
// inserted while this returns true.
func (t *Terminal) VTGround() (bool, error) {
	return terminalGetBool(t, vt.GHOSTTY_TERMINAL_DATA_VT_GROUND)
}

// VTProcessingError reports whether VT processing has ever encountered a
// non-gracefully handled failure that may have prevented a semantic update.
// Reset does not clear this informational flag.
func (t *Terminal) VTProcessingError() (bool, error) {
	return terminalGetBool(t, vt.GHOSTTY_TERMINAL_DATA_VT_PROCESSING_ERROR)
}

// ViewportActive reports whether the viewport is pinned to the active
// terminal area. It returns false when the user has scrolled into history.
func (t *Terminal) ViewportActive() (bool, error) {
	return terminalGetBool(t, vt.GHOSTTY_TERMINAL_DATA_VIEWPORT_ACTIVE)
}

// WidthPx returns the total terminal width in pixels
// (cols * cell_width_px as set by Resize).
func (t *Terminal) WidthPx() (uint32, error) {
	return terminalGet[uint32](t, vt.GHOSTTY_TERMINAL_DATA_WIDTH_PX)
}

// getColorRGB is a helper that reads a single ColorRGB value from the
// terminal. Returns nil (without error) when the result is NO_VALUE.
func (t *Terminal) getColorRGB(data vt.GhosttyTerminalData) (*ColorRGB, error) {
	c, err := terminalGet[vt.GhosttyColorRgb](t, data)
	if err != nil {
		var ge *Error
		if errors.As(err, &ge) && ge.Result == ResultNoValue {
			return nil, nil
		}
		return nil, err
	}
	return &ColorRGB{R: c.Fr, G: c.Fg, B: c.Fb}, nil
}

// getPalette is a helper that reads a full 256-color palette from the
// terminal.
func (t *Terminal) getPalette(data vt.GhosttyTerminalData) (*Palette, error) {
	cp, err := terminalGet[[PaletteSize]vt.GhosttyColorRgb](t, data)
	if err != nil {
		return nil, err
	}
	var p Palette
	for i, c := range cp {
		p[i] = ColorRGB{R: c.Fr, G: c.Fg, B: c.Fb}
	}
	return &p, nil
}

// getOptionalSize reads a size_t terminal value that uses NO_VALUE to mean
// unlimited.
func (t *Terminal) getOptionalSize(data vt.GhosttyTerminalData) (*uint, error) {
	v, err := terminalGet[uint64](t, data)
	if err != nil {
		var ge *Error
		if errors.As(err, &ge) && ge.Result == ResultNoValue {
			return nil, nil
		}
		return nil, err
	}
	value := uint(v)
	return &value, nil
}
