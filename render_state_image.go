package libghostty

// Render-state image iterator wrapping the
// GhosttyRenderStateImageIterator C APIs.

import (
	"errors"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// ImageProtocol identifies the graphics protocol a program used to send an
// image. It tells a renderer where to read the image's pixels.
//
// Later versions may add more protocols. A renderer should skip images
// with a protocol it doesn't support, so that it keeps working when that
// happens.
//
// C: GhosttyImageProtocol
type ImageProtocol int

const (
	// ImageProtocolKitty is the Kitty graphics protocol. Read the image's
	// pixels with [KittyGraphics.Image].
	ImageProtocolKitty ImageProtocol = vt.GHOSTTY_IMAGE_PROTOCOL_KITTY
)

// RenderStateKittyPlacementKind describes how a Kitty graphics image in a
// render state got its position. A renderer can draw every kind the same
// way. The kind is mostly useful for debugging.
//
// C: GhosttyRenderStateKittyPlacementKind
type RenderStateKittyPlacementKind int

const (
	// RenderStateKittyPlacementKindVirtual is one row of a virtual
	// placement. A program shows a virtual placement by printing
	// placeholder characters where the image should appear. Each run of
	// neighboring placeholder cells in a row becomes one image in the
	// render state, so this kind is always one row tall.
	RenderStateKittyPlacementKindVirtual RenderStateKittyPlacementKind = vt.GHOSTTY_RENDER_STATE_KITTY_PLACEMENT_KIND_VIRTUAL

	// RenderStateKittyPlacementKindRelative is a placement that a program
	// positioned relative to another placement, called its parent, so that
	// the two move together. The parent is a fixed or virtual placement,
	// either directly or through a chain of other relative placements.
	//
	// When the parent is a fixed placement, the position is measured from
	// the parent's cell. The image is listed whenever any part of it is in
	// the captured rows, even if its parent isn't.
	//
	// When the parent is a virtual placement, the position is measured
	// from the top-left corner of the parent's placeholder cells, which is
	// the leftmost column and topmost row that any of those cells use. The
	// image is left out when none of those cells are in the captured rows.
	//
	// See [RenderStateImageIterator] for what the captured rows are.
	RenderStateKittyPlacementKindRelative RenderStateKittyPlacementKind = vt.GHOSTTY_RENDER_STATE_KITTY_PLACEMENT_KIND_RELATIVE

	// RenderStateKittyPlacementKindFixed is a placement at a cell on the
	// screen. It scrolls with the text around it, like a line of output.
	// The program chooses its z-index, so it can be in any layer.
	RenderStateKittyPlacementKindFixed RenderStateKittyPlacementKind = vt.GHOSTTY_RENDER_STATE_KITTY_PLACEMENT_KIND_FIXED
)

// RenderStateImageData identifies one value of an image in a render state.
// Most callers should use [RenderStateImageIterator.Info] or the typed
// getters instead. These keys are for [RenderStateImageIterator.GetMulti].
//
// Each key's comment ends with the C type of its value in parentheses.
// Pixel values are based on the cell size set with [Terminal.Resize].
//
// Keys with a protocol in their name, such as
// [RenderStateImageDataKittyKind], only apply to images from that
// protocol. Reading one for any other image returns an error matching
// [ErrNoValue].
//
// C: GhosttyRenderStateImageData
type RenderStateImageData int

const (
	// RenderStateImageDataInvalid is not a valid key. Reading it returns
	// an error.
	RenderStateImageDataInvalid RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_INVALID

	// RenderStateImageDataProtocol is the protocol the image came from
	// (GhosttyImageProtocol).
	RenderStateImageDataProtocol RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_PROTOCOL

	// RenderStateImageDataImageID is the ID of the image (uint32_t).
	RenderStateImageDataImageID RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_IMAGE_ID

	// RenderStateImageDataImageGeneration is the generation of the image's
	// pixels (uint64_t).
	RenderStateImageDataImageGeneration RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_IMAGE_GENERATION

	// RenderStateImageDataLayer is the layer to draw the image in
	// (GhosttyImageLayer).
	RenderStateImageDataLayer RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_LAYER

	// RenderStateImageDataZ is the z-index of the image within its layer
	// (int32_t).
	RenderStateImageDataZ RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_Z

	// RenderStateImageDataViewportCol is the column of the image's
	// top-left cell (int32_t).
	RenderStateImageDataViewportCol RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_VIEWPORT_COL

	// RenderStateImageDataViewportRow is the row of the image's top-left
	// cell (int32_t).
	RenderStateImageDataViewportRow RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_VIEWPORT_ROW

	// RenderStateImageDataGridCols is the number of columns the image
	// covers (uint32_t).
	RenderStateImageDataGridCols RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_GRID_COLS

	// RenderStateImageDataGridRows is the number of rows the image covers
	// (uint32_t).
	RenderStateImageDataGridRows RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_GRID_ROWS

	// RenderStateImageDataCellOffsetX is the distance in pixels from the
	// left edge of the top-left cell to the image (uint32_t).
	RenderStateImageDataCellOffsetX RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_CELL_OFFSET_X

	// RenderStateImageDataCellOffsetY is the distance in pixels from the
	// top edge of the top-left cell to the image (uint32_t).
	RenderStateImageDataCellOffsetY RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_CELL_OFFSET_Y

	// RenderStateImageDataPixelWidth is the width to draw the image at, in
	// pixels (uint32_t).
	RenderStateImageDataPixelWidth RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_PIXEL_WIDTH

	// RenderStateImageDataPixelHeight is the height to draw the image at,
	// in pixels (uint32_t).
	RenderStateImageDataPixelHeight RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_PIXEL_HEIGHT

	// RenderStateImageDataSourceX is the left edge of the part of the
	// image to draw, in image pixels (uint32_t).
	RenderStateImageDataSourceX RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_X

	// RenderStateImageDataSourceY is the top edge of the part of the image
	// to draw, in image pixels (uint32_t).
	RenderStateImageDataSourceY RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_Y

	// RenderStateImageDataSourceWidth is the width of the part of the
	// image to draw, in image pixels (uint32_t).
	RenderStateImageDataSourceWidth RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_WIDTH

	// RenderStateImageDataSourceHeight is the height of the part of the
	// image to draw, in image pixels (uint32_t).
	RenderStateImageDataSourceHeight RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_HEIGHT

	// RenderStateImageDataKittyPlacementID is the Kitty placement ID, or 0
	// if there isn't one (uint32_t). Kitty images only.
	RenderStateImageDataKittyPlacementID RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_KITTY_PLACEMENT_ID

	// RenderStateImageDataKittyKind is how the Kitty image got its
	// position (GhosttyRenderStateKittyPlacementKind). Kitty images only.
	RenderStateImageDataKittyKind RenderStateImageData = vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_KITTY_KIND
)

// RenderStateImageInfo holds everything needed to draw one image from a
// render state. Read it with [RenderStateImageIterator.Info].
//
// It describes a single rectangle. Copy the source rectangle (SourceX,
// SourceY, SourceWidth and SourceHeight) out of the image and scale it to
// PixelWidth by PixelHeight. Place its top-left corner at this position,
// in pixels from the top-left corner of the viewport:
//
//	x := info.ViewportCol*cellWidth + int32(info.CellOffsetX)
//	y := info.ViewportRow*cellHeight + int32(info.CellOffsetY)
//
// Here cellWidth and cellHeight are the cell size passed to
// [Terminal.Resize].
type RenderStateImageInfo struct {
	// Protocol is the protocol the image came from. Skip images whose
	// protocol you don't support.
	Protocol ImageProtocol

	// ImageID identifies the image. IDs are only unique within one
	// protocol, so look the image up in that protocol's storage. For
	// Kitty images, use [KittyGraphics.Image].
	ImageID uint32

	// ImageGeneration changes whenever the image's pixels change, such as
	// when an animation moves to its next frame. Together with Protocol
	// and ImageID, it makes a good key for caching textures. For Kitty
	// images, it equals [KittyGraphicsImage.Generation]. To know when the
	// list as a whole changed, use [RenderState.ImagePlacementsGeneration].
	ImageGeneration uint64

	// Layer says whether to draw the image behind or in front of the text.
	Layer ImageLayer

	// Z orders images within the same layer. Draw lower values first.
	// Kitty virtual placements always use -1.
	Z int32

	// ViewportCol is the column of the image's top-left cell, where column
	// 0 is the left edge of the viewport. Fixed and relative placements
	// are listed at their full size, so it is negative when one starts to
	// the left of the viewport. Virtual placements always start inside it.
	ViewportCol int32

	// ViewportRow is the row of the image's top-left cell, where row 0 is
	// the top of the viewport. It is negative when the image starts above
	// the viewport, such as in an overscan row. Fixed and relative
	// placements can start above the first overscan row too. Virtual
	// placements always start inside the captured rows.
	ViewportRow int32

	// GridCols is the number of columns the image covers.
	GridCols uint32

	// GridRows is the number of rows the image covers.
	GridRows uint32

	// CellOffsetX is the distance in pixels from the left edge of the
	// top-left cell to the left edge of the image. For fixed and relative
	// placements, it is the offset the program asked for, and is always
	// less than one cell wide. For virtual placements, an image narrower
	// than its cells is centered, so the offset can be more than one cell
	// wide.
	CellOffsetX uint32

	// CellOffsetY is the distance in pixels from the top edge of the
	// top-left cell to the top edge of the image. For fixed and relative
	// placements, it is the offset the program asked for, and is always
	// less than one cell tall. For virtual placements, an image shorter
	// than its cells is centered, so the offset can be more than one cell
	// tall.
	CellOffsetY uint32

	// PixelWidth is the width to draw the image at, in pixels.
	PixelWidth uint32

	// PixelHeight is the height to draw the image at, in pixels.
	PixelHeight uint32

	// SourceX is the left edge of the part of the image to draw, in image
	// pixels.
	SourceX uint32

	// SourceY is the top edge of the part of the image to draw, in image
	// pixels.
	SourceY uint32

	// SourceWidth is the width of the part of the image to draw, in image
	// pixels.
	SourceWidth uint32

	// SourceHeight is the height of the part of the image to draw, in
	// image pixels.
	SourceHeight uint32

	// KittyPlacementID is the Kitty placement ID, or 0 if there isn't one.
	// It is only valid when Protocol is ImageProtocolKitty. See
	// [RenderStateImageIterator.KittyPlacementID] for details.
	KittyPlacementID uint32

	// KittyKind is how the Kitty image got its position. It is only valid
	// when Protocol is ImageProtocolKitty.
	KittyKind RenderStateKittyPlacementKind
}

// RenderStateImageIterator lists the images to draw for a render state.
//
// Create one iterator with [NewRenderStateImageIterator] and keep it for
// as long as you render. After each [RenderState.Update], fill it with
// [RenderState.ImageIterator]. Then call [RenderStateImageIterator.Next]
// in a loop and read each image with [RenderStateImageIterator.Info].
//
// Images are listed in no particular order. Sort them by layer and then by
// z-index before drawing.
//
// The list holds every Kitty graphics image on screen, however the
// program placed it:
//
//   - At a cell, so the image scrolls with the text
//     ([RenderStateKittyPlacementKindFixed]).
//   - With Unicode placeholder characters (U+10EEEE). Each row of
//     placeholder cells is listed as a separate image
//     ([RenderStateKittyPlacementKindVirtual]).
//   - Relative to another placement
//     ([RenderStateKittyPlacementKindRelative]).
//
// The render state works out the position of each image during every
// update, so a renderer can draw them all the same way. Don't also draw
// images from a [KittyGraphicsPlacementIterator], or they are drawn
// twice. Later versions may add more protocols and kinds of images.
//
// An image is listed when any part of it is in the captured rows, which
// are the viewport plus any overscan rows requested with
// [RenderState.SetOverscan]. Fixed and relative placements are listed at
// their full size even when only part of them is visible. They can start
// above or to the left of the viewport, and extend past its right or
// bottom edge. Clip each image to the area you draw, for example with a
// clip rectangle.
//
// Most frames don't change any image. To skip reading the list on those
// frames, use [RenderState.ImagePlacementsGeneration].
//
// The list is empty unless the terminal has a cell size in pixels, which
// you set with [Terminal.Resize].
//
// The render state does not copy image pixels. Read them from the terminal
// right after the update, while you still have exclusive access to it. An
// image only needs to be read again when its generation changes.
//
// The values in an iterator stay valid until the next update of the render
// state that filled it. Values returned by its methods are copies and can
// be kept longer.
//
// C: GhosttyRenderStateImageIterator
type RenderStateImageIterator struct {
	ptr vt.GhosttyRenderStateImageIterator

	// scratch is C memory of imageScratchSize bytes that InfoInto reads
	// every value of an image into with one get_multi call: the keys, the
	// output pointers and the values themselves. It is allocated once so
	// that reading the list each frame allocates nothing.
	scratch uintptr
}

// imageInfoKeys are the values InfoInto reads for every image, in the
// order they are stored in the scratch block.
var imageInfoKeys = [...]vt.GhosttyRenderStateImageData{
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_PROTOCOL,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_IMAGE_ID,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_IMAGE_GENERATION,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_LAYER,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_Z,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_VIEWPORT_COL,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_VIEWPORT_ROW,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_GRID_COLS,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_GRID_ROWS,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_CELL_OFFSET_X,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_CELL_OFFSET_Y,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_PIXEL_WIDTH,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_PIXEL_HEIGHT,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_X,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_Y,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_WIDTH,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_HEIGHT,
	// The Kitty values are only read when the protocol is Kitty, because
	// reading them for another protocol fails with GHOSTTY_NO_VALUE and
	// get_multi stops at the first error.
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_KITTY_PLACEMENT_ID,
	vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_KITTY_KIND,
}

// imageInfoValues is the layout of the values in the scratch block: every
// value is stored in its own 8-byte slot so that each is aligned, in the
// order of imageInfoKeys.
type imageInfoValues struct {
	protocol         int32
	_                int32
	imageID          uint32
	_                uint32
	generation       uint64
	layer            int32
	_                int32
	z                int32
	_                int32
	viewportCol      int32
	_                int32
	viewportRow      int32
	_                int32
	gridCols         uint32
	_                uint32
	gridRows         uint32
	_                uint32
	cellOffsetX      uint32
	_                uint32
	cellOffsetY      uint32
	_                uint32
	pixelWidth       uint32
	_                uint32
	pixelHeight      uint32
	_                uint32
	sourceX          uint32
	_                uint32
	sourceY          uint32
	_                uint32
	sourceWidth      uint32
	_                uint32
	sourceHeight     uint32
	_                uint32
	kittyPlacementID uint32
	_                uint32
	kittyKind        int32
	_                int32
}

// Offsets within the scratch block: the keys, then the output pointers,
// then the values.
const (
	imageKeysOffset   uintptr = 0
	imagePtrsOffset           = imageKeysOffset + uintptr(len(imageInfoKeys))*4
	imageValuesOffset         = imagePtrsOffset + uintptr(len(imageInfoKeys))*8
	imageScratchSize          = imageValuesOffset + unsafe.Sizeof(imageInfoValues{})
)

// NewRenderStateImageIterator returns a new, empty image iterator. Fill it
// with [RenderState.ImageIterator]. Call [RenderStateImageIterator.Close]
// when it is no longer needed.
func NewRenderStateImageIterator() (*RenderStateImageIterator, error) {
	tls := getTLS()
	defer putTLS(tls)
	out := tls.Alloc(handleSize)
	defer tls.Free(handleSize)
	if err := resultError(vt.Xghostty_render_state_image_iterator_new(tls, 0, out)); err != nil {
		return nil, err
	}
	it := &RenderStateImageIterator{ptr: *(*uintptr)(unsafe.Pointer(out))}

	// The block holds pointers, so it starts out zeroed.
	it.scratch = uintptr(allocZeroed(imageScratchSize))
	if it.scratch == 0 {
		vt.Xghostty_render_state_image_iterator_free(tls, it.ptr)
		return nil, &Error{Result: ResultOutOfMemory}
	}
	keys := (*[len(imageInfoKeys)]vt.GhosttyRenderStateImageData)(unsafe.Pointer(it.scratch + imageKeysOffset))
	ptrs := (*[len(imageInfoKeys)]uintptr)(unsafe.Pointer(it.scratch + imagePtrsOffset))
	for i, key := range imageInfoKeys {
		keys[i] = key
		ptrs[i] = it.scratch + imageValuesOffset + uintptr(i*8)
	}
	return it, nil
}

// Close frees the iterator. The iterator must not be used afterward.
// Calling Close more than once is safe.
func (it *RenderStateImageIterator) Close() {
	if it == nil || it.ptr == 0 {
		return
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_render_state_image_iterator_free(tls, it.ptr)
	it.ptr = 0
	Free(unsafe.Pointer(it.scratch), imageScratchSize)
	it.scratch = 0
}

// Next moves to the next image and reports whether there is one. Call it
// before reading the first image.
//
// Only images with something to draw are listed. For example, an image
// that the terminal has not fully received yet is skipped.
func (it *RenderStateImageIterator) Next() bool {
	tls := getTLS()
	defer putTLS(tls)
	return vt.Xghostty_render_state_image_iterator_next(tls, it.ptr) != 0
}

// Info returns every value of the current image. It is faster than
// calling the individual getters, because it reads all of the values in
// a single call into libghostty.
//
// Info returns a new [RenderStateImageInfo] each time. To reuse one
// instead, use [RenderStateImageIterator.InfoInto].
func (it *RenderStateImageIterator) Info() (*RenderStateImageInfo, error) {
	info := new(RenderStateImageInfo)
	if err := it.InfoInto(info); err != nil {
		return nil, err
	}
	return info, nil
}

// InfoInto is like [RenderStateImageIterator.Info], but writes the values
// to dst instead of returning a new struct. Renderers can use it to read
// the list without allocating, for example by keeping a slice of
// RenderStateImageInfo and reusing it every frame.
//
// If InfoInto returns an error, dst is unchanged.
func (it *RenderStateImageIterator) InfoInto(dst *RenderStateImageInfo) error {
	if dst == nil {
		return errors.New("libghostty: nil RenderStateImageInfo")
	}
	tls := getTLS()
	defer putTLS(tls)

	// The common values are read with one get_multi call. The
	// protocol-specific values are read with a second call only for that
	// protocol.
	const common = uintptr(len(imageInfoKeys)) - 2
	keys := it.scratch + imageKeysOffset
	ptrs := it.scratch + imagePtrsOffset
	v := (*imageInfoValues)(unsafe.Pointer(it.scratch + imageValuesOffset))
	*v = imageInfoValues{}
	if err := resultError(vt.Xghostty_render_state_image_get_multi(
		tls, it.ptr, uint64(common), keys, ptrs, 0,
	)); err != nil {
		return err
	}
	if v.protocol == vt.GHOSTTY_IMAGE_PROTOCOL_KITTY {
		if err := resultError(vt.Xghostty_render_state_image_get_multi(
			tls, it.ptr, 2, keys+common*4, ptrs+common*8, 0,
		)); err != nil {
			return err
		}
	}

	*dst = RenderStateImageInfo{
		Protocol:         ImageProtocol(v.protocol),
		ImageID:          v.imageID,
		ImageGeneration:  v.generation,
		Layer:            ImageLayer(v.layer),
		Z:                v.z,
		ViewportCol:      v.viewportCol,
		ViewportRow:      v.viewportRow,
		GridCols:         v.gridCols,
		GridRows:         v.gridRows,
		CellOffsetX:      v.cellOffsetX,
		CellOffsetY:      v.cellOffsetY,
		PixelWidth:       v.pixelWidth,
		PixelHeight:      v.pixelHeight,
		SourceX:          v.sourceX,
		SourceY:          v.sourceY,
		SourceWidth:      v.sourceWidth,
		SourceHeight:     v.sourceHeight,
		KittyPlacementID: v.kittyPlacementID,
		KittyKind:        RenderStateKittyPlacementKind(v.kittyKind),
	}
	return nil
}

// GetMulti reads several values of the current image in one call. Most
// callers should use [RenderStateImageIterator.Info] instead. GetMulti is
// for callers that need only a few values and want to avoid the cost of
// reading the rest.
//
// values[i] receives the value for keys[i]. Each element of values must
// point to a variable with the same size as the C type listed for its key
// on [RenderStateImageData]. For example, use uint32 for uint32_t, uint64
// for uint64_t, and int32 for an enum.
//
// GetMulti returns an error if keys and values have different lengths or
// if a key can't be read. Values for the keys before the one that failed
// may already be written.
//
// C: ghostty_render_state_image_get_multi
func (it *RenderStateImageIterator) GetMulti(keys []RenderStateImageData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[vt.GhosttyRenderStateImageData](&args, keys, values)
	if err != nil {
		return err
	}
	defer args.free()
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_render_state_image_get_multi(
		tls,
		it.ptr,
		uint64(len(keys)),
		uintptr(unsafe.Pointer(cKeys)),
		uintptr(unsafe.Pointer(args.values)),
		0,
	))
}

// get reads one value of the current image, whose C type is listed for
// data on [RenderStateImageData], into TLS stack memory and returns the
// 8 bytes there. The caller reads the value from the front of them.
func (it *RenderStateImageIterator) get(data vt.GhosttyRenderStateImageData) ([8]byte, error) {
	tls := getTLS()
	defer putTLS(tls)
	out := tls.Alloc(scratchSize)
	defer tls.Free(scratchSize)
	*(*[8]byte)(unsafe.Pointer(out)) = [8]byte{}
	if err := resultError(vt.Xghostty_render_state_image_get(tls, it.ptr, data, out)); err != nil {
		return [8]byte{}, err
	}
	return *(*[8]byte)(unsafe.Pointer(out)), nil
}

// getUint32 reads a uint32_t value of the current image.
func (it *RenderStateImageIterator) getUint32(data vt.GhosttyRenderStateImageData) (uint32, error) {
	b, err := it.get(data)
	return *(*uint32)(unsafe.Pointer(&b)), err
}

// getInt32 reads an int32_t or enum value of the current image.
func (it *RenderStateImageIterator) getInt32(data vt.GhosttyRenderStateImageData) (int32, error) {
	b, err := it.get(data)
	return *(*int32)(unsafe.Pointer(&b)), err
}

// Protocol returns the protocol the current image came from. See
// [RenderStateImageInfo.Protocol].
func (it *RenderStateImageIterator) Protocol() (ImageProtocol, error) {
	v, err := it.getInt32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_PROTOCOL)
	return ImageProtocol(v), err
}

