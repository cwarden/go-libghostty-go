package libghostty

// Key event representation and manipulation.
// Wraps the C APIs from key/event.h.

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// KeyEvent is an opaque handle representing a keyboard input event
// containing information about the physical key pressed, modifiers,
// and generated text. It is mutable and reusable, but not safe for
// concurrent use.
//
// C: GhosttyKeyEvent
type KeyEvent struct {
	ptr uintptr

	// utf8 is the C-owned copy of the text passed to [KeyEvent.SetUTF8], and
	// utf8Len is its size. libghostty keeps a pointer to the text rather than
	// copying it, and the garbage collector can't see that reference, so the
	// bytes live in C memory until the text is replaced or the event is
	// closed.
	utf8    unsafe.Pointer
	utf8Len uintptr
}

// KeyAction represents the type of keyboard input event (press, release,
// or repeat).
//
// C: GhosttyKeyAction
type KeyAction int

const (
	// KeyActionRelease indicates a key was released.
	KeyActionRelease KeyAction = vt.GHOSTTY_KEY_ACTION_RELEASE

	// KeyActionPress indicates a key was pressed.
	KeyActionPress KeyAction = vt.GHOSTTY_KEY_ACTION_PRESS

	// KeyActionRepeat indicates a key is being held down (repeat).
	KeyActionRepeat KeyAction = vt.GHOSTTY_KEY_ACTION_REPEAT
)

// Mods is a bitmask representing keyboard modifier keys.
// Use the Mods* constants to test and set individual modifiers.
//
// The values are those of the GHOSTTY_MODS_* macros in key/mods.h. The
// translated headers in internal/vt do not carry them because they are
// expressions rather than literals.
//
// C: GhosttyMods
type Mods uint16

const (
	// ModShift indicates the Shift key is pressed.
	ModShift Mods = 1 << 0

	// ModCtrl indicates the Control key is pressed.
	ModCtrl Mods = 1 << 1

	// ModAlt indicates the Alt/Option key is pressed.
	ModAlt Mods = 1 << 2

	// ModSuper indicates the Super/Command/Windows key is pressed.
	ModSuper Mods = 1 << 3

	// ModCapsLock indicates Caps Lock is active.
	ModCapsLock Mods = 1 << 4

	// ModNumLock indicates Num Lock is active.
	ModNumLock Mods = 1 << 5

	// ModShiftSide indicates right Shift is pressed (0 = left, 1 = right).
	// Only meaningful when ModShift is set.
	ModShiftSide Mods = 1 << 6

	// ModCtrlSide indicates right Ctrl is pressed (0 = left, 1 = right).
	// Only meaningful when ModCtrl is set.
	ModCtrlSide Mods = 1 << 7

	// ModAltSide indicates right Alt is pressed (0 = left, 1 = right).
	// Only meaningful when ModAlt is set.
	ModAltSide Mods = 1 << 8

	// ModSuperSide indicates right Super is pressed (0 = left, 1 = right).
	// Only meaningful when ModSuper is set.
	ModSuperSide Mods = 1 << 9
)

// Key represents a physical key code. These are layout-independent and
// based on the W3C UI Events KeyboardEvent code standard.
//
// C: GhosttyKey
type Key int

