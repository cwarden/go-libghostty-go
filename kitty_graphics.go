package libghostty

// Kitty graphics protocol bindings wrapping the C API from kitty_graphics.h.
// Provides access to images and placements stored via the Kitty graphics
// protocol.

import (
	"errors"
	"math"
	"time"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// KittyGraphicsData identifies a storage-level Kitty graphics query.
// C: GhosttyKittyGraphicsData
type KittyGraphicsData int

const (
	// KittyGraphicsDataInvalid is an invalid or sentinel query.
	KittyGraphicsDataInvalid KittyGraphicsData = vt.GHOSTTY_KITTY_GRAPHICS_DATA_INVALID

	// KittyGraphicsDataPlacementIterator populates a placement iterator.
	KittyGraphicsDataPlacementIterator KittyGraphicsData = vt.GHOSTTY_KITTY_GRAPHICS_DATA_PLACEMENT_ITERATOR

	// KittyGraphicsDataGeneration returns the storage-wide mutation stamp.
	KittyGraphicsDataGeneration KittyGraphicsData = vt.GHOSTTY_KITTY_GRAPHICS_DATA_GENERATION
)

// KittyGraphicsImageData identifies a data field for Kitty graphics
// image queries.
// C: GhosttyKittyGraphicsImageData
type KittyGraphicsImageData int

const (
	// KittyGraphicsImageDataInvalid is an invalid / sentinel value.
	KittyGraphicsImageDataInvalid KittyGraphicsImageData = vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_INVALID

	// KittyGraphicsImageDataID is the image ID (uint32_t).
	KittyGraphicsImageDataID KittyGraphicsImageData = vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_ID

	// KittyGraphicsImageDataNumber is the image number (uint32_t).
	KittyGraphicsImageDataNumber KittyGraphicsImageData = vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_NUMBER

	// KittyGraphicsImageDataWidth is the image width in pixels (uint32_t).
	KittyGraphicsImageDataWidth KittyGraphicsImageData = vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_WIDTH

	// KittyGraphicsImageDataHeight is the image height in pixels (uint32_t).
	KittyGraphicsImageDataHeight KittyGraphicsImageData = vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_HEIGHT

	// KittyGraphicsImageDataFormat is the pixel format of the image
	// (GhosttyKittyImageFormat).
	KittyGraphicsImageDataFormat KittyGraphicsImageData = vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_FORMAT

	// KittyGraphicsImageDataCompression is the compression of the image
	// (GhosttyKittyImageCompression).
	KittyGraphicsImageDataCompression KittyGraphicsImageData = vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_COMPRESSION

	// KittyGraphicsImageDataPixels is the image's pixel data (GhosttyString).
	// The string points into memory owned by the terminal and stays valid until
	// the terminal is next modified. It returns [ResultNoValue] if the image's
	// pixels have not arrived yet. Prefer [KittyGraphicsImage.Pixels].
	KittyGraphicsImageDataPixels KittyGraphicsImageData = vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_PIXELS

	// KittyGraphicsImageDataExpectedLen is the size of the image's pixel data
	// in bytes (size_t). It is available even before the pixels arrive.
	KittyGraphicsImageDataExpectedLen KittyGraphicsImageData = vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_EXPECTED_LEN

	// KittyGraphicsImageDataGeneration is the image's content generation
	// stamp (uint64_t).
	KittyGraphicsImageDataGeneration KittyGraphicsImageData = vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_GENERATION
)

// KittyGraphicsPlacementData identifies a data field for Kitty graphics
// placement queries.
// C: GhosttyKittyGraphicsPlacementData
type KittyGraphicsPlacementData int

const (
	// KittyGraphicsPlacementDataInvalid is an invalid / sentinel value.
	KittyGraphicsPlacementDataInvalid KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_INVALID

	// KittyGraphicsPlacementDataImageID is the image ID this placement
	// belongs to (uint32_t).
	KittyGraphicsPlacementDataImageID KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_IMAGE_ID

	// KittyGraphicsPlacementDataPlacementID is the placement ID (uint32_t).
	KittyGraphicsPlacementDataPlacementID KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_PLACEMENT_ID

	// KittyGraphicsPlacementDataIsVirtual indicates whether this is a
	// virtual placement (bool).
	KittyGraphicsPlacementDataIsVirtual KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_IS_VIRTUAL

	// KittyGraphicsPlacementDataXOffset is the pixel offset from the left
	// edge of the cell (uint32_t).
	KittyGraphicsPlacementDataXOffset KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_X_OFFSET

	// KittyGraphicsPlacementDataYOffset is the pixel offset from the top
	// edge of the cell (uint32_t).
	KittyGraphicsPlacementDataYOffset KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_Y_OFFSET

	// KittyGraphicsPlacementDataSourceX is the source rectangle x origin
	// in pixels (uint32_t).
	KittyGraphicsPlacementDataSourceX KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_X

	// KittyGraphicsPlacementDataSourceY is the source rectangle y origin
	// in pixels (uint32_t).
	KittyGraphicsPlacementDataSourceY KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_Y

	// KittyGraphicsPlacementDataSourceWidth is the source rectangle width
	// in pixels (uint32_t).
	KittyGraphicsPlacementDataSourceWidth KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_WIDTH

	// KittyGraphicsPlacementDataSourceHeight is the source rectangle height
	// in pixels (uint32_t).
	KittyGraphicsPlacementDataSourceHeight KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_HEIGHT

	// KittyGraphicsPlacementDataColumns is the number of columns this
	// placement occupies (uint32_t).
	KittyGraphicsPlacementDataColumns KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_COLUMNS

	// KittyGraphicsPlacementDataRows is the number of rows this placement
	// occupies (uint32_t).
	KittyGraphicsPlacementDataRows KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_ROWS

	// KittyGraphicsPlacementDataZ is the z-index for this placement
	// (int32_t).
	KittyGraphicsPlacementDataZ KittyGraphicsPlacementData = vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_Z
)

// KittyGraphics is a handle to the Kitty graphics image storage
// associated with a terminal's active screen. It is borrowed from the
// terminal and remains valid until the next mutating terminal call
// (for example [Terminal.VTWrite] or [Terminal.Reset]). Access to this
// handle must be serialized with mutations of the owning terminal.
//
// C: GhosttyKittyGraphics
type KittyGraphics struct {
	ptr vt.GhosttyKittyGraphics
}

// KittyGraphicsImage is a handle to a single Kitty graphics image.
// It is borrowed from the storage and remains valid until the next
// mutating terminal call. Access to this handle and any borrowed pixel
// data derived from it must be serialized with mutations of the owning
// terminal.
//
// C: GhosttyKittyGraphicsImage
type KittyGraphicsImage struct {
	ptr vt.GhosttyKittyGraphicsImage
}

// KittyGraphicsPlacementIterator iterates over placements in the
// Kitty graphics storage. It is independently owned and must be
// freed by calling Close, but the data it yields is only valid
// while the underlying terminal is not mutated. Access to the iterator
// must therefore be serialized with mutations of the terminal that
// produced it.
//
// The iterator lists every placement in the storage, including virtual
// placements and placements that are off screen. Use it in tools such as
// debuggers and tests that need to see what the storage holds.
//
// Renderers should use [RenderStateImageIterator] instead. It lists every
// image on screen with its position already worked out. Drawing from both
// lists draws each image twice.
//
// C: GhosttyKittyGraphicsPlacementIterator
type KittyGraphicsPlacementIterator struct {
	ptr vt.GhosttyKittyGraphicsPlacementIterator
}

// KittyImageFormat describes the pixel format of a Kitty graphics image.
//
// C: GhosttyKittyImageFormat
type KittyImageFormat int

const (
	// KittyImageFormatRGB is 24-bit RGB (3 bytes per pixel).
	KittyImageFormatRGB KittyImageFormat = vt.GHOSTTY_KITTY_IMAGE_FORMAT_RGB

	// KittyImageFormatRGBA is 32-bit RGBA (4 bytes per pixel).
	KittyImageFormatRGBA KittyImageFormat = vt.GHOSTTY_KITTY_IMAGE_FORMAT_RGBA

	// KittyImageFormatPNG is compressed PNG data.
	KittyImageFormatPNG KittyImageFormat = vt.GHOSTTY_KITTY_IMAGE_FORMAT_PNG

	// KittyImageFormatGrayAlpha is 16-bit gray+alpha (2 bytes per pixel).
	KittyImageFormatGrayAlpha KittyImageFormat = vt.GHOSTTY_KITTY_IMAGE_FORMAT_GRAY_ALPHA

	// KittyImageFormatGray is 8-bit grayscale (1 byte per pixel).
	KittyImageFormatGray KittyImageFormat = vt.GHOSTTY_KITTY_IMAGE_FORMAT_GRAY
)

// KittyImageCompression describes the compression of a Kitty graphics image.
//
// C: GhosttyKittyImageCompression
type KittyImageCompression int

const (
	// KittyImageCompressionNone means no compression.
	KittyImageCompressionNone KittyImageCompression = vt.GHOSTTY_KITTY_IMAGE_COMPRESSION_NONE

	// KittyImageCompressionZlibDeflate means zlib/deflate compression.
	KittyImageCompressionZlibDeflate KittyImageCompression = vt.GHOSTTY_KITTY_IMAGE_COMPRESSION_ZLIB_DEFLATE
)

// Selection represents a grid selection range defined by two grid
// references. Because Start and End are [GridRef] values, a Selection
// returned by Kitty graphics helpers is also a borrowed view and is
// invalidated by the next mutation of the owning terminal.
//
// C: GhosttySelection
type Selection struct {
	// Start is the start of the selection range (inclusive).
	Start GridRef

	// End is the end of the selection range (inclusive).
	End GridRef

	// Rectangle indicates whether the selection is rectangular (block)
	// rather than linear.
	Rectangle bool
}

// selectionFromC converts a C GhosttySelection to a Go Selection.
func selectionFromC(cs vt.GhosttySelection) Selection {
	return Selection{
		Start:     GridRef{ref: cs.Fstart},
		End:       GridRef{ref: cs.Fend},
		Rectangle: cs.Frectangle != 0,
	}
}

// KittyGraphicsImageInfo contains all image metadata in a single struct.
// This is a Go-only convenience type; it has no corresponding C struct.
// Populated via get_multi in a single call.
type KittyGraphicsImageInfo struct {
	// ID is the image ID.
	ID uint32

	// Number is the image number.
	Number uint32

	// Width is the image width in pixels.
	Width uint32

	// Height is the image height in pixels.
	Height uint32

	// Format is the pixel format of the image.
	Format KittyImageFormat

	// Compression is the compression of the image.
	Compression KittyImageCompression

	// Generation is the image's process-wide content generation stamp. It
	// changes when an animated image advances to another frame.
	Generation uint64

	// ExpectedLen is the size of the image's pixel data in bytes. It is
	// available even before the pixels arrive, in which case the image already
	// counts this many bytes against the storage limit.
	ExpectedLen uint

	// PixelsPending reports that metadata is resident but the pixels have not
	// arrived yet. Pixels is nil while this is true.
	PixelsPending bool

	// Pixels is a borrowed slice of the raw pixel data. Only valid until the
	// next mutating terminal call.
	Pixels []byte
}

// KittyGraphicsPlacementInfo contains all placement metadata in a single
// struct. This is a Go-only convenience type; it has no corresponding
// C struct. Populated via get_multi in a single call.
type KittyGraphicsPlacementInfo struct {
	// ImageID is the image ID this placement belongs to.
	ImageID uint32

	// PlacementID is the placement ID.
	PlacementID uint32

	// IsVirtual indicates whether this is a virtual placement (unicode placeholder).
	IsVirtual bool

	// XOffset is the pixel offset from the left edge of the cell.
	XOffset uint32

	// YOffset is the pixel offset from the top edge of the cell.
	YOffset uint32

	// SourceX is the source rectangle x origin in pixels.
	SourceX uint32

	// SourceY is the source rectangle y origin in pixels.
	SourceY uint32

	// SourceWidth is the source rectangle width in pixels (0 = full image width).
	SourceWidth uint32

	// SourceHeight is the source rectangle height in pixels (0 = full image height).
	SourceHeight uint32

	// Columns is the number of columns this placement occupies.
	Columns uint32

	// Rows is the number of rows this placement occupies.
	Rows uint32

	// Z is the z-index for this placement.
	Z int32
}

// KittyGraphicsPlacementRenderInfo contains all rendering geometry for a
// placement in a single struct. Combines pixel size, grid size, viewport
// position, and source rectangle into one call.
//
// Use it to inspect the storage. When drawing, use [RenderStateImageInfo]
// instead, which has the same values for every image on screen.
//
// C: GhosttyKittyGraphicsPlacementRenderInfo
type KittyGraphicsPlacementRenderInfo struct {
	// PixelWidth is the rendered width in pixels.
	PixelWidth uint32

	// PixelHeight is the rendered height in pixels.
	PixelHeight uint32

	// GridCols is the number of grid columns the placement occupies.
	GridCols uint32

	// GridRows is the number of grid rows the placement occupies.
	GridRows uint32

	// ViewportCol is the viewport-relative column (may be negative
	// for partially visible placements).
	ViewportCol int32

	// ViewportRow is the viewport-relative row (may be negative
	// for partially visible placements).
	ViewportRow int32

	// ViewportVisible reports whether any part of the placement is in the
	// viewport. It is false when the placement is entirely outside the
	// viewport. It is also false for virtual placements and placements
	// relative to them, because the storage can't tell where those are.
	// When ViewportVisible is false, ignore ViewportCol and ViewportRow.
	ViewportVisible bool

	// SourceX is the resolved source rectangle x origin in pixels.
	SourceX uint32

	// SourceY is the resolved source rectangle y origin in pixels.
	SourceY uint32

	// SourceWidth is the resolved source rectangle width in pixels.
	SourceWidth uint32

	// SourceHeight is the resolved source rectangle height in pixels.
	SourceHeight uint32
}

// outValue calls get with an 8-byte block of TLS stack memory for the
// value libghostty writes, and returns the block's bytes. The block is
// zeroed first so that a value narrower than 8 bytes can be read from
// the front of the result.
func outValue(get func(tls *libc.TLS, out uintptr) vt.GhosttyResult) ([8]byte, error) {
	tls := getTLS()
	defer putTLS(tls)
	out := tls.Alloc(scratchSize)
	defer tls.Free(scratchSize)
	*(*[8]byte)(unsafe.Pointer(out)) = [8]byte{}
	if err := resultError(get(tls, out)); err != nil {
		return [8]byte{}, err
	}
	return *(*[8]byte)(unsafe.Pointer(out)), nil
}

// PlacementIterator populates the given iterator with placement data
// from this storage. The iterator must have been created with
// NewKittyGraphicsPlacementIterator.
func (kg *KittyGraphics) PlacementIterator(iter *KittyGraphicsPlacementIterator) error {
	tls := getTLS()
	defer putTLS(tls)
	out := tls.Alloc(handleSize)
	defer tls.Free(handleSize)
	*(*uintptr)(unsafe.Pointer(out)) = iter.ptr
	if err := resultError(vt.Xghostty_kitty_graphics_get(
		tls,
		kg.ptr,
		vt.GHOSTTY_KITTY_GRAPHICS_DATA_PLACEMENT_ITERATOR,
		out,
	)); err != nil {
		return err
	}
	iter.ptr = *(*uintptr)(unsafe.Pointer(out))
	return nil
}

// Generation returns a value that changes whenever a program sends,
// replaces, places, or deletes an image. If it matches a value read
// earlier, the storage holds the same images and placements as it did
// then. A value is never reused, even by another screen, after a reset, or
// by another terminal, so it is safe to use as a cache key on its own.
//
// Placements can move on screen without changing this value, because
// scrolling and resizing don't change the storage. To know when to read
// the render state's image list again, use
// [RenderState.ImagePlacementsGeneration]. Generation is useful for work
// that depends only on what the storage holds, such as freeing the
// textures of deleted images.
func (kg *KittyGraphics) Generation() (uint64, error) {
	b, err := outValue(func(tls *libc.TLS, out uintptr) vt.GhosttyResult {
		return vt.Xghostty_kitty_graphics_get(tls, kg.ptr, vt.GHOSTTY_KITTY_GRAPHICS_DATA_GENERATION, out)
	})
	if err != nil {
		return 0, err
	}
	return *(*uint64)(unsafe.Pointer(&b)), nil
}

// AdvanceAnimations moves every playing Kitty graphics animation to the
// frame that is due at time now. It reports whether any animation is
// still playing and, if so, the time its next frame is due.
//
// Programs can send images with several frames, much like an animated
// GIF, and ask the terminal to play them. The terminal stores the frames
// but has no timer of its own. Instead, the renderer calls
// AdvanceAnimations once each time it prepares a frame to draw.
//
// Both now and next are measured on a clock that the caller chooses, as
// the time elapsed since some fixed starting point. A good choice is the
// time since the application started. Keep the same starting point for
// as long as the terminal exists, and never pass a time earlier than one
// passed before. Times are rounded down to whole milliseconds.
//
// If ok is true, schedule a single redraw at time next and call
// AdvanceAnimations again from that redraw. Don't schedule redraws
// closer together than the display's refresh interval, which is about
// 8 milliseconds at 120 Hz. If ok is false, nothing is playing and
// there is nothing to schedule. If libghostty was built without Kitty
// graphics support, this returns an error with result [ResultUnsupported].
//
// There is no need to watch for animations starting. A program starts an
// animation by writing to the terminal, and the redraw that follows that
// write calls AdvanceAnimations, which picks up the new animation.
//
// Each animation moves forward by at most one frame per call. If calls
// arrive late, playback slows down rather than skipping frames.
//
// When an image moves to a new frame, its [KittyGraphicsImage.Generation]
// changes and [KittyGraphicsImage.Pixels] returns the pixels of the new
// frame. A renderer that already caches textures by generation shows new
// frames with no extra work.
//
// AdvanceAnimations changes the stored images. Like [Terminal.VTWrite],
// it invalidates any [KittyGraphicsImage] and pixel data obtained before
// the call, so look images up again afterwards. kg itself stays valid.
func (kg *KittyGraphics) AdvanceAnimations(now time.Duration) (next time.Duration, ok bool, err error) {
	// The C API takes whole milliseconds as an unsigned integer. Treat a
	// negative time as the start of the clock rather than letting it wrap
	// around to a time far in the future.
	nowMs := max(now.Milliseconds(), 0)

	b, err := outValue(func(tls *libc.TLS, out uintptr) vt.GhosttyResult {
		return vt.Xghostty_kitty_graphics_animation_tick(tls, kg.ptr, uint64(nowMs), out)
	})
	if errors.Is(err, ErrNoValue) {
		// Nothing is playing. This is the normal idle state, not a failure.
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	nextMs := *(*uint64)(unsafe.Pointer(&b))

	// libghostty saturates the next time instead of overflowing. Do the
	// same when converting to a Duration, which has a smaller range.
	const maxMs = uint64(math.MaxInt64 / int64(time.Millisecond))
	if nextMs > maxMs {
		return time.Duration(math.MaxInt64), true, nil
	}
	return time.Duration(nextMs) * time.Millisecond, true, nil
}

// Image looks up a Kitty graphics image by its image ID. Returns nil
// if no image with the given ID exists.
func (kg *KittyGraphics) Image(imageID uint32) *KittyGraphicsImage {
	tls := getTLS()
	defer putTLS(tls)
	ptr := vt.Xghostty_kitty_graphics_image(tls, kg.ptr, imageID)
	if ptr == 0 {
		return nil
	}
	return &KittyGraphicsImage{ptr: ptr}
}

// get reads one value of the image, whose C type is listed for data on
// [KittyGraphicsImageData], into 8 bytes of TLS stack memory and returns
// them.
func (img *KittyGraphicsImage) get(data vt.GhosttyKittyGraphicsImageData) ([8]byte, error) {
	return outValue(func(tls *libc.TLS, out uintptr) vt.GhosttyResult {
		return vt.Xghostty_kitty_graphics_image_get(tls, img.ptr, data, out)
	})
}

// getUint32 reads a uint32_t value of the image.
func (img *KittyGraphicsImage) getUint32(data vt.GhosttyKittyGraphicsImageData) (uint32, error) {
	b, err := img.get(data)
	return *(*uint32)(unsafe.Pointer(&b)), err
}

// ID returns the image ID.
func (img *KittyGraphicsImage) ID() (uint32, error) {
	return img.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_ID)
}

// Number returns the image number.
func (img *KittyGraphicsImage) Number() (uint32, error) {
	return img.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_NUMBER)
}

