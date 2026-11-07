package libghostty

import (
	"errors"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// CellData identifies a data field for cell queries.
// C: GhosttyCellData
type CellData int

const (
	// CellDataInvalid is an invalid data type.
	CellDataInvalid CellData = vt.GHOSTTY_CELL_DATA_INVALID

	// CellDataCodepoint is the codepoint of the cell (uint32_t).
	CellDataCodepoint CellData = vt.GHOSTTY_CELL_DATA_CODEPOINT

	// CellDataContentTag is the content tag describing what kind of
	// content is in the cell (GhosttyCellContentTag).
	CellDataContentTag CellData = vt.GHOSTTY_CELL_DATA_CONTENT_TAG

	// CellDataWide is the wide property of the cell (GhosttyCellWide).
	CellDataWide CellData = vt.GHOSTTY_CELL_DATA_WIDE

	// CellDataHasText indicates whether the cell has text to render (bool).
	CellDataHasText CellData = vt.GHOSTTY_CELL_DATA_HAS_TEXT

	// CellDataHasStyling indicates whether the cell has non-default
	// styling (bool).
	CellDataHasStyling CellData = vt.GHOSTTY_CELL_DATA_HAS_STYLING

	// CellDataStyleID is the style ID for the cell (uint16_t).
	CellDataStyleID CellData = vt.GHOSTTY_CELL_DATA_STYLE_ID

	// CellDataHasHyperlink indicates whether the cell has a hyperlink
	// (bool).
	CellDataHasHyperlink CellData = vt.GHOSTTY_CELL_DATA_HAS_HYPERLINK

	// CellDataProtected indicates whether the cell is protected (bool).
	CellDataProtected CellData = vt.GHOSTTY_CELL_DATA_PROTECTED

	// CellDataSemanticContent is the semantic content type of the cell
	// (GhosttyCellSemanticContent).
	CellDataSemanticContent CellData = vt.GHOSTTY_CELL_DATA_SEMANTIC_CONTENT

	// CellDataColorPalette is the palette index for the cell's background
	// color (GhosttyColorPaletteIndex).
	CellDataColorPalette CellData = vt.GHOSTTY_CELL_DATA_COLOR_PALETTE

	// CellDataColorRGBValue is the RGB value for the cell's background
	// color (GhosttyColorRgb).
	CellDataColorRGBValue CellData = vt.GHOSTTY_CELL_DATA_COLOR_RGB
)

// RowData identifies a data field for row queries.
// C: GhosttyRowData
type RowData int

const (
	// RowDataInvalid is an invalid data type.
	RowDataInvalid RowData = vt.GHOSTTY_ROW_DATA_INVALID

	// RowDataWrap indicates whether the row is soft-wrapped (bool).
	RowDataWrap RowData = vt.GHOSTTY_ROW_DATA_WRAP

	// RowDataWrapContinuation indicates whether the row is a continuation
	// of a soft-wrapped row (bool).
	RowDataWrapContinuation RowData = vt.GHOSTTY_ROW_DATA_WRAP_CONTINUATION

	// RowDataGrapheme indicates whether any cells in the row have grapheme
	// clusters (bool).
	RowDataGrapheme RowData = vt.GHOSTTY_ROW_DATA_GRAPHEME

	// RowDataStyled indicates whether any cells in the row have styling
	// (bool).
	RowDataStyled RowData = vt.GHOSTTY_ROW_DATA_STYLED

	// RowDataHyperlink indicates whether any cells in the row have
	// hyperlinks (bool).
	RowDataHyperlink RowData = vt.GHOSTTY_ROW_DATA_HYPERLINK

	// RowDataSemanticPrompt is the semantic prompt state of the row
	// (GhosttyRowSemanticPrompt).
	RowDataSemanticPrompt RowData = vt.GHOSTTY_ROW_DATA_SEMANTIC_PROMPT

	// RowDataKittyVirtualPlaceholder indicates whether the row contains
	// a Kitty virtual placeholder (bool).
	RowDataKittyVirtualPlaceholder RowData = vt.GHOSTTY_ROW_DATA_KITTY_VIRTUAL_PLACEHOLDER

	// RowDataDirty indicates whether the row is dirty and requires a
	// redraw (bool).
	RowDataDirty RowData = vt.GHOSTTY_ROW_DATA_DIRTY
)

// Cell is a wrapper around a packed terminal grid cell value. Portable callers
// should use its typed getter methods. Callers that need to avoid per-field
// library calls may read [Cell.PackedValue] and decode it using the
// GhosttyCell layout in [TypeJSON]; the packed bit positions are not stable
// across versions.
//
// A Cell is a copied value snapshot, not a borrowed handle, so it may be
// retained after the [GridRef] or render-state iterator that produced it
// becomes invalid.
// C: GhosttyCell
type Cell struct {
	c vt.GhosttyCell
}

// CellsView is a borrowed view of contiguous packed terminal cell values.
// Views are produced by [RenderStateRowIterator.CellsRaw] and are only valid
// until the owning render state is updated. Individual cells returned by
// [CellsView.Cell] and all values returned by [CellsView.Clone] are copied and
// may be retained.
// C: GhosttyCellsView
type CellsView struct {
	// ptr is the C address of the first cell, or 0 when the view is empty.
	ptr uintptr
	len int
}

// cellsViewFromC converts a borrowed C GhosttyCellsView to its Go wrapper.
func cellsViewFromC(view vt.GhosttyCellsView) CellsView {
	return CellsView{
		ptr: view.Fptr,
		len: int(view.Flen1),
	}
}

// cells returns the view's cells as a slice over the borrowed C memory.
func (v CellsView) cells() []vt.GhosttyCell {
	return unsafe.Slice((*vt.GhosttyCell)(unsafe.Pointer(v.ptr)), v.len)
}

// Len returns the number of cells in the view.
func (v CellsView) Len() int {
	return v.len
}

// Cell returns a copied cell at index. It panics when index is outside the
// view, matching normal Go slice indexing behavior.
func (v CellsView) Cell(index int) *Cell {
	return &Cell{c: v.cells()[index]}
}

// Clone copies every cell in the view into Go-owned memory. The returned cells
// remain valid after the render state is updated.
func (v CellsView) Clone() []Cell {
	if v.len == 0 {
		return nil
	}

	raw := v.cells()
	result := make([]Cell, len(raw))
	for i, cell := range raw {
		result[i].c = cell
	}
	return result
}

// Row is a wrapper around an opaque terminal grid row value.
// Use getter methods to extract data from it. A Row is a copied value
// snapshot, not a borrowed handle, so it may be retained after the
// [GridRef] or render-state iterator that produced it becomes invalid.
// C: GhosttyRow
type Row struct {
	c vt.GhosttyRow
}

// PackedValue returns the raw packed cell value. Decode it using the
// GhosttyCell descriptor returned by [TypeJSON]; hardcoded bit positions are
// unsupported because the packed layout is not ABI-stable.
func (c *Cell) PackedValue() uint64 {
	return uint64(c.c)
}

// CellContentTag describes what kind of content a cell holds.
// C: GhosttyCellContentTag
type CellContentTag int

const (
	// CellContentCodepoint means a single codepoint (may be zero for empty).
	CellContentCodepoint CellContentTag = vt.GHOSTTY_CELL_CONTENT_CODEPOINT

	// CellContentCodepointGrapheme means a codepoint that is part of a
	// multi-codepoint grapheme cluster.
	CellContentCodepointGrapheme CellContentTag = vt.GHOSTTY_CELL_CONTENT_CODEPOINT_GRAPHEME

	// CellContentBgColorPalette means no text; background color from palette.
	CellContentBgColorPalette CellContentTag = vt.GHOSTTY_CELL_CONTENT_BG_COLOR_PALETTE

	// CellContentBgColorRGB means no text; background color as RGB.
	CellContentBgColorRGB CellContentTag = vt.GHOSTTY_CELL_CONTENT_BG_COLOR_RGB
)

// CellWide describes the width behavior of a cell.
// C: GhosttyCellWide
type CellWide int

const (
	// CellWideNarrow means not a wide character, cell width 1.
	CellWideNarrow CellWide = vt.GHOSTTY_CELL_WIDE_NARROW

	// CellWideWide means wide character, cell width 2.
	CellWideWide CellWide = vt.GHOSTTY_CELL_WIDE_WIDE

	// CellWideSpacerTail means spacer after wide character (do not render).
	CellWideSpacerTail CellWide = vt.GHOSTTY_CELL_WIDE_SPACER_TAIL

	// CellWideSpacerHead means spacer at end of soft-wrapped line for a
	// wide character.
	CellWideSpacerHead CellWide = vt.GHOSTTY_CELL_WIDE_SPACER_HEAD
)

// CellSemanticContent is the semantic content type of a cell,
// as set by OSC 133 sequences.
// C: GhosttyCellSemanticContent
type CellSemanticContent int

const (
	// CellSemanticOutput means regular output content (e.g. command output).
	CellSemanticOutput CellSemanticContent = vt.GHOSTTY_CELL_SEMANTIC_OUTPUT

	// CellSemanticInput means content that is part of user input.
	CellSemanticInput CellSemanticContent = vt.GHOSTTY_CELL_SEMANTIC_INPUT

	// CellSemanticPrompt means content that is part of a shell prompt.
	CellSemanticPrompt CellSemanticContent = vt.GHOSTTY_CELL_SEMANTIC_PROMPT
)

// RowSemanticPrompt indicates whether any cells in a row are part of
// a shell prompt, as reported by OSC 133 sequences.
// C: GhosttyRowSemanticPrompt
type RowSemanticPrompt int

const (
	// RowSemanticNone means no prompt cells in this row.
	RowSemanticNone RowSemanticPrompt = vt.GHOSTTY_ROW_SEMANTIC_NONE

	// RowSemanticPromptPrimary means prompt cells exist and this is a
	// primary prompt line.
	RowSemanticPromptPrimary RowSemanticPrompt = vt.GHOSTTY_ROW_SEMANTIC_PROMPT

	// RowSemanticPromptContinuation means prompt cells exist and this is
	// a continuation line.
	RowSemanticPromptContinuation RowSemanticPrompt = vt.GHOSTTY_ROW_SEMANTIC_PROMPT_CONTINUATION
)

// cellGet reads one cell data field of C type T. The value is read into TLS
// stack memory, because the translated code must not write through a pointer
// to a Go variable.
func cellGet[T any](c *Cell, data vt.GhosttyCellData) (T, error) {
	tls := getTLS()
	defer putTLS(tls)
	var zero T
	n := int(unsafe.Sizeof(zero))
	p := tls.Alloc(n)
	defer tls.Free(n)
	clear(unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
	if err := resultError(vt.Xghostty_cell_get(tls, c.c, data, p)); err != nil {
		return zero, err
	}
	return *(*T)(unsafe.Pointer(p)), nil
}

// rowGet reads one row data field of C type T, like cellGet.
func rowGet[T any](r *Row, data vt.GhosttyRowData) (T, error) {
	tls := getTLS()
	defer putTLS(tls)
	var zero T
	n := int(unsafe.Sizeof(zero))
	p := tls.Alloc(n)
	defer tls.Free(n)
	clear(unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
	if err := resultError(vt.Xghostty_row_get(tls, r.c, data, p)); err != nil {
		return zero, err
	}
	return *(*T)(unsafe.Pointer(p)), nil
}

// GetMulti queries multiple cell data fields in a single library call.
// This is a low-level function. Prefer the typed getters (Codepoint,
// Wide, HasText, etc.) for normal use. GetMulti is useful when you
// need many fields at once and want to avoid per-field call overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/screen.h, GhosttyCellData enum).
// Use a Go type with the same size as the C type, such as uint32
// for uint32_t, bool for bool, and int32 for an enum.
//
// GetMulti returns an error if keys and values have different lengths
// or if a key cannot be read. In the second case, values for the keys
// before it may already have been written.
//
// Example:
//
//	var cp uint32
//	var hasText bool
//	err := cell.GetMulti(
//		[]CellData{CellDataCodepoint, CellDataHasText},
//		[]unsafe.Pointer{unsafe.Pointer(&cp), unsafe.Pointer(&hasText)},
//	)
//
// C: ghostty_cell_get_multi
func (c *Cell) GetMulti(keys []CellData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[vt.GhosttyCellData](&args, keys, values)
	if err != nil {
		return err
	}
	defer args.free()
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_cell_get_multi(
		tls,
		c.c,
		uint64(len(keys)),
		uintptr(unsafe.Pointer(cKeys)),
		uintptr(unsafe.Pointer(args.values)),
		0,
	))
}

// Codepoint returns the codepoint of the cell (0 if empty).
func (c *Cell) Codepoint() (uint32, error) {
	return cellGet[uint32](c, vt.GHOSTTY_CELL_DATA_CODEPOINT)
}

// ContentTag returns the content tag describing what kind of content
// the cell holds.
func (c *Cell) ContentTag() (CellContentTag, error) {
	v, err := cellGet[vt.GhosttyCellContentTag](c, vt.GHOSTTY_CELL_DATA_CONTENT_TAG)
	return CellContentTag(v), err
}

// Wide returns the wide property of the cell.
func (c *Cell) Wide() (CellWide, error) {
	v, err := cellGet[vt.GhosttyCellWide](c, vt.GHOSTTY_CELL_DATA_WIDE)
	return CellWide(v), err
}

// HasText reports whether the cell has text to render.
func (c *Cell) HasText() (bool, error) {
	v, err := cellGet[uint8](c, vt.GHOSTTY_CELL_DATA_HAS_TEXT)
	return v != 0, err
}

// HasStyling reports whether the cell has non-default styling.
func (c *Cell) HasStyling() (bool, error) {
	v, err := cellGet[uint8](c, vt.GHOSTTY_CELL_DATA_HAS_STYLING)
	return v != 0, err
}

// StyleID returns the style ID for the cell.
func (c *Cell) StyleID() (uint16, error) {
	return cellGet[uint16](c, vt.GHOSTTY_CELL_DATA_STYLE_ID)
}

// HasHyperlink reports whether the cell has a hyperlink.
func (c *Cell) HasHyperlink() (bool, error) {
	v, err := cellGet[uint8](c, vt.GHOSTTY_CELL_DATA_HAS_HYPERLINK)
	return v != 0, err
}

// Protected reports whether the cell is protected.
func (c *Cell) Protected() (bool, error) {
	v, err := cellGet[uint8](c, vt.GHOSTTY_CELL_DATA_PROTECTED)
	return v != 0, err
}

// Semantic returns the semantic content type of the cell.
func (c *Cell) Semantic() (CellSemanticContent, error) {
	v, err := cellGet[vt.GhosttyCellSemanticContent](c, vt.GHOSTTY_CELL_DATA_SEMANTIC_CONTENT)
	return CellSemanticContent(v), err
}

// ColorPalette returns the palette index for the cell's background
// color. Only valid when the cell's content tag is CellContentBgColorPalette.
func (c *Cell) ColorPalette() (uint8, error) {
	v, err := cellGet[vt.GhosttyColorPaletteIndex](c, vt.GHOSTTY_CELL_DATA_COLOR_PALETTE)
	return uint8(v), err
}

// ColorRGB returns the RGB color for the cell's background color.
// Only valid when the cell's content tag is CellContentBgColorRGB.
func (c *Cell) ColorRGB() (ColorRGB, error) {
	v, err := cellGet[vt.GhosttyColorRgb](c, vt.GHOSTTY_CELL_DATA_COLOR_RGB)
	if err != nil {
		return ColorRGB{}, err
	}
	return ColorRGB{R: v.Fr, G: v.Fg, B: v.Fb}, nil
}

// GetMulti queries multiple row data fields in a single library call.
// This is a low-level function. Prefer the typed getters (Wrap,
// Grapheme, Styled, Semantic, etc.) for normal use. GetMulti is
// useful when you need many fields at once and want to avoid
// per-field call overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/screen.h, GhosttyRowData enum).
// Use a Go type with the same size as the C type, such as uint32
// for uint32_t, bool for bool, and int32 for an enum.
//
// GetMulti returns an error if keys and values have different lengths
// or if a key cannot be read. In the second case, values for the keys
// before it may already have been written.
//
// Example:
//
//	var wrap, styled bool
//	err := row.GetMulti(
//		[]RowData{RowDataWrap, RowDataStyled},
//		[]unsafe.Pointer{unsafe.Pointer(&wrap), unsafe.Pointer(&styled)},
//	)
//
// C: ghostty_row_get_multi
func (r *Row) GetMulti(keys []RowData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[vt.GhosttyRowData](&args, keys, values)
	if err != nil {
		return err
	}
	defer args.free()
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_row_get_multi(
		tls,
		r.c,
		uint64(len(keys)),
		uintptr(unsafe.Pointer(cKeys)),
		uintptr(unsafe.Pointer(args.values)),
		0,
	))
}

// Wrap reports whether the row is soft-wrapped.
func (r *Row) Wrap() (bool, error) {
	v, err := rowGet[uint8](r, vt.GHOSTTY_ROW_DATA_WRAP)
	return v != 0, err
}

// WrapContinuation reports whether the row is a continuation of
// a soft-wrapped row.
func (r *Row) WrapContinuation() (bool, error) {
	v, err := rowGet[uint8](r, vt.GHOSTTY_ROW_DATA_WRAP_CONTINUATION)
	return v != 0, err
}

// Grapheme reports whether any cells in the row have grapheme clusters.
func (r *Row) Grapheme() (bool, error) {
	v, err := rowGet[uint8](r, vt.GHOSTTY_ROW_DATA_GRAPHEME)
	return v != 0, err
}

// Styled reports whether any cells in the row have styling
// (may have false positives).
func (r *Row) Styled() (bool, error) {
	v, err := rowGet[uint8](r, vt.GHOSTTY_ROW_DATA_STYLED)
	return v != 0, err
}

// Hyperlink reports whether any cells in the row have hyperlinks
// (may have false positives).
func (r *Row) Hyperlink() (bool, error) {
	v, err := rowGet[uint8](r, vt.GHOSTTY_ROW_DATA_HYPERLINK)
	return v != 0, err
}

// Semantic returns the semantic prompt state of the row.
func (r *Row) Semantic() (RowSemanticPrompt, error) {
	v, err := rowGet[vt.GhosttyRowSemanticPrompt](r, vt.GHOSTTY_ROW_DATA_SEMANTIC_PROMPT)
	return RowSemanticPrompt(v), err
}

// KittyVirtualPlaceholder reports whether the row contains a
// Kitty virtual placeholder.
func (r *Row) KittyVirtualPlaceholder() (bool, error) {
	v, err := rowGet[uint8](r, vt.GHOSTTY_ROW_DATA_KITTY_VIRTUAL_PLACEHOLDER)
	return v != 0, err
}

// Dirty reports whether the row is dirty and requires a redraw.
func (r *Row) Dirty() (bool, error) {
	v, err := rowGet[uint8](r, vt.GHOSTTY_ROW_DATA_DIRTY)
	return v != 0, err
}