// ImageID returns the ID of the current image. See
// [RenderStateImageInfo.ImageID].
func (it *RenderStateImageIterator) ImageID() (uint32, error) {
	return it.getUint32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_IMAGE_ID)
}

// ImageGeneration returns the generation of the current image's pixels.
// See [RenderStateImageInfo.ImageGeneration].
func (it *RenderStateImageIterator) ImageGeneration() (uint64, error) {
	b, err := it.get(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_IMAGE_GENERATION)
	return *(*uint64)(unsafe.Pointer(&b)), err
}

// Layer returns the layer to draw the current image in.
func (it *RenderStateImageIterator) Layer() (ImageLayer, error) {
	v, err := it.getInt32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_LAYER)
	return ImageLayer(v), err
}

// Z returns the z-index of the current image within its layer. Draw lower
// values first.
func (it *RenderStateImageIterator) Z() (int32, error) {
	return it.getInt32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_Z)
}

// ViewportPos returns the column and row of the current image's top-left
// cell, counted from the top-left corner of the viewport. Either value is
// negative when the image starts outside the viewport.
func (it *RenderStateImageIterator) ViewportPos() (col, row int32, err error) {
	if col, err = it.getInt32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_VIEWPORT_COL); err != nil {
		return 0, 0, err
	}
	if row, err = it.getInt32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_VIEWPORT_ROW); err != nil {
		return 0, 0, err
	}
	return col, row, nil
}

