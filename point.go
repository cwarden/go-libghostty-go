package libghostty

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// PointCoordinate is an untagged coordinate in the terminal grid.
//
// It is used by APIs that already define which coordinate system applies,
// such as selection gesture autoscroll viewport positions. Use [Point] when
// an API needs the coordinate system tag alongside X and Y.
// C: GhosttyPointCoordinate
type PointCoordinate struct {
	// X is the column (0-indexed).
	X uint16

	// Y is the row (0-indexed). May exceed page size for screen/history tags.
	Y uint32
}

// toC converts a Go PointCoordinate to a C GhosttyPointCoordinate.
func (p PointCoordinate) toC() vt.GhosttyPointCoordinate {
	return vt.GhosttyPointCoordinate{
		Fx: p.X,
		Fy: p.Y,
	}
}

// PointTag determines which coordinate system a point uses.
// C: GhosttyPointTag
type PointTag int

const (
	// PointTagActive references the active area where the cursor can move.
	PointTagActive PointTag = vt.GHOSTTY_POINT_TAG_ACTIVE

	// PointTagViewport references the visible viewport (changes when scrolled).
	PointTagViewport PointTag = vt.GHOSTTY_POINT_TAG_VIEWPORT

	// PointTagScreen references the full screen including scrollback.
	PointTagScreen PointTag = vt.GHOSTTY_POINT_TAG_SCREEN

	// PointTagHistory references scrollback history only (before active area).
	PointTagHistory PointTag = vt.GHOSTTY_POINT_TAG_HISTORY
)

// Point is a tagged position in the terminal grid. The Tag determines
// which coordinate system X and Y refer to.
// C: GhosttyPoint
type Point struct {
	// Tag determines the coordinate system.
	Tag PointTag

	// X is the column (0-indexed).
	X uint16

	// Y is the row (0-indexed). May exceed page size for screen/history tags.
	Y uint32
}

// toC converts a Go Point to a C GhosttyPoint.
func (p Point) toC() vt.GhosttyPoint {
	var cp vt.GhosttyPoint
	cp.Ftag = vt.GhosttyPointTag(p.Tag)
	// Set the coordinate in the value union. Every member of the union is a
	// coordinate at offset zero.
	coord := (*vt.GhosttyPointCoordinate)(unsafe.Pointer(&cp.Fvalue))
	coord.Fx = p.X
	coord.Fy = p.Y
	return cp
}

// pointFromC converts a C coordinate plus its requested tag into a Go
// Point. The C APIs that resolve refs back to points return only the
// coordinate because the caller already selected the coordinate system.
func pointFromC(tag PointTag, coord vt.GhosttyPointCoordinate) Point {
	return Point{
		Tag: tag,
		X:   coord.Fx,
		Y:   coord.Fy,
	}
}
