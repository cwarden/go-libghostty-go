package libghostty

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// GridRef is a resolved reference to a specific cell position in the
// terminal's internal page structure. Obtain a GridRef from
// [Terminal.GridRef], then extract cell or row data from it.
//
// A GridRef is a borrowed view into terminal internals, so callers
// must use it under the same serialized access that protects the
// owning terminal. Any later terminal operation may invalidate the
// GridRef, even if it looks unrelated, so read and cache what you
// need immediately. Values returned by its getter methods are copied
// snapshots and may be retained after the GridRef itself becomes
// invalid.
// C: GhosttyGridRef
type GridRef struct {
	ref vt.GhosttyGridRef
}

// withRef copies the grid reference into TLS stack memory for one call
// and returns its address. The call must be made before fn returns.
func (g *GridRef) withRef(tls *libc.TLS, fn func(ref uintptr)) {
	n := int(unsafe.Sizeof(vt.GhosttyGridRef{}))
	p := tls.Alloc(n)
	defer tls.Free(n)
	*(*vt.GhosttyGridRef)(unsafe.Pointer(p)) = g.ref
	fn(p)
}

// Cell returns the cell at the grid reference's position.
func (g *GridRef) Cell() (*Cell, error) {
	tls := getTLS()
	defer putTLS(tls)
	n := int(unsafe.Sizeof(vt.GhosttyCell(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)

	var err error
	g.withRef(tls, func(ref uintptr) {
		err = resultError(vt.Xghostty_grid_ref_cell(tls, ref, out))
	})
	if err != nil {
		return nil, err
	}
	return &Cell{c: *(*vt.GhosttyCell)(unsafe.Pointer(out))}, nil
}

// Row returns the row at the grid reference's position.
func (g *GridRef) Row() (*Row, error) {
	tls := getTLS()
	defer putTLS(tls)
	n := int(unsafe.Sizeof(vt.GhosttyRow(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)

	var err error
	g.withRef(tls, func(ref uintptr) {
		err = resultError(vt.Xghostty_grid_ref_row(tls, ref, out))
	})
	if err != nil {
		return nil, err
	}
	return &Row{c: *(*vt.GhosttyRow)(unsafe.Pointer(out))}, nil
}

// Graphemes returns the full grapheme cluster codepoints for the cell
// at the grid reference's position. Returns nil if the cell has no text.
func (g *GridRef) Graphemes() ([]uint32, error) {
	tls := getTLS()
	defer putTLS(tls)
	const lenSize = int(unsafe.Sizeof(uint64(0)))
	outLen := tls.Alloc(lenSize)
	defer tls.Free(lenSize)
	*(*uint64)(unsafe.Pointer(outLen)) = 0

	// First call to get the required length.
	var err error
	g.withRef(tls, func(ref uintptr) {
		err = resultError(vt.Xghostty_grid_ref_graphemes(tls, ref, 0, 0, outLen))
	})
	if err != nil {
		// OUT_OF_SPACE means we need a bigger buffer; outLen has the size.
		ge, ok := err.(*Error)
		if !ok || ge.Result != ResultOutOfSpace {
			return nil, err
		}
	}

	count := *(*uint64)(unsafe.Pointer(outLen))
	if count == 0 {
		return nil, nil
	}

	// The output buffer is C memory: the translated code writes into it
	// while the goroutine stack may move.
	size := uintptr(count) * unsafe.Sizeof(uint32(0))
	buf := Alloc(size)
	if buf == nil {
		return nil, &Error{Result: ResultOutOfMemory}
	}
	defer Free(buf, size)

	g.withRef(tls, func(ref uintptr) {
		err = resultError(vt.Xghostty_grid_ref_graphemes(tls, ref, uintptr(buf), count, outLen))
	})
	if err != nil {
		return nil, err
	}

	written := *(*uint64)(unsafe.Pointer(outLen))
	return append([]uint32(nil), unsafe.Slice((*uint32)(buf), int(count))[:written]...), nil
}

// HyperlinkURI returns the URI for the hyperlink at the grid reference's
// position. It returns an empty string when the cell has no hyperlink.
func (g *GridRef) HyperlinkURI() (string, error) {
	tls := getTLS()
	defer putTLS(tls)
	const lenSize = int(unsafe.Sizeof(uint64(0)))
	outLen := tls.Alloc(lenSize)
	defer tls.Free(lenSize)
	*(*uint64)(unsafe.Pointer(outLen)) = 0

	// First call to get the required length. A cell with no hyperlink returns
	// success with a length of zero; a hyperlinked cell returns OUT_OF_SPACE
	// and reports the buffer size needed for the URI bytes.
	var err error
	g.withRef(tls, func(ref uintptr) {
		err = resultError(vt.Xghostty_grid_ref_hyperlink_uri(tls, ref, 0, 0, outLen))
	})
	if err != nil {
		ge, ok := err.(*Error)
		if !ok || ge.Result != ResultOutOfSpace {
			return "", err
		}
	}

	count := *(*uint64)(unsafe.Pointer(outLen))
	if count == 0 {
		return "", nil
	}

	buf := Alloc(uintptr(count))
	if buf == nil {
		return "", &Error{Result: ResultOutOfMemory}
	}
	defer Free(buf, uintptr(count))

	g.withRef(tls, func(ref uintptr) {
		err = resultError(vt.Xghostty_grid_ref_hyperlink_uri(tls, ref, uintptr(buf), count, outLen))
	})
	if err != nil {
		return "", err
	}

	written := *(*uint64)(unsafe.Pointer(outLen))
	return string(goBytes(uintptr(buf), int(written))), nil
}

// Style returns the style of the cell at the grid reference's position.
func (g *GridRef) Style() (*Style, error) {
	tls := getTLS()
	defer putTLS(tls)
	n := int(unsafe.Sizeof(vt.GhosttyStyle{}))
	out := tls.Alloc(n)
	defer tls.Free(n)

	var err error
	g.withRef(tls, func(ref uintptr) {
		err = resultError(vt.Xghostty_grid_ref_style(tls, ref, out))
	})
	if err != nil {
		return nil, err
	}
	return &Style{c: *(*vt.GhosttyStyle)(unsafe.Pointer(out))}, nil
}