// Width returns the image width in pixels.
func (img *KittyGraphicsImage) Width() (uint32, error) {
	return img.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_WIDTH)
}

// Height returns the image height in pixels.
func (img *KittyGraphicsImage) Height() (uint32, error) {
	return img.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_HEIGHT)
}

// GetMulti queries multiple image data fields in a single call.
// This is a low-level function. Prefer the typed getters (ID, Width,
// Height, Format, etc.) or Info() for normal use. GetMulti is useful
// when you need a custom subset of fields and want to avoid per-field
// call overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/kitty_graphics.h, GhosttyKittyGraphicsImageData
// enum).
// Use a Go type with the same size as the C type, such as uint32
// for uint32_t, bool for bool, and int32 for an enum.
//
// GetMulti returns an error if keys and values have different lengths
// or if a key cannot be read. In the second case, values for the keys
// before it may already have been written.
//
// Example:
//
//	var w, h uint32
//	err := img.GetMulti(
//		[]KittyGraphicsImageData{KittyGraphicsImageDataWidth, KittyGraphicsImageDataHeight},
//		[]unsafe.Pointer{unsafe.Pointer(&w), unsafe.Pointer(&h)},
//	)
//
// C: ghostty_kitty_graphics_image_get_multi
func (img *KittyGraphicsImage) GetMulti(keys []KittyGraphicsImageData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[vt.GhosttyKittyGraphicsImageData](&args, keys, values)
	if err != nil {
		return err
	}
	defer args.free()
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_kitty_graphics_image_get_multi(
		tls,
		img.ptr,
		uint64(len(keys)),
		uintptr(unsafe.Pointer(cKeys)),
		uintptr(unsafe.Pointer(args.values)),
		0,
	))
}

