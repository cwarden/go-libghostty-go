package libghostty

// Paste bindings for terminal-aware pasting plus terminal-free validation and
// encoding. Wraps the C APIs from paste.h.

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// PasteSource identifies why paste content is being inserted.
//
// C: GhosttyPasteSource
type PasteSource int

const (
	// PasteSourceClipboard identifies a user-initiated clipboard paste. When
	// Kitty paste events are enabled and a clipboard-read callback is
	// installed, the terminal sends an event instead of writing text directly.
	PasteSourceClipboard PasteSource = vt.GHOSTTY_PASTE_SOURCE_CLIPBOARD

	// PasteSourceText identifies text inserted by another mechanism such as an
	// IME commit, drag and drop, or scripted input. It is always written as text.
	PasteSourceText PasteSource = vt.GHOSTTY_PASTE_SOURCE_TEXT
)

// Paste describes content to paste into a terminal. MIMEs lists every
// available representation in preferred order. Reader is called at most once
// for the first text representation selected by libghostty and is not called
// when a Kitty paste event is emitted.
//
// C: GhosttyPaste
type Paste struct {
	// Location identifies the clipboard that supplied the content.
	Location ClipboardLocation

	// Source identifies why the paste happened.
	Source PasteSource

	// MIMEs lists the available representation types in preferred order.
	MIMEs []string

	// Reader streams a requested representation. It is required when MIMEs is
	// non-empty.
	Reader MIMEReaderFn

	// AllowUnsafe permits text that could inject commands. Call Paste with this
	// false first, confirm with the user on ResultRejected, then retry with it
	// true.
	AllowUnsafe bool
}

// Paste applies the terminal's current paste modes and writes either encoded
// text or a Kitty clipboard paste event through the terminal's write-pty
// callback. The returned boolean reports whether anything was written. Paste
// returns an error matching [ResultInvalidValue] if the terminal has no
// write-pty callback.
//
// Paste refuses text that could run commands without the user meaning to.
// Normally that is text with a newline or carriage return, because the shell
// treats either one like the user pressing Enter. When the program has turned
// on bracketed paste mode, line breaks are safe, and only the sequence that
// ends a bracketed paste is refused. Refused text returns an error matching
// [ResultRejected] and nothing is written. Set Paste.AllowUnsafe to paste it
// anyway.
//
// If the reader fails, the returned error matches both [ResultIOError] and
// the error the reader returned.
//
// C: ghostty_terminal_paste
func (t *Terminal) Paste(paste Paste) (bool, error) {
	mimeValues := make([][]byte, len(paste.MIMEs))
	for i, mime := range paste.MIMEs {
		mimeValues[i] = []byte(mime)
	}
	mimes, err := newCGhosttyStringArray(mimeValues)
	if err != nil {
		return false, err
	}
	defer mimes.close()

	bridge := &ghosttyMIMEReaderBridge{}
	if len(paste.MIMEs) > 0 {
		if paste.Reader == nil {
			return false, &Error{Result: ResultInvalidValue}
		}
		bridge.reader = paste.Reader
		bridge.handle = newHandle(bridge)
		defer bridge.close()
	}

	tls := getTLS()
	defer putTLS(tls)

	// One block holds the request followed by the written flag.
	const requestSize = int(unsafe.Sizeof(vt.GhosttyPaste{}))
	const n = requestSize + 1
	p := tls.Alloc(n)
	defer tls.Free(n)
	clear(unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
	request := (*vt.GhosttyPaste)(unsafe.Pointer(p))
	request.Fsize = uint64(requestSize)
	request.Flocation = vt.GhosttyClipboardLocation(paste.Location)
	request.Fsource = vt.GhosttyPasteSource(paste.Source)
	request.Fmimes = mimes.ptr
	request.Fmimes_len = uint64(len(paste.MIMEs))
	request.Fallow_unsafe = cBool(paste.AllowUnsafe)
	if bridge.handle != 0 {
		request.Freader = bridge.cMIMEReader()
	}
	written := p + uintptr(requestSize)

	result := vt.Xghostty_terminal_paste(tls, t.ptr, p, written)
	if err := resultErrorWithCallback(result, bridge.err); err != nil {
		return false, err
	}
	return *(*uint8)(unsafe.Pointer(written)) != 0, nil
}

// PasteIsSafe reports whether data is safe to paste into a terminal.
//
// Data is unsafe if it contains a newline or carriage return, because the
// shell treats either one like the user pressing Enter and may run a
// command. Data is also unsafe if it contains the sequence that ends a
// bracketed paste (ESC [201~), because it lets the pasted text end the paste
// early and run whatever follows as a command. Empty data is safe.
//
// PasteIsSafe is stricter than [Terminal.Paste]. It doesn't know whether the
// terminal is in bracketed paste mode, so it treats line breaks as unsafe
// even when Paste would allow them. Use it to warn the user about every
// paste that contains a line break, whatever mode the terminal is in.
func PasteIsSafe(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	tls := getTLS()
	defer putTLS(tls)

	addr, unpin := pinBytes(data)
	defer unpin()
	return vt.Xghostty_paste_is_safe(tls, addr, uint64(len(data))) != 0
}

// PasteEncode prepares data for writing to a terminal pty.
//
// The encoder applies Ghostty's terminal-input paste rules: unsafe control
// bytes are replaced with spaces, bracketed paste markers are added when
// bracketed is true, and newlines become carriage returns when bracketed is
// false. The input slice is copied before calling into Ghostty because the C
// encoder mutates its data buffer in place.
func PasteEncode(data []byte, bracketed bool) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// The current encoder only grows output by adding bracketed-paste markers,
	// but intentionally treat buffer sizing as an implementation detail of the C
	// API. Starting with modest slack keeps the common path allocation-light
	// while still honoring GHOSTTY_OUT_OF_SPACE below if the encoder changes.
	in := append([]byte(nil), data...)
	buf := make([]byte, len(data)+32)
	out, required, err := pasteEncodeInto(in, bracketed, buf)
	if err == nil {
		return out, nil
	}

	ge, ok := err.(*Error)
	if !ok || ge.Result != ResultOutOfSpace || required <= 0 {
		return nil, err
	}

	in = append([]byte(nil), data...)
	buf = make([]byte, required)
	out, _, err = pasteEncodeInto(in, bracketed, buf)
	return out, err
}

func pasteEncodeInto(
	data []byte,
	bracketed bool,
	buf []byte,
) ([]byte, int, error) {
	tls := getTLS()
	defer putTLS(tls)

	const n = int(unsafe.Sizeof(uint64(0)))
	written := tls.Alloc(n)
	defer tls.Free(n)
	*(*uint64)(unsafe.Pointer(written)) = 0

	dataAddr, unpinData := pinBytes(data)
	defer unpinData()
	bufAddr, unpinBuf := pinBytes(buf)
	defer unpinBuf()
	result := vt.Xghostty_paste_encode(
		tls,
		dataAddr,
		uint64(len(data)),
		cBool(bracketed),
		bufAddr,
		uint64(len(buf)),
		written,
	)
	count := int(*(*uint64)(unsafe.Pointer(written)))
	if result == vt.GHOSTTY_SUCCESS {
		return buf[:count], count, nil
	}

	return nil, count, &Error{Result: Result(result)}
}