// GridSize returns the number of columns and rows the current image
// covers.
func (it *RenderStateImageIterator) GridSize() (cols, rows uint32, err error) {
	if cols, err = it.getUint32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_GRID_COLS); err != nil {
		return 0, 0, err
	}
	if rows, err = it.getUint32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_GRID_ROWS); err != nil {
		return 0, 0, err
	}
	return cols, rows, nil
}

// CellOffset returns the distance in pixels from the top-left corner of
// the current image's top-left cell to the image itself. See
// [RenderStateImageInfo.CellOffsetX].
func (it *RenderStateImageIterator) CellOffset() (x, y uint32, err error) {
	if x, err = it.getUint32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_CELL_OFFSET_X); err != nil {
		return 0, 0, err
	}
	if y, err = it.getUint32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_CELL_OFFSET_Y); err != nil {
		return 0, 0, err
	}
	return x, y, nil
}

// PixelSize returns the size to draw the current image at, in pixels.
func (it *RenderStateImageIterator) PixelSize() (width, height uint32, err error) {
	if width, err = it.getUint32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_PIXEL_WIDTH); err != nil {
		return 0, 0, err
	}
	if height, err = it.getUint32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_PIXEL_HEIGHT); err != nil {
		return 0, 0, err
	}
	return width, height, nil
}