// Format returns the pixel format of the image.
func (img *KittyGraphicsImage) Format() (KittyImageFormat, error) {
	b, err := img.get(vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_FORMAT)
	return KittyImageFormat(*(*int32)(unsafe.Pointer(&b))), err
}

// Compression returns the compression of the image.
func (img *KittyGraphicsImage) Compression() (KittyImageCompression, error) {
	b, err := img.get(vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_COMPRESSION)
	return KittyImageCompression(*(*int32)(unsafe.Pointer(&b))), err
}

// Generation returns the process-wide content generation stamp. It changes
// when the image is added or replaced and whenever an animated image advances
// to another frame.
func (img *KittyGraphicsImage) Generation() (uint64, error) {
	b, err := img.get(vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_GENERATION)
	return *(*uint64)(unsafe.Pointer(&b)), err
}

// Info returns all image metadata and pixels in a single call. This is more
// efficient than calling every getter individually. Pending images still return
// their metadata and ExpectedLen, with PixelsPending=true and Pixels nil.
func (img *KittyGraphicsImage) Info() (*KittyGraphicsImageInfo, error) {
	// Output variables — one per field, typed to match the C API. The
	// pinner in getMultiArgs keeps them on the heap, where libghostty can
	// write to them.
	var (
		id          uint32
		number      uint32
		width       uint32
		height      uint32
		format      vt.GhosttyKittyImageFormat
		compression vt.GhosttyKittyImageCompression
		generation  uint64
		expectedLen uint64
		pixels      vt.GhosttyString
	)

	// Keys identify which fields to fetch; order must match values. Pixels
	// must stay last. get_multi stops at the first key that fails, and a pending
	// image fails only on its pixels, so every other value is still written.
	keys := [...]vt.GhosttyKittyGraphicsImageData{
		vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_ID,
		vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_NUMBER,
		vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_WIDTH,
		vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_HEIGHT,
		vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_FORMAT,
		vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_COMPRESSION,
		vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_GENERATION,
		vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_EXPECTED_LEN,
		vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_PIXELS,
	}

	// Each value pointer receives the corresponding field from libghostty.
	values := [...]unsafe.Pointer{
		unsafe.Pointer(&id),
		unsafe.Pointer(&number),
		unsafe.Pointer(&width),
		unsafe.Pointer(&height),
		unsafe.Pointer(&format),
		unsafe.Pointer(&compression),
		unsafe.Pointer(&generation),
		unsafe.Pointer(&expectedLen),
		unsafe.Pointer(&pixels),
	}

	// Copy the output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	if err := args.alloc(values[:]); err != nil {
		return nil, err
	}
	defer args.free()

	tls := getTLS()
	defer putTLS(tls)
	// The keys and the written count live in TLS stack memory.
	const keysSize = len(keys) * 4
	cKeys := tls.Alloc(keysSize + 8)
	defer tls.Free(keysSize + 8)
	*(*[len(keys)]vt.GhosttyKittyGraphicsImageData)(unsafe.Pointer(cKeys)) = keys
	written := cKeys + uintptr(keysSize)
	*(*uint64)(unsafe.Pointer(written)) = 0
	result := vt.Xghostty_kitty_graphics_image_get_multi(
		tls,
		img.ptr,
		uint64(len(keys)),
		cKeys,
		uintptr(unsafe.Pointer(args.values)),
		written,
	)

	// NO_VALUE from the last key means the image is pending. From any other key
	// it would leave fields unset, so it is a real error like any other result.
	pending := result == vt.GHOSTTY_NO_VALUE && int(*(*uint64)(unsafe.Pointer(written))) == len(keys)-1
	if err := resultError(result); err != nil && !pending {
		return nil, err
	}

	info := &KittyGraphicsImageInfo{
		ID:            id,
		Number:        number,
		Width:         width,
		Height:        height,
		Format:        KittyImageFormat(format),
		Compression:   KittyImageCompression(compression),
		Generation:    generation,
		ExpectedLen:   uint(expectedLen),
		PixelsPending: pending,
	}
	if pending {
		return info, nil
	}

	var err error
	if info.Pixels, err = borrowGhosttyBytes(pixels); err != nil {
		return nil, err
	}
	return info, nil
}

