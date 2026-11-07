package libghostty

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// ghosttySizeToInt converts a C size to a Go slice length without allowing
// an overflowing conversion to produce an invalid unsafe.Slice length.
func ghosttySizeToInt(size uint64) (int, bool) {
	if size > uint64(^uint(0)>>1) {
		return 0, false
	}
	return int(size), true
}

// copyGhosttyString copies a borrowed, binary-safe GhosttyString into Go
// memory. Library-produced empty strings point to valid storage, but no
// bytes need to be read when the length is zero.
func copyGhosttyString(value vt.GhosttyString) ([]byte, bool) {
	length, ok := ghosttySizeToInt(value.Flen1)
	if !ok {
		return nil, false
	}
	if length == 0 {
		return []byte{}, true
	}
	if value.Fptr == 0 {
		return nil, false
	}

	return append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(value.Fptr)), length)...), true
}

// copyGhosttyStringArray copies a borrowed GhosttyString array into Go-owned
// strings. A nonzero count requires a non-NULL array pointer.
func copyGhosttyStringArray(values uintptr, count uint64) ([]string, bool) {
	length, ok := ghosttySizeToInt(count)
	if !ok || (length > 0 && values == 0) {
		return nil, false
	}

	result := make([]string, length)
	if length == 0 {
		return result, true
	}
	for i, value := range unsafe.Slice((*vt.GhosttyString)(unsafe.Pointer(values)), length) {
		data, valid := copyGhosttyString(value)
		if !valid {
			return nil, false
		}
		result[i] = string(data)
	}
	return result, true
}

// cGhosttyStringArray owns a C-allocated array of GhosttyString descriptors
// and the C-allocated byte buffers referenced by those descriptors. Reply
// structs passed synchronously to libghostty point to this C memory, which
// the garbage collector does not move.
type cGhosttyStringArray struct {
	// ptr is the C address of the descriptor array, or 0 when empty.
	ptr            uintptr
	descriptorSize uintptr
	buffers        []cGhosttyStringBuffer
}

// cGhosttyStringBuffer records one allocation so it can be released with the
// exact size required by ghostty_free.
type cGhosttyStringBuffer struct {
	ptr  unsafe.Pointer
	size uintptr
}

// newCGhosttyStringArray copies values into a C-owned GhosttyString array.
// The caller must close the returned owner after the synchronous C call.
func newCGhosttyStringArray(values [][]byte) (*cGhosttyStringArray, error) {
	a := &cGhosttyStringArray{}
	if len(values) == 0 {
		return a, nil
	}

	elementSize := unsafe.Sizeof(vt.GhosttyString{})
	if uintptr(len(values)) > ^uintptr(0)/elementSize {
		return nil, &Error{Result: ResultLimitExceeded}
	}
	a.descriptorSize = uintptr(len(values)) * elementSize
	// Each GhosttyString in the array contains a pointer, so the array
	// must start out zeroed. See allocZeroed.
	descriptorPtr := allocZeroed(a.descriptorSize)
	if descriptorPtr == nil {
		return nil, &Error{Result: ResultOutOfMemory}
	}
	a.ptr = uintptr(descriptorPtr)
	a.buffers = make([]cGhosttyStringBuffer, 0, len(values))

	descriptors := unsafe.Slice((*vt.GhosttyString)(descriptorPtr), len(values))
	for i, value := range values {
		if len(value) == 0 {
			descriptors[i] = vt.GhosttyString{}
			continue
		}

		size := uintptr(len(value))
		ptr := Alloc(size)
		if ptr == nil {
			a.close()
			return nil, &Error{Result: ResultOutOfMemory}
		}
		copy(unsafe.Slice((*byte)(ptr), len(value)), value)
		a.buffers = append(a.buffers, cGhosttyStringBuffer{ptr: ptr, size: size})
		descriptors[i] = vt.GhosttyString{
			Fptr:  uintptr(ptr),
			Flen1: uint64(len(value)),
		}
	}

	return a, nil
}

// close releases every descriptor and data allocation owned by a.
func (a *cGhosttyStringArray) close() {
	if a == nil {
		return
	}
	for _, buffer := range a.buffers {
		Free(buffer.ptr, buffer.size)
	}
	Free(unsafe.Pointer(a.ptr), a.descriptorSize)
	a.ptr = 0
	a.descriptorSize = 0
	a.buffers = nil
}

