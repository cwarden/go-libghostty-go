package libghostty

// Render-state row cell iterator wrapping the
// GhosttyRenderStateRowCells C APIs.

import (
	"errors"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// RenderStateRowCellsData identifies a data field for render state cell
// queries.
// C: GhosttyRenderStateRowCellsData
type RenderStateRowCellsData int

const (
	// RenderStateRowCellsDataInvalid is an invalid / sentinel value.
	RenderStateRowCellsDataInvalid RenderStateRowCellsData = vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_INVALID

	// RenderStateRowCellsDataRaw is the raw cell value (GhosttyCell).
	RenderStateRowCellsDataRaw RenderStateRowCellsData = vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_RAW

	// RenderStateRowCellsDataStyle is the style for the current cell
	// (GhosttyStyle).
	RenderStateRowCellsDataStyle RenderStateRowCellsData = vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_STYLE

	// RenderStateRowCellsDataGraphemesLen is the total number of grapheme
	// codepoints including the base codepoint (uint32_t).
	RenderStateRowCellsDataGraphemesLen RenderStateRowCellsData = vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_LEN

	// RenderStateRowCellsDataBgColor is the resolved background color of
	// the cell (GhosttyColorRgb).
	RenderStateRowCellsDataBgColor RenderStateRowCellsData = vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_BG_COLOR

	// RenderStateRowCellsDataFgColor is the resolved foreground color of
	// the cell (GhosttyColorRgb).
	RenderStateRowCellsDataFgColor RenderStateRowCellsData = vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_FG_COLOR

	// RenderStateRowCellsDataSelected indicates whether the cell is contained
	// within the current selection (bool).
	RenderStateRowCellsDataSelected RenderStateRowCellsData = vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_SELECTED

	// RenderStateRowCellsDataHasStyling indicates whether the cell has any
	// explicit styling (bool).
	RenderStateRowCellsDataHasStyling RenderStateRowCellsData = vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_HAS_STYLING

	// RenderStateRowCellsDataGraphemesUTF8 encodes the current cell's full
	// grapheme cluster as UTF-8 into a caller-provided GhosttyBuffer.
	RenderStateRowCellsDataGraphemesUTF8 RenderStateRowCellsData = vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_UTF8
)

// RenderCellStyle is a reusable, resolved style snapshot for the current
// render-state cell. It corresponds to the data returned by
// ghostty_render_state_row_cells_get for style, foreground color, background
// color, and has-styling, flattened into one Go value for hot render paths.
type RenderCellStyle struct {
	// Foreground is the resolved foreground color when HasForeground is true.
	Foreground ColorRGB

	// Background is the resolved background color when HasBackground is true.
	Background ColorRGB

	// HasForeground reports whether Foreground contains a resolved foreground
	// color. When false, the renderer should use its default foreground.
	HasForeground bool

	// HasBackground reports whether Background contains a resolved background
	// color. When false, the renderer should use its default background.
	HasBackground bool

	// HasStyling reports whether the cell has any explicit non-default styling.
	HasStyling bool

	// Bold reports whether bold text is set.
	Bold bool

	// Faint reports whether faint/dim text is set.
	Faint bool

	// Italic reports whether italic text is set.
	Italic bool

	// Underline reports whether any underline style is set.
	Underline bool

	// Strikethrough reports whether strikethrough text is set.
	Strikethrough bool

	// Inverse reports whether inverse video is set.
	Inverse bool
}

// The per-cell getters are called for every cell of every frame, so each
// RenderStateRowCells owns one block of C memory for the values libghostty
// writes, allocated once, instead of allocating per call. These are the
// offsets of the values within the block, each aligned for its type.
const (
	// scratchBuffer holds a GhosttyBuffer or GhosttyCodepointBuffer.
	scratchBuffer = 0
	// scratchStyle holds a GhosttyStyle.
	scratchStyle = 32
	// scratchColor holds a GhosttyColorRgb.
	scratchColor = 48
	// scratchRaw holds a GhosttyCell.
	scratchRaw = 56
	// scratchWide holds a GhosttyCellWide.
	scratchWide = 64
	// scratchBool holds a bool.
	scratchBool = 68
	// cellsScratchSize is the size of the block.
	cellsScratchSize = 72
)

// RenderStateRowCells iterates over cells in a render-state row.
// Create one with NewRenderStateRowCells, populate it via
// [RenderStateRowIterator.Cells], then advance with
// [RenderStateRowCells.Next] (or jump with [RenderStateRowCells.Select])
// and read data with getter methods.
//
// A single instance can be reused across rows to avoid repeated
// allocation. The cells view is only valid until the next call to
// [RenderState.Update]. Do not use it while [RenderState.Update] may
// run on the same render state. Extracted [Cell], [Style], color
// values, and grapheme slices are copied values and may be retained.
//
// C: GhosttyRenderStateRowCells
type RenderStateRowCells struct {
	ptr vt.GhosttyRenderStateRowCells

	// scratch is C memory of cellsScratchSize bytes that the getters read
	// values into. See the scratch* offsets.
	scratch uintptr
}

// NewRenderStateRowCells creates a new row cells instance. The
// instance is empty until populated via RenderStateRowIterator.Cells.
func NewRenderStateRowCells() (*RenderStateRowCells, error) {
	tls := getTLS()
	defer putTLS(tls)
	out := tls.Alloc(handleSize)
	defer tls.Free(handleSize)
	if err := resultError(vt.Xghostty_render_state_row_cells_new(tls, 0, out)); err != nil {
		return nil, err
	}
	rc := &RenderStateRowCells{ptr: *(*uintptr)(unsafe.Pointer(out))}
	// The block holds a buffer pointer, so it starts out zeroed.
	rc.scratch = uintptr(allocZeroed(cellsScratchSize))
	if rc.scratch == 0 {
		vt.Xghostty_render_state_row_cells_free(tls, rc.ptr)
		return nil, &Error{Result: ResultOutOfMemory}
	}
	return rc, nil
}

// Close frees the underlying row cells handle. After this call,
// the instance must not be used, except that calling Close again is a
// safe no-op.
func (rc *RenderStateRowCells) Close() {
	if rc == nil || rc.ptr == 0 {
		return
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_render_state_row_cells_free(tls, rc.ptr)
	rc.ptr = 0
	Free(unsafe.Pointer(rc.scratch), cellsScratchSize)
	rc.scratch = 0
}

// Next advances the iterator to the next cell. Returns true if the
// iterator moved successfully and cell data is available. Returns
// false when there are no more cells.
func (rc *RenderStateRowCells) Next() bool {
	tls := getTLS()
	defer putTLS(tls)
	return vt.Xghostty_render_state_row_cells_next(tls, rc.ptr) != 0
}

// Select positions the iterator at the given column index (0-based)
// so that subsequent reads return data for that cell.
func (rc *RenderStateRowCells) Select(x uint16) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_render_state_row_cells_select(tls, rc.ptr, x))
}

