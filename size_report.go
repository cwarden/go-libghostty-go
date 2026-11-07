package libghostty

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// SizeReportStyle determines the output format for a terminal size report.
// C: GhosttySizeReportStyle
type SizeReportStyle int

const (
	// SizeReportMode2048 is the in-band size report (mode 2048).
	SizeReportMode2048 SizeReportStyle = vt.GHOSTTY_SIZE_REPORT_MODE_2048

	// SizeReportCSI14T is the XTWINOPS text area size in pixels.
	SizeReportCSI14T SizeReportStyle = vt.GHOSTTY_SIZE_REPORT_CSI_14_T

	// SizeReportCSI16T is the XTWINOPS cell size in pixels.
	SizeReportCSI16T SizeReportStyle = vt.GHOSTTY_SIZE_REPORT_CSI_16_T

	// SizeReportCSI18T is the XTWINOPS text area size in characters.
	SizeReportCSI18T SizeReportStyle = vt.GHOSTTY_SIZE_REPORT_CSI_18_T

	// SizeReportCSI19T is the XTWINOPS screen size in characters, sent as
	// ESC [ 9 ; rows ; cols t. The terminal can't know the size of the
	// physical screen, so it reports the same size as [SizeReportCSI18T].
	SizeReportCSI19T SizeReportStyle = vt.GHOSTTY_SIZE_REPORT_CSI_19_T
)

// SizeReportSize holds terminal geometry for XTWINOPS size queries.
// C: GhosttySizeReportSize
type SizeReportSize struct {
	// Rows is the terminal row count in cells.
	Rows uint16

	// Columns is the terminal column count in cells.
	Columns uint16

	// CellWidth is the width of a single terminal cell in pixels.
	CellWidth uint32

	// CellHeight is the height of a single terminal cell in pixels.
	CellHeight uint32
}

// toC converts a Go size to GhosttySizeReportSize.
func (s SizeReportSize) toC() vt.GhosttySizeReportSize {
	return vt.GhosttySizeReportSize{
		Frows:        s.Rows,
		Fcolumns:     s.Columns,
		Fcell_width:  s.CellWidth,
		Fcell_height: s.CellHeight,
	}
}

// SizeReportEncode encodes a terminal size report in the requested style.
func SizeReportEncode(style SizeReportStyle, size SizeReportSize) ([]byte, error) {
	encode := func(tls *libc.TLS, buf uintptr, bufLen uint64, written uintptr) vt.GhosttyResult {
		const n = int(unsafe.Sizeof(vt.GhosttySizeReportSize{}))
		csize := tls.Alloc(n)
		defer tls.Free(n)
		*(*vt.GhosttySizeReportSize)(unsafe.Pointer(csize)) = size.toC()
		return vt.Xghostty_size_report_encode(
			tls,
			vt.GhosttySizeReportStyle(style),
			csize,
			buf,
			bufLen,
			written,
		)
	}

	buf := make([]byte, 64)
	result, written := encodeInto(buf, encode)
	if result == vt.GHOSTTY_SUCCESS {
		return append([]byte(nil), buf[:written]...), nil
	}
	if result != vt.GHOSTTY_OUT_OF_SPACE {
		return nil, &Error{Result: Result(result)}
	}

	out := make([]byte, written)
	result, outWritten := encodeInto(out, encode)
	if err := resultError(result); err != nil {
		return nil, err
	}
	return out[:outWritten], nil
}
