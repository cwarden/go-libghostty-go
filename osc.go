package libghostty

// OSC parser bindings from osc.h.

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// OSCCommandType identifies a parsed Operating System Command.
// C: GhosttyOscCommandType
type OSCCommandType int

const (
	// OSCCommandInvalid identifies an invalid or unsupported command.
	OSCCommandInvalid OSCCommandType = vt.GHOSTTY_OSC_COMMAND_INVALID

	// OSCCommandChangeWindowTitle changes the window title.
	OSCCommandChangeWindowTitle OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CHANGE_WINDOW_TITLE

	// OSCCommandChangeWindowIcon changes the window icon.
	OSCCommandChangeWindowIcon OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CHANGE_WINDOW_ICON

	// OSCCommandSemanticPrompt carries semantic prompt metadata.
	OSCCommandSemanticPrompt OSCCommandType = vt.GHOSTTY_OSC_COMMAND_SEMANTIC_PROMPT

	// OSCCommandClipboardContents accesses clipboard contents.
	OSCCommandClipboardContents OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CLIPBOARD_CONTENTS

	// OSCCommandReportPwd reports the present working directory.
	OSCCommandReportPwd OSCCommandType = vt.GHOSTTY_OSC_COMMAND_REPORT_PWD

	// OSCCommandMouseShape changes the pointer shape.
	OSCCommandMouseShape OSCCommandType = vt.GHOSTTY_OSC_COMMAND_MOUSE_SHAPE

	// OSCCommandColorOperation performs a color query or update.
	OSCCommandColorOperation OSCCommandType = vt.GHOSTTY_OSC_COMMAND_COLOR_OPERATION

	// OSCCommandKittyColorProtocol uses Kitty's color protocol.
	OSCCommandKittyColorProtocol OSCCommandType = vt.GHOSTTY_OSC_COMMAND_KITTY_COLOR_PROTOCOL

	// OSCCommandShowDesktopNotification requests a desktop notification.
	OSCCommandShowDesktopNotification OSCCommandType = vt.GHOSTTY_OSC_COMMAND_SHOW_DESKTOP_NOTIFICATION

	// OSCCommandHyperlinkStart begins an OSC 8 hyperlink.
	OSCCommandHyperlinkStart OSCCommandType = vt.GHOSTTY_OSC_COMMAND_HYPERLINK_START

	// OSCCommandHyperlinkEnd ends an OSC 8 hyperlink.
	OSCCommandHyperlinkEnd OSCCommandType = vt.GHOSTTY_OSC_COMMAND_HYPERLINK_END

	// OSCCommandConEmuSleep is a ConEmu sleep command.
	OSCCommandConEmuSleep OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CONEMU_SLEEP

	// OSCCommandConEmuShowMessageBox is a ConEmu message-box command.
	OSCCommandConEmuShowMessageBox OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CONEMU_SHOW_MESSAGE_BOX

	// OSCCommandConEmuChangeTabTitle is a ConEmu tab-title command.
	OSCCommandConEmuChangeTabTitle OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CONEMU_CHANGE_TAB_TITLE

	// OSCCommandConEmuProgressReport is a ConEmu progress command.
	OSCCommandConEmuProgressReport OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CONEMU_PROGRESS_REPORT

	// OSCCommandConEmuWaitInput is a ConEmu wait-input command.
	OSCCommandConEmuWaitInput OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CONEMU_WAIT_INPUT

	// OSCCommandConEmuGuiMacro is a ConEmu GUI macro command.
	OSCCommandConEmuGuiMacro OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CONEMU_GUIMACRO

	// OSCCommandConEmuRunProcess is a ConEmu process command.
	OSCCommandConEmuRunProcess OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CONEMU_RUN_PROCESS

	// OSCCommandConEmuOutputEnvironmentVariable is a ConEmu environment
	// variable command.
	OSCCommandConEmuOutputEnvironmentVariable OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CONEMU_OUTPUT_ENVIRONMENT_VARIABLE

	// OSCCommandConEmuXtermEmulation is a ConEmu xterm-emulation command.
	OSCCommandConEmuXtermEmulation OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CONEMU_XTERM_EMULATION

	// OSCCommandConEmuComment is a ConEmu comment command.
	OSCCommandConEmuComment OSCCommandType = vt.GHOSTTY_OSC_COMMAND_CONEMU_COMMENT

	// OSCCommandKittyTextSizing uses Kitty's text-sizing protocol.
	OSCCommandKittyTextSizing OSCCommandType = vt.GHOSTTY_OSC_COMMAND_KITTY_TEXT_SIZING

	// OSCCommandKittyClipboardProtocol uses Kitty's clipboard protocol.
	OSCCommandKittyClipboardProtocol OSCCommandType = vt.GHOSTTY_OSC_COMMAND_KITTY_CLIPBOARD_PROTOCOL

	// OSCCommandKittyDNDProtocol uses Kitty's drag-and-drop protocol.
	OSCCommandKittyDNDProtocol OSCCommandType = vt.GHOSTTY_OSC_COMMAND_KITTY_DND_PROTOCOL

	// OSCCommandHierarchicalContext is an OSC 3008 hierarchical context
	// command. Programs use it to tell the terminal when they enter or
	// leave a context, such as a shell session or a command, and to
	// attach details about that context. Contexts can nest inside each
	// other. See https://uapi-group.org/specifications/specs/osc_context/.
	OSCCommandHierarchicalContext OSCCommandType = vt.GHOSTTY_OSC_COMMAND_HIERARCHICAL_CONTEXT

	// OSCCommandContextSignal is the old name for
	// OSCCommandHierarchicalContext.
	//
	// Deprecated: Use [OSCCommandHierarchicalContext] instead.
	OSCCommandContextSignal = OSCCommandHierarchicalContext

	// OSCCommandKittyDesktopNotification uses Kitty's desktop notification protocol.
	OSCCommandKittyDesktopNotification OSCCommandType = vt.GHOSTTY_OSC_COMMAND_KITTY_DESKTOP_NOTIFICATION

	// OSCCommandUnknown identifies an OSC command whose number the parser
	// does not recognize. Enable capture with [OSCParser.SetUnknownMaxBytes].
	OSCCommandUnknown OSCCommandType = vt.GHOSTTY_OSC_COMMAND_UNKNOWN

	// OSCCommandProgramStatus identifies an OSC 7501 program status report,
	// or a program asking whether the terminal supports those reports. The
	// parser doesn't decode the report itself. To read reports, create a
	// [Terminal] with [WithProgramStatus].
	OSCCommandProgramStatus OSCCommandType = vt.GHOSTTY_OSC_COMMAND_PROGRAM_STATUS
)