// GetMulti queries multiple render-state cell data fields in a single
// call. This is a low-level function. Prefer the typed getters
// (Raw, Style, Graphemes, BgColor, FgColor) for normal use. GetMulti
// is useful when you need many fields at once and want to avoid
// per-field call overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/render.h, GhosttyRenderStateRowCellsData enum).
// Use a Go type with the same size as the C type, such as uint32
// for uint32_t, bool for bool, and int32 for an enum.
//
// GetMulti returns an error if keys and values have different lengths
// or if a key cannot be read. In the second case, values for the keys
// before it may already have been written.
//
// Example:
//
//	var graphemesLen uint32
//	var hasStyling bool
//	err := rc.GetMulti(
//		[]RenderStateRowCellsData{RenderStateRowCellsDataGraphemesLen, RenderStateRowCellsDataHasStyling},
//		[]unsafe.Pointer{unsafe.Pointer(&graphemesLen), unsafe.Pointer(&hasStyling)},
//	)
//
// C: ghostty_render_state_row_cells_get_multi
func (rc *RenderStateRowCells) GetMulti(keys []RenderStateRowCellsData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[vt.GhosttyRenderStateRowCellsData](&args, keys, values)
	if err != nil {
		return err
	}
	defer args.free()
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_render_state_row_cells_get_multi(
		tls,
		rc.ptr,
		uint64(len(keys)),
		uintptr(unsafe.Pointer(cKeys)),
		uintptr(unsafe.Pointer(args.values)),
		0,
	))
}

