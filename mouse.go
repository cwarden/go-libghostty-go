package libghostty

import "go.mitchellh.com/libghostty/internal/vt"

// MouseShape identifies a mouse pointer shape requested by the program running
// in the terminal. [Terminal.MouseShape] returns the current request.
//
// The shapes correspond to W3C cursor names. The host application chooses the
// native pointer to display, since not every platform supports every shape.
// These values describe the mouse pointer rather than the terminal text cursor.
//
// C: GhosttyMouseShape
type MouseShape int

const (
	// MouseShapeDefault is the default pointer, usually an arrow.
	MouseShapeDefault MouseShape = vt.GHOSTTY_MOUSE_SHAPE_DEFAULT

	// MouseShapeContextMenu indicates that a context menu is available.
	MouseShapeContextMenu MouseShape = vt.GHOSTTY_MOUSE_SHAPE_CONTEXT_MENU

	// MouseShapeHelp indicates that help is available.
	MouseShapeHelp MouseShape = vt.GHOSTTY_MOUSE_SHAPE_HELP

	// MouseShapePointer indicates a link, usually with a pointing hand.
	MouseShapePointer MouseShape = vt.GHOSTTY_MOUSE_SHAPE_POINTER

	// MouseShapeProgress indicates work in progress while interaction remains
	// available.
	MouseShapeProgress MouseShape = vt.GHOSTTY_MOUSE_SHAPE_PROGRESS

	// MouseShapeWait indicates that the program is busy and interaction must wait.
	MouseShapeWait MouseShape = vt.GHOSTTY_MOUSE_SHAPE_WAIT

	// MouseShapeCell indicates that cells can be selected.
	MouseShapeCell MouseShape = vt.GHOSTTY_MOUSE_SHAPE_CELL

	// MouseShapeCrosshair is a crosshair for precise selection.
	MouseShapeCrosshair MouseShape = vt.GHOSTTY_MOUSE_SHAPE_CROSSHAIR

	// MouseShapeText indicates selectable horizontal text.
	MouseShapeText MouseShape = vt.GHOSTTY_MOUSE_SHAPE_TEXT

	// MouseShapeVerticalText indicates selectable vertical text.
	MouseShapeVerticalText MouseShape = vt.GHOSTTY_MOUSE_SHAPE_VERTICAL_TEXT

	// MouseShapeAlias indicates that a shortcut or alias will be created.
	MouseShapeAlias MouseShape = vt.GHOSTTY_MOUSE_SHAPE_ALIAS

	// MouseShapeCopy indicates that an item will be copied.
	MouseShapeCopy MouseShape = vt.GHOSTTY_MOUSE_SHAPE_COPY

	// MouseShapeMove indicates that an item will be moved.
	MouseShapeMove MouseShape = vt.GHOSTTY_MOUSE_SHAPE_MOVE

	// MouseShapeNoDrop indicates that an item cannot be dropped here.
	MouseShapeNoDrop MouseShape = vt.GHOSTTY_MOUSE_SHAPE_NO_DROP

	// MouseShapeNotAllowed indicates that the requested action is not allowed.
	MouseShapeNotAllowed MouseShape = vt.GHOSTTY_MOUSE_SHAPE_NOT_ALLOWED

	// MouseShapeGrab indicates that an item can be grabbed.
	MouseShapeGrab MouseShape = vt.GHOSTTY_MOUSE_SHAPE_GRAB

	// MouseShapeGrabbing indicates that an item is being dragged.
	MouseShapeGrabbing MouseShape = vt.GHOSTTY_MOUSE_SHAPE_GRABBING

	// MouseShapeAllScroll indicates that scrolling is available in any direction.
	MouseShapeAllScroll MouseShape = vt.GHOSTTY_MOUSE_SHAPE_ALL_SCROLL

	// MouseShapeColumnResize indicates that a column can be resized horizontally.
	MouseShapeColumnResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_COL_RESIZE

	// MouseShapeRowResize indicates that a row can be resized vertically.
	MouseShapeRowResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_ROW_RESIZE

	// MouseShapeNorthResize indicates resizing toward the north.
	MouseShapeNorthResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_N_RESIZE

	// MouseShapeEastResize indicates resizing toward the east.
	MouseShapeEastResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_E_RESIZE

	// MouseShapeSouthResize indicates resizing toward the south.
	MouseShapeSouthResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_S_RESIZE

	// MouseShapeWestResize indicates resizing toward the west.
	MouseShapeWestResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_W_RESIZE

	// MouseShapeNorthEastResize indicates resizing toward the northeast.
	MouseShapeNorthEastResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_NE_RESIZE

	// MouseShapeNorthWestResize indicates resizing toward the northwest.
	MouseShapeNorthWestResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_NW_RESIZE

	// MouseShapeSouthEastResize indicates resizing toward the southeast.
	MouseShapeSouthEastResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_SE_RESIZE

	// MouseShapeSouthWestResize indicates resizing toward the southwest.
	MouseShapeSouthWestResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_SW_RESIZE

	// MouseShapeEastWestResize indicates resizing toward the east or west.
	MouseShapeEastWestResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_EW_RESIZE

	// MouseShapeNorthSouthResize indicates resizing toward the north or south.
	MouseShapeNorthSouthResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_NS_RESIZE

	// MouseShapeNorthEastSouthWestResize indicates resizing toward the
	// northeast or southwest.
	MouseShapeNorthEastSouthWestResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_NESW_RESIZE

	// MouseShapeNorthWestSouthEastResize indicates resizing toward the
	// northwest or southeast.
	MouseShapeNorthWestSouthEastResize MouseShape = vt.GHOSTTY_MOUSE_SHAPE_NWSE_RESIZE

	// MouseShapeZoomIn indicates that the view can be enlarged.
	MouseShapeZoomIn MouseShape = vt.GHOSTTY_MOUSE_SHAPE_ZOOM_IN

	// MouseShapeZoomOut indicates that the view can be reduced.
	MouseShapeZoomOut MouseShape = vt.GHOSTTY_MOUSE_SHAPE_ZOOM_OUT
)
