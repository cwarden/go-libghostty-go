package libghostty

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// ColorRGB represents an RGB color value.
// C: GhosttyColorRgb
type ColorRGB struct {
	// R is the red component.
	R uint8

	// G is the green component.
	G uint8

	// B is the blue component.
	B uint8
}

// PaletteSize is the number of entries in a terminal color palette.
const PaletteSize = 256

// Palette is a 256-color palette.
type Palette [PaletteSize]ColorRGB

// ColorPaletteMask identifies palette entries that palette generation should
// preserve from its base palette.
// C: GhosttyColorPaletteMask
type ColorPaletteMask struct {
	// Bits stores the 256-bit mask as four consecutive 64-bit words.
	Bits [4]uint64
}

// Set marks index as preserved during palette generation.
func (m *ColorPaletteMask) Set(index uint8) {
	m.Bits[index>>6] |= uint64(1) << (index & 63)
}

// Unset clears index from the set preserved during palette generation.
func (m *ColorPaletteMask) Unset(index uint8) {
	m.Bits[index>>6] &^= uint64(1) << (index & 63)
}

// IsSet reports whether index is preserved during palette generation.
func (m *ColorPaletteMask) IsSet(index uint8) bool {
	return m.Bits[index>>6]&(uint64(1)<<(index&63)) != 0
}

// X11Color is one entry in Ghostty's embedded X11 rgb.txt table.
// C: GhosttyColorX11Entry
type X11Color struct {
	// Name is the exact spelling stored in the X11 color table.
	Name string

	// Color is the RGB value associated with Name.
	Color ColorRGB
}

// Named color palette indices.
// C: GHOSTTY_COLOR_NAMED_*
const (
	ColorNamedBlack         = vt.GHOSTTY_COLOR_NAMED_BLACK
	ColorNamedRed           = vt.GHOSTTY_COLOR_NAMED_RED
	ColorNamedGreen         = vt.GHOSTTY_COLOR_NAMED_GREEN
	ColorNamedYellow        = vt.GHOSTTY_COLOR_NAMED_YELLOW
	ColorNamedBlue          = vt.GHOSTTY_COLOR_NAMED_BLUE
	ColorNamedMagenta       = vt.GHOSTTY_COLOR_NAMED_MAGENTA
	ColorNamedCyan          = vt.GHOSTTY_COLOR_NAMED_CYAN
	ColorNamedWhite         = vt.GHOSTTY_COLOR_NAMED_WHITE
	ColorNamedBrightBlack   = vt.GHOSTTY_COLOR_NAMED_BRIGHT_BLACK
	ColorNamedBrightRed     = vt.GHOSTTY_COLOR_NAMED_BRIGHT_RED
	ColorNamedBrightGreen   = vt.GHOSTTY_COLOR_NAMED_BRIGHT_GREEN
	ColorNamedBrightYellow  = vt.GHOSTTY_COLOR_NAMED_BRIGHT_YELLOW
	ColorNamedBrightBlue    = vt.GHOSTTY_COLOR_NAMED_BRIGHT_BLUE
	ColorNamedBrightMagenta = vt.GHOSTTY_COLOR_NAMED_BRIGHT_MAGENTA
	ColorNamedBrightCyan    = vt.GHOSTTY_COLOR_NAMED_BRIGHT_CYAN
	ColorNamedBrightWhite   = vt.GHOSTTY_COLOR_NAMED_BRIGHT_WHITE
)

// Sizes of the C structures this file places in TLS stack memory.
const (
	colorRGBSize     = int(unsafe.Sizeof(vt.GhosttyColorRgb{}))
	colorPaletteSize = PaletteSize * colorRGBSize
	colorMaskSize    = int(unsafe.Sizeof(vt.GhosttyColorPaletteMask{}))
)

// Components returns the red, green, and blue components through
// ghostty_color_rgb_get. The fields on ColorRGB expose the same values
// directly; this method is useful for consumers mirroring the C API.
func (c ColorRGB) Components() (r, g, b uint8) {
	tls := getTLS()
	defer putTLS(tls)

	// One block holds the input color and the three output bytes.
	const n = colorRGBSize + 3
	p := tls.Alloc(n)
	defer tls.Free(n)
	*(*vt.GhosttyColorRgb)(unsafe.Pointer(p)) = c.toC()
	out := p + uintptr(colorRGBSize)
	vt.Xghostty_color_rgb_get(tls, p, out, out+1, out+2)
	bytes := unsafe.Slice((*byte)(unsafe.Pointer(out)), 3)
	return bytes[0], bytes[1], bytes[2]
}

