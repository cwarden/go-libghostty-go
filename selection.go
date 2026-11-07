package libghostty

// Selection helpers wrapping the C API from selection.h.

import (
	"errors"
	"runtime"
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// SelectionOrder describes the ordering of selection endpoints in
// terminal coordinates.
// C: GhosttySelectionOrder
type SelectionOrder int

const (
	// SelectionOrderForward means Start is before End in top-left to
	// bottom-right order.
	SelectionOrderForward SelectionOrder = vt.GHOSTTY_SELECTION_ORDER_FORWARD

	// SelectionOrderReverse means End is before Start in top-left to
	// bottom-right order.
	SelectionOrderReverse SelectionOrder = vt.GHOSTTY_SELECTION_ORDER_REVERSE

	// SelectionOrderMirroredForward means a rectangular selection from
	// top-right to bottom-left.
	SelectionOrderMirroredForward SelectionOrder = vt.GHOSTTY_SELECTION_ORDER_MIRRORED_FORWARD

	// SelectionOrderMirroredReverse means a rectangular selection from
	// bottom-left to top-right.
	SelectionOrderMirroredReverse SelectionOrder = vt.GHOSTTY_SELECTION_ORDER_MIRRORED_REVERSE
)

// SelectionAdjust describes an operation used to adjust a selection's
// logical end endpoint.
// C: GhosttySelectionAdjust
type SelectionAdjust int

const (
	// SelectionAdjustLeft moves left to the previous non-empty cell,
	// wrapping upward.
	SelectionAdjustLeft SelectionAdjust = vt.GHOSTTY_SELECTION_ADJUST_LEFT

	// SelectionAdjustRight moves right to the next non-empty cell,
	// wrapping downward.
	SelectionAdjustRight SelectionAdjust = vt.GHOSTTY_SELECTION_ADJUST_RIGHT

	// SelectionAdjustUp moves up one row at the current column, or to the
	// beginning of the line if already at the top.
	SelectionAdjustUp SelectionAdjust = vt.GHOSTTY_SELECTION_ADJUST_UP

	// SelectionAdjustDown moves down to the next non-blank row at the
	// current column, or to the end of the line if none exists.
	SelectionAdjustDown SelectionAdjust = vt.GHOSTTY_SELECTION_ADJUST_DOWN

	// SelectionAdjustHome moves to the top-left cell of the screen.
	SelectionAdjustHome SelectionAdjust = vt.GHOSTTY_SELECTION_ADJUST_HOME

	// SelectionAdjustEnd moves to the right edge of the last non-blank row
	// on the screen.
	SelectionAdjustEnd SelectionAdjust = vt.GHOSTTY_SELECTION_ADJUST_END

	// SelectionAdjustPageUp moves up by one terminal page height, or to
	// home if that would move past the top.
	SelectionAdjustPageUp SelectionAdjust = vt.GHOSTTY_SELECTION_ADJUST_PAGE_UP

	// SelectionAdjustPageDown moves down by one terminal page height, or
	// to end if that would move past the bottom.
	SelectionAdjustPageDown SelectionAdjust = vt.GHOSTTY_SELECTION_ADJUST_PAGE_DOWN

	// SelectionAdjustBeginningOfLine moves to the left edge of the current
	// line.
	SelectionAdjustBeginningOfLine SelectionAdjust = vt.GHOSTTY_SELECTION_ADJUST_BEGINNING_OF_LINE

	// SelectionAdjustEndOfLine moves to the right edge of the current line.
	SelectionAdjustEndOfLine SelectionAdjust = vt.GHOSTTY_SELECTION_ADJUST_END_OF_LINE
)

// SelectionGesture is an opaque handle to mutable state for interpreting
// pointer gestures as terminal selection snapshots. It is not safe for
// concurrent use and should be serialized with the terminal it is used with.
// C: GhosttySelectionGesture
type SelectionGesture struct {
	ptr vt.GhosttySelectionGesture
}

// SelectionGestureEvent is an opaque, reusable input event for a selection
// gesture. Its event type is fixed at creation time, while its option values
// may be changed between calls.
// C: GhosttySelectionGestureEvent
type SelectionGestureEvent struct {
	ptr vt.GhosttySelectionGestureEvent
}

// SelectionGestureBehavior is the selection behavior chosen for a gesture's
// click sequence.
// C: GhosttySelectionGestureBehavior
type SelectionGestureBehavior int

const (
	// SelectionGestureBehaviorCell performs cell-granular drag selection.
	SelectionGestureBehaviorCell SelectionGestureBehavior = vt.GHOSTTY_SELECTION_GESTURE_BEHAVIOR_CELL

	// SelectionGestureBehaviorWord performs word selection on press and
	// word-granular drag selection.
	SelectionGestureBehaviorWord SelectionGestureBehavior = vt.GHOSTTY_SELECTION_GESTURE_BEHAVIOR_WORD

	// SelectionGestureBehaviorLine performs line selection on press and
	// line-granular drag selection.
	SelectionGestureBehaviorLine SelectionGestureBehavior = vt.GHOSTTY_SELECTION_GESTURE_BEHAVIOR_LINE

	// SelectionGestureBehaviorOutput performs semantic command-output selection
	// on press and drag.
	SelectionGestureBehaviorOutput SelectionGestureBehavior = vt.GHOSTTY_SELECTION_GESTURE_BEHAVIOR_OUTPUT
)

// SelectionGestureBehaviors configures the behavior selected for single-,
// double-, and triple-click gesture sequences.
// C: GhosttySelectionGestureBehaviors
type SelectionGestureBehaviors struct {
	// SingleClick is the behavior for single-click selection gestures.
	SingleClick SelectionGestureBehavior

	// DoubleClick is the behavior for double-click selection gestures.
	DoubleClick SelectionGestureBehavior

	// TripleClick is the behavior for triple-click selection gestures.
	TripleClick SelectionGestureBehavior
}

// toC converts Go selection gesture behaviors to C.
func (b SelectionGestureBehaviors) toC() vt.GhosttySelectionGestureBehaviors {
	return vt.GhosttySelectionGestureBehaviors{
		Fsingle_click: vt.GhosttySelectionGestureBehavior(b.SingleClick),
		Fdouble_click: vt.GhosttySelectionGestureBehavior(b.DoubleClick),
		Ftriple_click: vt.GhosttySelectionGestureBehavior(b.TripleClick),
	}
}

// SelectionGestureGeometry is the display geometry used to interpret drag and
// autoscroll selection gesture events.
// C: GhosttySelectionGestureGeometry
type SelectionGestureGeometry struct {
	// Columns is the number of columns in the rendered terminal grid. It must
	// be non-zero.
	Columns uint32

	// CellWidth is the width of one terminal cell in surface pixels. It must be
	// non-zero.
	CellWidth uint32

	// PaddingLeft is the left padding before the terminal grid begins in surface
	// pixels.
	PaddingLeft uint32

	// ScreenHeight is the height of the rendered terminal surface in surface
	// pixels. It must be non-zero.
	ScreenHeight uint32
}

// toC converts Go selection gesture geometry to C.
func (g SelectionGestureGeometry) toC() vt.GhosttySelectionGestureGeometry {
	return vt.GhosttySelectionGestureGeometry{
		Fcolumns:       g.Columns,
		Fcell_width:    g.CellWidth,
		Fpadding_left:  g.PaddingLeft,
		Fscreen_height: g.ScreenHeight,
	}
}

// SelectionGestureAutoscroll is the current autoscroll direction for an active
// selection drag gesture.
// C: GhosttySelectionGestureAutoscroll
type SelectionGestureAutoscroll int

const (
	// SelectionGestureAutoscrollNone means no selection autoscroll is requested.
	SelectionGestureAutoscrollNone SelectionGestureAutoscroll = vt.GHOSTTY_SELECTION_GESTURE_AUTOSCROLL_NONE

	// SelectionGestureAutoscrollUp means selection dragging should autoscroll
	// the viewport upward.
	SelectionGestureAutoscrollUp SelectionGestureAutoscroll = vt.GHOSTTY_SELECTION_GESTURE_AUTOSCROLL_UP

	// SelectionGestureAutoscrollDown means selection dragging should autoscroll
	// the viewport downward.
	SelectionGestureAutoscrollDown SelectionGestureAutoscroll = vt.GHOSTTY_SELECTION_GESTURE_AUTOSCROLL_DOWN
)

// SelectionGestureData identifies data fields readable from a selection
// gesture.
// C: GhosttySelectionGestureData
type SelectionGestureData int

const (
	// SelectionGestureDataClickCount reads the current click count as *uint8.
	SelectionGestureDataClickCount SelectionGestureData = vt.GHOSTTY_SELECTION_GESTURE_DATA_CLICK_COUNT

	// SelectionGestureDataDragged reads whether the current or last left-click
	// gesture dragged as *bool.
	SelectionGestureDataDragged SelectionGestureData = vt.GHOSTTY_SELECTION_GESTURE_DATA_DRAGGED

	// SelectionGestureDataAutoscroll reads the current autoscroll request as a
	// *GhosttySelectionGestureAutoscroll-compatible value.
	SelectionGestureDataAutoscroll SelectionGestureData = vt.GHOSTTY_SELECTION_GESTURE_DATA_AUTOSCROLL

	// SelectionGestureDataBehavior reads the current gesture behavior as a
	// *GhosttySelectionGestureBehavior-compatible value.
	SelectionGestureDataBehavior SelectionGestureData = vt.GHOSTTY_SELECTION_GESTURE_DATA_BEHAVIOR

	// SelectionGestureDataAnchor reads the current left-click anchor as a
	// *GhosttyGridRef-compatible value. It returns ResultNoValue when there is
	// no valid active anchor.
	SelectionGestureDataAnchor SelectionGestureData = vt.GHOSTTY_SELECTION_GESTURE_DATA_ANCHOR
)

// SelectionGestureEventType is the fixed type of a selection gesture event.
// C: GhosttySelectionGestureEventType
type SelectionGestureEventType int

const (
	// SelectionGestureEventTypePress is a press event.
	SelectionGestureEventTypePress SelectionGestureEventType = vt.GHOSTTY_SELECTION_GESTURE_EVENT_TYPE_PRESS

	// SelectionGestureEventTypeRelease is a release event.
	SelectionGestureEventTypeRelease SelectionGestureEventType = vt.GHOSTTY_SELECTION_GESTURE_EVENT_TYPE_RELEASE

	// SelectionGestureEventTypeDrag is a drag event.
	SelectionGestureEventTypeDrag SelectionGestureEventType = vt.GHOSTTY_SELECTION_GESTURE_EVENT_TYPE_DRAG

	// SelectionGestureEventTypeAutoscrollTick is an autoscroll tick event.
	SelectionGestureEventTypeAutoscrollTick SelectionGestureEventType = vt.GHOSTTY_SELECTION_GESTURE_EVENT_TYPE_AUTOSCROLL_TICK

	// SelectionGestureEventTypeDeepPress is a deep press event.
	SelectionGestureEventTypeDeepPress SelectionGestureEventType = vt.GHOSTTY_SELECTION_GESTURE_EVENT_TYPE_DEEP_PRESS
)

// SelectionGestureEventOption identifies an option stored on a reusable
// selection gesture event.
// C: GhosttySelectionGestureEventOption
type SelectionGestureEventOption int

const (
	// SelectionGestureEventOptionRef stores a *GridRef pointer under the
	// pointer. It is required for press and drag events and optional for release.
	SelectionGestureEventOptionRef SelectionGestureEventOption = vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_REF

	// SelectionGestureEventOptionPosition stores a *SurfacePosition pointer
	// position.
	SelectionGestureEventOptionPosition SelectionGestureEventOption = vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_POSITION

	// SelectionGestureEventOptionRepeatDistance stores a *float64 maximum
	// repeat-click distance in pixels.
	SelectionGestureEventOptionRepeatDistance SelectionGestureEventOption = vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_REPEAT_DISTANCE

	// SelectionGestureEventOptionTimeNS stores a *uint64 monotonic event time in
	// nanoseconds.
	SelectionGestureEventOptionTimeNS SelectionGestureEventOption = vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_TIME_NS

	// SelectionGestureEventOptionRepeatIntervalNS stores a *uint64 maximum
	// interval between repeat clicks in nanoseconds.
	SelectionGestureEventOptionRepeatIntervalNS SelectionGestureEventOption = vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_REPEAT_INTERVAL_NS

	// SelectionGestureEventOptionWordBoundaryCodepoints stores borrowed word
	// boundary codepoints that libghostty copies into the event.
	SelectionGestureEventOptionWordBoundaryCodepoints SelectionGestureEventOption = vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_WORD_BOUNDARY_CODEPOINTS

	// SelectionGestureEventOptionBehaviors stores a *SelectionGestureBehaviors
	// behavior table.
	SelectionGestureEventOptionBehaviors SelectionGestureEventOption = vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_BEHAVIORS

	// SelectionGestureEventOptionRectangle stores a *bool indicating whether a
	// drag or autoscroll tick should produce a rectangular selection.
	SelectionGestureEventOptionRectangle SelectionGestureEventOption = vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_RECTANGLE

	// SelectionGestureEventOptionGeometry stores a *SelectionGestureGeometry
	// drag display geometry.
	SelectionGestureEventOptionGeometry SelectionGestureEventOption = vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_GEOMETRY

	// SelectionGestureEventOptionViewport stores a *PointCoordinate viewport
	// coordinate for an autoscroll tick.
	SelectionGestureEventOptionViewport SelectionGestureEventOption = vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_VIEWPORT
)

// SelectWordOptions configures [Terminal.SelectWord].
// C: GhosttyTerminalSelectWordOptions
type SelectWordOptions struct {
	// Ref is the grid reference under which to derive the word selection.
	Ref *GridRef

	// BoundaryCodepoints optionally replaces Ghostty's default word-boundary
	// codepoints. Leave nil to use Ghostty's defaults.
	BoundaryCodepoints []uint32
}

// SelectWordBetweenOptions configures [Terminal.SelectWordBetween].
// C: GhosttyTerminalSelectWordBetweenOptions
type SelectWordBetweenOptions struct {
	// Start is the starting grid reference for the inclusive search range.
	Start *GridRef

	// End is the ending grid reference for the inclusive search range.
	End *GridRef

	// BoundaryCodepoints optionally replaces Ghostty's default word-boundary
	// codepoints. Leave nil to use Ghostty's defaults.
	BoundaryCodepoints []uint32
}

// SelectLineOptions configures [Terminal.SelectLine].
// C: GhosttyTerminalSelectLineOptions
type SelectLineOptions struct {
	// Ref is the grid reference under which to derive the line selection.
	Ref *GridRef

	// Whitespace optionally replaces Ghostty's default line-trim whitespace
	// codepoints. Leave nil to use Ghostty's defaults.
	Whitespace []uint32

	// SemanticPromptBoundary bounds line selection at semantic prompt state
	// changes when true.
	SemanticPromptBoundary bool
}

// SelectOutputOptions configures [Terminal.SelectOutput].
type SelectOutputOptions struct {
	// Ref is a grid reference within the command output to select.
	Ref *GridRef
}

// toC converts a Go Selection into a C GhosttySelection.
func (s Selection) toC() vt.GhosttySelection {
	var cs vt.GhosttySelection
	cs.Fstart = s.Start.ref
	cs.Fend = s.End.ref
	cs.Frectangle = cBool(s.Rectangle)
	return cs
}

// cValue copies a C value into TLS stack memory and returns its address.
// The memory holds n bytes and must be released with tls.Free(n) after the
// call that uses it, in LIFO order with other TLS allocations.
func cValue[T any](tls *libc.TLS, value T) (p uintptr, n int) {
	n = int(unsafe.Sizeof(value))
	p = tls.Alloc(n)
	*(*T)(unsafe.Pointer(p)) = value
	return p, n
}

// cOut reserves zeroed TLS stack memory for an output value of type T. The
// memory must be released with tls.Free(n), as for cValue.
func cOut[T any](tls *libc.TLS) (p uintptr, n int) {
	var zero T
	return cValue(tls, zero)
}

// noValueNilSelection turns C APIs that report GHOSTTY_NO_VALUE into a nil
// selection result, matching the existing Terminal.Selection behavior.
func noValueNilSelection(err error) (*Selection, error) {
	if err == nil {
		return nil, nil
	}
	var ge *Error
	if errors.As(err, &ge) && ge.Result == ResultNoValue {
		return nil, nil
	}
	return nil, err
}

// NewSelectionGesture creates a selection gesture state machine. The returned
// gesture must be freed with [SelectionGesture.Close] when no longer needed.
func NewSelectionGesture() (*SelectionGesture, error) {
	tls := getTLS()
	defer putTLS(tls)
	out, n := cOut[uintptr](tls)
	defer tls.Free(n)
	if err := resultError(vt.Xghostty_selection_gesture_new(tls, 0, out)); err != nil {
		return nil, err
	}
	return &SelectionGesture{ptr: *(*uintptr)(unsafe.Pointer(out))}, nil
}

// Close frees the selection gesture. If the terminal most recently used with
// the gesture is still alive, pass it so libghostty can release tracked
// terminal references. Pass nil only after the terminal has already been freed.
func (g *SelectionGesture) Close(t *Terminal) {
	var terminal vt.GhosttyTerminal
	if t != nil {
		terminal = t.ptr
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_selection_gesture_free(tls, g.ptr, terminal)
}

// Reset cancels active selection gesture state and releases tracked terminal
// references without freeing the gesture.
func (g *SelectionGesture) Reset(t *Terminal) {
	var terminal vt.GhosttyTerminal
	if t != nil {
		terminal = t.ptr
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_selection_gesture_reset(tls, g.ptr, terminal)
}

// Event applies a selection gesture event to the terminal and returns the
// resulting selection snapshot. It returns nil when the event updates gesture
// state but does not currently produce a selection.
func (g *SelectionGesture) Event(t *Terminal, event *SelectionGestureEvent) (*Selection, error) {
	tls := getTLS()
	defer putTLS(tls)
	out, n := cOut[vt.GhosttySelection](tls)
	defer tls.Free(n)
	if err := resultError(vt.Xghostty_selection_gesture_event(tls, g.ptr, t.ptr, event.ptr, out)); err != nil {
		return noValueNilSelection(err)
	}

	sel := selectionFromC(*(*vt.GhosttySelection)(unsafe.Pointer(out)))
	return &sel, nil
}

// Get reads a raw data field from the selection gesture into value. The value
// pointer type must match the documented type for the data key; prefer the
// typed helpers such as [SelectionGesture.ClickCount] and
// [SelectionGesture.Anchor] when possible.
//
// C: ghostty_selection_gesture_get
func (g *SelectionGesture) Get(t *Terminal, data SelectionGestureData, value unsafe.Pointer) error {
	// Pinning keeps the value on the heap, where it does not move while the
	// translated code writes to it, and pinned there until the call returns.
	var pin runtime.Pinner
	pin.Pin(value)
	defer pin.Unpin()
	tls := getTLS()
	defer putTLS(tls)
	return g.get(tls, t, data, uintptr(value))
}

// get reads a raw data field into the C memory at value.
func (g *SelectionGesture) get(tls *libc.TLS, t *Terminal, data SelectionGestureData, value uintptr) error {
	return resultError(vt.Xghostty_selection_gesture_get(
		tls,
		g.ptr,
		t.ptr,
		vt.GhosttySelectionGestureData(data),
		value,
	))
}

// gestureGet reads one gesture data field of C type T into TLS stack memory,
// because the translated code must not write through a pointer to a Go
// variable.
func gestureGet[T any](g *SelectionGesture, t *Terminal, data SelectionGestureData) (T, error) {
	tls := getTLS()
	defer putTLS(tls)
	p, n := cOut[T](tls)
	defer tls.Free(n)
	var zero T
	if err := g.get(tls, t, data, p); err != nil {
		return zero, err
	}
	return *(*T)(unsafe.Pointer(p)), nil
}

// GetMulti reads multiple raw data fields from the selection gesture in one
// library call. Each element in values must point to storage with the type
// documented by the corresponding key. On partial failure, written is the
// failing index.
// C: ghostty_selection_gesture_get_multi
func (g *SelectionGesture) GetMulti(t *Terminal, keys []SelectionGestureData, values []unsafe.Pointer) (written int, err error) {
	if len(keys) != len(values) {
		return 0, errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return 0, nil
	}

	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[vt.GhosttySelectionGestureData](&args, keys, values)
	if err != nil {
		return 0, err
	}
	defer args.free()
	tls := getTLS()
	defer putTLS(tls)
	cWritten, n := cOut[uint64](tls)
	defer tls.Free(n)
	err = resultError(vt.Xghostty_selection_gesture_get_multi(
		tls,
		g.ptr,
		t.ptr,
		uint64(len(keys)),
		uintptr(unsafe.Pointer(cKeys)),
		uintptr(unsafe.Pointer(args.values)),
		cWritten,
	))
	return int(*(*uint64)(unsafe.Pointer(cWritten))), err
}

// ClickCount returns the current gesture click count. Zero means inactive.
func (g *SelectionGesture) ClickCount(t *Terminal) (uint8, error) {
	return gestureGet[uint8](g, t, SelectionGestureDataClickCount)
}

// Dragged reports whether the current or last left-click gesture dragged.
func (g *SelectionGesture) Dragged(t *Terminal) (bool, error) {
	v, err := gestureGet[uint8](g, t, SelectionGestureDataDragged)
	return v != 0, err
}

// Autoscroll returns the current autoscroll request for the gesture.
func (g *SelectionGesture) Autoscroll(t *Terminal) (SelectionGestureAutoscroll, error) {
	v, err := gestureGet[vt.GhosttySelectionGestureAutoscroll](g, t, SelectionGestureDataAutoscroll)
	return SelectionGestureAutoscroll(v), err
}

// Behavior returns the current gesture behavior.
func (g *SelectionGesture) Behavior(t *Terminal) (SelectionGestureBehavior, error) {
	v, err := gestureGet[vt.GhosttySelectionGestureBehavior](g, t, SelectionGestureDataBehavior)
	return SelectionGestureBehavior(v), err
}

// Anchor returns the current left-click anchor as an untracked grid reference
// snapshot. It returns nil when there is no valid active anchor.
func (g *SelectionGesture) Anchor(t *Terminal) (*GridRef, error) {
	ref, err := gestureGet[vt.GhosttyGridRef](g, t, SelectionGestureDataAnchor)
	if err != nil {
		var ge *Error
		if errors.As(err, &ge) && ge.Result == ResultNoValue {
			return nil, nil
		}
		return nil, err
	}
	return &GridRef{ref: ref}, nil
}

// NewSelectionGestureEvent creates a reusable gesture event of a fixed type.
// The returned event must be freed with [SelectionGestureEvent.Close].
func NewSelectionGestureEvent(eventType SelectionGestureEventType) (*SelectionGestureEvent, error) {
	tls := getTLS()
	defer putTLS(tls)
	out, n := cOut[uintptr](tls)
	defer tls.Free(n)
	if err := resultError(vt.Xghostty_selection_gesture_event_new(
		tls,
		0,
		out,
		vt.GhosttySelectionGestureEventType(eventType),
	)); err != nil {
		return nil, err
	}
	return &SelectionGestureEvent{ptr: *(*uintptr)(unsafe.Pointer(out))}, nil
}

// Close frees the selection gesture event. After this call, the event must not
// be used, except that calling Close again is a safe no-op.
func (e *SelectionGestureEvent) Close() {
	if e == nil || e.ptr == 0 {
		return
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_selection_gesture_event_free(tls, e.ptr)
	e.ptr = 0
}

// set writes one event option from a C value, or clears it when value is 0.
func (e *SelectionGestureEvent) set(tls *libc.TLS, option vt.GhosttySelectionGestureEventOption, value uintptr) error {
	return resultError(vt.Xghostty_selection_gesture_event_set(tls, e.ptr, option, value))
}

// setValue copies value into TLS stack memory and writes it as option.
func setEventValue[T any](e *SelectionGestureEvent, option vt.GhosttySelectionGestureEventOption, value T) error {
	tls := getTLS()
	defer putTLS(tls)
	p, n := cValue(tls, value)
	defer tls.Free(n)
	return e.set(tls, option, p)
}

// ClearOption clears a selection gesture event option.
func (e *SelectionGestureEvent) ClearOption(option SelectionGestureEventOption) error {
	tls := getTLS()
	defer putTLS(tls)
	return e.set(tls, vt.GhosttySelectionGestureEventOption(option), 0)
}

// SetRef sets the grid reference under the pointer for a press, drag, or
// release event.
func (e *SelectionGestureEvent) SetRef(ref *GridRef) error {
	return setEventValue(e, vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_REF, ref.ref)
}

// ClearRef clears the optional grid reference under the pointer.
func (e *SelectionGestureEvent) ClearRef() error {
	return e.ClearOption(SelectionGestureEventOptionRef)
}

// SetPosition sets the surface-space pointer position for a press, drag, or
// autoscroll tick event.
func (e *SelectionGestureEvent) SetPosition(position SurfacePosition) error {
	return setEventValue(e, vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_POSITION, position.toC())
}

// ClearPosition clears the surface-space pointer position.
func (e *SelectionGestureEvent) ClearPosition() error {
	return e.ClearOption(SelectionGestureEventOptionPosition)
}

// SetRepeatDistance sets the maximum repeat-click distance in pixels.
func (e *SelectionGestureEvent) SetRepeatDistance(distance float64) error {
	return setEventValue(e, vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_REPEAT_DISTANCE, distance)
}

// ClearRepeatDistance clears the maximum repeat-click distance.
func (e *SelectionGestureEvent) ClearRepeatDistance() error {
	return e.ClearOption(SelectionGestureEventOptionRepeatDistance)
}

// SetTimeNS sets the optional monotonic event time in nanoseconds.
func (e *SelectionGestureEvent) SetTimeNS(ns uint64) error {
	return setEventValue(e, vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_TIME_NS, ns)
}

// ClearTimeNS clears the optional monotonic event time.
func (e *SelectionGestureEvent) ClearTimeNS() error {
	return e.ClearOption(SelectionGestureEventOptionTimeNS)
}

// SetRepeatIntervalNS sets the maximum interval between repeat clicks in
// nanoseconds.
func (e *SelectionGestureEvent) SetRepeatIntervalNS(ns uint64) error {
	return setEventValue(e, vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_REPEAT_INTERVAL_NS, ns)
}

// ClearRepeatIntervalNS clears the maximum interval between repeat clicks.
func (e *SelectionGestureEvent) ClearRepeatIntervalNS() error {
	return e.ClearOption(SelectionGestureEventOptionRepeatIntervalNS)
}

// SetWordBoundaryCodepoints sets custom word-boundary codepoints for events
// that derive word selections. Libghostty copies the values into the event.
func (e *SelectionGestureEvent) SetWordBoundaryCodepoints(codepoints []uint32) error {
	ptr, size := cUint32Slice(codepoints)
	if len(codepoints) > 0 && ptr == nil {
		return &Error{Result: ResultOutOfMemory}
	}
	if size > 0 {
		defer Free(ptr, size)
	}

	return setEventValue(e, vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_WORD_BOUNDARY_CODEPOINTS, vt.GhosttyCodepoints{
		Fptr:  uintptr(ptr),
		Flen1: uint64(len(codepoints)),
	})
}

// ClearWordBoundaryCodepoints clears custom word-boundary codepoints so the
// event uses Ghostty's defaults again.
func (e *SelectionGestureEvent) ClearWordBoundaryCodepoints() error {
	return e.ClearOption(SelectionGestureEventOptionWordBoundaryCodepoints)
}

// SetBehaviors sets the behavior table for a press event.
func (e *SelectionGestureEvent) SetBehaviors(behaviors SelectionGestureBehaviors) error {
	return setEventValue(e, vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_BEHAVIORS, behaviors.toC())
}

// ClearBehaviors clears custom press behaviors so Ghostty's default behavior
// table is used again.
func (e *SelectionGestureEvent) ClearBehaviors() error {
	return e.ClearOption(SelectionGestureEventOptionBehaviors)
}

// SetRectangle sets whether a drag or autoscroll tick should produce a
// rectangular selection.
func (e *SelectionGestureEvent) SetRectangle(rectangle bool) error {
	return setEventValue(e, vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_RECTANGLE, cBool(rectangle))
}

// ClearRectangle clears the rectangular-selection setting.
func (e *SelectionGestureEvent) ClearRectangle() error {
	return e.ClearOption(SelectionGestureEventOptionRectangle)
}

// SetGeometry sets the display geometry required by drag and autoscroll tick
// events.
func (e *SelectionGestureEvent) SetGeometry(geometry SelectionGestureGeometry) error {
	return setEventValue(e, vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_GEOMETRY, geometry.toC())
}

// ClearGeometry clears the display geometry.
func (e *SelectionGestureEvent) ClearGeometry() error {
	return e.ClearOption(SelectionGestureEventOptionGeometry)
}

// SetViewport sets the viewport coordinate required by autoscroll tick events.
func (e *SelectionGestureEvent) SetViewport(viewport PointCoordinate) error {
	return setEventValue(e, vt.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_VIEWPORT, viewport.toC())
}

// ClearViewport clears the autoscroll tick viewport coordinate.
func (e *SelectionGestureEvent) ClearViewport() error {
	return e.ClearOption(SelectionGestureEventOptionViewport)
}

// cUint32Slice copies codepoints into C-owned memory for the duration of a
// call, so that no C option struct points at Go memory. It returns nil and
// the size needed when the allocation fails, and nil and 0 for no
// codepoints.
func cUint32Slice(codepoints []uint32) (unsafe.Pointer, uintptr) {
	if len(codepoints) == 0 {
		return nil, 0
	}
	size := uintptr(len(codepoints)) * unsafe.Sizeof(uint32(0))
	ptr := Alloc(size)
	if ptr == nil {
		return nil, size
	}
	copy(unsafe.Slice((*uint32)(ptr), len(codepoints)), codepoints)
	return ptr, size
}

// selectResult reads the selection written at out, turning NO_VALUE into a
// nil selection.
func selectResult(err error, out uintptr) (*Selection, error) {
	if err != nil {
		return noValueNilSelection(err)
	}
	sel := selectionFromC(*(*vt.GhosttySelection)(unsafe.Pointer(out)))
	return &sel, nil
}

// SelectWord derives a word selection snapshot from a terminal grid reference.
// The returned selection is not installed as the terminal's active selection.
// It returns nil when the reference has no selectable word content.
//
// Options must contain a Ref. BoundaryCodepoints optionally replaces
// Ghostty's default word-boundary codepoints.
func (t *Terminal) SelectWord(options SelectWordOptions) (*Selection, error) {
	// A sized struct (GHOSTTY_INIT_SIZED): the size field tells the library
	// which fields exist.
	opts := vt.GhosttyTerminalSelectWordOptions{Fsize: uint64(unsafe.Sizeof(vt.GhosttyTerminalSelectWordOptions{}))}
	opts.Fref = options.Ref.ref
	if len(options.BoundaryCodepoints) > 0 {
		ptr, size := cUint32Slice(options.BoundaryCodepoints)
		if ptr == nil {
			return nil, &Error{Result: ResultOutOfMemory}
		}
		defer Free(ptr, size)
		opts.Fboundary_codepoints = uintptr(ptr)
		opts.Fboundary_codepoints_len = uint64(len(options.BoundaryCodepoints))
	}

	tls := getTLS()
	defer putTLS(tls)
	copts, n := cValue(tls, opts)
	defer tls.Free(n)
	out, m := cOut[vt.GhosttySelection](tls)
	defer tls.Free(m)
	return selectResult(resultError(vt.Xghostty_terminal_select_word(tls, t.ptr, copts, out)), out)
}

// SelectWordBetween derives the nearest word selection snapshot between two
// terminal grid references. The search starts at start and moves toward end,
// inclusive. It returns nil when no selectable word content exists between
// the references.
//
// Options must contain Start and End references. BoundaryCodepoints
// optionally replaces Ghostty's default word-boundary codepoints.
func (t *Terminal) SelectWordBetween(options SelectWordBetweenOptions) (*Selection, error) {
	opts := vt.GhosttyTerminalSelectWordBetweenOptions{Fsize: uint64(unsafe.Sizeof(vt.GhosttyTerminalSelectWordBetweenOptions{}))}
	opts.Fstart = options.Start.ref
	opts.Fend = options.End.ref
	if len(options.BoundaryCodepoints) > 0 {
		ptr, size := cUint32Slice(options.BoundaryCodepoints)
		if ptr == nil {
			return nil, &Error{Result: ResultOutOfMemory}
		}
		defer Free(ptr, size)
		opts.Fboundary_codepoints = uintptr(ptr)
		opts.Fboundary_codepoints_len = uint64(len(options.BoundaryCodepoints))
	}

	tls := getTLS()
	defer putTLS(tls)
	copts, n := cValue(tls, opts)
	defer tls.Free(n)
	out, m := cOut[vt.GhosttySelection](tls)
	defer tls.Free(m)
	return selectResult(resultError(vt.Xghostty_terminal_select_word_between(tls, t.ptr, copts, out)), out)
}

// SelectLine derives a line selection snapshot from a terminal grid reference.
// The returned selection is not installed as the terminal's active selection.
// It returns nil when the reference has no selectable line content.
//
// Options must contain a Ref. Whitespace optionally replaces Ghostty's default
// line-trim whitespace codepoints. SemanticPromptBoundary controls whether
// semantic prompt state changes bound the line selection.
func (t *Terminal) SelectLine(options SelectLineOptions) (*Selection, error) {
	opts := vt.GhosttyTerminalSelectLineOptions{Fsize: uint64(unsafe.Sizeof(vt.GhosttyTerminalSelectLineOptions{}))}
	opts.Fref = options.Ref.ref
	opts.Fsemantic_prompt_boundary = cBool(options.SemanticPromptBoundary)
	if len(options.Whitespace) > 0 {
		ptr, size := cUint32Slice(options.Whitespace)
		if ptr == nil {
			return nil, &Error{Result: ResultOutOfMemory}
		}
		defer Free(ptr, size)
		opts.Fwhitespace = uintptr(ptr)
		opts.Fwhitespace_len = uint64(len(options.Whitespace))
	}

	tls := getTLS()
	defer putTLS(tls)
	copts, n := cValue(tls, opts)
	defer tls.Free(n)
	out, m := cOut[vt.GhosttySelection](tls)
	defer tls.Free(m)
	return selectResult(resultError(vt.Xghostty_terminal_select_line(tls, t.ptr, copts, out)), out)
}

// SelectAll derives a selection snapshot covering all selectable terminal
// content. The returned selection is not installed as the terminal's active
// selection. It returns nil when no selectable content exists.
func (t *Terminal) SelectAll() (*Selection, error) {
	tls := getTLS()
	defer putTLS(tls)
	out, n := cOut[vt.GhosttySelection](tls)
	defer tls.Free(n)
	return selectResult(resultError(vt.Xghostty_terminal_select_all(tls, t.ptr, out)), out)
}

// SelectOutput derives a command-output selection snapshot from a terminal
// grid reference. The returned selection is not installed as the terminal's
// active selection. It returns nil when the reference is not selectable
// command output.
func (t *Terminal) SelectOutput(options SelectOutputOptions) (*Selection, error) {
	tls := getTLS()
	defer putTLS(tls)
	out, n := cOut[vt.GhosttySelection](tls)
	defer tls.Free(n)
	return selectResult(resultError(vt.Xghostty_terminal_select_output(tls, t.ptr, options.Ref.ref, out)), out)
}

// SelectionAdjust mutates a selection snapshot using terminal selection
// semantics. It adjusts the logical end endpoint regardless of whether the
// selection is visually forward or reversed.
func (t *Terminal) SelectionAdjust(sel *Selection, adjustment SelectionAdjust) error {
	tls := getTLS()
	defer putTLS(tls)
	cs, n := cValue(tls, sel.toC())
	defer tls.Free(n)
	if err := resultError(vt.Xghostty_terminal_selection_adjust(
		tls,
		t.ptr,
		cs,
		vt.GhosttySelectionAdjust(adjustment),
	)); err != nil {
		return err
	}

	*sel = selectionFromC(*(*vt.GhosttySelection)(unsafe.Pointer(cs)))
	return nil
}

// SelectionOrder returns the current endpoint ordering of a selection
// snapshot.
func (t *Terminal) SelectionOrder(sel *Selection) (SelectionOrder, error) {
	tls := getTLS()
	defer putTLS(tls)
	cs, n := cValue(tls, sel.toC())
	defer tls.Free(n)
	order, m := cOut[vt.GhosttySelectionOrder](tls)
	defer tls.Free(m)
	if err := resultError(vt.Xghostty_terminal_selection_order(tls, t.ptr, cs, order)); err != nil {
		return 0, err
	}
	return SelectionOrder(*(*vt.GhosttySelectionOrder)(unsafe.Pointer(order))), nil
}

// SelectionOrdered returns a fresh selection snapshot with endpoints ordered
// as requested. Mirrored desired orders are accepted by Ghostty but normalize
// the same as forward ordering.
func (t *Terminal) SelectionOrdered(sel *Selection, desired SelectionOrder) (*Selection, error) {
	tls := getTLS()
	defer putTLS(tls)
	cs, n := cValue(tls, sel.toC())
	defer tls.Free(n)
	out, m := cOut[vt.GhosttySelection](tls)
	defer tls.Free(m)
	if err := resultError(vt.Xghostty_terminal_selection_ordered(
		tls,
		t.ptr,
		cs,
		vt.GhosttySelectionOrder(desired),
		out,
	)); err != nil {
		return nil, err
	}

	ordered := selectionFromC(*(*vt.GhosttySelection)(unsafe.Pointer(out)))
	return &ordered, nil
}

// SelectionContains reports whether point is inside a selection snapshot,
// using the terminal's linear and rectangular selection semantics.
func (t *Terminal) SelectionContains(sel *Selection, point Point) (bool, error) {
	tls := getTLS()
	defer putTLS(tls)
	cs, n := cValue(tls, sel.toC())
	defer tls.Free(n)
	contains, m := cOut[uint8](tls)
	defer tls.Free(m)
	if err := resultError(vt.Xghostty_terminal_selection_contains(
		tls,
		t.ptr,
		cs,
		point.toC(),
		contains,
	)); err != nil {
		return false, err
	}
	return *(*uint8)(unsafe.Pointer(contains)) != 0, nil
}

// SelectionEqual reports whether two selection snapshots are equal according
// to the terminal's internal selection semantics.
func (t *Terminal) SelectionEqual(a *Selection, b *Selection) (bool, error) {
	tls := getTLS()
	defer putTLS(tls)
	ca, n := cValue(tls, a.toC())
	defer tls.Free(n)
	cb, m := cValue(tls, b.toC())
	defer tls.Free(m)
	equal, k := cOut[uint8](tls)
	defer tls.Free(k)
	if err := resultError(vt.Xghostty_terminal_selection_equal(tls, t.ptr, ca, cb, equal)); err != nil {
		return false, err
	}
	return *(*uint8)(unsafe.Pointer(equal)) != 0, nil
}
