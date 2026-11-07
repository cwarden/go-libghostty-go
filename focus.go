package libghostty

// Focus encoding — encode focus in/out events into terminal escape
// sequences (CSI I / CSI O) for focus reporting mode (mode 1004).
// Wraps the C API from focus.h.

import (
	"encoding"
	"fmt"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// Compile-time assertions that FocusEvent implements the standard
// text marshaling interfaces.
var (
	_ encoding.TextMarshaler   = FocusEvent(0)
	_ encoding.TextUnmarshaler = (*FocusEvent)(nil)
)

// FocusEvent represents a focus gained or lost event for focus
// reporting mode (mode 1004).
//
// C: GhosttyFocusEvent
type FocusEvent int

const (
	// FocusGained indicates the terminal window gained focus.
	FocusGained FocusEvent = vt.GHOSTTY_FOCUS_GAINED

	// FocusLost indicates the terminal window lost focus.
	FocusLost FocusEvent = vt.GHOSTTY_FOCUS_LOST
)

// String returns a human-friendly name for the focus event:
// "gained" or "lost". Unknown values render as "unknown".
func (f FocusEvent) String() string {
	switch f {
	case FocusGained:
		return "gained"
	case FocusLost:
		return "lost"
	default:
		return "unknown"
	}
}

// MarshalText implements encoding.TextMarshaler. The output is the
// same as String() so the type integrates with encoding/json and
// other text-based encoders.
func (f FocusEvent) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler. It parses a
// focus event name ("gained" or "lost") and stores the
// corresponding FocusEvent value in the receiver. Returns an error
// if the name is not recognized.
func (f *FocusEvent) UnmarshalText(text []byte) error {
	switch string(text) {
	case "gained":
		*f = FocusGained
	case "lost":
		*f = FocusLost
	default:
		return fmt.Errorf("libghostty: unknown focus event %q", string(text))
	}
	return nil
}

// ParseFocusEvent returns the FocusEvent value for the given name
// ("gained" or "lost"). Returns an error if the name is not
// recognized.
func ParseFocusEvent(s string) (FocusEvent, error) {
	var f FocusEvent
	if err := f.UnmarshalText([]byte(s)); err != nil {
		return 0, err
	}
	return f, nil
}

// FocusEncode encodes a focus event into a terminal escape sequence
// and returns the result as a byte slice.
func FocusEncode(event FocusEvent) ([]byte, error) {
	encode := func(tls *libc.TLS, buf uintptr, bufLen uint64, written uintptr) vt.GhosttyResult {
		return vt.Xghostty_focus_encode(
			tls,
			vt.GhosttyFocusEvent(event),
			buf,
			bufLen,
			written,
		)
	}

	// Focus sequences are short (CSI I or CSI O = 3 bytes).
	buf := make([]byte, 16)
	result, outLen := encodeInto(buf, encode)

	if result == vt.GHOSTTY_SUCCESS {
		if outLen == 0 {
			return nil, nil
		}
		out := make([]byte, outLen)
		copy(out, buf[:outLen])
		return out, nil
	}

	if result == vt.GHOSTTY_OUT_OF_SPACE {
		dynBuf := make([]byte, outLen)
		result, written := encodeInto(dynBuf, encode)
		if err := resultError(result); err != nil {
			return nil, err
		}
		if written == 0 {
			return nil, nil
		}
		return dynBuf[:written], nil
	}

	return nil, &Error{Result: Result(result)}
}