// Writing System Keys (W3C § 3.1.1)
const (
	KeyUnidentified  Key = vt.GHOSTTY_KEY_UNIDENTIFIED
	KeyBackquote     Key = vt.GHOSTTY_KEY_BACKQUOTE
	KeyBackslash     Key = vt.GHOSTTY_KEY_BACKSLASH
	KeyBracketLeft   Key = vt.GHOSTTY_KEY_BRACKET_LEFT
	KeyBracketRight  Key = vt.GHOSTTY_KEY_BRACKET_RIGHT
	KeyComma         Key = vt.GHOSTTY_KEY_COMMA
	KeyDigit0        Key = vt.GHOSTTY_KEY_DIGIT_0
	KeyDigit1        Key = vt.GHOSTTY_KEY_DIGIT_1
	KeyDigit2        Key = vt.GHOSTTY_KEY_DIGIT_2
	KeyDigit3        Key = vt.GHOSTTY_KEY_DIGIT_3
	KeyDigit4        Key = vt.GHOSTTY_KEY_DIGIT_4
	KeyDigit5        Key = vt.GHOSTTY_KEY_DIGIT_5
	KeyDigit6        Key = vt.GHOSTTY_KEY_DIGIT_6
	KeyDigit7        Key = vt.GHOSTTY_KEY_DIGIT_7
	KeyDigit8        Key = vt.GHOSTTY_KEY_DIGIT_8
	KeyDigit9        Key = vt.GHOSTTY_KEY_DIGIT_9
	KeyEqual         Key = vt.GHOSTTY_KEY_EQUAL
	KeyIntlBackslash Key = vt.GHOSTTY_KEY_INTL_BACKSLASH
	KeyIntlRo        Key = vt.GHOSTTY_KEY_INTL_RO
	KeyIntlYen       Key = vt.GHOSTTY_KEY_INTL_YEN
	KeyA             Key = vt.GHOSTTY_KEY_A
	KeyB             Key = vt.GHOSTTY_KEY_B
	KeyC             Key = vt.GHOSTTY_KEY_C
	KeyD             Key = vt.GHOSTTY_KEY_D
	KeyE             Key = vt.GHOSTTY_KEY_E
	KeyF             Key = vt.GHOSTTY_KEY_F
	KeyG             Key = vt.GHOSTTY_KEY_G
	KeyH             Key = vt.GHOSTTY_KEY_H
	KeyI             Key = vt.GHOSTTY_KEY_I
	KeyJ             Key = vt.GHOSTTY_KEY_J
	KeyK             Key = vt.GHOSTTY_KEY_K
	KeyL             Key = vt.GHOSTTY_KEY_L
	KeyM             Key = vt.GHOSTTY_KEY_M
	KeyN             Key = vt.GHOSTTY_KEY_N
	KeyO             Key = vt.GHOSTTY_KEY_O
	KeyP             Key = vt.GHOSTTY_KEY_P
	KeyQ             Key = vt.GHOSTTY_KEY_Q
	KeyR             Key = vt.GHOSTTY_KEY_R
	KeyS             Key = vt.GHOSTTY_KEY_S
	KeyT             Key = vt.GHOSTTY_KEY_T
	KeyU             Key = vt.GHOSTTY_KEY_U
	KeyV             Key = vt.GHOSTTY_KEY_V
	KeyW             Key = vt.GHOSTTY_KEY_W
	KeyX             Key = vt.GHOSTTY_KEY_X
	KeyY             Key = vt.GHOSTTY_KEY_Y
	KeyZ             Key = vt.GHOSTTY_KEY_Z
	KeyMinus         Key = vt.GHOSTTY_KEY_MINUS
	KeyPeriod        Key = vt.GHOSTTY_KEY_PERIOD
	KeyQuote         Key = vt.GHOSTTY_KEY_QUOTE
	KeySemicolon     Key = vt.GHOSTTY_KEY_SEMICOLON
	KeySlash         Key = vt.GHOSTTY_KEY_SLASH
)

// Functional Keys (W3C § 3.1.2)
const (
	KeyAltLeft      Key = vt.GHOSTTY_KEY_ALT_LEFT
	KeyAltRight     Key = vt.GHOSTTY_KEY_ALT_RIGHT
	KeyBackspace    Key = vt.GHOSTTY_KEY_BACKSPACE
	KeyCapsLock     Key = vt.GHOSTTY_KEY_CAPS_LOCK
	KeyContextMenu  Key = vt.GHOSTTY_KEY_CONTEXT_MENU
	KeyControlLeft  Key = vt.GHOSTTY_KEY_CONTROL_LEFT
	KeyControlRight Key = vt.GHOSTTY_KEY_CONTROL_RIGHT
	KeyEnter        Key = vt.GHOSTTY_KEY_ENTER
	KeyMetaLeft     Key = vt.GHOSTTY_KEY_META_LEFT
	KeyMetaRight    Key = vt.GHOSTTY_KEY_META_RIGHT
	KeyShiftLeft    Key = vt.GHOSTTY_KEY_SHIFT_LEFT
	KeyShiftRight   Key = vt.GHOSTTY_KEY_SHIFT_RIGHT
	KeySpace        Key = vt.GHOSTTY_KEY_SPACE
	KeyTab          Key = vt.GHOSTTY_KEY_TAB
	KeyConvert      Key = vt.GHOSTTY_KEY_CONVERT
	KeyKanaMode     Key = vt.GHOSTTY_KEY_KANA_MODE
	KeyNonConvert   Key = vt.GHOSTTY_KEY_NON_CONVERT
)