// Pixels returns a borrowed slice of the raw pixel data for the current
// animation frame. The slice is only valid until the next mutating terminal
// call. A pending image returns an error with [ResultNoValue]; use
// [KittyGraphicsImage.ExpectedLen] to inspect its expected size. Use
// [KittyGraphicsImage.Generation] to invalidate frame caches.
func (img *KittyGraphicsImage) Pixels() ([]byte, error) {
	tls := getTLS()
	defer putTLS(tls)
	size := int(unsafe.Sizeof(vt.GhosttyString{}))
	out := tls.Alloc(size)
	defer tls.Free(size)
	*(*vt.GhosttyString)(unsafe.Pointer(out)) = vt.GhosttyString{}
	if err := resultError(vt.Xghostty_kitty_graphics_image_get(
		tls,
		img.ptr,
		vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_PIXELS,
		out,
	)); err != nil {
		return nil, err
	}
	return borrowGhosttyBytes(*(*vt.GhosttyString)(unsafe.Pointer(out)))
}

// ExpectedLen returns the size of the image's pixel data in bytes, which is
// width * height * bytes per pixel for the reported format. For a pending image
// this is the size reserved against the storage limit even though
// [KittyGraphicsImage.Pixels] returns [ResultNoValue].
func (img *KittyGraphicsImage) ExpectedLen() (uint, error) {
	b, err := img.get(vt.GHOSTTY_KITTY_GRAPHICS_IMAGE_DATA_EXPECTED_LEN)
	return uint(*(*uint64)(unsafe.Pointer(&b))), err
}