// ParseColor parses a color using Ghostty's config and theme syntax.
func ParseColor(value string) (ColorRGB, error) {
	tls := getTLS()
	defer putTLS(tls)

	data, free := cBytes(tls, []byte(value))
	defer free()
	out := tls.Alloc(colorRGBSize)
	defer tls.Free(colorRGBSize)
	if err := resultError(vt.Xghostty_color_parse(
		tls,
		data,
		uint64(len(value)),
		out,
	)); err != nil {
		return ColorRGB{}, err
	}
	return colorRGBFromC(*(*vt.GhosttyColorRgb)(unsafe.Pointer(out))), nil
}

// ParseX11Color parses an X11 color name using Ghostty's embedded rgb.txt
// table. Matching is ASCII case-insensitive.
func ParseX11Color(name string) (ColorRGB, error) {
	tls := getTLS()
	defer putTLS(tls)

	data, free := cBytes(tls, []byte(name))
	defer free()
	out := tls.Alloc(colorRGBSize)
	defer tls.Free(colorRGBSize)
	if err := resultError(vt.Xghostty_color_parse_x11(
		tls,
		data,
		uint64(len(name)),
		out,
	)); err != nil {
		return ColorRGB{}, err
	}
	return colorRGBFromC(*(*vt.GhosttyColorRgb)(unsafe.Pointer(out))), nil
}

// ParsePaletteEntry parses a Ghostty palette override in INDEX=COLOR form.
func ParsePaletteEntry(value string) (uint8, ColorRGB, error) {
	tls := getTLS()
	defer putTLS(tls)

	data, free := cBytes(tls, []byte(value))
	defer free()
	// One block holds the output color followed by the output index.
	const n = colorRGBSize + 1
	p := tls.Alloc(n)
	defer tls.Free(n)
	index := p + uintptr(colorRGBSize)
	if err := resultError(vt.Xghostty_color_parse_palette_entry(
		tls,
		data,
		uint64(len(value)),
		index,
		p,
	)); err != nil {
		return 0, ColorRGB{}, err
	}
	return *(*uint8)(unsafe.Pointer(index)), colorRGBFromC(*(*vt.GhosttyColorRgb)(unsafe.Pointer(p))), nil
}

// DefaultPalette returns Ghostty's built-in 256-color palette.
func DefaultPalette() Palette {
	tls := getTLS()
	defer putTLS(tls)

	out := tls.Alloc(colorPaletteSize)
	defer tls.Free(colorPaletteSize)
	vt.Xghostty_color_palette_default(tls, out)
	return paletteFromC(out)
}

// GeneratePalette derives the 216-color cube and grayscale ramp from a base
// palette, background, and foreground. A nil base uses Ghostty's default
// palette. A nil skip mask preserves no additional indices.
func GeneratePalette(
	base *Palette,
	skip *ColorPaletteMask,
	background ColorRGB,
	foreground ColorRGB,
	harmonious bool,
) Palette {
	tls := getTLS()
	defer putTLS(tls)

	// One block holds the base palette, the skip mask, the background and
	// foreground colors, and the output palette, in that order.
	const n = colorPaletteSize + colorMaskSize + 2*colorRGBSize + colorPaletteSize
	p := tls.Alloc(n)
	defer tls.Free(n)
	cbase := p
	cskip := cbase + uintptr(colorPaletteSize)
	bg := cskip + uintptr(colorMaskSize)
	fg := bg + uintptr(colorRGBSize)
	out := fg + uintptr(colorRGBSize)

	var basePtr uintptr
	if base != nil {
		paletteToC(base, cbase)
		basePtr = cbase
	}

	var skipPtr uintptr
	if skip != nil {
		mask := (*vt.GhosttyColorPaletteMask)(unsafe.Pointer(cskip))
		for i, bits := range skip.Bits {
			mask.Fbits.Farray[i] = bits
		}
		skipPtr = cskip
	}

	*(*vt.GhosttyColorRgb)(unsafe.Pointer(bg)) = background.toC()
	*(*vt.GhosttyColorRgb)(unsafe.Pointer(fg)) = foreground.toC()
	vt.Xghostty_color_palette_generate(
		tls,
		basePtr,
		skipPtr,
		bg,
		fg,
		cBool(harmonious),
		out,
	)
	return paletteFromC(out)
}