// SourceRect returns the part of the current image to draw, in image
// pixels. Scale it to the size returned by
// [RenderStateImageIterator.PixelSize].
func (it *RenderStateImageIterator) SourceRect() (x, y, width, height uint32, err error) {
	keys := [...]vt.GhosttyRenderStateImageData{
		vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_X,
		vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_Y,
		vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_WIDTH,
		vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_HEIGHT,
	}
	var outs [4]uint32
	for i, key := range keys {
		if outs[i], err = it.getUint32(key); err != nil {
			return 0, 0, 0, 0, err
		}
	}
	return outs[0], outs[1], outs[2], outs[3], nil
}

// KittyPlacementID returns the Kitty placement ID of the current image, or
// 0 if there isn't one.
//
// For a virtual placement, this is the placement ID that the placeholder
// cells ask for. Cells that don't ask for a specific placement report 0.
// For a fixed or relative placement, this is the ID the program gave it,
// or 0 if it didn't give one.
//
// KittyPlacementID returns an error matching [ErrNoValue] when the image
// is not a Kitty image.
func (it *RenderStateImageIterator) KittyPlacementID() (uint32, error) {
	return it.getUint32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_KITTY_PLACEMENT_ID)
}

// KittyKind returns how the current Kitty image got its position. It
// returns an error matching [ErrNoValue] when the image is not a Kitty
// image.
func (it *RenderStateImageIterator) KittyKind() (RenderStateKittyPlacementKind, error) {
	v, err := it.getInt32(vt.GHOSTTY_RENDER_STATE_IMAGE_DATA_KITTY_KIND)
	return RenderStateKittyPlacementKind(v), err
}