// borrowGhosttyBytes views a borrowed GhosttyString as a Go byte slice without
// copying it. The slice is only valid for as long as the string is. It returns
// nil for an empty string.
func borrowGhosttyBytes(value vt.GhosttyString) ([]byte, error) {
	length, valid := ghosttySizeToInt(value.Flen1)
	if !valid {
		return nil, &Error{Result: ResultLimitExceeded}
	}
	if value.Fptr == 0 || length == 0 {
		return nil, nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(value.Fptr)), length), nil
}

// NewKittyGraphicsPlacementIterator creates a new placement iterator.
// Call KittyGraphics.PlacementIterator to populate it with data, then
// iterate with Next and read fields with the getter methods.
// The iterator must be freed by calling Close.
func NewKittyGraphicsPlacementIterator() (*KittyGraphicsPlacementIterator, error) {
	tls := getTLS()
	defer putTLS(tls)
	out := tls.Alloc(handleSize)
	defer tls.Free(handleSize)
	if err := resultError(vt.Xghostty_kitty_graphics_placement_iterator_new(tls, 0, out)); err != nil {
		return nil, err
	}
	return &KittyGraphicsPlacementIterator{ptr: *(*uintptr)(unsafe.Pointer(out))}, nil
}

// Close frees the placement iterator. After this call, the iterator
// must not be used, except that calling Close again is a safe no-op.
func (it *KittyGraphicsPlacementIterator) Close() {
	if it == nil || it.ptr == 0 {
		return
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_kitty_graphics_placement_iterator_free(tls, it.ptr)
	it.ptr = 0
}

// SetLayers filters the placements returned by Next to the layers in mask,
// which can select more than one layer:
//
//	iter.SetLayers(ImageLayerMaskBelowText | ImageLayerMaskAboveText)
//
// The default is no filter. A mask of zero also means no filter, not "nothing",
// so a mask built by removing layers must not end up empty if it should match
// nothing. A mask with a bit that matches no layer returns an error and keeps
// the current filter.
//
// This filter only applies to the storage iterator. When drawing, use
// [RenderStateImageInfo.Layer] instead. It follows the same rules and also
// covers virtual placements.
func (it *KittyGraphicsPlacementIterator) SetLayers(mask ImageLayerMask) error {
	tls := getTLS()
	defer putTLS(tls)
	v := tls.Alloc(scratchSize)
	defer tls.Free(scratchSize)
	*(*uint32)(unsafe.Pointer(v)) = uint32(mask)
	return resultError(vt.Xghostty_kitty_graphics_placement_iterator_set(
		tls,
		it.ptr,
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_ITERATOR_OPT_LAYER,
		v,
	))
}

// Next advances the iterator to the next placement. Returns true if
// a placement is available, false when iteration is complete.
func (it *KittyGraphicsPlacementIterator) Next() bool {
	tls := getTLS()
	defer putTLS(tls)
	return vt.Xghostty_kitty_graphics_placement_next(tls, it.ptr) != 0
}

// GetMulti queries multiple placement data fields in a single
// call. This is a low-level function. Prefer the typed getters
// (ImageID, PlacementID, Z, etc.) or Info() for normal use. GetMulti
// is useful when you need a custom subset of fields and want to avoid
// per-field call overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/kitty_graphics.h,
// GhosttyKittyGraphicsPlacementData enum).
// Use a Go type with the same size as the C type, such as uint32
// for uint32_t, bool for bool, and int32 for an enum.
//
// GetMulti returns an error if keys and values have different lengths
// or if a key cannot be read. In the second case, values for the keys
// before it may already have been written.
//
// Example:
//
//	var imageID uint32
//	var z int32
//	err := it.GetMulti(
//		[]KittyGraphicsPlacementData{KittyGraphicsPlacementDataImageID, KittyGraphicsPlacementDataZ},
//		[]unsafe.Pointer{unsafe.Pointer(&imageID), unsafe.Pointer(&z)},
//	)
//
// C: ghostty_kitty_graphics_placement_get_multi
func (it *KittyGraphicsPlacementIterator) GetMulti(keys []KittyGraphicsPlacementData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[vt.GhosttyKittyGraphicsPlacementData](&args, keys, values)
	if err != nil {
		return err
	}
	defer args.free()
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_kitty_graphics_placement_get_multi(
		tls,
		it.ptr,
		uint64(len(keys)),
		uintptr(unsafe.Pointer(cKeys)),
		uintptr(unsafe.Pointer(args.values)),
		0,
	))
}

