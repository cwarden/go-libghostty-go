package libghostty

// SGR parser bindings from sgr.h.

import (
	"fmt"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// SGRAttributeTag identifies the active value in an [SGRAttribute].
// C: GhosttySgrAttributeTag
type SGRAttributeTag int

const (
	// SGRAttrUnset is the zero-value attribute tag.
	SGRAttrUnset SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_UNSET

	// SGRAttrUnknown identifies an unknown or malformed parameter sequence.
	SGRAttrUnknown SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_UNKNOWN

	// SGRAttrBold enables bold text.
	SGRAttrBold SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_BOLD

	// SGRAttrResetBold disables bold text.
	SGRAttrResetBold SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_RESET_BOLD

	// SGRAttrItalic enables italic text.
	SGRAttrItalic SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_ITALIC

	// SGRAttrResetItalic disables italic text.
	SGRAttrResetItalic SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_RESET_ITALIC

	// SGRAttrFaint enables faint text.
	SGRAttrFaint SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_FAINT

	// SGRAttrUnderline sets the underline style in Underline.
	SGRAttrUnderline SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_UNDERLINE

	// SGRAttrUnderlineColor sets a direct RGB underline color in Color.
	SGRAttrUnderlineColor SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_UNDERLINE_COLOR

	// SGRAttrUnderlineColor256 sets a palette underline color in PaletteIndex.
	SGRAttrUnderlineColor256 SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_UNDERLINE_COLOR_256

	// SGRAttrResetUnderlineColor resets the underline color.
	SGRAttrResetUnderlineColor SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_RESET_UNDERLINE_COLOR

	// SGRAttrOverline enables overline.
	SGRAttrOverline SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_OVERLINE

	// SGRAttrResetOverline disables overline.
	SGRAttrResetOverline SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_RESET_OVERLINE

	// SGRAttrBlink enables blinking text.
	SGRAttrBlink SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_BLINK

	// SGRAttrResetBlink disables blinking text.
	SGRAttrResetBlink SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_RESET_BLINK

	// SGRAttrInverse enables inverse video.
	SGRAttrInverse SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_INVERSE

	// SGRAttrResetInverse disables inverse video.
	SGRAttrResetInverse SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_RESET_INVERSE

	// SGRAttrInvisible enables invisible text.
	SGRAttrInvisible SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_INVISIBLE

	// SGRAttrResetInvisible disables invisible text.
	SGRAttrResetInvisible SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_RESET_INVISIBLE

	// SGRAttrStrikethrough enables strikethrough.
	SGRAttrStrikethrough SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_STRIKETHROUGH

	// SGRAttrResetStrikethrough disables strikethrough.
	SGRAttrResetStrikethrough SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_RESET_STRIKETHROUGH

	// SGRAttrDirectColorFG sets a direct RGB foreground color in Color.
	SGRAttrDirectColorFG SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_DIRECT_COLOR_FG

	// SGRAttrDirectColorBG sets a direct RGB background color in Color.
	SGRAttrDirectColorBG SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_DIRECT_COLOR_BG

	// SGRAttrBG8 sets an eight-color background in PaletteIndex.
	SGRAttrBG8 SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_BG_8

	// SGRAttrFG8 sets an eight-color foreground in PaletteIndex.
	SGRAttrFG8 SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_FG_8

	// SGRAttrResetFG resets the foreground color.
	SGRAttrResetFG SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_RESET_FG

	// SGRAttrResetBG resets the background color.
	SGRAttrResetBG SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_RESET_BG

	// SGRAttrBrightBG8 sets a bright eight-color background in PaletteIndex.
	SGRAttrBrightBG8 SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_BRIGHT_BG_8

	// SGRAttrBrightFG8 sets a bright eight-color foreground in PaletteIndex.
	SGRAttrBrightFG8 SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_BRIGHT_FG_8

	// SGRAttrBG256 sets a 256-color background in PaletteIndex.
	SGRAttrBG256 SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_BG_256

	// SGRAttrFG256 sets a 256-color foreground in PaletteIndex.
	SGRAttrFG256 SGRAttributeTag = vt.GHOSTTY_SGR_ATTR_FG_256
)

// SGRUnderline identifies an SGR underline style.
// C: GhosttySgrUnderline
type SGRUnderline int

// SGRUnknown contains a copied unknown SGR parameter sequence.
// C: GhosttySgrUnknown
type SGRUnknown struct {
	// Full is the complete parameter list supplied to the parser.
	Full []uint16

	// Partial is the portion at which parsing encountered the unknown value.
	Partial []uint16
}

// SGRAttribute is one parsed SGR operation. Callers should inspect Tag and
// then use the corresponding field documented by that tag.
// C: GhosttySgrAttribute
type SGRAttribute struct {
	// Tag identifies the attribute operation and active value.
	Tag SGRAttributeTag

	// Unknown is populated when Tag is SGRAttrUnknown.
	Unknown SGRUnknown

	// Underline is populated when Tag is SGRAttrUnderline.
	Underline SGRUnderline

	// Color is populated for direct RGB foreground, background, or underline
	// color attributes.
	Color ColorRGB

	// PaletteIndex is populated for 8-, 16-, and 256-color attributes.
	PaletteIndex uint8
}

// SGRParser parses CSI SGR parameter lists into semantic attributes.
// C: GhosttySgrParser
type SGRParser struct {
	ptr vt.GhosttySgrParser
}

// NewSGRParser creates a reusable SGR parser.
func NewSGRParser() (*SGRParser, error) {
	tls := getTLS()
	defer putTLS(tls)

	const n = int(unsafe.Sizeof(uintptr(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)
	*(*uintptr)(unsafe.Pointer(out)) = 0
	if err := resultError(vt.Xghostty_sgr_new(tls, 0, out)); err != nil {
		return nil, err
	}
	return &SGRParser{ptr: *(*uintptr)(unsafe.Pointer(out))}, nil
}

// Close frees the parser. Calling Close again is a safe no-op.
func (p *SGRParser) Close() {
	if p == nil || p.ptr == 0 {
		return
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_sgr_free(tls, p.ptr)
	p.ptr = 0
}

// Reset restarts iteration at the beginning of the current parameter list.
func (p *SGRParser) Reset() {
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_sgr_reset(tls, p.ptr)
}

// SetParams copies a CSI SGR parameter list into the parser. separators may
// be nil to treat every separator as a semicolon; otherwise it must contain
// one ';' or ':' byte for every parameter.
func (p *SGRParser) SetParams(params []uint16, separators []byte) error {
	if len(separators) != 0 && len(separators) != len(params) {
		return fmt.Errorf(
			"libghostty: SGR separators length %d does not match params length %d",
			len(separators),
			len(params),
		)
	}
	tls := getTLS()
	defer putTLS(tls)

	// The parser copies both lists, so they are passed from TLS memory for
	// the duration of the call.
	var paramsPtr uintptr
	if len(params) > 0 {
		n := len(params) * 2
		paramsPtr = tls.Alloc(n)
		defer tls.Free(n)
		copy(unsafe.Slice((*uint16)(unsafe.Pointer(paramsPtr)), len(params)), params)
	}
	separatorsPtr, free := cBytes(tls, separators)
	defer free()

	return resultError(vt.Xghostty_sgr_set_params(
		tls,
		p.ptr,
		paramsPtr,
		separatorsPtr,
		uint64(len(params)),
	))
}

// Next returns the next parsed attribute. The boolean is false when the
// current parameter list is exhausted.
func (p *SGRParser) Next() (SGRAttribute, bool) {
	tls := getTLS()
	defer putTLS(tls)

	const n = int(unsafe.Sizeof(vt.GhosttySgrAttribute{}))
	raw := tls.Alloc(n)
	defer tls.Free(n)
	clear(unsafe.Slice((*byte)(unsafe.Pointer(raw)), n))
	if vt.Xghostty_sgr_next(tls, p.ptr, raw) == 0 {
		return SGRAttribute{}, false
	}

	attr := SGRAttribute{
		Tag: SGRAttributeTag(vt.Xghostty_sgr_attribute_tag(tls, raw)),
	}
	value := vt.Xghostty_sgr_attribute_value(tls, raw)

	switch attr.Tag {
	case SGRAttrUnknown:
		unknown := *(*vt.GhosttySgrUnknown)(unsafe.Pointer(value))
		attr.Unknown = sgrUnknownFromC(tls, unknown)

	case SGRAttrUnderline:
		attr.Underline = SGRUnderline(
			*(*vt.GhosttySgrUnderline)(unsafe.Pointer(value)),
		)

	case SGRAttrUnderlineColor, SGRAttrDirectColorFG, SGRAttrDirectColorBG:
		attr.Color = colorRGBFromC(
			*(*vt.GhosttyColorRgb)(unsafe.Pointer(value)),
		)

	case SGRAttrUnderlineColor256,
		SGRAttrBG8,
		SGRAttrFG8,
		SGRAttrBrightBG8,
		SGRAttrBrightFG8,
		SGRAttrBG256,
		SGRAttrFG256:
		attr.PaletteIndex = *(*vt.GhosttyColorPaletteIndex)(unsafe.Pointer(value))
	}

	return attr, true
}

// sgrUnknownFromC copies both borrowed parameter slices from an unknown
// attribute into Go-owned memory.
func sgrUnknownFromC(tls *libc.TLS, unknown vt.GhosttySgrUnknown) SGRUnknown {
	// The out-parameter receives the address of each list.
	const n = int(unsafe.Sizeof(uintptr(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)

	*(*uintptr)(unsafe.Pointer(out)) = 0
	fullLen := vt.Xghostty_sgr_unknown_full(tls, unknown, out)
	fullPtr := *(*uintptr)(unsafe.Pointer(out))
	full := make([]uint16, int(fullLen))
	if fullLen > 0 {
		copy(full, unsafe.Slice((*uint16)(unsafe.Pointer(fullPtr)), int(fullLen)))
	}

	*(*uintptr)(unsafe.Pointer(out)) = 0
	partialLen := vt.Xghostty_sgr_unknown_partial(tls, unknown, out)
	partialPtr := *(*uintptr)(unsafe.Pointer(out))
	partial := make([]uint16, int(partialLen))
	if partialLen > 0 {
		copy(partial, unsafe.Slice((*uint16)(unsafe.Pointer(partialPtr)), int(partialLen)))
	}

	return SGRUnknown{
		Full:    full,
		Partial: partial,
	}
}
