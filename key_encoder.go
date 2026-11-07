package libghostty

// Key encoder — encodes key events into terminal escape sequences.
// Wraps the C APIs from key/encoder.h.

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// KeyEncoder encodes key events into terminal escape sequences,
// supporting both legacy encoding and the Kitty Keyboard Protocol.
// It maintains mutable encoding options and is not safe for concurrent
// use.
//
// Basic usage:
//  1. Create an encoder with NewKeyEncoder.
//  2. Configure options with the Set* methods or SetFromTerminal.
//  3. Create key events, encode them with Encode, and free them.
//  4. Free the encoder with Close when done.
//
// C: GhosttyKeyEncoder
type KeyEncoder struct {
	ptr uintptr
}

// KittyKeyFlags is a bitmask of Kitty keyboard protocol flags. The values
// are those of the GHOSTTY_KITTY_KEY_* macros in key/encoder.h; the
// translated headers in internal/vt carry only the literal one.
// C: GhosttyKittyKeyFlags
type KittyKeyFlags uint8

const (
	// KittyKeyDisabled disables the Kitty keyboard protocol (all flags off).
	KittyKeyDisabled KittyKeyFlags = vt.GHOSTTY_KITTY_KEY_DISABLED

	// KittyKeyDisambiguate enables disambiguating escape codes.
	KittyKeyDisambiguate KittyKeyFlags = 1 << 0

	// KittyKeyReportEvents enables reporting key press and release events.
	KittyKeyReportEvents KittyKeyFlags = 1 << 1

	// KittyKeyReportAlternates enables reporting alternate key codes.
	KittyKeyReportAlternates KittyKeyFlags = 1 << 2

	// KittyKeyReportAll reports all key events including those normally
	// handled by the terminal.
	KittyKeyReportAll KittyKeyFlags = 1 << 3

	// KittyKeyReportAssociated reports associated text with key events.
	KittyKeyReportAssociated KittyKeyFlags = 1 << 4

	// KittyKeyAll enables all Kitty keyboard protocol flags.
	KittyKeyAll KittyKeyFlags = KittyKeyDisambiguate | KittyKeyReportEvents | KittyKeyReportAlternates | KittyKeyReportAll | KittyKeyReportAssociated
)

// OptionAsAlt determines whether the macOS "option" key is treated as
// "alt".
//
// C: GhosttyOptionAsAlt
type OptionAsAlt int

const (
	// OptionAsAltFalse means the option key is not treated as alt.
	OptionAsAltFalse OptionAsAlt = vt.GHOSTTY_OPTION_AS_ALT_FALSE

	// OptionAsAltTrue means the option key is treated as alt.
	OptionAsAltTrue OptionAsAlt = vt.GHOSTTY_OPTION_AS_ALT_TRUE

	// OptionAsAltLeft means only the left option key is treated as alt.
	OptionAsAltLeft OptionAsAlt = vt.GHOSTTY_OPTION_AS_ALT_LEFT

	// OptionAsAltRight means only the right option key is treated as alt.
	OptionAsAltRight OptionAsAlt = vt.GHOSTTY_OPTION_AS_ALT_RIGHT
)

// KeyEncoderOption identifies an encoder configuration option. Most take a
// bool and are set with [KeyEncoder.SetBool]. KeyEncoderOptKittyFlags and
// KeyEncoderOptMacOSOptionAsAlt have their own setters.
//
// C: GhosttyKeyEncoderOption
type KeyEncoderOption int

const (
	// KeyEncoderOptCursorKeyApplication sets DEC mode 1: cursor key
	// application mode (value: bool).
	KeyEncoderOptCursorKeyApplication KeyEncoderOption = vt.GHOSTTY_KEY_ENCODER_OPT_CURSOR_KEY_APPLICATION

	// KeyEncoderOptKeypadKeyApplication sets DEC mode 66: keypad key
	// application mode (value: bool).
	KeyEncoderOptKeypadKeyApplication KeyEncoderOption = vt.GHOSTTY_KEY_ENCODER_OPT_KEYPAD_KEY_APPLICATION

	// KeyEncoderOptIgnoreKeypadWithNumlock sets DEC mode 1035: ignore
	// keypad with numlock (value: bool).
	KeyEncoderOptIgnoreKeypadWithNumlock KeyEncoderOption = vt.GHOSTTY_KEY_ENCODER_OPT_IGNORE_KEYPAD_WITH_NUMLOCK

	// KeyEncoderOptAltEscPrefix sets DEC mode 1036: alt sends escape
	// prefix (value: bool).
	KeyEncoderOptAltEscPrefix KeyEncoderOption = vt.GHOSTTY_KEY_ENCODER_OPT_ALT_ESC_PREFIX

	// KeyEncoderOptModifyOtherKeysState2 sets xterm modifyOtherKeys
	// mode 2 (value: bool).
	KeyEncoderOptModifyOtherKeysState2 KeyEncoderOption = vt.GHOSTTY_KEY_ENCODER_OPT_MODIFY_OTHER_KEYS_STATE_2

	// KeyEncoderOptKittyFlags sets Kitty keyboard protocol flags
	// (value: KittyKeyFlags bitmask).
	KeyEncoderOptKittyFlags KeyEncoderOption = vt.GHOSTTY_KEY_ENCODER_OPT_KITTY_FLAGS

	// KeyEncoderOptMacOSOptionAsAlt sets the macOS option-as-alt
	// setting (value: OptionAsAlt).
	KeyEncoderOptMacOSOptionAsAlt KeyEncoderOption = vt.GHOSTTY_KEY_ENCODER_OPT_MACOS_OPTION_AS_ALT

	// KeyEncoderOptBackarrowKeyMode sets backarrow key mode (value: bool).
	// When false (default), backspace emits 0x7f; when true, 0x08.
	KeyEncoderOptBackarrowKeyMode KeyEncoderOption = vt.GHOSTTY_KEY_ENCODER_OPT_BACKARROW_KEY_MODE
)

