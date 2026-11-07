package libghostty

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// StyleColorTag identifies the type of color in a style color.
// C: GhosttyStyleColorTag
type StyleColorTag int

const (
	// StyleColorNone means no color is set.
	StyleColorNone StyleColorTag = vt.GHOSTTY_STYLE_COLOR_NONE

	// StyleColorPalette means the color is a palette index.
	StyleColorPalette StyleColorTag = vt.GHOSTTY_STYLE_COLOR_PALETTE

	// StyleColorRGB means the color is a direct RGB value.
	StyleColorRGB StyleColorTag = vt.GHOSTTY_STYLE_COLOR_RGB
)

// StyleColor is a tagged union representing a color in a style attribute.
// Check Tag to determine which field is valid.
// C: GhosttyStyleColor
type StyleColor struct {
	// Tag identifies the type of color.
	Tag StyleColorTag

	// Palette is the palette index (valid when Tag is StyleColorPalette).
	Palette uint8

	// RGB is the direct RGB color (valid when Tag is StyleColorRGB).
	RGB ColorRGB
}

// SGR underline styles.
// C: GhosttySgrUnderline
const (
	// UnderlineNone disables underlining.
	UnderlineNone SGRUnderline = vt.GHOSTTY_SGR_UNDERLINE_NONE

	// UnderlineSingle draws a single underline.
	UnderlineSingle SGRUnderline = vt.GHOSTTY_SGR_UNDERLINE_SINGLE

	// UnderlineDouble draws a double underline.
	UnderlineDouble SGRUnderline = vt.GHOSTTY_SGR_UNDERLINE_DOUBLE

	// UnderlineCurly draws a curly underline.
	UnderlineCurly SGRUnderline = vt.GHOSTTY_SGR_UNDERLINE_CURLY

	// UnderlineDotted draws a dotted underline.
	UnderlineDotted SGRUnderline = vt.GHOSTTY_SGR_UNDERLINE_DOTTED

	// UnderlineDashed draws a dashed underline.
	UnderlineDashed SGRUnderline = vt.GHOSTTY_SGR_UNDERLINE_DASHED
)

// The GHOSTTY_STYLE_FLAG_* and GHOSTTY_STYLE_UNDERLINE_* macros of style.h,
// which the generated constants do not include because they are macros
// rather than enumerators.
const (
	styleFlagBold          vt.GhosttyStyleFlags = 1 << 0
	styleFlagItalic        vt.GhosttyStyleFlags = 1 << 1
	styleFlagFaint         vt.GhosttyStyleFlags = 1 << 2
	styleFlagBlink         vt.GhosttyStyleFlags = 1 << 3
	styleFlagInverse       vt.GhosttyStyleFlags = 1 << 4
	styleFlagInvisible     vt.GhosttyStyleFlags = 1 << 5
	styleFlagStrikethrough vt.GhosttyStyleFlags = 1 << 6
	styleFlagOverline      vt.GhosttyStyleFlags = 1 << 7
	styleUnderlineShift                         = 8
	styleUnderlineMask     vt.GhosttyStyleFlags = 0x7 << styleUnderlineShift
)

// styleSize is the size of GhosttyStyle, for TLS stack allocations.
const styleSize = int(unsafe.Sizeof(vt.GhosttyStyle{}))

// Style is a thin wrapper around the copied C GhosttyStyle value. It
// provides getter methods to access individual style attributes
// without copying the entire struct upfront. A Style is a value
// snapshot and may be retained after the terminal, [GridRef], or
// render-state iterator that produced it becomes invalid.
// C: GhosttyStyle
type Style struct {
	c vt.GhosttyStyle
}

// DefaultStyle returns a new style initialized to Ghostty's default values.
func DefaultStyle() *Style {
	tls := getTLS()
	defer putTLS(tls)

	p := tls.Alloc(styleSize)
	defer tls.Free(styleSize)
	vt.Xghostty_style_default(tls, p)
	return &Style{c: *(*vt.GhosttyStyle)(unsafe.Pointer(p))}
}

// IsDefault reports whether the style is the default style
// (no colors, no flags).
func (s *Style) IsDefault() bool {
	tls := getTLS()
	defer putTLS(tls)

	p := tls.Alloc(styleSize)
	defer tls.Free(styleSize)
	*(*vt.GhosttyStyle)(unsafe.Pointer(p)) = s.c
	return vt.Xghostty_style_is_default(tls, p) != 0
}

// FgColor returns the foreground color.
func (s *Style) FgColor() StyleColor {
	return styleColorFromC(s.c.Ffg_color)
}

// BgColor returns the background color.
func (s *Style) BgColor() StyleColor {
	return styleColorFromC(s.c.Fbg_color)
}

// UnderlineColor returns the underline color.
func (s *Style) UnderlineColor() StyleColor {
	return styleColorFromC(s.c.Funderline_color)
}

// hasFlag reports whether the on/off attribute flag is set in the style's
// packed flags.
func (s *Style) hasFlag(flag vt.GhosttyStyleFlags) bool {
	return s.c.Fflags&flag != 0
}

// Bold reports whether bold is set.
func (s *Style) Bold() bool {
	return s.hasFlag(styleFlagBold)
}

// Italic reports whether italic is set.
func (s *Style) Italic() bool {
	return s.hasFlag(styleFlagItalic)
}

// Faint reports whether faint (dim) is set.
func (s *Style) Faint() bool {
	return s.hasFlag(styleFlagFaint)
}

// Blink reports whether blink is set.
func (s *Style) Blink() bool {
	return s.hasFlag(styleFlagBlink)
}

// Inverse reports whether inverse video is set.
func (s *Style) Inverse() bool {
	return s.hasFlag(styleFlagInverse)
}

// Invisible reports whether invisible is set.
func (s *Style) Invisible() bool {
	return s.hasFlag(styleFlagInvisible)
}

// Strikethrough reports whether strikethrough is set.
func (s *Style) Strikethrough() bool {
	return s.hasFlag(styleFlagStrikethrough)
}

// Overline reports whether overline is set.
func (s *Style) Overline() bool {
	return s.hasFlag(styleFlagOverline)
}

// Underline returns the underline style (one of the Underline* constants).
func (s *Style) Underline() SGRUnderline {
	// This is the masking that the GHOSTTY_STYLE_UNDERLINE macro does.
	return SGRUnderline((s.c.Fflags & styleUnderlineMask) >> styleUnderlineShift)
}

// styleColorFromC converts a C GhosttyStyleColor to a Go StyleColor.
func styleColorFromC(c vt.GhosttyStyleColor) StyleColor {
	sc := StyleColor{Tag: StyleColorTag(c.Ftag)}
	switch sc.Tag {
	case StyleColorPalette:
		sc.Palette = *(*vt.GhosttyColorPaletteIndex)(unsafe.Pointer(&c.Fvalue))
	case StyleColorRGB:
		rgb := *(*vt.GhosttyColorRgb)(unsafe.Pointer(&c.Fvalue))
		sc.RGB = ColorRGB{R: rgb.Fr, G: rgb.Fg, B: rgb.Fb}
	}
	return sc
}