// OSCCommandData identifies typed data extractable from an OSC command.
// C: GhosttyOscCommandData
type OSCCommandData int

const (
	// OSCDataInvalid is an invalid data query.
	OSCDataInvalid OSCCommandData = vt.GHOSTTY_OSC_DATA_INVALID

	// OSCDataChangeWindowTitleString extracts a window title as a GhosttyString.
	// The bytes are not null-terminated and belong to the parser, so they remain
	// valid only until the next operation on it. Use [OSCCommand.WindowTitle] for
	// a copy.
	OSCDataChangeWindowTitleString OSCCommandData = vt.GHOSTTY_OSC_DATA_CHANGE_WINDOW_TITLE_STR

	// OSCDataUnknownContent extracts an unknown command's body as a
	// GhosttyString. The bytes belong to the parser and remain valid only
	// until the next operation on it. Use [OSCCommand.UnknownContent] for a copy.
	OSCDataUnknownContent OSCCommandData = vt.GHOSTTY_OSC_DATA_UNKNOWN_CONTENT

	// OSCDataUnknownTruncated extracts whether an unknown command's content
	// was shortened by the capture limit or an allocation failure (bool).
	OSCDataUnknownTruncated OSCCommandData = vt.GHOSTTY_OSC_DATA_UNKNOWN_TRUNCATED

	// OSCDataUnknownTerminator extracts how an unknown command ended
	// (GhosttyOscTerminator).
	OSCDataUnknownTerminator OSCCommandData = vt.GHOSTTY_OSC_DATA_UNKNOWN_TERMINATOR
)

// OSCTerminator identifies how an OSC sequence ended. Replies should use the
// same terminator as the request.
//
// These constants identify terminators but are not byte values. [OSCParser.End]
// takes the terminating byte instead.
//
// C: GhosttyOscTerminator
type OSCTerminator int