// NewKeyEncoder creates a new key encoder with default options.
// The encoder must be freed with Close when no longer needed.
func NewKeyEncoder() (*KeyEncoder, error) {
	tls := getTLS()
	defer putTLS(tls)

	const n = int(unsafe.Sizeof(uintptr(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)
	if err := resultError(vt.Xghostty_key_encoder_new(tls, 0, out)); err != nil {
		return nil, err
	}
	return &KeyEncoder{ptr: *(*uintptr)(unsafe.Pointer(out))}, nil
}

// Close frees the underlying key encoder handle. After this call,
// the encoder must not be used, except that calling Close again is a
// safe no-op.
func (enc *KeyEncoder) Close() {
	if enc == nil || enc.ptr == 0 {
		return
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_key_encoder_free(tls, enc.ptr)
	enc.ptr = 0
}

// set writes one encoder option whose value has type T.
func setKeyEncoderScalar[T any](enc *KeyEncoder, opt vt.GhosttyKeyEncoderOption, v T) error {
	tls := getTLS()
	defer putTLS(tls)

	n := int(unsafe.Sizeof(v))
	p := tls.Alloc(n)
	defer tls.Free(n)
	*(*T)(unsafe.Pointer(p)) = v
	return resultError(vt.Xghostty_key_encoder_set(tls, enc.ptr, opt, p))
}

// SetBool sets a boolean encoder option. Use this for options that
// accept a bool value (most encoder options). It returns an error with
// result [ResultInvalidValue] for an option that takes another type.
func (enc *KeyEncoder) SetBool(opt KeyEncoderOption, val bool) error {
	// libghostty reads the value as the option's own type, and the option-as-alt
	// enum is wider than a bool. It can't detect the mismatch, so reject it here.
	switch opt {
	case KeyEncoderOptKittyFlags, KeyEncoderOptMacOSOptionAsAlt:
		return &Error{Result: ResultInvalidValue}
	}

	return setKeyEncoderScalar(enc, vt.GhosttyKeyEncoderOption(opt), cBool(val))
}

// SetKittyFlags sets the Kitty keyboard protocol flags on the encoder.
func (enc *KeyEncoder) SetKittyFlags(flags KittyKeyFlags) error {
	return setKeyEncoderScalar(enc, vt.GHOSTTY_KEY_ENCODER_OPT_KITTY_FLAGS, vt.GhosttyKittyKeyFlags(flags))
}

// SetOptionAsAlt sets the macOS option-as-alt behavior on the encoder.
func (enc *KeyEncoder) SetOptionAsAlt(val OptionAsAlt) error {
	return setKeyEncoderScalar(enc, vt.GHOSTTY_KEY_ENCODER_OPT_MACOS_OPTION_AS_ALT, vt.GhosttyOptionAsAlt(val))
}

// SetFromTerminal reads the terminal's current modes and flags and
// applies them to the encoder's options. This sets cursor key
// application mode, keypad mode, alt escape prefix, backarrow key mode,
// ignore keypad with numlock, modifyOtherKeys state, and Kitty keyboard
// protocol flags from the terminal state, replacing any values set
// earlier.
//
// Note that the macOS option-as-alt option cannot be determined from
// terminal state and is reset to OptionAsAltFalse by this call. Use
// SetOptionAsAlt afterward if needed. The caller must serialize
// access to both the encoder and the terminal during this call.
func (enc *KeyEncoder) SetFromTerminal(t *Terminal) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_key_encoder_set_from_terminal(tls, enc.ptr, t.ptr))
}

// Encode encodes a key event into a terminal escape sequence and
// returns the result as a byte slice. Not all key events produce
// output (e.g. unmodified modifier keys); in that case, a nil slice
// and nil error are returned.
func (enc *KeyEncoder) Encode(event *KeyEvent) ([]byte, error) {
	tls := getTLS()
	defer putTLS(tls)

	// Most escape sequences fit in 128 bytes. Encode into a buffer on the
	// C stack first; fall back to a larger allocation if needed.
	const bufLen = 128
	const n = bufLen + int(unsafe.Sizeof(uint64(0)))
	buf := tls.Alloc(n)
	defer tls.Free(n)
	outLen := buf + bufLen
	*(*uint64)(unsafe.Pointer(outLen)) = 0
	result := vt.Xghostty_key_encoder_encode(
		tls,
		enc.ptr,
		event.ptr,
		buf,
		bufLen,
		outLen,
	)
	length := *(*uint64)(unsafe.Pointer(outLen))

	if result == vt.GHOSTTY_SUCCESS {
		if length == 0 {
			return nil, nil
		}
		out := make([]byte, length)
		copy(out, unsafe.Slice((*byte)(unsafe.Pointer(buf)), length))
		return out, nil
	}

	if result == vt.GHOSTTY_OUT_OF_SPACE {
		// length is the required buffer size.
		dynBuf := make([]byte, length)
		addr, unpin := pinBytes(dynBuf)
		defer unpin()
		*(*uint64)(unsafe.Pointer(outLen)) = 0
		if err := resultError(vt.Xghostty_key_encoder_encode(
			tls,
			enc.ptr,
			event.ptr,
			addr,
			length,
			outLen,
		)); err != nil {
			return nil, err
		}
		written := *(*uint64)(unsafe.Pointer(outLen))
		if written == 0 {
			return nil, nil
		}
		return dynBuf[:written], nil
	}

	return nil, &Error{Result: Result(result)}
}