// get reads data into the scratch block at offset and returns the error.
func (rc *RenderStateRowCells) get(tls *libc.TLS, data vt.GhosttyRenderStateRowCellsData, offset uintptr) error {
	return resultError(vt.Xghostty_render_state_row_cells_get(tls, rc.ptr, data, rc.scratch+offset))
}

// Raw returns the raw Cell value for the current iterator position.
// The returned Cell can be used with the same getter methods as cells
// obtained from GridRef.
func (rc *RenderStateRowCells) Raw() (*Cell, error) {
	tls := getTLS()
	defer putTLS(tls)
	if err := rc.get(tls, vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_RAW, scratchRaw); err != nil {
		return nil, err
	}
	return &Cell{c: *(*vt.GhosttyCell)(unsafe.Pointer(rc.scratch + scratchRaw))}, nil
}

// Style returns the style for the current cell.
func (rc *RenderStateRowCells) Style() (*Style, error) {
	tls := getTLS()
	defer putTLS(tls)
	if err := rc.get(tls, vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_STYLE, scratchStyle); err != nil {
		return nil, err
	}
	return &Style{c: *(*vt.GhosttyStyle)(unsafe.Pointer(rc.scratch + scratchStyle))}, nil
}

// StyleInto fills dst with the reusable resolved style snapshot for the current
// cell. This is the allocation-reusing form of querying Style, FgColor,
// BgColor, and HasStyling separately: it reads the values into memory the
// iterator already owns, stores colors as values, and sets
// HasForeground/HasBackground when the corresponding color is present.
func (rc *RenderStateRowCells) StyleInto(dst *RenderCellStyle) error {
	if dst == nil {
		return errors.New("libghostty: nil RenderCellStyle")
	}

	tls := getTLS()
	defer putTLS(tls)
	var snap RenderCellStyle

	if err := rc.get(tls, vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_HAS_STYLING, scratchBool); err != nil {
		return err
	}
	snap.HasStyling = *(*uint8)(unsafe.Pointer(rc.scratch + scratchBool)) != 0

	if snap.HasStyling {
		*(*vt.GhosttyStyle)(unsafe.Pointer(rc.scratch + scratchStyle)) = vt.GhosttyStyle{}
		if err := rc.get(tls, vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_STYLE, scratchStyle); err != nil {
			return err
		}
		style := Style{c: *(*vt.GhosttyStyle)(unsafe.Pointer(rc.scratch + scratchStyle))}
		snap.Bold = style.Bold()
		snap.Faint = style.Faint()
		snap.Italic = style.Italic()
		// GHOSTTY_SGR_UNDERLINE_NONE is zero.
		snap.Underline = style.Underline() != vt.GHOSTTY_SGR_UNDERLINE_NONE
		snap.Strikethrough = style.Strikethrough()
		snap.Inverse = style.Inverse()
	}

	color := (*vt.GhosttyColorRgb)(unsafe.Pointer(rc.scratch + scratchColor))
	err := rc.get(tls, vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_FG_COLOR, scratchColor)
	switch {
	case err == nil:
		snap.HasForeground = true
		snap.Foreground = colorRGBFromC(*color)
	case !errors.Is(err, ErrInvalidValue):
		return err
	}

	err = rc.get(tls, vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_BG_COLOR, scratchColor)
	switch {
	case err == nil:
		snap.HasBackground = true
		snap.Background = colorRGBFromC(*color)
	case !errors.Is(err, ErrInvalidValue):
		return err
	}

	*dst = snap
	return nil
}

