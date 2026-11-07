package libshim

import "modernc.org/libc"

// Xccgo_unreachable takes the place of __builtin_unreachable in the
// translated code. Zig's ReleaseFast builds compile `unreachable` to it,
// so reaching it means a path the Zig code excludes was taken. libc's
// implementation exits the process; a panic gives the Go traceback,
// whose function names carry the Zig source names.
func Xccgo_unreachable(tls *libc.TLS) {
	panic("libghostty: reached unreachable code")
}
