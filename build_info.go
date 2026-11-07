package libghostty

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// OptimizeMode identifies the optimization mode the library was built with.
// C: GhosttyOptimizeMode
type OptimizeMode int

const (
	// OptimizeDebug is the debug optimization mode.
	OptimizeDebug OptimizeMode = vt.GHOSTTY_OPTIMIZE_DEBUG

	// OptimizeReleaseSafe is the release-safe optimization mode.
	OptimizeReleaseSafe OptimizeMode = vt.GHOSTTY_OPTIMIZE_RELEASE_SAFE

	// OptimizeReleaseSmall is the release-small optimization mode.
	OptimizeReleaseSmall OptimizeMode = vt.GHOSTTY_OPTIMIZE_RELEASE_SMALL

	// OptimizeReleaseFast is the release-fast optimization mode.
	OptimizeReleaseFast OptimizeMode = vt.GHOSTTY_OPTIMIZE_RELEASE_FAST
)

// BuildInfo holds compile-time build configuration of libghostty-vt.
// All values are constant for the lifetime of the process.
// C: GhosttyBuildInfo (query enum)
type BuildInfo struct {
	// SIMD reports whether SIMD-accelerated code paths are enabled.
	SIMD bool

	// KittyGraphics reports whether Kitty graphics protocol support
	// is available.
	KittyGraphics bool

	// TmuxControlMode reports whether tmux control mode support
	// is available.
	TmuxControlMode bool

	// Optimize is the optimization mode the library was built with.
	Optimize OptimizeMode

	// VersionString is the full version string
	// (e.g. "1.2.3" or "1.2.3-dev+abcdef").
	VersionString string

	// VersionMajor is the major version number.
	VersionMajor uint

	// VersionMinor is the minor version number.
	VersionMinor uint

	// VersionPatch is the patch version number.
	VersionPatch uint

	// VersionPre is the pre-release metadata string (e.g. "alpha",
	// "beta", or "dev"). Empty if no pre-release metadata is present.
	VersionPre string

	// VersionBuild is the build metadata string (e.g. commit hash).
	// Empty if no build metadata is present.
	VersionBuild string

	// FeatureSnapshot reports whether terminal snapshots are available.
	FeatureSnapshot bool

	// FeatureFormatter reports whether the formatter is available.
	FeatureFormatter bool

	// FeatureSelection reports whether selection APIs are available.
	FeatureSelection bool

	// FeatureSearch reports whether terminal search is available.
	FeatureSearch bool

	// FeatureRenderState reports whether the render state is available.
	FeatureRenderState bool

	// FeatureInputEncode reports whether key, mouse, focus, and paste
	// encoding are available.
	FeatureInputEncode bool

	// FeatureColor reports whether color utilities are available.
	FeatureColor bool

	// FeatureGridIntrospection reports whether grid references and the
	// cell and row accessors are available.
	FeatureGridIntrospection bool

	// FeatureGlyphProtocol reports whether applications can register
	// custom glyphs with the Glyph Protocol.
	FeatureGlyphProtocol bool
}

// buildInfoQuery runs one ghostty_build_info query. out is a C buffer large
// enough for the key's output type; it is zeroed before the call.
func buildInfoQuery(tls *libc.TLS, key vt.GhosttyBuildInfo, out uintptr, size int) vt.GhosttyResult {
	clear(unsafe.Slice((*byte)(unsafe.Pointer(out)), size))
	return vt.Xghostty_build_info(tls, key, out)
}