// Control Pad Section (W3C § 3.2)
const (
	KeyDelete   Key = vt.GHOSTTY_KEY_DELETE
	KeyEnd      Key = vt.GHOSTTY_KEY_END
	KeyHelp     Key = vt.GHOSTTY_KEY_HELP
	KeyHome     Key = vt.GHOSTTY_KEY_HOME
	KeyInsert   Key = vt.GHOSTTY_KEY_INSERT
	KeyPageDown Key = vt.GHOSTTY_KEY_PAGE_DOWN
	KeyPageUp   Key = vt.GHOSTTY_KEY_PAGE_UP
)

// Arrow Pad Section (W3C § 3.3)
const (
	KeyArrowDown  Key = vt.GHOSTTY_KEY_ARROW_DOWN
	KeyArrowLeft  Key = vt.GHOSTTY_KEY_ARROW_LEFT
	KeyArrowRight Key = vt.GHOSTTY_KEY_ARROW_RIGHT
	KeyArrowUp    Key = vt.GHOSTTY_KEY_ARROW_UP
)

// Numpad Section (W3C § 3.4)
const (
	KeyNumLock            Key = vt.GHOSTTY_KEY_NUM_LOCK
	KeyNumpad0            Key = vt.GHOSTTY_KEY_NUMPAD_0
	KeyNumpad1            Key = vt.GHOSTTY_KEY_NUMPAD_1
	KeyNumpad2            Key = vt.GHOSTTY_KEY_NUMPAD_2
	KeyNumpad3            Key = vt.GHOSTTY_KEY_NUMPAD_3
	KeyNumpad4            Key = vt.GHOSTTY_KEY_NUMPAD_4
	KeyNumpad5            Key = vt.GHOSTTY_KEY_NUMPAD_5
	KeyNumpad6            Key = vt.GHOSTTY_KEY_NUMPAD_6
	KeyNumpad7            Key = vt.GHOSTTY_KEY_NUMPAD_7
	KeyNumpad8            Key = vt.GHOSTTY_KEY_NUMPAD_8
	KeyNumpad9            Key = vt.GHOSTTY_KEY_NUMPAD_9
	KeyNumpadAdd          Key = vt.GHOSTTY_KEY_NUMPAD_ADD
	KeyNumpadBackspace    Key = vt.GHOSTTY_KEY_NUMPAD_BACKSPACE
	KeyNumpadClear        Key = vt.GHOSTTY_KEY_NUMPAD_CLEAR
	KeyNumpadClearEntry   Key = vt.GHOSTTY_KEY_NUMPAD_CLEAR_ENTRY
	KeyNumpadComma        Key = vt.GHOSTTY_KEY_NUMPAD_COMMA
	KeyNumpadDecimal      Key = vt.GHOSTTY_KEY_NUMPAD_DECIMAL
	KeyNumpadDivide       Key = vt.GHOSTTY_KEY_NUMPAD_DIVIDE
	KeyNumpadEnter        Key = vt.GHOSTTY_KEY_NUMPAD_ENTER
	KeyNumpadEqual        Key = vt.GHOSTTY_KEY_NUMPAD_EQUAL
	KeyNumpadMemoryAdd    Key = vt.GHOSTTY_KEY_NUMPAD_MEMORY_ADD
	KeyNumpadMemoryClear  Key = vt.GHOSTTY_KEY_NUMPAD_MEMORY_CLEAR
	KeyNumpadMemoryRecall Key = vt.GHOSTTY_KEY_NUMPAD_MEMORY_RECALL
	KeyNumpadMemoryStore  Key = vt.GHOSTTY_KEY_NUMPAD_MEMORY_STORE
	KeyNumpadMemorySub    Key = vt.GHOSTTY_KEY_NUMPAD_MEMORY_SUBTRACT
	KeyNumpadMultiply     Key = vt.GHOSTTY_KEY_NUMPAD_MULTIPLY
	KeyNumpadParenLeft    Key = vt.GHOSTTY_KEY_NUMPAD_PAREN_LEFT
	KeyNumpadParenRight   Key = vt.GHOSTTY_KEY_NUMPAD_PAREN_RIGHT
	KeyNumpadSubtract     Key = vt.GHOSTTY_KEY_NUMPAD_SUBTRACT
	KeyNumpadSeparator    Key = vt.GHOSTTY_KEY_NUMPAD_SEPARATOR
	KeyNumpadUp           Key = vt.GHOSTTY_KEY_NUMPAD_UP
	KeyNumpadDown         Key = vt.GHOSTTY_KEY_NUMPAD_DOWN
	KeyNumpadRight        Key = vt.GHOSTTY_KEY_NUMPAD_RIGHT
	KeyNumpadLeft         Key = vt.GHOSTTY_KEY_NUMPAD_LEFT
	KeyNumpadBegin        Key = vt.GHOSTTY_KEY_NUMPAD_BEGIN
	KeyNumpadHome         Key = vt.GHOSTTY_KEY_NUMPAD_HOME
	KeyNumpadEnd          Key = vt.GHOSTTY_KEY_NUMPAD_END
	KeyNumpadInsert       Key = vt.GHOSTTY_KEY_NUMPAD_INSERT
	KeyNumpadDelete       Key = vt.GHOSTTY_KEY_NUMPAD_DELETE
	KeyNumpadPageUp       Key = vt.GHOSTTY_KEY_NUMPAD_PAGE_UP
	KeyNumpadPageDown     Key = vt.GHOSTTY_KEY_NUMPAD_PAGE_DOWN
)

