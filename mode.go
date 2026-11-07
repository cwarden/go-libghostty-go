package libghostty

import (
	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// Mode is a packed 16-bit terminal mode identifier. It encodes a mode
// value (bits 0–14) and an ANSI flag (bit 15). DEC private modes have
// the ANSI bit clear; standard ANSI modes have it set.
// C: GhosttyMode
type Mode uint16

// ghostty_mode_new, ghostty_mode_value and ghostty_mode_ansi are static
// inline functions of modes.h, so the library does not export them and
// the packing is done here.

// modeValueMask selects the mode value; the remaining bit is the ANSI flag.
const (
	modeValueMask Mode = 0x7FFF
	modeANSIBit   Mode = 1 << 15
)

// NewMode packs a numeric mode value and ANSI flag into a Mode. Values are
// limited to 15 bits, matching GhosttyMode.
// C: ghostty_mode_new
func NewMode(value uint16, ansi bool) Mode {
	m := Mode(value) & modeValueMask
	if ansi {
		m |= modeANSIBit
	}
	return m
}

// Value returns the numeric mode value (0–32767).
// C: ghostty_mode_value
func (m Mode) Value() uint16 {
	return uint16(m & modeValueMask)
}

// ANSI reports whether this is a standard ANSI mode. If false, it is
// a DEC private mode (?-prefixed).
// C: ghostty_mode_ansi
func (m Mode) ANSI() bool {
	return m&modeANSIBit != 0
}

// The GHOSTTY_MODE_* macros of modes.h expand to ghostty_mode_new calls
// with the mode numbers below, so the Go values are built the same way.

// ANSI modes.
var (
	// ModeKAM is keyboard action mode (disable keyboard).
	ModeKAM = NewMode(2, true)

	// ModeInsert is insert mode.
	ModeInsert = NewMode(4, true)

	// ModeSRM is send/receive mode.
	ModeSRM = NewMode(12, true)

	// ModeLinefeed is linefeed/new line mode.
	ModeLinefeed = NewMode(20, true)
)

// DEC private modes.
var (
	// ModeDECCKM is cursor keys mode.
	ModeDECCKM = NewMode(1, false)

	// Mode132Column is 132/80 column mode.
	Mode132Column = NewMode(3, false)

	// ModeSlowScroll is slow scroll mode.
	ModeSlowScroll = NewMode(4, false)

	// ModeReverseColors is reverse video mode.
	ModeReverseColors = NewMode(5, false)

	// ModeOrigin is origin mode.
	ModeOrigin = NewMode(6, false)

	// ModeWraparound is auto-wrap mode.
	ModeWraparound = NewMode(7, false)

	// ModeAutorepeat is auto-repeat keys mode.
	ModeAutorepeat = NewMode(8, false)

	// ModeX10Mouse is X10 mouse reporting mode.
	ModeX10Mouse = NewMode(9, false)

	// ModeCursorBlinking is cursor blink mode.
	ModeCursorBlinking = NewMode(12, false)

	// ModeCursorVisible is cursor visible mode (DECTCEM).
	ModeCursorVisible = NewMode(25, false)

	// ModeEnableMode3 allows 132 column mode.
	ModeEnableMode3 = NewMode(40, false)

	// ModeReverseWrap is reverse wrap mode.
	ModeReverseWrap = NewMode(45, false)

	// ModeAltScreenLegacy is alternate screen (legacy, mode 47).
	ModeAltScreenLegacy = NewMode(47, false)

	// ModeKeypadKeys is application keypad mode.
	ModeKeypadKeys = NewMode(66, false)

	// ModeBackarrowKeyMode is backarrow key mode (DECBKM).
	ModeBackarrowKeyMode = NewMode(67, false)

	// ModeLeftRightMargin is left/right margin mode.
	ModeLeftRightMargin = NewMode(69, false)

	// ModeNormalMouse is normal mouse tracking mode.
	ModeNormalMouse = NewMode(1000, false)

	// ModeButtonMouse is button-event mouse tracking mode.
	ModeButtonMouse = NewMode(1002, false)

	// ModeAnyMouse is any-event mouse tracking mode.
	ModeAnyMouse = NewMode(1003, false)

	// ModeFocusEvent enables focus in/out events.
	ModeFocusEvent = NewMode(1004, false)

	// ModeUTF8Mouse is UTF-8 mouse format mode.
	ModeUTF8Mouse = NewMode(1005, false)

	// ModeSGRMouse is SGR mouse format mode.
	ModeSGRMouse = NewMode(1006, false)

	// ModeAltScroll is alternate scroll mode.
	ModeAltScroll = NewMode(1007, false)

	// ModeURxvtMouse is URxvt mouse format mode.
	ModeURxvtMouse = NewMode(1015, false)

	// ModeSGRPixelsMouse is SGR-Pixels mouse format mode.
	ModeSGRPixelsMouse = NewMode(1016, false)

	// ModeNumlockKeypad ignores keypad with NumLock.
	ModeNumlockKeypad = NewMode(1035, false)

	// ModeAltEscPrefix makes Alt key send ESC prefix.
	ModeAltEscPrefix = NewMode(1036, false)

	// ModeAltSendsEsc makes Alt send escape.
	ModeAltSendsEsc = NewMode(1039, false)

	// ModeReverseWrapExt is extended reverse wrap mode.
	ModeReverseWrapExt = NewMode(1045, false)

	// ModeAltScreen is alternate screen mode (mode 1047).
	ModeAltScreen = NewMode(1047, false)

	// ModeSaveCursor saves cursor position (DECSC, mode 1048).
	ModeSaveCursor = NewMode(1048, false)

	// ModeAltScreenSave is alt screen + save cursor + clear (mode 1049).
	ModeAltScreenSave = NewMode(1049, false)

	// ModeBracketedPaste is bracketed paste mode.
	ModeBracketedPaste = NewMode(2004, false)

	// ModeSyncOutput is synchronized output mode.
	ModeSyncOutput = NewMode(2026, false)

	// ModeGraphemeCluster is grapheme cluster mode.
	ModeGraphemeCluster = NewMode(2027, false)

	// ModeColorSchemeReport enables color scheme reporting.
	ModeColorSchemeReport = NewMode(2031, false)

	// ModeVisibilityReport enables terminal visibility reporting.
	ModeVisibilityReport = NewMode(2033, false)

	// ModeInBandResize enables in-band size reports.
	ModeInBandResize = NewMode(2048, false)

	// ModePasteEvents enables Kitty clipboard protocol paste events.
	ModePasteEvents = NewMode(5522, false)
)

// ModeReportState represents DECRPM report state values (Ps2 parameter).
// C: GhosttyModeReportState
type ModeReportState int

const (
	// ModeReportNotRecognized means the mode is not recognized.
	ModeReportNotRecognized ModeReportState = vt.GHOSTTY_MODE_REPORT_NOT_RECOGNIZED

	// ModeReportSet means the mode is set (enabled).
	ModeReportSet ModeReportState = vt.GHOSTTY_MODE_REPORT_SET

	// ModeReportReset means the mode is reset (disabled).
	ModeReportReset ModeReportState = vt.GHOSTTY_MODE_REPORT_RESET

	// ModeReportPermanentlySet means the mode is permanently set.
	ModeReportPermanentlySet ModeReportState = vt.GHOSTTY_MODE_REPORT_PERMANENTLY_SET

	// ModeReportPermanentlyReset means the mode is permanently reset.
	ModeReportPermanentlyReset ModeReportState = vt.GHOSTTY_MODE_REPORT_PERMANENTLY_RESET
)

// ModeReportEncode encodes a DECRPM response for mode and state.
func ModeReportEncode(mode Mode, state ModeReportState) ([]byte, error) {
	encode := func(tls *libc.TLS, buf uintptr, bufLen uint64, written uintptr) vt.GhosttyResult {
		return vt.Xghostty_mode_report_encode(
			tls,
			uint16(mode),
			vt.GhosttyModeReportState(state),
			buf,
			bufLen,
			written,
		)
	}

	buf := make([]byte, 32)
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