// Luminance returns the W3C relative luminance in the range 0 to 1.
func (c ColorRGB) Luminance() float64 {
	tls := getTLS()
	defer putTLS(tls)

	p := tls.Alloc(colorRGBSize)
	defer tls.Free(colorRGBSize)
	*(*vt.GhosttyColorRgb)(unsafe.Pointer(p)) = c.toC()
	return float64(vt.Xghostty_color_luminance(tls, p))
}

// PerceivedLuminance returns Ghostty's perceived luminance in the range 0
// to 1. Ghostty treats backgrounds above 0.5 as light.
func (c ColorRGB) PerceivedLuminance() float64 {
	tls := getTLS()
	defer putTLS(tls)

	p := tls.Alloc(colorRGBSize)
	defer tls.Free(colorRGBSize)
	*(*vt.GhosttyColorRgb)(unsafe.Pointer(p)) = c.toC()
	return float64(vt.Xghostty_color_perceived_luminance(tls, p))
}

// Contrast returns the WCAG contrast ratio between c and other.
func (c ColorRGB) Contrast(other ColorRGB) float64 {
	tls := getTLS()
	defer putTLS(tls)

	const n = 2 * colorRGBSize
	p := tls.Alloc(n)
	defer tls.Free(n)
	a := p
	b := p + uintptr(colorRGBSize)
	*(*vt.GhosttyColorRgb)(unsafe.Pointer(a)) = c.toC()
	*(*vt.GhosttyColorRgb)(unsafe.Pointer(b)) = other.toC()
	return float64(vt.Xghostty_color_contrast(tls, a, b))
}

// X11Colors returns a Go-owned copy of Ghostty's X11 color name table in
// rgb.txt order.
func X11Colors() []X11Color {
	tls := getTLS()
	defer putTLS(tls)

	count := int(vt.Xghostty_color_x11_name_count(tls))
	entries := unsafe.Slice((*vt.GhosttyColorX11Entry)(unsafe.Pointer(vt.Xghostty_color_x11_names(tls))), count)
	result := make([]X11Color, count)
	for i, entry := range entries {
		result[i] = X11Color{
			Name:  libc.GoString(entry.Fname),
			Color: colorRGBFromC(entry.Fcolor),
		}
	}
	return result
}

// toC converts a Go RGB value to GhosttyColorRgb.
func (c ColorRGB) toC() vt.GhosttyColorRgb {
	return vt.GhosttyColorRgb{
		Fr: c.R,
		Fg: c.G,
		Fb: c.B,
	}
}

// colorRGBFromC converts GhosttyColorRgb to its Go representation.
func colorRGBFromC(c vt.GhosttyColorRgb) ColorRGB {
	return ColorRGB{R: c.Fr, G: c.Fg, B: c.Fb}
}

// paletteToC copies a Go palette into the C palette storage at dst, which
// holds PaletteSize GhosttyColorRgb values.
func paletteToC(src *Palette, dst uintptr) {
	entries := unsafe.Slice((*vt.GhosttyColorRgb)(unsafe.Pointer(dst)), PaletteSize)
	for i, color := range src {
		entries[i] = color.toC()
	}
}

// paletteFromC copies the C palette storage at src, which holds PaletteSize
// GhosttyColorRgb values, into a Go palette.
func paletteFromC(src uintptr) Palette {
	var result Palette
	for i, color := range unsafe.Slice((*vt.GhosttyColorRgb)(unsafe.Pointer(src)), PaletteSize) {
		result[i] = colorRGBFromC(color)
	}
	return result
}

// cBytes copies data into TLS stack memory for the duration of a call and
// returns its address and the function that frees it. An empty slice gives
// address 0 and frees nothing. Unlike pinBytes it also serves data that
// was derived from a string.
func cBytes(tls *libc.TLS, data []byte) (addr uintptr, free func()) {
	if len(data) == 0 {
		return 0, func() {}
	}
	p := tls.Alloc(len(data))
	copy(unsafe.Slice((*byte)(unsafe.Pointer(p)), len(data)), data)
	return p, func() { tls.Free(len(data)) }
}