// Function Section (W3C § 3.5)
const (
	KeyEscape      Key = vt.GHOSTTY_KEY_ESCAPE
	KeyF1          Key = vt.GHOSTTY_KEY_F1
	KeyF2          Key = vt.GHOSTTY_KEY_F2
	KeyF3          Key = vt.GHOSTTY_KEY_F3
	KeyF4          Key = vt.GHOSTTY_KEY_F4
	KeyF5          Key = vt.GHOSTTY_KEY_F5
	KeyF6          Key = vt.GHOSTTY_KEY_F6
	KeyF7          Key = vt.GHOSTTY_KEY_F7
	KeyF8          Key = vt.GHOSTTY_KEY_F8
	KeyF9          Key = vt.GHOSTTY_KEY_F9
	KeyF10         Key = vt.GHOSTTY_KEY_F10
	KeyF11         Key = vt.GHOSTTY_KEY_F11
	KeyF12         Key = vt.GHOSTTY_KEY_F12
	KeyF13         Key = vt.GHOSTTY_KEY_F13
	KeyF14         Key = vt.GHOSTTY_KEY_F14
	KeyF15         Key = vt.GHOSTTY_KEY_F15
	KeyF16         Key = vt.GHOSTTY_KEY_F16
	KeyF17         Key = vt.GHOSTTY_KEY_F17
	KeyF18         Key = vt.GHOSTTY_KEY_F18
	KeyF19         Key = vt.GHOSTTY_KEY_F19
	KeyF20         Key = vt.GHOSTTY_KEY_F20
	KeyF21         Key = vt.GHOSTTY_KEY_F21
	KeyF22         Key = vt.GHOSTTY_KEY_F22
	KeyF23         Key = vt.GHOSTTY_KEY_F23
	KeyF24         Key = vt.GHOSTTY_KEY_F24
	KeyF25         Key = vt.GHOSTTY_KEY_F25
	KeyFn          Key = vt.GHOSTTY_KEY_FN
	KeyFnLock      Key = vt.GHOSTTY_KEY_FN_LOCK
	KeyPrintScreen Key = vt.GHOSTTY_KEY_PRINT_SCREEN
	KeyScrollLock  Key = vt.GHOSTTY_KEY_SCROLL_LOCK
	KeyPause       Key = vt.GHOSTTY_KEY_PAUSE
)