// Graphemes returns the full grapheme cluster codepoints for the
// current cell. The base codepoint is first, followed by any extra
// codepoints. Returns nil if the cell has no text.
func (rc *RenderStateRowCells) Graphemes() ([]uint32, error) {
	graphemes, err := rc.GraphemesInto(nil)
	if err != nil {
		return nil, err
	}
	if len(graphemes) == 0 {
		return nil, nil
	}
	return graphemes, nil
}

// GraphemesInto appends the full grapheme cluster codepoints for the
// current cell to dst and returns the extended slice. The base codepoint
// is appended first, followed by any extra codepoints. If the cell has no
// text, dst is returned unchanged.
//
// This is the allocation-reusing form of [RenderStateRowCells.Graphemes].
// Callers that want a per-cell scratch buffer should pass scratch[:0].
func (rc *RenderStateRowCells) GraphemesInto(dst []uint32) ([]uint32, error) {
	tls := getTLS()
	defer putTLS(tls)

	// Write into whatever spare capacity dst already has. A cell usually holds
	// one codepoint, so a reused buffer finishes here in a single call. The
	// result reports the total count either way, which sizes the retry below.
	oldLen := len(dst)
	result, n := rc.graphemesAppend(tls, dst[oldLen:cap(dst)])
	if result == vt.GHOSTTY_SUCCESS {
		return dst[:oldLen+n], nil
	}
	if result != vt.GHOSTTY_OUT_OF_SPACE {
		return dst, resultError(result)
	}

	// The cluster didn't fit, so grow to the size just reported and read again.
	grown := make([]uint32, oldLen, oldLen+n)
	copy(grown, dst)
	result, n = rc.graphemesAppend(tls, grown[oldLen:cap(grown)])
	if err := resultError(result); err != nil {
		return dst, err
	}
	return grown[:oldLen+n], nil
}

// graphemesAppend asks libghostty for the current cell's codepoints into
// dst, the spare capacity of the caller's buffer, and returns the result
// and the count libghostty reports: the number written, or the number
// needed with GHOSTTY_OUT_OF_SPACE.
func (rc *RenderStateRowCells) graphemesAppend(tls *libc.TLS, dst []uint32) (vt.GhosttyResult, int) {
	buf := (*vt.GhosttyCodepointBuffer)(unsafe.Pointer(rc.scratch + scratchBuffer))
	*buf = vt.GhosttyCodepointBuffer{}
	if len(dst) > 0 {
		var pin unpinFunc
		buf.Fptr, pin = pinUint32s(dst)
		defer pin()
		buf.Fcap1 = uint64(len(dst))
	}
	result := vt.Xghostty_render_state_row_cells_get(
		tls,
		rc.ptr,
		vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_BUF,
		uintptr(unsafe.Pointer(buf)),
	)
	return result, int(buf.Flen1)
}

// AppendGraphemes appends the current cell's grapheme cluster encoded as
// UTF-8 to dst and returns the extended byte slice. If the cell has no text,
// dst is returned unchanged.
//
// This is intended for renderers and text extractors that ultimately need
// bytes or strings and want to avoid allocating a temporary codepoint slice
// and per-cell string. The upstream UTF-8 getter keeps grapheme scratch storage
// on the libghostty side, so Go allocates nothing per cell in hot render
// loops.
func (rc *RenderStateRowCells) AppendGraphemes(dst []byte) ([]byte, error) {
	tls := getTLS()
	defer putTLS(tls)
	oldLen := len(dst)
	for {
		available := cap(dst) - oldLen
		result, n := rc.graphemesUTF8Append(tls, dst[oldLen:cap(dst)])
		switch result {
		case vt.GHOSTTY_SUCCESS:
			return dst[:oldLen+n], nil
		case vt.GHOSTTY_OUT_OF_SPACE:
			needed := n
			if needed <= available {
				return dst, resultError(result)
			}
			required := oldLen + needed
			newCap := cap(dst) * 2
			if newCap < 64 {
				newCap = 64
			}
			if newCap < required {
				newCap = required
			}
			grown := make([]byte, oldLen, newCap)
			copy(grown, dst[:oldLen])
			dst = grown
		default:
			return dst, resultError(result)
		}
	}
}

