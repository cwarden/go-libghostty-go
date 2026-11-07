package libghostty

// Render-state row iterator wrapping the
// GhosttyRenderStateRowIterator C APIs.

import (
	"errors"
	"slices"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// RenderStateRowData identifies a data field for render state row queries.
// C: GhosttyRenderStateRowData
type RenderStateRowData int

const (
	// RenderStateRowDataInvalid is an invalid / sentinel value.
	RenderStateRowDataInvalid RenderStateRowData = vt.GHOSTTY_RENDER_STATE_ROW_DATA_INVALID

	// RenderStateRowDataDirty indicates whether the current row is dirty
	// (bool).
	RenderStateRowDataDirty RenderStateRowData = vt.GHOSTTY_RENDER_STATE_ROW_DATA_DIRTY

	// RenderStateRowDataRaw is the raw row value (GhosttyRow).
	RenderStateRowDataRaw RenderStateRowData = vt.GHOSTTY_RENDER_STATE_ROW_DATA_RAW

	// RenderStateRowDataCells populates a pre-allocated row cells instance
	// (GhosttyRenderStateRowCells).
	RenderStateRowDataCells RenderStateRowData = vt.GHOSTTY_RENDER_STATE_ROW_DATA_CELLS

	// RenderStateRowDataSelection is the row-local selected cell range
	// (GhosttyRenderStateRowSelection).
	RenderStateRowDataSelection RenderStateRowData = vt.GHOSTTY_RENDER_STATE_ROW_DATA_SELECTION

	// RenderStateRowDataCellsRaw is a borrowed view of all packed cell values
	// in the current row (GhosttyCellsView).
	RenderStateRowDataCellsRaw RenderStateRowData = vt.GHOSTTY_RENDER_STATE_ROW_DATA_CELLS_RAW

	// RenderStateRowDataViewportY is the signed viewport-relative row position (int32_t).
	RenderStateRowDataViewportY RenderStateRowData = vt.GHOSTTY_RENDER_STATE_ROW_DATA_VIEWPORT_Y

	// RenderStateRowDataID is the stable row identity (GhosttyRenderStateRowId).
	RenderStateRowDataID RenderStateRowData = vt.GHOSTTY_RENDER_STATE_ROW_DATA_ID
)

// RenderStateRowID identifies a row across render state updates, even when
// scrolling moves the row to a different position. It supports equality
// comparisons and can be used as a map key. The zero value is not a valid ID.
//
// A renderer can use an ID to cache work for a row. Reuse a cached result only
// if the ID appears in the new update and [RenderStateRowIterator.Dirty] is
// false. Discard entries for IDs that no longer appear. IDs are never reused
// for a different row.
//
// An ID can disappear when its row leaves the captured area, is removed from
// scrollback, or is replaced by a terminal operation.
//
// C: GhosttyRenderStateRowId
type RenderStateRowID struct {
	bits [2]uint64
}

// RenderStateRowSelection is the row-local selected cell range.
// C: GhosttyRenderStateRowSelection
type RenderStateRowSelection struct {
	// StartX is the start column of the row-local selection range,
	// inclusive.
	StartX uint16

	// EndX is the end column of the row-local selection range, inclusive.
	EndX uint16
}

// RenderStateRowIterator iterates over rows in a render state.
// Create one with NewRenderStateRowIterator, populate it via
// [RenderState.RowIterator], then advance with [RenderStateRowIterator.Next]
// and read data with getter methods.
//
// The iterator's current position is only valid as long as the
// underlying render state is not updated. Do not use the iterator
// while [RenderState.Update] may run on the same render state.
// Extracted [Row] values are copied snapshots and may be retained after
// the iterator itself becomes invalid.
//
// C: GhosttyRenderStateRowIterator
type RenderStateRowIterator struct {
	ptr vt.GhosttyRenderStateRowIterator
}

// NewRenderStateRowIterator creates a new row iterator instance.
// The iterator is empty until populated via RenderState.RowIterator.
func NewRenderStateRowIterator() (*RenderStateRowIterator, error) {
	tls := getTLS()
	defer putTLS(tls)
	out := tls.Alloc(handleSize)
	defer tls.Free(handleSize)
	if err := resultError(vt.Xghostty_render_state_row_iterator_new(tls, 0, out)); err != nil {
		return nil, err
	}
	return &RenderStateRowIterator{ptr: *(*uintptr)(unsafe.Pointer(out))}, nil
}

// Close frees the underlying row iterator handle. After this call,
// the iterator must not be used, except that calling Close again is a
// safe no-op.
func (ri *RenderStateRowIterator) Close() {
	if ri == nil || ri.ptr == 0 {
		return
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_render_state_row_iterator_free(tls, ri.ptr)
	ri.ptr = 0
}

// Next advances to the next row and reports whether row data is available.
// It returns false when no rows remain. Rows are visited from top to bottom,
// including any overscan rows. Use [RenderStateRowIterator.ViewportY] to
// position the current row relative to the viewport.
func (ri *RenderStateRowIterator) Next() bool {
	tls := getTLS()
	defer putTLS(tls)
	return vt.Xghostty_render_state_row_iterator_next(tls, ri.ptr) != 0
}

// NextDirty advances to the next row that requires a redraw. It returns the
// row's index and true, or zero and false when no such rows remain.
//
// The index counts from the first captured row, including overscan. Without
// overscan, it equals the row's position in the viewport. With overscan, use
// [RenderStateRowIterator.ViewportY] to position the row instead.
//
// When the entire render state is dirty, NextDirty visits every remaining row.
// Otherwise it skips clean rows. It does not clear any dirty flags.
//
// C: ghostty_render_state_row_iterator_next_dirty
func (ri *RenderStateRowIterator) NextDirty() (index uint16, ok bool) {
	tls := getTLS()
	defer putTLS(tls)
	y := tls.Alloc(scratchSize)
	defer tls.Free(scratchSize)
	if vt.Xghostty_render_state_row_iterator_next_dirty(tls, ri.ptr, y) == 0 {
		return 0, false
	}
	return *(*uint16)(unsafe.Pointer(y)), true
}

// GetMulti queries multiple render-state row data fields in a single
// call. This is a low-level function. Prefer the typed getters
// (Dirty, Raw, Cells) for normal use. GetMulti is useful when you
// need many fields at once and want to avoid per-field call overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/render.h, GhosttyRenderStateRowData enum).
// Use a Go type with the same size as the C type, such as uint32
// for uint32_t, bool for bool, and int32 for an enum.
//
// GetMulti returns an error if keys and values have different lengths
// or if a key cannot be read. In the second case, values for the keys
// before it may already have been written.
//
// Example:
//
//	var dirty bool
//	var viewportY int32
//	err := ri.GetMulti(
//		[]RenderStateRowData{RenderStateRowDataDirty, RenderStateRowDataViewportY},
//		[]unsafe.Pointer{unsafe.Pointer(&dirty), unsafe.Pointer(&viewportY)},
//	)
//
// C: ghostty_render_state_row_get_multi
func (ri *RenderStateRowIterator) GetMulti(keys []RenderStateRowData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[vt.GhosttyRenderStateRowData](&args, keys, values)
	if err != nil {
		return err
	}
	defer args.free()
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_render_state_row_get_multi(
		tls,
		ri.ptr,
		uint64(len(keys)),
		uintptr(unsafe.Pointer(cKeys)),
		uintptr(unsafe.Pointer(args.values)),
		0,
	))
}

// get reads data into TLS stack memory of size bytes and passes the block
// to read, which copies the value out. The block is not zeroed.
func (ri *RenderStateRowIterator) get(data vt.GhosttyRenderStateRowData, size int, read func(out uintptr)) error {
	tls := getTLS()
	defer putTLS(tls)
	out := tls.Alloc(size)
	defer tls.Free(size)
	if err := resultError(vt.Xghostty_render_state_row_get(tls, ri.ptr, data, out)); err != nil {
		return err
	}
	read(out)
	return nil
}

// Dirty reports whether the current row is dirty and requires a
// redraw.
func (ri *RenderStateRowIterator) Dirty() (bool, error) {
	var v bool
	err := ri.get(vt.GHOSTTY_RENDER_STATE_ROW_DATA_DIRTY, scratchSize, func(out uintptr) { v = *(*uint8)(unsafe.Pointer(out)) != 0 })
	return v, err
}

// ViewportY returns the current row's position in rows relative to the top of
// the viewport. The first visible row is zero. Overscan rows above the viewport
// have negative positions. Those below it start at the height returned by
// [RenderState.Rows].
func (ri *RenderStateRowIterator) ViewportY() (int32, error) {
	var v int32
	err := ri.get(vt.GHOSTTY_RENDER_STATE_ROW_DATA_VIEWPORT_Y, scratchSize, func(out uintptr) { v = *(*int32)(unsafe.Pointer(out)) })
	return v, err
}

// ID returns the current row's identity. The value can be kept after the
// iterator advances or the render state is updated. See [RenderStateRowID]
// for how to use IDs when caching rendered rows.
func (ri *RenderStateRowIterator) ID() (RenderStateRowID, error) {
	var v RenderStateRowID
	err := ri.get(vt.GHOSTTY_RENDER_STATE_ROW_DATA_ID, scratchSize, func(out uintptr) {
		v.bits = (*vt.GhosttyRenderStateRowId)(unsafe.Pointer(out)).Fbits
	})
	return v, err
}

// SetDirty sets the dirty state for the current row.
func (ri *RenderStateRowIterator) SetDirty(dirty bool) error {
	tls := getTLS()
	defer putTLS(tls)
	v := tls.Alloc(scratchSize)
	defer tls.Free(scratchSize)
	*(*uint8)(unsafe.Pointer(v)) = cBool(dirty)
	return resultError(vt.Xghostty_render_state_row_set(tls, ri.ptr, vt.GHOSTTY_RENDER_STATE_ROW_OPT_DIRTY, v))
}

// Raw returns the raw Row value for the current iterator position.
// The returned Row can be used with the same getter methods as rows
// obtained from GridRef.
func (ri *RenderStateRowIterator) Raw() (*Row, error) {
	var v vt.GhosttyRow
	err := ri.get(vt.GHOSTTY_RENDER_STATE_ROW_DATA_RAW, scratchSize, func(out uintptr) { v = *(*vt.GhosttyRow)(unsafe.Pointer(out)) })
	if err != nil {
		return nil, err
	}
	return &Row{c: v}, nil
}

// Selection returns the row-local selected cell range for the current row. It
// returns nil when the current row does not intersect the terminal's active
// selection at the time the render state was updated.
func (ri *RenderStateRowIterator) Selection() (*RenderStateRowSelection, error) {
	tls := getTLS()
	defer putTLS(tls)
	size := int(unsafe.Sizeof(vt.GhosttyRenderStateRowSelection{}))
	out := tls.Alloc(size)
	defer tls.Free(size)
	// GHOSTTY_INIT_SIZED: a zeroed struct whose size field is set.
	v := (*vt.GhosttyRenderStateRowSelection)(unsafe.Pointer(out))
	*v = vt.GhosttyRenderStateRowSelection{Fsize: uint64(size)}
	err := resultError(vt.Xghostty_render_state_row_get(
		tls,
		ri.ptr,
		vt.GHOSTTY_RENDER_STATE_ROW_DATA_SELECTION,
		out,
	))
	if err != nil {
		var ge *Error
		if errors.As(err, &ge) && ge.Result == ResultNoValue {
			return nil, nil
		}
		return nil, err
	}

	return &RenderStateRowSelection{
		StartX: v.Fstart_x,
		EndX:   v.Fend_x,
	}, nil
}

// Cells populates a pre-allocated row cells instance with cell data
// for the current row. The cells instance can then be advanced with
// Next or positioned with Select.
//
// The cells instance can be reused across rows. Cell data is only
// valid until the next call to RenderState.Update.
func (ri *RenderStateRowIterator) Cells(rc *RenderStateRowCells) error {
	tls := getTLS()
	defer putTLS(tls)
	return ri.cells(tls, rc)
}

// cells loads the current row into rc using the caller's TLS. libghostty
// updates the object the handle points to and writes the handle back.
func (ri *RenderStateRowIterator) cells(tls *libc.TLS, rc *RenderStateRowCells) error {
	out := tls.Alloc(handleSize)
	defer tls.Free(handleSize)
	*(*uintptr)(unsafe.Pointer(out)) = rc.ptr
	if err := resultError(vt.Xghostty_render_state_row_get(
		tls,
		ri.ptr,
		vt.GHOSTTY_RENDER_STATE_ROW_DATA_CELLS,
		out,
	)); err != nil {
		return err
	}
	rc.ptr = *(*uintptr)(unsafe.Pointer(out))
	return nil
}

// AppendText appends the text of the current row to dst as UTF-8 and returns
// the extended slice. Any existing content in dst is kept.
//
// The text follows what the row shows on screen:
//
//   - Empty cells between characters become spaces.
//   - Empty cells at the end of the row are left out. Spaces the program
//     actually printed are text and are kept, even at the end of the row.
//   - A wide character, such as a CJK character or most emoji, appears once
//     even though it covers two cells.
//   - A character made of several code points, such as an emoji with a skin
//     tone modifier, is kept whole.
//
// No newline is added. When the terminal wraps a long line onto several
// rows, each row is returned separately. To join them, check [Row.Wrap] on
// the value returned by [RenderStateRowIterator.Raw].
//
// rc is working storage. AppendText loads the current row into it and leaves
// it at an unspecified cell. As with [RenderStateRowIterator.Cells], one rc
// can be reused for every row.
//
// AppendText is much faster than calling [RenderStateRowCells.AppendGraphemes]
// on each cell, because it reads each cell's text with one call into
// libghostty and no per-cell Go allocation.
// To avoid allocations, reuse one buffer across rows by passing buf[:0].
func (ri *RenderStateRowIterator) AppendText(dst []byte, rc *RenderStateRowCells) ([]byte, error) {
	tls := getTLS()
	defer putTLS(tls)
	var x uint16
	for {
		oldLen := len(dst)
		result := ri.appendText(tls, rc, x, dst[oldLen:cap(dst)])
		dst = dst[:oldLen+result.len]
		switch result.result {
		case vt.GHOSTTY_SUCCESS:
			return dst, nil

		case vt.GHOSTTY_OUT_OF_SPACE:
			// Grow dst so the next cell fits and continue from where the
			// loop stopped.
			x = result.nextX
			dst = slices.Grow(dst, result.needed)

		default:
			return dst, resultError(result.result)
		}
	}
}

// rowTextResult describes the outcome of appendText.
type rowTextResult struct {
	// result is GHOSTTY_SUCCESS when the rest of the row was written.
	// GHOSTTY_OUT_OF_SPACE means dst filled up first, and the caller should
	// grow dst and call again starting at nextX. Any other value is an
	// error from libghostty.
	result vt.GhosttyResult

	// len is the number of bytes written to dst.
	len int

	// nextX is the column to start from on the next call, set only with
	// GHOSTTY_OUT_OF_SPACE. Every cell before it has been handled.
	nextX uint16

	// needed is the minimum number of extra bytes dst must have for the next
	// call to make progress, set only with GHOSTTY_OUT_OF_SPACE.
	needed int
}

// appendText writes the text of the current row to dst as UTF-8, starting
// at column startX. dst is the spare capacity of the caller's buffer, and
// is pinned for the duration of the call. Each cell's text comes from the
// GRAPHEMES_UTF8 getter, so encoding matches the per-cell API exactly.
//
// The text rules are documented on AppendText.
//
// Empty cells are held back as a pending run of spaces and written only when
// a cell with text follows them. This is what drops empty cells at the end
// of the row. If dst fills up while spaces are pending, nextX points at the
// first pending space rather than at the cell with text, so the next call
// writes those spaces again along with the text that follows them.
//
// The cells handle is reset to the current row. Its position afterwards is
// unspecified.
func (ri *RenderStateRowIterator) appendText(tls *libc.TLS, rc *RenderStateRowCells, startX uint16, dst []byte) rowTextResult {
	out := rowTextResult{result: vt.GHOSTTY_SUCCESS, nextX: startX}

	// Load the current row into the cells handle.
	if err := ri.cells(tls, rc); err != nil {
		out.result = vt.GhosttyResult(err.(*Error).Result)
		return out
	}

	vt.Xghostty_render_state_row_cells_select(tls, rc.ptr, startX)

	var dstAddr uintptr
	if len(dst) > 0 {
		var unpin func()
		dstAddr, unpin = pinBytes(dst)
		defer unpin()
	}
	room0 := len(dst)
	buf := (*vt.GhosttyBuffer)(unsafe.Pointer(rc.scratch + scratchBuffer))

	// blanks counts the pending empty cells, and blankX is the column of
	// the first one.
	blanks := 0
	blankX := startX
	x := startX
	for {
		// Leave room for the pending spaces and ask for the cell's text
		// right after them. If even the spaces don't fit, pass a NULL
		// buffer. An empty cell still succeeds with len=0, and a cell with
		// text fails with GHOSTTY_OUT_OF_SPACE and reports the size it needs.
		room := room0 - out.len
		*buf = vt.GhosttyBuffer{}
		if room > blanks {
			buf.Fptr = dstAddr + uintptr(out.len+blanks)
			buf.Fcap1 = uint64(room - blanks)
		}
		result := vt.Xghostty_render_state_row_cells_get(
			tls,
			rc.ptr,
			vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_UTF8,
			uintptr(unsafe.Pointer(buf)),
		)
		if result == vt.GHOSTTY_OUT_OF_SPACE {
			out.result = result
			out.nextX = x
			if blanks > 0 {
				out.nextX = blankX
			}
			out.needed = blanks + int(buf.Flen1)
			return out
		}
		if result != vt.GHOSTTY_SUCCESS {
			out.result = result
			return out
		}

		switch {
		case buf.Flen1 == 0:
			// An empty cell adds to the pending run of spaces, unless it is
			// a spacer cell. A spacer cell is either the second half of a
			// wide character or the leftover cell where a wide character
			// didn't fit at the end of a row, so it produces no text.
			raw := rc.scratch + scratchRaw
			result = vt.Xghostty_render_state_row_cells_get(
				tls,
				rc.ptr,
				vt.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_RAW,
				raw,
			)
			if result != vt.GHOSTTY_SUCCESS {
				out.result = result
				return out
			}

			// This is part of the optional grid introspection feature, so
			// it can return GHOSTTY_UNSUPPORTED without writing to wide.
			wide := rc.scratch + scratchWide
			result = vt.Xghostty_cell_get(tls, *(*vt.GhosttyCell)(unsafe.Pointer(raw)), vt.GHOSTTY_CELL_DATA_WIDE, wide)
			if result != vt.GHOSTTY_SUCCESS {
				out.result = result
				return out
			}

			switch *(*vt.GhosttyCellWide)(unsafe.Pointer(wide)) {
			case vt.GHOSTTY_CELL_WIDE_SPACER_TAIL, vt.GHOSTTY_CELL_WIDE_SPACER_HEAD:
			default:
				if blanks == 0 {
					blankX = x
				}
				blanks++
			}

		default:
			// A cell with text was written after the reserved room, so fill
			// that room with the pending spaces.
			for i := 0; i < blanks; i++ {
				dst[out.len+i] = ' '
			}
			out.len += blanks + int(buf.Flen1)
			blanks = 0
		}

		x++
		if vt.Xghostty_render_state_row_cells_next(tls, rc.ptr) == 0 {
			return out
		}
	}
}

// CellsRaw returns a borrowed, contiguous view of the packed cell values in
// the current row. It avoids one call into libghostty per cell and is
// intended for render paths that decode [Cell.PackedValue] using the layout
// from [TypeJSON].
//
// The view is invalidated by the next [RenderState.Update]. Call
// [CellsView.Clone] when the values must outlive the current render state
// snapshot.
func (ri *RenderStateRowIterator) CellsRaw() (CellsView, error) {
	var view vt.GhosttyCellsView
	err := ri.get(vt.GHOSTTY_RENDER_STATE_ROW_DATA_CELLS_RAW, scratchSize, func(out uintptr) {
		view = *(*vt.GhosttyCellsView)(unsafe.Pointer(out))
	})
	if err != nil {
		return CellsView{}, err
	}
	return cellsViewFromC(view), nil
}