// Media Keys (W3C § 3.6)
const (
	KeyBrowserBack        Key = vt.GHOSTTY_KEY_BROWSER_BACK
	KeyBrowserFavorites   Key = vt.GHOSTTY_KEY_BROWSER_FAVORITES
	KeyBrowserForward     Key = vt.GHOSTTY_KEY_BROWSER_FORWARD
	KeyBrowserHome        Key = vt.GHOSTTY_KEY_BROWSER_HOME
	KeyBrowserRefresh     Key = vt.GHOSTTY_KEY_BROWSER_REFRESH
	KeyBrowserSearch      Key = vt.GHOSTTY_KEY_BROWSER_SEARCH
	KeyBrowserStop        Key = vt.GHOSTTY_KEY_BROWSER_STOP
	KeyEject              Key = vt.GHOSTTY_KEY_EJECT
	KeyLaunchApp1         Key = vt.GHOSTTY_KEY_LAUNCH_APP_1
	KeyLaunchApp2         Key = vt.GHOSTTY_KEY_LAUNCH_APP_2
	KeyLaunchMail         Key = vt.GHOSTTY_KEY_LAUNCH_MAIL
	KeyMediaPlayPause     Key = vt.GHOSTTY_KEY_MEDIA_PLAY_PAUSE
	KeyMediaSelect        Key = vt.GHOSTTY_KEY_MEDIA_SELECT
	KeyMediaStop          Key = vt.GHOSTTY_KEY_MEDIA_STOP
	KeyMediaTrackNext     Key = vt.GHOSTTY_KEY_MEDIA_TRACK_NEXT
	KeyMediaTrackPrevious Key = vt.GHOSTTY_KEY_MEDIA_TRACK_PREVIOUS
	KeyPower              Key = vt.GHOSTTY_KEY_POWER
	KeySleep              Key = vt.GHOSTTY_KEY_SLEEP
	KeyAudioVolumeDown    Key = vt.GHOSTTY_KEY_AUDIO_VOLUME_DOWN
	KeyAudioVolumeMute    Key = vt.GHOSTTY_KEY_AUDIO_VOLUME_MUTE
	KeyAudioVolumeUp      Key = vt.GHOSTTY_KEY_AUDIO_VOLUME_UP
	KeyWakeUp             Key = vt.GHOSTTY_KEY_WAKE_UP
)

// Legacy, Non-standard, and Special Keys (W3C § 3.7)
const (
	KeyCopy  Key = vt.GHOSTTY_KEY_COPY
	KeyCut   Key = vt.GHOSTTY_KEY_CUT
	KeyPaste Key = vt.GHOSTTY_KEY_PASTE
)

// NewKeyEvent creates a new key event with default values. The event
// must be freed with Close when no longer needed.
func NewKeyEvent() (*KeyEvent, error) {
	tls := getTLS()
	defer putTLS(tls)

	const n = int(unsafe.Sizeof(uintptr(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)
	if err := resultError(vt.Xghostty_key_event_new(tls, 0, out)); err != nil {
		return nil, err
	}
	return &KeyEvent{ptr: *(*uintptr)(unsafe.Pointer(out))}, nil
}

// Close frees the underlying key event handle. After this call, the
// key event must not be used, except that calling Close again is a safe
// no-op.
func (e *KeyEvent) Close() {
	if e == nil || e.ptr == 0 {
		return
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_key_event_free(tls, e.ptr)
	e.ptr = 0
	e.freeUTF8()
}

// freeUTF8 releases the event's copy of its text, if it has one.
func (e *KeyEvent) freeUTF8() {
	if e.utf8 != nil {
		Free(e.utf8, e.utf8Len)
	}
	e.utf8 = nil
	e.utf8Len = 0
}

// SetAction sets the key action (press, release, repeat).
func (e *KeyEvent) SetAction(action KeyAction) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_key_event_set_action(tls, e.ptr, vt.GhosttyKeyAction(action)))
}

// Action returns the key action (press, release, repeat).
func (e *KeyEvent) Action() KeyAction {
	tls := getTLS()
	defer putTLS(tls)
	return KeyAction(vt.Xghostty_key_event_get_action(tls, e.ptr))
}

// SetKey sets the physical key code.
func (e *KeyEvent) SetKey(key Key) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_key_event_set_key(tls, e.ptr, vt.GhosttyKey(key)))
}