// get reads one value of the current placement, whose C type is listed
// for data on [KittyGraphicsPlacementData], into 8 bytes of TLS stack
// memory and returns them.
func (it *KittyGraphicsPlacementIterator) get(data vt.GhosttyKittyGraphicsPlacementData) ([8]byte, error) {
	return outValue(func(tls *libc.TLS, out uintptr) vt.GhosttyResult {
		return vt.Xghostty_kitty_graphics_placement_get(tls, it.ptr, data, out)
	})
}

// getUint32 reads a uint32_t value of the current placement.
func (it *KittyGraphicsPlacementIterator) getUint32(data vt.GhosttyKittyGraphicsPlacementData) (uint32, error) {
	b, err := it.get(data)
	return *(*uint32)(unsafe.Pointer(&b)), err
}

// ImageID returns the image ID of the current placement.
func (it *KittyGraphicsPlacementIterator) ImageID() (uint32, error) {
	return it.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_IMAGE_ID)
}

// PlacementID returns the placement ID of the current placement.
func (it *KittyGraphicsPlacementIterator) PlacementID() (uint32, error) {
	return it.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_PLACEMENT_ID)
}

// IsVirtual reports whether the current placement is a virtual
// (unicode placeholder) placement.
//
// A virtual placement has no position of its own. It is drawn wherever
// the terminal shows placeholder cells for it. [RenderStateImageIterator]
// finds those cells and positions the image like any other, so renderers
// don't need to check IsVirtual.
func (it *KittyGraphicsPlacementIterator) IsVirtual() (bool, error) {
	b, err := it.get(vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_IS_VIRTUAL)
	return b[0] != 0, err
}

// XOffset returns the pixel offset from the left edge of the cell.
func (it *KittyGraphicsPlacementIterator) XOffset() (uint32, error) {
	return it.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_X_OFFSET)
}