const (
	// OSCTerminatorST is the string terminator: ESC (0x1b) followed by a backslash (0x5c).
	OSCTerminatorST OSCTerminator = vt.GHOSTTY_OSC_TERMINATOR_ST

	// OSCTerminatorBEL is the bell byte (0x07).
	OSCTerminatorBEL OSCTerminator = vt.GHOSTTY_OSC_TERMINATOR_BEL
)

// OSCParser incrementally parses the bytes inside an OSC sequence.
// C: GhosttyOscParser
type OSCParser struct {
	ptr vt.GhosttyOscParser
}

// OSCCommand is a borrowed command produced by [OSCParser.End]. It remains
// valid until the next parser operation other than command introspection.
// C: GhosttyOscCommand
type OSCCommand struct {
	ptr vt.GhosttyOscCommand
}

// NewOSCParser creates a reusable OSC parser.
func NewOSCParser() (*OSCParser, error) {
	tls := getTLS()
	defer putTLS(tls)

	const n = int(unsafe.Sizeof(uintptr(0)))
	out := tls.Alloc(n)
	defer tls.Free(n)
	*(*uintptr)(unsafe.Pointer(out)) = 0
	if err := resultError(vt.Xghostty_osc_new(tls, 0, out)); err != nil {
		return nil, err
	}
	return &OSCParser{ptr: *(*uintptr)(unsafe.Pointer(out))}, nil
}