// CursorStyle is the shape of the cursor. Programs change it with the DECSCUSR
// escape sequence (CSI Ps SP q). [RenderState.CursorVisualStyle] reports the
// shape to draw, and [Terminal.SetDefaultCursorStyle] sets the shape that
// DECSCUSR 0 restores.
//
// C: GhosttyCursorStyle
type CursorStyle int

const (
	// CursorStyleBar is a bar cursor (DECSCUSR 5, 6).
	CursorStyleBar CursorStyle = vt.GHOSTTY_CURSOR_STYLE_BAR

	// CursorStyleBlock is a block cursor (DECSCUSR 1, 2).
	CursorStyleBlock CursorStyle = vt.GHOSTTY_CURSOR_STYLE_BLOCK

	// CursorStyleUnderline is an underline cursor (DECSCUSR 3, 4).
	CursorStyleUnderline CursorStyle = vt.GHOSTTY_CURSOR_STYLE_UNDERLINE

	// CursorStyleBlockHollow is a hollow block cursor.
	CursorStyleBlockHollow CursorStyle = vt.GHOSTTY_CURSOR_STYLE_BLOCK_HOLLOW
)

// ImageLayer says whether an image is drawn behind or in front of the
// terminal's text. The Kitty graphics protocol assigns each placement a layer
// based on its z-index, as described on each value below.
//
// Draw a frame from back to front in this order:
//
//  1. Images in [ImageLayerBelowBackground].
//  2. Cell background colors.
//  3. Images in [ImageLayerBelowText].
//  4. Text.
//  5. Images in [ImageLayerAboveText].
//
// Within one layer, draw images with a lower z-index first.
//
// C: GhosttyImageLayer
type ImageLayer int

const (
	// ImageLayerBelowBackground draws the image behind the cell background
	// colors (z < INT32_MIN/2).
	ImageLayerBelowBackground ImageLayer = vt.GHOSTTY_IMAGE_LAYER_BELOW_BACKGROUND

	// ImageLayerBelowText draws the image in front of the cell background
	// colors and behind the text (INT32_MIN/2 <= z < 0).
	ImageLayerBelowText ImageLayer = vt.GHOSTTY_IMAGE_LAYER_BELOW_TEXT

	// ImageLayerAboveText draws the image in front of the text (z >= 0).
	ImageLayerAboveText ImageLayer = vt.GHOSTTY_IMAGE_LAYER_ABOVE_TEXT
)

// ImageLayerMask selects any number of image layers. Combine values with |.
// Use it with [KittyGraphicsPlacementIterator.SetLayers].
//
// C: GHOSTTY_IMAGE_LAYER_MASK_*
type ImageLayerMask uint32

const (
	// ImageLayerMaskBelowBackground selects [ImageLayerBelowBackground].
	ImageLayerMaskBelowBackground ImageLayerMask = vt.GHOSTTY_IMAGE_LAYER_MASK_BELOW_BACKGROUND

	// ImageLayerMaskBelowText selects [ImageLayerBelowText].
	ImageLayerMaskBelowText ImageLayerMask = vt.GHOSTTY_IMAGE_LAYER_MASK_BELOW_TEXT

	// ImageLayerMaskAboveText selects [ImageLayerAboveText].
	ImageLayerMaskAboveText ImageLayerMask = vt.GHOSTTY_IMAGE_LAYER_MASK_ABOVE_TEXT

	// ImageLayerMaskAll selects every layer.
	ImageLayerMaskAll = ImageLayerMaskBelowBackground | ImageLayerMaskBelowText | ImageLayerMaskAboveText
)

// TypeJSON returns the process-lifetime, versioned libghostty-vt C type
// manifest for the current target. The manifest describes all public types,
// including layouts, enum values, union fields, and packed bit layouts.
// Consumers should reject unknown schema versions and use this metadata rather
// than hardcoding ABI details such as [Cell] bit positions.
func TypeJSON() string {
	tls := getTLS()
	defer putTLS(tls)
	return libc.GoString(vt.Xghostty_type_json(tls))
}

// SurfacePosition is an x/y position in rendered surface pixel space.
//
// This is not a terminal grid coordinate. The origin is the top-left of the
// rendered terminal surface.
// C: GhosttySurfacePosition
type SurfacePosition struct {
	// X is the horizontal position in surface pixels.
	X float64

	// Y is the vertical position in surface pixels.
	Y float64
}

// toC converts a Go SurfacePosition to a C GhosttySurfacePosition.
func (p SurfacePosition) toC() vt.GhosttySurfacePosition {
	return vt.GhosttySurfacePosition{
		Fx: p.X,
		Fy: p.Y,
	}
}