// YOffset returns the pixel offset from the top edge of the cell.
func (it *KittyGraphicsPlacementIterator) YOffset() (uint32, error) {
	return it.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_Y_OFFSET)
}

// SourceX returns the source rectangle x origin in pixels.
func (it *KittyGraphicsPlacementIterator) SourceX() (uint32, error) {
	return it.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_X)
}

// SourceY returns the source rectangle y origin in pixels.
func (it *KittyGraphicsPlacementIterator) SourceY() (uint32, error) {
	return it.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_Y)
}

// SourceWidth returns the source rectangle width in pixels
// (0 = full image width).
func (it *KittyGraphicsPlacementIterator) SourceWidth() (uint32, error) {
	return it.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_WIDTH)
}

// SourceHeight returns the source rectangle height in pixels
// (0 = full image height).
func (it *KittyGraphicsPlacementIterator) SourceHeight() (uint32, error) {
	return it.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_HEIGHT)
}

// Columns returns the number of columns this placement occupies.
func (it *KittyGraphicsPlacementIterator) Columns() (uint32, error) {
	return it.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_COLUMNS)
}

// Rows returns the number of rows this placement occupies.
func (it *KittyGraphicsPlacementIterator) Rows() (uint32, error) {
	return it.getUint32(vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_ROWS)
}

// Z returns the z-index of the current placement.
func (it *KittyGraphicsPlacementIterator) Z() (int32, error) {
	b, err := it.get(vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_Z)
	return *(*int32)(unsafe.Pointer(&b)), err
}

// Info returns all placement metadata in a single call. This is more
// efficient than calling ImageID, PlacementID, IsVirtual, XOffset,
// YOffset, SourceX, SourceY, SourceWidth, SourceHeight, Columns,
// Rows, and Z individually. Uses the get_multi C API to fetch all
// fields in one round-trip.
func (it *KittyGraphicsPlacementIterator) Info() (*KittyGraphicsPlacementInfo, error) {
	// Output variables — one per field, typed to match the C API. The
	// pinner in getMultiArgs keeps them on the heap, where libghostty can
	// write to them.
	var (
		imageID      uint32
		placementID  uint32
		isVirtual    uint8
		xOffset      uint32
		yOffset      uint32
		sourceX      uint32
		sourceY      uint32
		sourceWidth  uint32
		sourceHeight uint32
		columns      uint32
		rows         uint32
		z            int32
	)

	// Keys identify which fields to fetch; order must match values.
	keys := [...]vt.GhosttyKittyGraphicsPlacementData{
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_IMAGE_ID,
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_PLACEMENT_ID,
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_IS_VIRTUAL,
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_X_OFFSET,
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_Y_OFFSET,
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_X,
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_Y,
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_WIDTH,
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_HEIGHT,
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_COLUMNS,
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_ROWS,
		vt.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_Z,
	}

	// Each value pointer receives the corresponding field from libghostty.
	values := [...]unsafe.Pointer{
		unsafe.Pointer(&imageID),
		unsafe.Pointer(&placementID),
		unsafe.Pointer(&isVirtual),
		unsafe.Pointer(&xOffset),
		unsafe.Pointer(&yOffset),
		unsafe.Pointer(&sourceX),
		unsafe.Pointer(&sourceY),
		unsafe.Pointer(&sourceWidth),
		unsafe.Pointer(&sourceHeight),
		unsafe.Pointer(&columns),
		unsafe.Pointer(&rows),
		unsafe.Pointer(&z),
	}

	// Copy the output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	if err := args.alloc(values[:]); err != nil {
		return nil, err
	}
	defer args.free()

	tls := getTLS()
	defer putTLS(tls)
	// The keys live in TLS stack memory.
	const keysSize = len(keys) * 4
	cKeys := tls.Alloc(keysSize)
	defer tls.Free(keysSize)
	*(*[len(keys)]vt.GhosttyKittyGraphicsPlacementData)(unsafe.Pointer(cKeys)) = keys

	if err := resultError(vt.Xghostty_kitty_graphics_placement_get_multi(
		tls,
		it.ptr,
		uint64(len(keys)),
		cKeys,
		uintptr(unsafe.Pointer(args.values)),
		0,
	)); err != nil {
		return nil, err
	}

	return &KittyGraphicsPlacementInfo{
		ImageID:      imageID,
		PlacementID:  placementID,
		IsVirtual:    isVirtual != 0,
		XOffset:      xOffset,
		YOffset:      yOffset,
		SourceX:      sourceX,
		SourceY:      sourceY,
		SourceWidth:  sourceWidth,
		SourceHeight: sourceHeight,
		Columns:      columns,
		Rows:         rows,
		Z:            z,
	}, nil
}

// RenderInfo returns all rendering geometry for the current placement
// in a single call. This combines the results of PixelSize, GridSize,
// ViewportPos, and SourceRect into one call.
//
// Use it to inspect the storage. When drawing, use
// [RenderStateImageIterator] instead. It has the same values for every
// image on screen, including virtual placements and placements relative
// to them. Drawing from both draws each image twice.
//
// RenderInfo succeeds even when the placement isn't in the viewport.
// Check ViewportVisible before using ViewportCol and ViewportRow.
func (it *KittyGraphicsPlacementIterator) RenderInfo(img *KittyGraphicsImage, t *Terminal) (*KittyGraphicsPlacementRenderInfo, error) {
	tls := getTLS()
	defer putTLS(tls)
	size := int(unsafe.Sizeof(vt.GhosttyKittyGraphicsPlacementRenderInfo{}))
	out := tls.Alloc(size)
	defer tls.Free(size)
	// GHOSTTY_INIT_SIZED: a zeroed struct whose size field is set.
	ci := (*vt.GhosttyKittyGraphicsPlacementRenderInfo)(unsafe.Pointer(out))
	*ci = vt.GhosttyKittyGraphicsPlacementRenderInfo{Fsize: uint64(size)}
	if err := resultError(vt.Xghostty_kitty_graphics_placement_render_info(
		tls,
		it.ptr,
		img.ptr,
		t.ptr,
		out,
	)); err != nil {
		return nil, err
	}

	return &KittyGraphicsPlacementRenderInfo{
		PixelWidth:      ci.Fpixel_width,
		PixelHeight:     ci.Fpixel_height,
		GridCols:        ci.Fgrid_cols,
		GridRows:        ci.Fgrid_rows,
		ViewportCol:     ci.Fviewport_col,
		ViewportRow:     ci.Fviewport_row,
		ViewportVisible: ci.Fviewport_visible != 0,
		SourceX:         ci.Fsource_x,
		SourceY:         ci.Fsource_y,
		SourceWidth:     ci.Fsource_width,
		SourceHeight:    ci.Fsource_height,
	}, nil
}