// Close frees the parser. Commands borrowed from it become invalid. Calling
// Close again is a safe no-op.
func (p *OSCParser) Close() {
	if p == nil || p.ptr == 0 {
		return
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_osc_free(tls, p.ptr)
	p.ptr = 0
}

// Reset clears the current sequence without changing parser options.
// Call Reset before parsing another sequence. Commands returned by earlier
// calls to [OSCParser.End] are no longer valid.
func (p *OSCParser) Reset() {
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_osc_reset(tls, p.ptr)
}

// Next feeds one byte from the OSC sequence body into the parser.
func (p *OSCParser) Next(b byte) {
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_osc_next(tls, p.ptr, b)
}

// End finishes parsing the current sequence and returns its command.
// The terminator is the byte that ended the sequence. This is usually BEL
// (0x07), or the backslash (0x5c) that ends an ESC \ string terminator.
// Commands that send a reply, such as color queries, end the reply the same
// way the request ended.
//
// A program can also cancel a sequence partway through by sending CAN (0x18)
// or SUB (0x1a). Pass that byte as the terminator. The sequence is then
// discarded, even if the bytes before the cancel form a complete command.
// This matches the behavior of xterm.
//
// If the sequence is invalid or was cancelled, the returned command has type
// [OSCCommandInvalid]. The command is valid until the next call to a method
// on p. Call [OSCParser.Reset] before parsing the next sequence.
func (p *OSCParser) End(terminator byte) OSCCommand {
	tls := getTLS()
	defer putTLS(tls)
	return OSCCommand{
		ptr: vt.Xghostty_osc_end(tls, p.ptr, terminator),
	}
}

// Type returns the parsed command type.
func (c OSCCommand) Type() OSCCommandType {
	tls := getTLS()
	defer putTLS(tls)
	return OSCCommandType(vt.Xghostty_osc_command_type(tls, c.ptr))
}

// Data extracts a low-level typed value into out. The pointed-to Go value
// must match the output type documented for data in osc.h. Only a nil error
// writes to out. An error with result [ResultNoValue] means the command has no
// data of this kind, for example asking a window title command for unknown
// content.
//
// out may point into Go memory: the value is written to a C buffer of size
// bytes and then copied, so only the first size bytes of *out are changed.
// The size is the size of the documented output type.
func (c OSCCommand) Data(data OSCCommandData, out unsafe.Pointer) error {
	return c.data(data, out, oscDataSize(data))
}

// oscDataSize returns the size in bytes of the output type documented for
// data in osc.h.
func oscDataSize(data OSCCommandData) int {
	switch data {
	case OSCDataChangeWindowTitleString, OSCDataUnknownContent:
		return int(unsafe.Sizeof(vt.GhosttyString{}))
	case OSCDataUnknownTruncated:
		return 1
	case OSCDataUnknownTerminator:
		return int(unsafe.Sizeof(vt.GhosttyOscTerminator(0)))
	default:
		return 0
	}
}

// data extracts the value for data into a C buffer of size bytes and copies
// it to out on success.
func (c OSCCommand) data(data OSCCommandData, out unsafe.Pointer, size int) error {
	tls := getTLS()
	defer putTLS(tls)

	var buf uintptr
	if size > 0 {
		buf = tls.Alloc(size)
		defer tls.Free(size)
		clear(unsafe.Slice((*byte)(unsafe.Pointer(buf)), size))
	}
	err := resultError(vt.Xghostty_osc_command_data(
		tls,
		c.ptr,
		vt.GhosttyOscCommandData(data),
		buf,
	))
	if err == nil && size > 0 && out != nil {
		copy(unsafe.Slice((*byte)(out), size), unsafe.Slice((*byte)(unsafe.Pointer(buf)), size))
	}
	return err
}

// WindowTitle returns the title carried by a change-window-title command.
// The string is copied into Go-owned memory.
func (c OSCCommand) WindowTitle() (string, bool) {
	var v vt.GhosttyString
	if c.Data(OSCDataChangeWindowTitleString, unsafe.Pointer(&v)) != nil {
		return "", false
	}
	title, ok := copyGhosttyString(v)
	if !ok {
		return "", false
	}
	return string(title), true
}

// SetUnknownMaxBytes sets the maximum number of bytes retained for each OSC
// sequence whose command number the parser does not recognize. Zero, the
// default, disables capture. A positive limit enables [OSCCommandUnknown].
//
// Longer sequences are still reported, but contain only the first limit bytes.
// [OSCCommand.UnknownTruncated] reports whether content was lost. Recognized
// commands are never reported as unknown, even when their contents are invalid.
//
// The limit remains set across calls to [OSCParser.Reset]. Set it before feeding
// a sequence, since a sequence already in progress may keep the previous limit.
// Limits up to 2048 bytes use the parser's existing buffer. Larger limits may
// allocate memory for each unknown sequence.
//
// C: ghostty_osc_set, GHOSTTY_OSC_OPT_UNKNOWN_MAX_BYTES
func (p *OSCParser) SetUnknownMaxBytes(limit uint) error {
	tls := getTLS()
	defer putTLS(tls)

	const n = int(unsafe.Sizeof(uint64(0)))
	v := tls.Alloc(n)
	defer tls.Free(n)
	*(*uint64)(unsafe.Pointer(v)) = uint64(limit)
	return resultError(vt.Xghostty_osc_set(tls, p.ptr, vt.GHOSTTY_OSC_OPT_UNKNOWN_MAX_BYTES, v))
}

// UnknownContent returns a copy of an unknown command's body, including its
// command number and excluding the sequence delimiters. For example, the OSC
// sequence "\x1b]7400;hello\a" has the body "7400;hello". The bytes may be kept
// or modified after the parser is reset or closed.
//
// It returns nil, false if c is not [OSCCommandUnknown] or the content cannot
// be represented as a Go byte slice.
func (c OSCCommand) UnknownContent() (content []byte, ok bool) {
	var v vt.GhosttyString
	if c.Data(OSCDataUnknownContent, unsafe.Pointer(&v)) != nil {
		return nil, false
	}
	return copyGhosttyString(v)
}

// UnknownTruncated reports whether an unknown command's content was shortened
// by the capture limit or a memory allocation failure. When truncated is true,
// [OSCCommand.UnknownContent] returns only the beginning of the sequence.
// The ok result is false if c is not [OSCCommandUnknown].
func (c OSCCommand) UnknownTruncated() (truncated bool, ok bool) {
	var v uint8
	if c.Data(OSCDataUnknownTruncated, unsafe.Pointer(&v)) != nil {
		return false, false
	}
	return v != 0, true
}

// UnknownTerminator returns how an unknown command ended. Use the same
// terminator when replying to the command. The ok result is false if c is not
// [OSCCommandUnknown].
func (c OSCCommand) UnknownTerminator() (terminator OSCTerminator, ok bool) {
	var v vt.GhosttyOscTerminator
	if c.Data(OSCDataUnknownTerminator, unsafe.Pointer(&v)) != nil {
		return 0, false
	}
	return OSCTerminator(v), true
}