// GetBuildInfo queries all compile-time build configuration values
// and returns them in a single BuildInfo struct.
func GetBuildInfo() (BuildInfo, error) {
	var info BuildInfo
	tls := getTLS()
	defer putTLS(tls)

	// A GhosttyString is the largest output type; one block serves every
	// query.
	const n = int(unsafe.Sizeof(vt.GhosttyString{}))
	out := tls.Alloc(n)
	defer tls.Free(n)
	boolAt := func() bool { return *(*uint8)(unsafe.Pointer(out)) != 0 }
	sizeAt := func() uint { return uint(*(*uint64)(unsafe.Pointer(out))) }
	stringAt := func() string {
		s := *(*vt.GhosttyString)(unsafe.Pointer(out))
		return string(goBytes(s.Fptr, int(s.Flen1)))
	}

	if err := resultError(buildInfoQuery(tls, vt.GHOSTTY_BUILD_INFO_SIMD, out, n)); err != nil {
		return info, err
	}
	info.SIMD = boolAt()

	if err := resultError(buildInfoQuery(tls, vt.GHOSTTY_BUILD_INFO_KITTY_GRAPHICS, out, n)); err != nil {
		return info, err
	}
	info.KittyGraphics = boolAt()

	if err := resultError(buildInfoQuery(tls, vt.GHOSTTY_BUILD_INFO_TMUX_CONTROL_MODE, out, n)); err != nil {
		return info, err
	}
	info.TmuxControlMode = boolAt()

	if err := resultError(buildInfoQuery(tls, vt.GHOSTTY_BUILD_INFO_OPTIMIZE, out, n)); err != nil {
		return info, err
	}
	info.Optimize = OptimizeMode(*(*vt.GhosttyOptimizeMode)(unsafe.Pointer(out)))

	if err := resultError(buildInfoQuery(tls, vt.GHOSTTY_BUILD_INFO_VERSION_STRING, out, n)); err != nil {
		return info, err
	}
	info.VersionString = stringAt()

	if err := resultError(buildInfoQuery(tls, vt.GHOSTTY_BUILD_INFO_VERSION_MAJOR, out, n)); err != nil {
		return info, err
	}
	info.VersionMajor = sizeAt()

	if err := resultError(buildInfoQuery(tls, vt.GHOSTTY_BUILD_INFO_VERSION_MINOR, out, n)); err != nil {
		return info, err
	}
	info.VersionMinor = sizeAt()

	if err := resultError(buildInfoQuery(tls, vt.GHOSTTY_BUILD_INFO_VERSION_PATCH, out, n)); err != nil {
		return info, err
	}
	info.VersionPatch = sizeAt()

	if err := resultError(buildInfoQuery(tls, vt.GHOSTTY_BUILD_INFO_VERSION_PRE, out, n)); err != nil {
		return info, err
	}
	info.VersionPre = stringAt()

	if err := resultError(buildInfoQuery(tls, vt.GHOSTTY_BUILD_INFO_VERSION_BUILD, out, n)); err != nil {
		return info, err
	}
	info.VersionBuild = stringAt()

	// Each optional feature has a key that reports whether the library was
	// built with it. A call into a missing feature returns [ErrUnsupported].
	features := [...]struct {
		key vt.GhosttyBuildInfo
		out *bool
	}{
		{vt.GHOSTTY_BUILD_INFO_FEATURE_SNAPSHOT, &info.FeatureSnapshot},
		{vt.GHOSTTY_BUILD_INFO_FEATURE_FORMATTER, &info.FeatureFormatter},
		{vt.GHOSTTY_BUILD_INFO_FEATURE_SELECTION, &info.FeatureSelection},
		{vt.GHOSTTY_BUILD_INFO_FEATURE_SEARCH, &info.FeatureSearch},
		{vt.GHOSTTY_BUILD_INFO_FEATURE_RENDER_STATE, &info.FeatureRenderState},
		{vt.GHOSTTY_BUILD_INFO_FEATURE_INPUT_ENCODE, &info.FeatureInputEncode},
		{vt.GHOSTTY_BUILD_INFO_FEATURE_COLOR, &info.FeatureColor},
		{vt.GHOSTTY_BUILD_INFO_FEATURE_GRID_INTROSPECTION, &info.FeatureGridIntrospection},
		{vt.GHOSTTY_BUILD_INFO_FEATURE_GLYPH_PROTOCOL, &info.FeatureGlyphProtocol},
	}
	for _, feature := range features {
		// A library older than the headers doesn't know a newer key,
		// which means it doesn't have that feature, so that isn't a failure.
		result := buildInfoQuery(tls, feature.key, out, n)
		if result == vt.GHOSTTY_UNSUPPORTED {
			continue
		}
		if err := resultError(result); err != nil {
			return info, err
		}
		*feature.out = boolAt()
	}

	return info, nil
}

// Version returns the full version of the library, such as
// "0.1.0-dev+9a018f5cc". It is the same text as [BuildInfo.VersionString], but
// needs no query and cannot fail.
//
// Compare it with [HeaderVersion] to see what this package was generated
// from. The library is built into this package, so the two agree.
// C: ghostty_version
func Version() string {
	tls := getTLS()
	defer putTLS(tls)
	return libc.GoString(vt.Xghostty_version(tls))
}
