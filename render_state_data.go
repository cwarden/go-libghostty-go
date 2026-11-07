package libghostty

// Render state data getters and setters wrapping
// ghostty_render_state_get() and ghostty_render_state_set().
// Functions are ordered alphabetically.
//
// Every value is read into TLS stack memory, which the translated
// libghostty may address, and copied out from there. A Go local variable
// cannot be used as the out-parameter: the translated code runs on the
// goroutine's stack, which the runtime may move during the call.

import (
	"errors"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// RenderStateData identifies a data field for render state queries.
// C: GhosttyRenderStateData
type RenderStateData int

const (
	// RenderStateDataInvalid is an invalid / sentinel value.
	RenderStateDataInvalid RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_INVALID

	// RenderStateDataCols is the viewport width in cells (uint16_t).
	RenderStateDataCols RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_COLS

	// RenderStateDataRows is the viewport height in cells, excluding overscan (uint16_t).
	RenderStateDataRows RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_ROWS

	// RenderStateDataOverscan is the number of extra rows captured by the last update
	// (GhosttyRenderStateOverscan).
	RenderStateDataOverscan RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_OVERSCAN

	// RenderStateDataOverscanRequest is the request used for the next update
	// (GhosttyRenderStateOverscan).
	RenderStateDataOverscanRequest RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_OVERSCAN_REQUEST

	// RenderStateDataDirty is the current dirty state
	// (GhosttyRenderStateDirty).
	RenderStateDataDirty RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_DIRTY

	// RenderStateDataRowIterator populates a pre-allocated row iterator
	// (GhosttyRenderStateRowIterator).
	RenderStateDataRowIterator RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_ROW_ITERATOR

	// RenderStateDataImageIterator fills an existing image iterator with
	// the image placements from the last update
	// (GhosttyRenderStateImageIterator).
	RenderStateDataImageIterator RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_IMAGE_ITERATOR

	// RenderStateDataImagePlacementsGeneration changes whenever an update
	// changes the images listed by [RenderStateDataImageIterator]
	// (uint64_t). See [RenderState.ImagePlacementsGeneration].
	RenderStateDataImagePlacementsGeneration RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_IMAGE_PLACEMENTS_GENERATION

	// RenderStateDataColorBackground is the default/current background
	// color (GhosttyColorRgb).
	RenderStateDataColorBackground RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_COLOR_BACKGROUND

	// RenderStateDataColorForeground is the default/current foreground
	// color (GhosttyColorRgb).
	RenderStateDataColorForeground RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_COLOR_FOREGROUND

	// RenderStateDataColorCursor is the cursor color when explicitly set
	// (GhosttyColorRgb).
	RenderStateDataColorCursor RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_COLOR_CURSOR

	// RenderStateDataColorCursorHasValue indicates whether an explicit
	// cursor color is set (bool).
	RenderStateDataColorCursorHasValue RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_COLOR_CURSOR_HAS_VALUE

	// RenderStateDataColorPalette is the active 256-color palette
	// (GhosttyColorRgb[256]).
	RenderStateDataColorPalette RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_COLOR_PALETTE

	// RenderStateDataCursorVisualStyle is the visual style of the cursor
	// (GhosttyCursorStyle).
	RenderStateDataCursorVisualStyle RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_VISUAL_STYLE

	// RenderStateDataCursorVisible indicates whether the cursor is visible
	// based on terminal modes (bool).
	RenderStateDataCursorVisible RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_VISIBLE

	// RenderStateDataCursorBlinking indicates whether the cursor should
	// blink based on terminal modes (bool).
	RenderStateDataCursorBlinking RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_BLINKING

	// RenderStateDataCursorPasswordInput indicates whether the cursor is
	// at a password input field (bool).
	RenderStateDataCursorPasswordInput RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_PASSWORD_INPUT

	// RenderStateDataCursorViewportHasValue indicates whether the cursor
	// is visible within the viewport (bool).
	RenderStateDataCursorViewportHasValue RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_HAS_VALUE

	// RenderStateDataCursorViewportX is the cursor viewport x position
	// in cells (uint16_t).
	RenderStateDataCursorViewportX RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_X

	// RenderStateDataCursorViewportY is the cursor viewport y position
	// in cells (uint16_t).
	RenderStateDataCursorViewportY RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_Y

	// RenderStateDataCursorViewportWideTail indicates whether the cursor
	// is on the tail of a wide character (bool).
	RenderStateDataCursorViewportWideTail RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_WIDE_TAIL

	// RenderStateDataCursor is all cursor state in one sized struct
	// (GhosttyRenderStateCursor).
	RenderStateDataCursor RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_CURSOR

	// RenderStateDataColors is all render-state colors in one sized struct
	// (GhosttyRenderStateColors).
	RenderStateDataColors RenderStateData = vt.GHOSTTY_RENDER_STATE_DATA_COLORS
)

// scratchSize is the size of the TLS stack block that the scalar getters
// below read their values into. It holds any scalar or small struct value.
const scratchSize = 16

// get reads data into TLS stack memory of size bytes and passes the block
// to read, which copies the value out. The block is not zeroed.
func (rs *RenderState) get(data vt.GhosttyRenderStateData, size int, read func(tls *libc.TLS, out uintptr)) error {
	tls := getTLS()
	defer putTLS(tls)
	out := tls.Alloc(size)
	defer tls.Free(size)
	if err := resultError(vt.Xghostty_render_state_get(tls, rs.ptr, data, out)); err != nil {
		return err
	}
	read(tls, out)
	return nil
}

// getUint16 reads a uint16_t value.
func (rs *RenderState) getUint16(data vt.GhosttyRenderStateData) (uint16, error) {
	var v uint16
	err := rs.get(data, scratchSize, func(_ *libc.TLS, out uintptr) { v = *(*uint16)(unsafe.Pointer(out)) })
	return v, err
}

// getBool reads a bool value.
func (rs *RenderState) getBool(data vt.GhosttyRenderStateData) (bool, error) {
	var v bool
	err := rs.get(data, scratchSize, func(_ *libc.TLS, out uintptr) { v = *(*uint8)(unsafe.Pointer(out)) != 0 })
	return v, err
}

// getInt32 reads an int32_t or enum value.
func (rs *RenderState) getInt32(data vt.GhosttyRenderStateData) (int32, error) {
	var v int32
	err := rs.get(data, scratchSize, func(_ *libc.TLS, out uintptr) { v = *(*int32)(unsafe.Pointer(out)) })
	return v, err
}

// getColor reads a GhosttyColorRgb value.
func (rs *RenderState) getColor(data vt.GhosttyRenderStateData) (ColorRGB, error) {
	var v ColorRGB
	err := rs.get(data, scratchSize, func(_ *libc.TLS, out uintptr) { v = colorRGBFromC(*(*vt.GhosttyColorRgb)(unsafe.Pointer(out))) })
	return v, err
}

// Cols returns the viewport width in cells.
func (rs *RenderState) Cols() (uint16, error) {
	return rs.getUint16(vt.GHOSTTY_RENDER_STATE_DATA_COLS)
}

// ColorBackground returns the default/current background color.
func (rs *RenderState) ColorBackground() (ColorRGB, error) {
	return rs.getColor(vt.GHOSTTY_RENDER_STATE_DATA_COLOR_BACKGROUND)
}

// ColorCursor returns the cursor color when explicitly set by terminal
// state. Returns nil (without error) when no explicit cursor color is set.
func (rs *RenderState) ColorCursor() (*ColorRGB, error) {
	// Check whether a cursor color is set first.
	has, err := rs.getBool(vt.GHOSTTY_RENDER_STATE_DATA_COLOR_CURSOR_HAS_VALUE)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	c, err := rs.getColor(vt.GHOSTTY_RENDER_STATE_DATA_COLOR_CURSOR)
	if err != nil {
		var ge *Error
		if errors.As(err, &ge) && ge.Result == ResultInvalidValue {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// ColorForeground returns the default/current foreground color.
func (rs *RenderState) ColorForeground() (ColorRGB, error) {
	return rs.getColor(vt.GHOSTTY_RENDER_STATE_DATA_COLOR_FOREGROUND)
}

// ColorPalette returns the active 256-color palette.
func (rs *RenderState) ColorPalette() (*Palette, error) {
	var p Palette
	size := int(unsafe.Sizeof([PaletteSize]vt.GhosttyColorRgb{}))
	err := rs.get(vt.GHOSTTY_RENDER_STATE_DATA_COLOR_PALETTE, size, func(_ *libc.TLS, out uintptr) {
		for i, c := range (*[PaletteSize]vt.GhosttyColorRgb)(unsafe.Pointer(out)) {
			p[i] = colorRGBFromC(c)
		}
	})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Colors returns all color information from the render state in a
// single call using the sized-struct API.
func (rs *RenderState) Colors() (*RenderStateColors, error) {
	tls := getTLS()
	defer putTLS(tls)
	size := int(unsafe.Sizeof(vt.GhosttyRenderStateColors{}))
	out := tls.Alloc(size)
	defer tls.Free(size)
	// GHOSTTY_INIT_SIZED: a zeroed struct whose size field is set.
	cc := (*vt.GhosttyRenderStateColors)(unsafe.Pointer(out))
	*cc = vt.GhosttyRenderStateColors{Fsize: uint64(size)}
	if err := resultError(vt.Xghostty_render_state_get(
		tls,
		rs.ptr,
		vt.GHOSTTY_RENDER_STATE_DATA_COLORS,
		out,
	)); err != nil {
		return nil, err
	}

	result := &RenderStateColors{
		Background:     colorRGBFromC(cc.Fbackground),
		Foreground:     colorRGBFromC(cc.Fforeground),
		Cursor:         colorRGBFromC(cc.Fcursor),
		CursorHasValue: cc.Fcursor_has_value != 0,
	}

	for i, c := range cc.Fpalette {
		result.Palette[i] = colorRGBFromC(c)
	}
	return result, nil
}

// Cursor returns all cursor information from the render state in one call.
// This is the preferred cursor read for render loops because it avoids the
// separate call into libghostty required by each scalar cursor getter.
func (rs *RenderState) Cursor() (*RenderStateCursor, error) {
	tls := getTLS()
	defer putTLS(tls)
	size := int(unsafe.Sizeof(vt.GhosttyRenderStateCursor{}))
	out := tls.Alloc(size)
	defer tls.Free(size)
	// GHOSTTY_INIT_SIZED: a zeroed struct whose size field is set.
	cc := (*vt.GhosttyRenderStateCursor)(unsafe.Pointer(out))
	*cc = vt.GhosttyRenderStateCursor{Fsize: uint64(size)}
	if err := resultError(vt.Xghostty_render_state_get(
		tls,
		rs.ptr,
		vt.GHOSTTY_RENDER_STATE_DATA_CURSOR,
		out,
	)); err != nil {
		return nil, err
	}

	return &RenderStateCursor{
		ViewportHasValue: cc.Fviewport_has_value != 0,
		ViewportX:        cc.Fviewport_x,
		ViewportY:        cc.Fviewport_y,
		WideTail:         cc.Fwide_tail != 0,
		Visible:          cc.Fvisible != 0,
		Blinking:         cc.Fblinking != 0,
		PasswordInput:    cc.Fpassword_input != 0,
		VisualStyle:      CursorStyle(cc.Fvisual_style),
	}, nil
}

// GetMulti queries multiple render state data fields in a single
// call. This is a low-level function. Prefer the typed getters (Cols,
// Rows, CursorVisible, etc.) for normal use. GetMulti is useful when
// you need many fields at once and want to avoid per-field call overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/render.h, GhosttyRenderStateData enum).
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
//	err := rs.GetMulti(
//		[]RenderStateData{RenderStateDataCols, RenderStateDataRows},
//		[]unsafe.Pointer{unsafe.Pointer(&cols), unsafe.Pointer(&rows)},
//	)
//
// C: ghostty_render_state_get_multi
func (rs *RenderState) GetMulti(keys []RenderStateData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[vt.GhosttyRenderStateData](&args, keys, values)
	if err != nil {
		return err
	}
	defer args.free()
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_render_state_get_multi(
		tls,
		rs.ptr,
		uint64(len(keys)),
		uintptr(unsafe.Pointer(cKeys)),
		uintptr(unsafe.Pointer(args.values)),
		0,
	))
}

// CursorBlinking reports whether the cursor should blink based on
// terminal modes.
func (rs *RenderState) CursorBlinking() (bool, error) {
	return rs.getBool(vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_BLINKING)
}

// CursorPasswordInput reports whether the cursor is at a password
// input field.
func (rs *RenderState) CursorPasswordInput() (bool, error) {
	return rs.getBool(vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_PASSWORD_INPUT)
}

// CursorVisible reports whether the cursor is visible based on
// terminal modes.
func (rs *RenderState) CursorVisible() (bool, error) {
	return rs.getBool(vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_VISIBLE)
}

// CursorVisualStyle returns the visual style of the cursor.
func (rs *RenderState) CursorVisualStyle() (CursorStyle, error) {
	v, err := rs.getInt32(vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_VISUAL_STYLE)
	return CursorStyle(v), err
}

// CursorViewportHasValue reports whether the cursor is visible within
// the viewport. If false, the cursor viewport position values are
// undefined.
func (rs *RenderState) CursorViewportHasValue() (bool, error) {
	return rs.getBool(vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_HAS_VALUE)
}

// CursorViewportWideTail reports whether the cursor is on the tail
// of a wide character. Only valid when CursorViewportHasValue
// returns true.
func (rs *RenderState) CursorViewportWideTail() (bool, error) {
	return rs.getBool(vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_WIDE_TAIL)
}

// CursorViewportX returns the cursor viewport x position in cells.
// Only valid when CursorViewportHasValue returns true.
func (rs *RenderState) CursorViewportX() (uint16, error) {
	return rs.getUint16(vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_X)
}

// CursorViewportY returns the cursor viewport y position in cells.
// Only valid when CursorViewportHasValue returns true.
func (rs *RenderState) CursorViewportY() (uint16, error) {
	return rs.getUint16(vt.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_Y)
}

// Dirty returns the current dirty state.
func (rs *RenderState) Dirty() (RenderStateDirty, error) {
	v, err := rs.getInt32(vt.GHOSTTY_RENDER_STATE_DATA_DIRTY)
	return RenderStateDirty(v), err
}

// ImageIterator fills iter with the images to draw from the last update.
// Call [RenderStateImageIterator.Next] to move to the first image before
// reading any values.
//
// The values in iter stay valid until the next call to [RenderState.Update].
// After each update, call ImageIterator again before reading iter. The
// same iterator can be reused each time. To skip reading the list on
// frames where no image changed, use [RenderState.ImagePlacementsGeneration].
func (rs *RenderState) ImageIterator(iter *RenderStateImageIterator) error {
	return rs.getHandle(vt.GHOSTTY_RENDER_STATE_DATA_IMAGE_ITERATOR, &iter.ptr)
}

// getHandle passes an existing handle to libghostty, which fills the
// object it points to, and stores the handle it writes back.
func (rs *RenderState) getHandle(data vt.GhosttyRenderStateData, handle *uintptr) error {
	tls := getTLS()
	defer putTLS(tls)
	out := tls.Alloc(handleSize)
	defer tls.Free(handleSize)
	*(*uintptr)(unsafe.Pointer(out)) = *handle
	if err := resultError(vt.Xghostty_render_state_get(tls, rs.ptr, data, out)); err != nil {
		return err
	}
	*handle = *(*uintptr)(unsafe.Pointer(out))
	return nil
}

// ImagePlacementsGeneration returns a value that changes whenever an
// update changes the images listed by [RenderState.ImageIterator].
//
// Most frames don't change any image. A renderer can save this value each
// time it reads the image list, then skip reading the list on later
// frames while the value stays the same. See the example.
//
// The value changes when an update adds, removes, or moves an image,
// changes which part of an image is shown, or changes an image's pixels.
// It stays the same when only text changes, unless the rows that changed
// held Kitty placeholder cells before or after the change.
//
// A new value means the list may have changed, not that it did. For
// example, the value also changes when a program rewrites a row that holds
// placeholder cells, or changes any image while images shown with
// placeholder cells are on screen, even if every image stays where it was.
//
// Compare values only for equality, and only with values from the same
// render state. The value wraps around, so a larger value is not
// necessarily a newer one. It is always 0 when Kitty graphics are
// disabled.
//
// The value describes the list as a whole. To know when the pixels of one
// image changed, use [RenderStateImageInfo.ImageGeneration].
//
// ImagePlacementsGeneration doesn't allocate, so it is cheap to call
// every frame.
//
// C: GHOSTTY_RENDER_STATE_DATA_IMAGE_PLACEMENTS_GENERATION
func (rs *RenderState) ImagePlacementsGeneration() (uint64, error) {
	var v uint64
	err := rs.get(vt.GHOSTTY_RENDER_STATE_DATA_IMAGE_PLACEMENTS_GENERATION, scratchSize, func(_ *libc.TLS, out uintptr) {
		v = *(*uint64)(unsafe.Pointer(out))
	})
	return v, err
}

// RowIterator initializes ri to iterate over the rows from the last update.
// Advance it with [RenderStateRowIterator.Next] before reading row data.
// Rows are ordered from top to bottom and include any overscan rows. Use
// [RenderStateRowIterator.ViewportY] to position each row in the viewport.
//
// The iterator can be reused by calling RowIterator again. Its data is valid
// only until the next call to [RenderState.Update].
func (rs *RenderState) RowIterator(ri *RenderStateRowIterator) error {
	return rs.getHandle(vt.GHOSTTY_RENDER_STATE_DATA_ROW_ITERATOR, &ri.ptr)
}

// Rows returns the viewport height in cells, excluding overscan.
func (rs *RenderState) Rows() (uint16, error) {
	return rs.getUint16(vt.GHOSTTY_RENDER_STATE_DATA_ROWS)
}

// SetDirty sets the dirty state.
func (rs *RenderState) SetDirty(dirty RenderStateDirty) error {
	tls := getTLS()
	defer putTLS(tls)
	v := tls.Alloc(scratchSize)
	defer tls.Free(scratchSize)
	*(*vt.GhosttyRenderStateDirty)(unsafe.Pointer(v)) = vt.GhosttyRenderStateDirty(dirty)
	return resultError(vt.Xghostty_render_state_set(tls, rs.ptr, vt.GHOSTTY_RENDER_STATE_OPT_DIRTY, v))
}

// SetOverscan sets the number of extra rows to capture above and below the
// viewport on subsequent updates. The request remains in effect until changed.
// A zero [RenderStateOverscan] captures only the visible viewport.
//
// Changing the request leaves the current row data valid. The next update
// applies the request and marks the entire render state as dirty. Use
// [RenderState.Overscan] after that update to find how many rows were available.
//
// C: GHOSTTY_RENDER_STATE_OPT_OVERSCAN
func (rs *RenderState) SetOverscan(request RenderStateOverscan) error {
	tls := getTLS()
	defer putTLS(tls)
	v := tls.Alloc(scratchSize)
	defer tls.Free(scratchSize)
	*(*vt.GhosttyRenderStateOverscan)(unsafe.Pointer(v)) = vt.GhosttyRenderStateOverscan{Fabove: request.Above, Fbelow: request.Below}
	return resultError(vt.Xghostty_render_state_set(tls, rs.ptr, vt.GHOSTTY_RENDER_STATE_OPT_OVERSCAN, v))
}

// Overscan returns the number of extra rows captured by the last update.
// Each count is at most the corresponding count set by [RenderState.SetOverscan].
// Fewer rows are captured when the viewport is near the start or end of the
// terminal's contents. Below is zero when the viewport is scrolled to the bottom.
//
// [RenderState.RowIterator] includes these extra rows. Cursor coordinates still
// refer to the visible viewport and are available only when the cursor is
// inside it.
func (rs *RenderState) Overscan() (RenderStateOverscan, error) {
	return rs.getOverscan(vt.GHOSTTY_RENDER_STATE_DATA_OVERSCAN)
}

// OverscanRequest returns the counts most recently set by [RenderState.SetOverscan].
// Both counts are zero if SetOverscan has not been called. The next update uses
// this request. Use [RenderState.Overscan] for the counts from the last update.
func (rs *RenderState) OverscanRequest() (RenderStateOverscan, error) {
	return rs.getOverscan(vt.GHOSTTY_RENDER_STATE_DATA_OVERSCAN_REQUEST)
}

// getOverscan reads either the requested or captured overscan counts.
func (rs *RenderState) getOverscan(data vt.GhosttyRenderStateData) (RenderStateOverscan, error) {
	var v RenderStateOverscan
	err := rs.get(data, scratchSize, func(_ *libc.TLS, out uintptr) {
		c := (*vt.GhosttyRenderStateOverscan)(unsafe.Pointer(out))
		v = RenderStateOverscan{Above: c.Fabove, Below: c.Fbelow}
	})
	return v, err
}