// Key returns the physical key code.
func (e *KeyEvent) Key() Key {
	tls := getTLS()
	defer putTLS(tls)
	return Key(vt.Xghostty_key_event_get_key(tls, e.ptr))
}

// SetMods sets the modifier keys bitmask.
func (e *KeyEvent) SetMods(mods Mods) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_key_event_set_mods(tls, e.ptr, vt.GhosttyMods(mods)))
}

// Mods returns the modifier keys bitmask.
func (e *KeyEvent) Mods() Mods {
	tls := getTLS()
	defer putTLS(tls)
	return Mods(vt.Xghostty_key_event_get_mods(tls, e.ptr))
}

// SetConsumedMods sets the consumed modifiers bitmask.
func (e *KeyEvent) SetConsumedMods(mods Mods) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_key_event_set_consumed_mods(tls, e.ptr, vt.GhosttyMods(mods)))
}

// ConsumedMods returns the consumed modifiers bitmask.
func (e *KeyEvent) ConsumedMods() Mods {
	tls := getTLS()
	defer putTLS(tls)
	return Mods(vt.Xghostty_key_event_get_consumed_mods(tls, e.ptr))
}

// SetComposing sets whether the key event is part of a composition sequence.
func (e *KeyEvent) SetComposing(composing bool) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_key_event_set_composing(tls, e.ptr, cBool(composing)))
}

// Composing reports whether the key event is part of a composition sequence.
func (e *KeyEvent) Composing() bool {
	tls := getTLS()
	defer putTLS(tls)
	return vt.Xghostty_key_event_get_composing(tls, e.ptr) != 0
}

// SetUTF8 sets the UTF-8 text generated by the key for the current
// keyboard layout. Must contain the unmodified character before any
// Ctrl/Meta transformations. Do not pass C0 control characters or
// platform function key codes; pass an empty string instead and let
// the encoder use the logical key.
//
// The event keeps its own copy of the text, so s does not need to outlive
// this call.
func (e *KeyEvent) SetUTF8(s string) error {
	// An empty string clears the text. Its length is zero, so libghostty never
	// reads the pointer and there is nothing to copy.
	length := uintptr(len(s))
	var ptr unsafe.Pointer
	if length > 0 {
		ptr = Alloc(length)
		if ptr == nil {
			return &Error{Result: ResultOutOfMemory}
		}
		copy(unsafe.Slice((*byte)(ptr), len(s)), s)
	}

	tls := getTLS()
	defer putTLS(tls)
	if err := resultError(vt.Xghostty_key_event_set_utf8(tls, e.ptr, vt.GhosttyString{
		Fptr:  uintptr(ptr),
		Flen1: uint64(length),
	})); err != nil {
		// The event still points at its previous text, so keep that copy.
		if ptr != nil {
			Free(ptr, length)
		}
		return err
	}

	e.freeUTF8()
	e.utf8 = ptr
	e.utf8Len = length
	return nil
}

// UTF8 returns the UTF-8 text generated by the key event.
func (e *KeyEvent) UTF8() string {
	tls := getTLS()
	defer putTLS(tls)
	s := vt.Xghostty_key_event_get_utf8(tls, e.ptr)
	if s.Flen1 == 0 {
		return ""
	}
	return string(goBytes(s.Fptr, int(s.Flen1)))
}

// SetUnshiftedCodepoint sets the unshifted Unicode codepoint.
func (e *KeyEvent) SetUnshiftedCodepoint(cp rune) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_key_event_set_unshifted_codepoint(tls, e.ptr, uint32(cp)))
}

// UnshiftedCodepoint returns the unshifted Unicode codepoint.
func (e *KeyEvent) UnshiftedCodepoint() rune {
	tls := getTLS()
	defer putTLS(tls)
	return rune(vt.Xghostty_key_event_get_unshifted_codepoint(tls, e.ptr))
}
