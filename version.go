package libghostty

import "go.mitchellh.com/libghostty/internal/vt"

// The version of the libghostty-vt headers this package was generated from.
// These are fixed at compile time. [Version] and [GetBuildInfo] report the
// library that is built in, which has the same version.
//
// Until 1.0, any 0.N release may change the ABI, so a program compiled against
// the 0.1 headers needs a 0.1 library: the minor versions must match. From 1.0,
// only a new major version can break programs.
//
// These always hold the number of the release, even in a development build.
// For example, a development build before 0.2.0 reports "0.2.0" here while
// [Version] returns the full "0.2.0-dev+abc1234".
const (
	// HeaderVersionMajor is the major version of the headers.
	// C: GHOSTTY_VT_VERSION_MAJOR
	HeaderVersionMajor = vt.GHOSTTY_VT_VERSION_MAJOR

	// HeaderVersionMinor is the minor version of the headers.
	// C: GHOSTTY_VT_VERSION_MINOR
	HeaderVersionMinor = vt.GHOSTTY_VT_VERSION_MINOR

	// HeaderVersionPatch is the patch version of the headers.
	// C: GHOSTTY_VT_VERSION_PATCH
	HeaderVersionPatch = vt.GHOSTTY_VT_VERSION_PATCH

	// HeaderVersion is the version of the headers as a string, such as "0.1.0".
	// It never has a pre-release or build suffix such as "-dev".
	// C: GHOSTTY_VT_VERSION
	HeaderVersion = vt.GHOSTTY_VT_VERSION
)