// Rect computes the grid rectangle occupied by the current placement.
// Virtual placements (unicode placeholders) return an error with
// ResultNoValue.
func (it *KittyGraphicsPlacementIterator) Rect(img *KittyGraphicsImage, t *Terminal) (*Selection, error) {
	tls := getTLS()
	defer putTLS(tls)
	size := int(unsafe.Sizeof(vt.GhosttySelection{}))
	out := tls.Alloc(size)
	defer tls.Free(size)
	*(*vt.GhosttySelection)(unsafe.Pointer(out)) = vt.GhosttySelection{}
	if err := resultError(vt.Xghostty_kitty_graphics_placement_rect(
		tls,
		it.ptr,
		img.ptr,
		t.ptr,
		out,
	)); err != nil {
		return nil, err
	}
	sel := selectionFromC(*(*vt.GhosttySelection)(unsafe.Pointer(out)))
	return &sel, nil
}

// pair calls get with two 4-byte output slots in TLS stack memory and
// returns their contents. The geometry getters below read two uint32_t
// or int32_t values this way.
func pair(get func(tls *libc.TLS, a, b uintptr) vt.GhosttyResult) (a, b [4]byte, err error) {
	tls := getTLS()
	defer putTLS(tls)
	out := tls.Alloc(scratchSize)
	defer tls.Free(scratchSize)
	*(*[8]byte)(unsafe.Pointer(out)) = [8]byte{}
	if err := resultError(get(tls, out, out+4)); err != nil {
		return [4]byte{}, [4]byte{}, err
	}
	return *(*[4]byte)(unsafe.Pointer(out)), *(*[4]byte)(unsafe.Pointer(out + 4)), nil
}

// PixelSize computes the rendered pixel dimensions of the current
// placement, accounting for the source rectangle, specified
// columns/rows, and aspect ratio.
func (it *KittyGraphicsPlacementIterator) PixelSize(img *KittyGraphicsImage, t *Terminal) (width, height uint32, err error) {
	w, h, err := pair(func(tls *libc.TLS, a, b uintptr) vt.GhosttyResult {
		return vt.Xghostty_kitty_graphics_placement_pixel_size(tls, it.ptr, img.ptr, t.ptr, a, b)
	})
	if err != nil {
		return 0, 0, err
	}
	return *(*uint32)(unsafe.Pointer(&w)), *(*uint32)(unsafe.Pointer(&h)), nil
}

// GridSize computes the number of grid columns and rows the current
// placement occupies.
func (it *KittyGraphicsPlacementIterator) GridSize(img *KittyGraphicsImage, t *Terminal) (cols, rows uint32, err error) {
	c, r, err := pair(func(tls *libc.TLS, a, b uintptr) vt.GhosttyResult {
		return vt.Xghostty_kitty_graphics_placement_grid_size(tls, it.ptr, img.ptr, t.ptr, a, b)
	})
	if err != nil {
		return 0, 0, err
	}
	return *(*uint32)(unsafe.Pointer(&c)), *(*uint32)(unsafe.Pointer(&r)), nil
}

// ViewportPos returns the column and row of the current placement's
// top-left cell, relative to the viewport. The row is negative when the
// top of the placement has scrolled above the viewport.
//
// ViewportPos returns an error matching [ErrNoValue] when the placement
// is entirely outside the viewport, is a virtual placement, or is
// relative to a virtual placement.
//
// Use it to inspect the storage. When drawing, use
// [RenderStateImageIterator] instead, which positions every image on
// screen. Drawing from both draws each image twice.
func (it *KittyGraphicsPlacementIterator) ViewportPos(img *KittyGraphicsImage, t *Terminal) (col, row int32, err error) {
	c, r, err := pair(func(tls *libc.TLS, a, b uintptr) vt.GhosttyResult {
		return vt.Xghostty_kitty_graphics_placement_viewport_pos(tls, it.ptr, img.ptr, t.ptr, a, b)
	})
	if err != nil {
		return 0, 0, err
	}
	return *(*int32)(unsafe.Pointer(&c)), *(*int32)(unsafe.Pointer(&r)), nil
}

// SourceRect returns the resolved source rectangle for the current
// placement in pixels, clamped to the actual image bounds. A width
// or height of 0 in the placement means "use the full image dimension".
func (it *KittyGraphicsPlacementIterator) SourceRect(img *KittyGraphicsImage) (x, y, width, height uint32, err error) {
	tls := getTLS()
	defer putTLS(tls)
	out := tls.Alloc(scratchSize)
	defer tls.Free(scratchSize)
	v := (*[4]uint32)(unsafe.Pointer(out))
	*v = [4]uint32{}
	if err := resultError(vt.Xghostty_kitty_graphics_placement_source_rect(
		tls,
		it.ptr,
		img.ptr,
		out,
		out+4,
		out+8,
		out+12,
	)); err != nil {
		return 0, 0, 0, 0, err
	}
	return v[0], v[1], v[2], v[3], nil
}
