package libghostty

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
)

// TrackedGridRef is an owned grid reference that follows its cell as the
// terminal changes. Obtain one from [Terminal.TrackGridRef] and release it
// with [TrackedGridRef.Close].
//
// A tracked reference may lose its value if the referenced grid contents are
// discarded, such as after terminal reset or scrollback pruning. In that state
// [TrackedGridRef.HasValue] returns false and Snapshot or Point return an error
// with ResultNoValue. The same handle can be moved to a new terminal point with
// Set.
// C: GhosttyTrackedGridRef
type TrackedGridRef struct {
	ptr vt.GhosttyTrackedGridRef
}

// Close frees the tracked grid reference. Passing an already-closed tracked
// reference is safe; after Close, the reference must not be used again.
func (g *TrackedGridRef) Close() {
	if g == nil || g.ptr == 0 {
		return
	}
	tls := getTLS()
	defer putTLS(tls)
	vt.Xghostty_tracked_grid_ref_free(tls, g.ptr)
	g.ptr = 0
}

// HasValue reports whether the tracked grid reference currently has a
// meaningful location. It returns false after the owning terminal is closed or
// after the tracked cell is discarded.
func (g *TrackedGridRef) HasValue() bool {
	tls := getTLS()
	defer putTLS(tls)
	return vt.Xghostty_tracked_grid_ref_has_value(tls, g.ptr) != 0
}

// Point converts the tracked grid reference to a point in the requested
// coordinate system. If the reference no longer has a value, or cannot be
// represented in that coordinate system, this returns an error with
// ResultNoValue.
func (g *TrackedGridRef) Point(tag PointTag) (Point, error) {
	tls := getTLS()
	defer putTLS(tls)
	n := int(unsafe.Sizeof(vt.GhosttyPointCoordinate{}))
	out := tls.Alloc(n)
	defer tls.Free(n)
	if err := resultError(vt.Xghostty_tracked_grid_ref_point(
		tls,
		g.ptr,
		vt.GhosttyPointTag(tag),
		out,
	)); err != nil {
		return Point{}, err
	}
	return pointFromC(tag, *(*vt.GhosttyPointCoordinate)(unsafe.Pointer(out))), nil
}

// Set moves the tracked grid reference to a new point in its owning terminal.
// The terminal must be the same terminal that originally created the tracked
// reference. On success, any prior no-value state is cleared.
func (g *TrackedGridRef) Set(t *Terminal, point Point) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_tracked_grid_ref_set(
		tls,
		g.ptr,
		t.ptr,
		point.toC(),
	))
}

// Snapshot returns an untracked snapshot of the tracked grid reference's
// current location. The returned GridRef has the same borrowed lifetime rules
// as [Terminal.GridRef]: use it immediately and do not retain it across later
// terminal mutations.
func (g *TrackedGridRef) Snapshot() (*GridRef, error) {
	tls := getTLS()
	defer putTLS(tls)
	n := int(unsafe.Sizeof(vt.GhosttyGridRef{}))
	out := tls.Alloc(n)
	defer tls.Free(n)
	if err := resultError(vt.Xghostty_tracked_grid_ref_snapshot(tls, g.ptr, out)); err != nil {
		return nil, err
	}
	return &GridRef{ref: *(*vt.GhosttyGridRef)(unsafe.Pointer(out))}, nil
}
