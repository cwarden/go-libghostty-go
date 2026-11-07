package libghostty

// Trampolines for terminal effects.
//
// Each top-level Go function here is passed to libghostty as a function
// pointer through vt.Fp. libghostty calls it with a userdata void*, which is
// a handle (see ccgo.go) that identifies the owning *Terminal. The
// trampoline recovers the Terminal and dispatches to the user-supplied Go
// effect handler. Each trampoline's signature is the translation of the C
// function pointer type: the TLS first, then the terminal, the userdata and
// the C parameters, with bool as uint8 and pointers as uintptr.

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// syncEffects registers or clears each C effect based on whether
// the corresponding Go effect handler is set.
func (t *Terminal) syncEffects() {
	tls := getTLS()
	defer putTLS(tls)

	// Install userdata before registering the first callback. Terminals without
	// callbacks do not need a handle. The handle is the userdata value itself,
	// not the address of anything.
	if t.handle == 0 && t.hasEffects() {
		t.handle = newHandle(t)
		vt.Xghostty_terminal_set(tls, t.ptr, vt.GHOSTTY_TERMINAL_OPT_USERDATA, uintptr(t.handle))
	}

	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_WRITE_PTY, t.onWritePty != nil, vt.Fp(goWritePtyTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_BELL, t.onBell != nil, vt.Fp(goBellTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_CLIPBOARD_WRITE, t.onClipboardWrite != nil, vt.Fp(goClipboardWriteTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_CLIPBOARD_READ, t.onClipboardRead != nil, vt.Fp(goClipboardReadTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_DESKTOP_NOTIFICATION, t.onDesktopNotification != nil, vt.Fp(goDesktopNotificationTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_TITLE_CHANGED, t.onTitleChanged != nil, vt.Fp(goTitleChangedTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_PWD_CHANGED, t.onPwdChanged != nil, vt.Fp(goPwdChangedTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_PROGRESS_REPORT, t.onProgressReport != nil, vt.Fp(goProgressReportTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_PROGRAM_STATUS, t.onProgramStatus != nil, vt.Fp(goProgramStatusTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_ENQUIRY, t.onEnquiry != nil, vt.Fp(goEnquiryTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_XTVERSION, t.onXtversion != nil, vt.Fp(goXtversionTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_SIZE, t.onSize != nil, vt.Fp(goSizeTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_COLOR_SCHEME, t.onColorScheme != nil, vt.Fp(goColorSchemeTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_DEVICE_ATTRIBUTES, t.onDeviceAttributes != nil, vt.Fp(goDeviceAttributesTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_UNKNOWN_SEQUENCE, t.onUnknownSequence != nil, vt.Fp(goUnknownSequenceTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_RENDER_HOLD, t.onRenderHold != nil, vt.Fp(goRenderHoldTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_SEMANTIC_PROMPT, t.onSemanticPrompt != nil, vt.Fp(goSemanticPromptTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_RESET, t.onReset != nil, vt.Fp(goResetTrampoline))
	t.setEffect(tls, vt.GHOSTTY_TERMINAL_OPT_RESIZE_WINDOW, t.onWindowResize != nil, vt.Fp(goWindowResizeTrampoline))
}

// setEffect registers fp for option when set is true, and clears the
// option otherwise.
func (t *Terminal) setEffect(tls *libc.TLS, option vt.GhosttyTerminalOption, set bool, fp uintptr) {
	if !set {
		fp = 0
	}
	vt.Xghostty_terminal_set(tls, t.ptr, option, fp)
}

// hasEffects reports whether any native effect trampoline needs to recover
// this Terminal through userdata.
func (t *Terminal) hasEffects() bool {
	return t.onWritePty != nil ||
		t.onBell != nil ||
		t.onClipboardWrite != nil ||
		t.onClipboardRead != nil ||
		t.onDesktopNotification != nil ||
		t.onTitleChanged != nil ||
		t.onPwdChanged != nil ||
		t.onProgressReport != nil ||
		t.onProgramStatus != nil ||
		t.onEnquiry != nil ||
		t.onXtversion != nil ||
		t.onSize != nil ||
		t.onColorScheme != nil ||
		t.onDeviceAttributes != nil ||
		t.onUnknownSequence != nil ||
		t.onRenderHold != nil ||
		t.onSemanticPrompt != nil ||
		t.onReset != nil ||
		t.onWindowResize != nil
}

// terminalFromUserdata recovers a *Terminal from the C userdata value.
func terminalFromUserdata(userdata uintptr) *Terminal {
	return handle(userdata).value().(*Terminal)
}

// goWritePtyTrampoline is the GhosttyTerminalWritePtyFn.
func goWritePtyTrampoline(_ *libc.TLS, _ uintptr, userdata uintptr, data uintptr, length uint64) {
	t := terminalFromUserdata(userdata)
	if t.onWritePty != nil {
		t.onWritePty(t, goBytes(data, int(length)))
	}
}

// goBellTrampoline is the GhosttyTerminalBellFn.
func goBellTrampoline(_ *libc.TLS, _ uintptr, userdata uintptr) {
	t := terminalFromUserdata(userdata)
	if t.onBell != nil {
		t.onBell(t)
	}
}

// goClipboardWriteTrampoline is the GhosttyTerminalClipboardWriteFn. write
// is a borrowed *GhosttyClipboardWrite.
func goClipboardWriteTrampoline(tls *libc.TLS, _ uintptr, userdata uintptr, write uintptr) {
	t := terminalFromUserdata(userdata)

	if write == 0 {
		return
	}
	if t.onClipboardWrite == nil {
		replyClipboardWrite(tls, write, ClipboardWriteReply{Result: ClipboardWriteUnsupported})
		return
	}

	w := (*vt.GhosttyClipboardWrite)(unsafe.Pointer(write))
	contents, ok := copyClipboardContents(w.Fcontents, w.Fcontents_len)
	if !ok {
		replyClipboardWrite(tls, write, ClipboardWriteReply{Result: ClipboardWriteInvalidData})
		return
	}
	name, ok := copyGhosttyString(w.Fname)
	if !ok {
		replyClipboardWrite(tls, write, ClipboardWriteReply{Result: ClipboardWriteInvalidData})
		return
	}

	reply := t.onClipboardWrite(t, ClipboardWrite{
		Location:    ClipboardLocation(w.Flocation),
		Contents:    contents,
		Name:        string(name),
		Granted:     w.Fgranted != 0,
		CanRemember: w.Fcan_remember != 0,
	})
	replyClipboardWrite(tls, write, reply)
}

// replyClipboardWrite answers a borrowed C clipboard-write request before its
// callback lifetime ends. The reply struct is built in TLS memory and passed
// to the request's reply function pointer.
func replyClipboardWrite(tls *libc.TLS, write uintptr, reply ClipboardWriteReply) {
	w := (*vt.GhosttyClipboardWrite)(unsafe.Pointer(write))
	if w.Freply == 0 {
		return
	}
	const size = int(unsafe.Sizeof(vt.GhosttyClipboardWriteReply{}))
	p := tls.Alloc(size)
	defer tls.Free(size)
	*(*vt.GhosttyClipboardWriteReply)(unsafe.Pointer(p)) = vt.GhosttyClipboardWriteReply{
		Fsize:     uint64(size),
		Fresult:   vt.GhosttyClipboardWriteResult(reply.Result),
		Fremember: cBool(reply.Remember),
	}
	(*(*func(*libc.TLS, uintptr, uintptr))(unsafe.Pointer(&struct{ uintptr }{w.Freply})))(tls, write, p)
}

// goClipboardReadTrampoline is the GhosttyTerminalClipboardReadFn. read is
// a borrowed *GhosttyClipboardRead.
func goClipboardReadTrampoline(tls *libc.TLS, _ uintptr, userdata uintptr, read uintptr) {
	t := terminalFromUserdata(userdata)

	// Without a descriptor there is no reply function pointer, so returning
	// without a reply deliberately selects libghostty's denied fallback.
	if read == 0 {
		return
	}
	if t.onClipboardRead == nil {
		replyClipboardRead(tls, read, ClipboardReadReply{Result: ClipboardReadUnsupported})
		return
	}

	r := (*vt.GhosttyClipboardRead)(unsafe.Pointer(read))
	mimes, ok := copyGhosttyStringArray(r.Fmimes, r.Fmimes_len)
	if !ok {
		replyClipboardRead(tls, read, ClipboardReadReply{Result: ClipboardReadIOError})
		return
	}
	name, ok := copyGhosttyString(r.Fname)
	if !ok {
		replyClipboardRead(tls, read, ClipboardReadReply{Result: ClipboardReadIOError})
		return
	}

	reply := t.onClipboardRead(t, ClipboardRead{
		Location:    ClipboardLocation(r.Flocation),
		MIMEs:       mimes,
		List:        r.Flist != 0,
		Name:        string(name),
		Granted:     r.Fgranted != 0,
		CanRemember: r.Fcan_remember != 0,
	})
	replyClipboardRead(tls, read, reply)
}

// replyClipboardRead copies a Go reply into C-owned temporary buffers and
// invokes the request's synchronous reply function. Allocation failures are
// reported to the running program as clipboard I/O errors.
func replyClipboardRead(tls *libc.TLS, read uintptr, reply ClipboardReadReply) {
	r := (*vt.GhosttyClipboardRead)(unsafe.Pointer(read))
	if r.Freply == 0 {
		return
	}
	const size = int(unsafe.Sizeof(vt.GhosttyClipboardReadReply{}))
	p := tls.Alloc(size)
	defer tls.Free(size)
	cReply := (*vt.GhosttyClipboardReadReply)(unsafe.Pointer(p))
	*cReply = vt.GhosttyClipboardReadReply{
		Fsize:     uint64(size),
		Fresult:   vt.GhosttyClipboardReadResult(reply.Result),
		Fremember: cBool(reply.Remember),
	}

	var contents *cGhosttyClipboardContents
	var available *cGhosttyStringArray
	if reply.Result == ClipboardReadSuccess {
		var err error
		contents, err = newCGhosttyClipboardContents(reply.Contents)
		if err == nil {
			availableValues := make([][]byte, len(reply.Available))
			for i, mime := range reply.Available {
				availableValues[i] = []byte(mime)
			}
			available, err = newCGhosttyStringArray(availableValues)
		}
		if err != nil {
			if contents != nil {
				contents.close()
				contents = nil
			}
			cReply.Fresult = vt.GHOSTTY_CLIPBOARD_READ_RESULT_IO_ERROR
			cReply.Fremember = cFalse
		} else {
			cReply.Fcontents = contents.ptr
			cReply.Fcontents_len = uint64(len(reply.Contents))
			cReply.Favailable = available.ptr
			cReply.Favailable_len = uint64(len(reply.Available))
		}
	}

	(*(*func(*libc.TLS, uintptr, uintptr))(unsafe.Pointer(&struct{ uintptr }{r.Freply})))(tls, read, p)
	if available != nil {
		available.close()
	}
	if contents != nil {
		contents.close()
	}
}

// copyClipboardContents copies borrowed C clipboard representations into
// Go-owned memory. values is the C address of a GhosttyClipboardContent
// array.
func copyClipboardContents(values uintptr, count uint64) ([]ClipboardContent, bool) {
	length, ok := ghosttySizeToInt(count)
	if !ok || (length > 0 && values == 0) {
		return nil, false
	}

	result := make([]ClipboardContent, length)
	if length == 0 {
		return result, true
	}
	for i, content := range unsafe.Slice((*vt.GhosttyClipboardContent)(unsafe.Pointer(values)), length) {
		mime, valid := copyGhosttyString(content.Fmime)
		if !valid {
			return nil, false
		}
		data, valid := copyGhosttyString(content.Fdata)
		if !valid {
			return nil, false
		}
		result[i] = ClipboardContent{MIME: string(mime), Data: data}
	}
	return result, true
}

// cGhosttyClipboardContents owns a C array of clipboard-content descriptors
// and every C string referenced by the array.
type cGhosttyClipboardContents struct {
	// ptr is the C address of the descriptor array, or 0 when empty.
	ptr            uintptr
	descriptorSize uintptr
	strings        *cGhosttyStringArray
}

// newCGhosttyClipboardContents copies Go clipboard representations into C
// memory suitable for a synchronous GhosttyClipboardReadReply.
func newCGhosttyClipboardContents(values []ClipboardContent) (*cGhosttyClipboardContents, error) {
	owner := &cGhosttyClipboardContents{}
	if len(values) == 0 {
		return owner, nil
	}
	if len(values) > int(^uint(0)>>1)/2 {
		return nil, &Error{Result: ResultLimitExceeded}
	}

	strings := make([][]byte, len(values)*2)
	for i, value := range values {
		strings[i*2] = []byte(value.MIME)
		strings[i*2+1] = value.Data
	}
	stringOwner, err := newCGhosttyStringArray(strings)
	if err != nil {
		return nil, err
	}
	owner.strings = stringOwner

	elementSize := unsafe.Sizeof(vt.GhosttyClipboardContent{})
	if uintptr(len(values)) > ^uintptr(0)/elementSize {
		owner.close()
		return nil, &Error{Result: ResultLimitExceeded}
	}
	owner.descriptorSize = uintptr(len(values)) * elementSize
	// Each GhosttyClipboardContent in the array contains pointers, so the
	// array must start out zeroed. See allocZeroed.
	descriptorPtr := allocZeroed(owner.descriptorSize)
	if descriptorPtr == nil {
		owner.close()
		return nil, &Error{Result: ResultOutOfMemory}
	}
	owner.ptr = uintptr(descriptorPtr)

	descriptors := unsafe.Slice((*vt.GhosttyClipboardContent)(descriptorPtr), len(values))
	stringDescriptors := unsafe.Slice((*vt.GhosttyString)(unsafe.Pointer(stringOwner.ptr)), len(strings))
	for i := range values {
		descriptors[i] = vt.GhosttyClipboardContent{
			Fmime: stringDescriptors[i*2],
			Fdata: stringDescriptors[i*2+1],
		}
	}
	return owner, nil
}

// close releases the descriptor array and all of its referenced strings.
func (c *cGhosttyClipboardContents) close() {
	if c == nil {
		return
	}
	Free(unsafe.Pointer(c.ptr), c.descriptorSize)
	if c.strings != nil {
		c.strings.close()
	}
	c.ptr = 0
	c.descriptorSize = 0
	c.strings = nil
}

// goDesktopNotificationTrampoline is the
// GhosttyTerminalDesktopNotificationFn. notification is a borrowed
// *GhosttyTerminalDesktopNotification.
func goDesktopNotificationTrampoline(
	_ *libc.TLS,
	_ uintptr,
	userdata uintptr,
	notification uintptr,
) {
	t := terminalFromUserdata(userdata)
	if t.onDesktopNotification == nil {
		return
	}

	if notification == 0 {
		return
	}

	n := (*vt.GhosttyTerminalDesktopNotification)(unsafe.Pointer(notification))
	title, ok := copyGhosttyString(n.Ftitle)
	if !ok {
		return
	}
	body, ok := copyGhosttyString(n.Fbody)
	if !ok {
		return
	}

	t.onDesktopNotification(t, TerminalDesktopNotification{
		Title: string(title),
		Body:  string(body),
	})
}

// goTitleChangedTrampoline is the GhosttyTerminalTitleChangedFn.
func goTitleChangedTrampoline(_ *libc.TLS, _ uintptr, userdata uintptr) {
	t := terminalFromUserdata(userdata)
	if t.onTitleChanged != nil {
		t.onTitleChanged(t)
	}
}

// goPwdChangedTrampoline is the GhosttyTerminalPwdChangedFn.
func goPwdChangedTrampoline(_ *libc.TLS, _ uintptr, userdata uintptr) {
	t := terminalFromUserdata(userdata)
	if t.onPwdChanged != nil {
		t.onPwdChanged(t)
	}
}

// goProgressReportTrampoline is the GhosttyTerminalProgressReportFn.
// report is a borrowed *GhosttyTerminalProgressReport.
func goProgressReportTrampoline(
	_ *libc.TLS,
	_ uintptr,
	userdata uintptr,
	report uintptr,
) {
	t := terminalFromUserdata(userdata)
	if t.onProgressReport == nil {
		return
	}

	if report == 0 {
		return
	}

	r := (*vt.GhosttyTerminalProgressReport)(unsafe.Pointer(report))
	t.onProgressReport(t, TerminalProgressReport{
		State:    TerminalProgressState(r.Fstate),
		Progress: int8(r.Fprogress),
	})
}

// goProgramStatusTrampoline is the GhosttyTerminalProgramStatusFn. report
// is a borrowed *GhosttyTerminalProgramStatus.
func goProgramStatusTrampoline(
	_ *libc.TLS,
	_ uintptr,
	userdata uintptr,
	reportPtr uintptr,
) {
	t := terminalFromUserdata(userdata)
	if t.onProgramStatus == nil {
		return
	}

	if reportPtr == 0 {
		return
	}
	report := (*vt.GhosttyTerminalProgramStatus)(unsafe.Pointer(reportPtr))

	// The strings are only valid during the callback, so copy them all into
	// Go memory. A malformed string drops the whole report rather than
	// delivering a partial record, since reports replace records entirely.
	id, ok := copyGhosttyString(report.Fid)
	if !ok {
		return
	}
	app, ok := copyGhosttyString(report.Fapp)
	if !ok {
		return
	}
	title, ok := copyGhosttyString(report.Ftitle)
	if !ok {
		return
	}
	message, ok := copyGhosttyString(report.Fmessage)
	if !ok {
		return
	}

	t.onProgramStatus(t, ProgramStatus{
		State:    ProgramStatusState(report.Fstate),
		Kind:     ProgramStatusKind(report.Fkind),
		Progress: int8(report.Fprogress),
		ID:       string(id),
		App:      string(app),
		Title:    string(title),
		Message:  string(message),
	})
}

// goUnknownSequenceTrampoline is the GhosttyTerminalUnknownSequenceFn.
// sequence is a borrowed *GhosttyTerminalUnknownSequence.
func goUnknownSequenceTrampoline(
	_ *libc.TLS,
	_ uintptr,
	userdata uintptr,
	sequence uintptr,
) {
	t := terminalFromUserdata(userdata)
	if t.onUnknownSequence == nil || sequence == 0 {
		return
	}

	seq := (*vt.GhosttyTerminalUnknownSequence)(unsafe.Pointer(sequence))
	value := TerminalUnknownSequence{
		Tag: TerminalUnknownSequenceTag(seq.Ftag),
	}
	// Only the union member that matches the tag is read; every member
	// starts at the union's address.
	if value.Tag == TerminalUnknownSequenceAPC {
		apc := (*vt.GhosttyTerminalUnknownStringSequence)(unsafe.Pointer(&seq.Fvalue))
		content, ok := copyGhosttyString(apc.Fcontent)
		if !ok {
			return
		}
		value.APC = TerminalUnknownStringSequence{
			Truncated: apc.Ftruncated != 0,
			Content:   content,
		}
	} else if value.Tag == TerminalUnknownSequenceOSC {
		osc := (*vt.GhosttyTerminalUnknownOscSequence)(unsafe.Pointer(&seq.Fvalue))
		content, ok := copyGhosttyString(osc.Fcontent)
		if !ok {
			return
		}
		value.OSC = TerminalUnknownOSCSequence{
			Truncated:  osc.Ftruncated != 0,
			Content:    content,
			Terminator: OSCTerminator(osc.Fterminator),
		}
	}

	t.onUnknownSequence(t, value)
}

// goRenderHoldTrampoline is the GhosttyTerminalRenderHoldFn.
func goRenderHoldTrampoline(_ *libc.TLS, _ uintptr, userdata uintptr, held uint8) {
	t := terminalFromUserdata(userdata)
	if t.onRenderHold != nil {
		t.onRenderHold(t, held != 0)
	}
}

// goSemanticPromptTrampoline is the GhosttyTerminalSemanticPromptFn. event
// is a borrowed *GhosttyTerminalSemanticPrompt.
func goSemanticPromptTrampoline(
	_ *libc.TLS,
	_ uintptr,
	userdata uintptr,
	eventPtr uintptr,
) {
	t := terminalFromUserdata(userdata)
	if t.onSemanticPrompt == nil {
		return
	}

	if eventPtr == 0 {
		return
	}
	event := (*vt.GhosttyTerminalSemanticPrompt)(unsafe.Pointer(eventPtr))

	// The strings are only valid during this call, so copy them into Go
	// memory before passing the event to the application.
	command, ok := copyGhosttyString(event.Fcommand)
	if !ok {
		return
	}
	errText, ok := copyGhosttyString(event.Ferror1)
	if !ok {
		return
	}

	t.onSemanticPrompt(t, TerminalSemanticPrompt{
		Kind:        SemanticPromptKind(event.Fkind),
		PromptKind:  PromptKind(event.Fprompt_kind),
		HasExitCode: event.Fhas_exit_code != 0,
		ExitCode:    int(event.Fexit_code),
		Command:     string(command),
		Error:       string(errText),
	})
}

// goResetTrampoline is the GhosttyTerminalResetFn.
func goResetTrampoline(_ *libc.TLS, _ uintptr, userdata uintptr) {
	t := terminalFromUserdata(userdata)
	if t.onReset != nil {
		t.onReset(t)
	}
}

// goWindowResizeTrampoline is the GhosttyTerminalResizeWindowFn. request
// is a borrowed *GhosttyTerminalResizeWindow.
func goWindowResizeTrampoline(
	_ *libc.TLS,
	_ uintptr,
	userdata uintptr,
	requestPtr uintptr,
) {
	t := terminalFromUserdata(userdata)
	if t.onWindowResize == nil || requestPtr == 0 {
		return
	}
	request := (*vt.GhosttyTerminalResizeWindow)(unsafe.Pointer(requestPtr))

	// Only read the union member that matches the tag. Requests with an
	// unknown tag are still delivered, with only Kind set, so the callback
	// can see and ignore them.
	resize := WindowResize{Kind: WindowResizeKind(request.Ftag)}
	if resize.Kind == WindowResizeGrid {
		grid := (*vt.GhosttyTerminalResizeWindowGrid)(unsafe.Pointer(&request.Fvalue))
		resize.Rows = grid.Frows
		resize.Columns = grid.Fcolumns
	}

	t.onWindowResize(t, resize)
}

// goEnquiryTrampoline is the GhosttyTerminalEnquiryFn.
func goEnquiryTrampoline(_ *libc.TLS, _ uintptr, userdata uintptr) vt.GhosttyString {
	t := terminalFromUserdata(userdata)
	if t.onEnquiry == nil {
		return vt.GhosttyString{}
	}
	return t.effectString(t.onEnquiry(t))
}

// goXtversionTrampoline is the GhosttyTerminalXtversionFn.
func goXtversionTrampoline(_ *libc.TLS, _ uintptr, userdata uintptr) vt.GhosttyString {
	t := terminalFromUserdata(userdata)
	if t.onXtversion == nil {
		return vt.GhosttyString{}
	}
	return t.effectString([]byte(t.onXtversion(t)))
}

// goSizeTrampoline is the GhosttyTerminalSizeFn. outSize is a
// *GhosttySizeReportSize to fill in.
func goSizeTrampoline(_ *libc.TLS, _ uintptr, userdata uintptr, outSize uintptr) uint8 {
	t := terminalFromUserdata(userdata)
	if t.onSize == nil {
		return cFalse
	}
	size, ok := t.onSize(t)
	if !ok {
		return cFalse
	}
	out := (*vt.GhosttySizeReportSize)(unsafe.Pointer(outSize))
	out.Frows = size.Rows
	out.Fcolumns = size.Columns
	out.Fcell_width = size.CellWidth
	out.Fcell_height = size.CellHeight
	return cTrue
}

// goColorSchemeTrampoline is the GhosttyTerminalColorSchemeFn. outScheme is
// a *GhosttyColorScheme to fill in.
func goColorSchemeTrampoline(_ *libc.TLS, _ uintptr, userdata uintptr, outScheme uintptr) uint8 {
	t := terminalFromUserdata(userdata)
	if t.onColorScheme == nil {
		return cFalse
	}
	scheme, ok := t.onColorScheme(t)
	if !ok {
		return cFalse
	}
	*(*vt.GhosttyColorScheme)(unsafe.Pointer(outScheme)) = vt.GhosttyColorScheme(scheme)
	return cTrue
}

// goDeviceAttributesTrampoline is the GhosttyTerminalDeviceAttributesFn.
// outAttrs is a *GhosttyDeviceAttributes to fill in.
func goDeviceAttributesTrampoline(_ *libc.TLS, _ uintptr, userdata uintptr, outAttrs uintptr) uint8 {
	t := terminalFromUserdata(userdata)
	if t.onDeviceAttributes == nil {
		return cFalse
	}
	attrs, ok := t.onDeviceAttributes(t)
	if !ok {
		return cFalse
	}
	out := (*vt.GhosttyDeviceAttributes)(unsafe.Pointer(outAttrs))

	// Primary (DA1).
	out.Fprimary.Fconformance_level = attrs.Primary.ConformanceLevel
	out.Fprimary.Fnum_features = uint64(attrs.Primary.NumFeatures)
	for i := 0; i < attrs.Primary.NumFeatures && i < 64; i++ {
		out.Fprimary.Ffeatures[i] = attrs.Primary.Features[i]
	}

	// Secondary (DA2).
	out.Fsecondary.Fdevice_type = attrs.Secondary.DeviceType
	out.Fsecondary.Ffirmware_version = attrs.Secondary.FirmwareVersion
	out.Fsecondary.From_cartridge = attrs.Secondary.ROMCartridge

	// Tertiary (DA3).
	out.Ftertiary.Funit_id = attrs.Tertiary.UnitID

	return cTrue
}

// effectString copies data into C memory allocated via the libghostty
// allocator, updates effectBuf/effectBufLen, and returns a
// GhosttyString pointing to it. The previous effectBuf is freed.
// Returns a zero-length GhosttyString if data is empty or the
// allocation fails.
func (t *Terminal) effectString(data []byte) vt.GhosttyString {
	if t.effectBuf != nil {
		Free(t.effectBuf, t.effectBufLen)
		t.effectBuf = nil
		t.effectBufLen = 0
	}

	if len(data) == 0 {
		return vt.GhosttyString{}
	}

	n := uintptr(len(data))
	cmem := Alloc(n)
	if cmem == nil {
		// The callers are C callbacks that can only return a string,
		// so there is no way to report an error. An empty string
		// sends no reply, which is better than crashing when memory
		// runs out.
		return vt.GhosttyString{}
	}
	copy(unsafe.Slice((*byte)(cmem), n), data)
	t.effectBuf = cmem
	t.effectBufLen = n
	return vt.GhosttyString{
		Fptr:  uintptr(cmem),
		Flen1: uint64(n),
	}
}