// graphemesUTF8Append asks libghostty for the current cell's text into
// dst, the spare capacity of the caller's buffer, and returns the result
// and the count libghostty reports: the number of bytes written, or the
// number needed with GHOSTTY_OUT_OF_SPACE.
func (rc *RenderStateRowCells) graphemesUTF8Append(tls *libc.TLS, dst []byte) (vt.GhosttyResult, int) {
	buf := (*vt.GhosttyBuffer)(unsafe.Pointer(rc.scratch + scratchBuffer))
	*buf = vt.GhosttyBuffer{}
	if len(dst) > 0 {
		var unpin func()
		buf.Fptr, unpin = pinBytes(dst)
		defer unpin()
		buf.Fcap1 = uint64(len(dst))
	}
	result := vt.Xghostty_render_state_row_cells_get(
		tls,
		rc.ptr,
		vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_UTF8,
		uintptr(unsafe.Pointer(buf)),
	)
	return result, int(buf.Flen1)
}

// BgColor returns the resolved background color for the current cell.
// Returns nil (without error) when the cell has no background color,
// in which case the caller should use the terminal default background.
func (rc *RenderStateRowCells) BgColor() (*ColorRGB, error) {
	return rc.color(vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_BG_COLOR)
}

// FgColor returns the resolved foreground color for the current cell.
// Returns nil (without error) when the cell has no explicit foreground
// color, in which case the caller should use the terminal default
// foreground. Bold color handling is not applied.
func (rc *RenderStateRowCells) FgColor() (*ColorRGB, error) {
	return rc.color(vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_FG_COLOR)
}

// color reads one of the cell's resolved colors, or nil when the cell has
// none.
func (rc *RenderStateRowCells) color(data vt.GhosttyRenderStateRowCellsData) (*ColorRGB, error) {
	tls := getTLS()
	defer putTLS(tls)
	err := rc.get(tls, data, scratchColor)
	if err != nil {
		var ge *Error
		if errors.As(err, &ge) && ge.Result == ResultInvalidValue {
			return nil, nil
		}
		return nil, err
	}
	c := colorRGBFromC(*(*vt.GhosttyColorRgb)(unsafe.Pointer(rc.scratch + scratchColor)))
	return &c, nil
}

// Selected reports whether the current cell is contained within the
// terminal selection snapshot that was current when the render state
// was updated. Rendering policy for selected cells is left to the
// caller.
func (rc *RenderStateRowCells) Selected() (bool, error) {
	return rc.getBool(vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_SELECTED)
}

// HasStyling reports whether the current cell has any explicit styling.
// This is equivalent to querying the raw Cell's HasStyling value, but
// avoids materializing the raw cell when renderers only need to know
// whether fetching the full style is necessary.
func (rc *RenderStateRowCells) HasStyling() (bool, error) {
	return rc.getBool(vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_HAS_STYLING)
}

// getBool reads a bool value of the current cell.
func (rc *RenderStateRowCells) getBool(data vt.GhosttyRenderStateRowCellsData) (bool, error) {
	tls := getTLS()
	defer putTLS(tls)
	if err := rc.get(tls, data, scratchBool); err != nil {
		return false, err
	}
	return *(*uint8)(unsafe.Pointer(rc.scratch + scratchBool)) != 0, nil
}

// unpinFunc releases memory pinned by pinUint32s.
type unpinFunc = func()

// pinUint32s is pinBytes for a []uint32: it returns the address of the
// first element, kept valid until the returned function is called. p must
// not be empty.
func pinUint32s(p []uint32) (addr uintptr, unpin unpinFunc) {
	return pinBytes(unsafe.Slice((*byte)(unsafe.Pointer(&p[0])), len(p)*4))
}
