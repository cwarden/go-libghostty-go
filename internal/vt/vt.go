// Package vt is the Go translation of libghostty-vt, produced by ccgo/v4
// from the C that Zig's C backend emits. See generator.go in the module
// root. The translated files, vt_<goos>_<goarch>_<n>.go, and the
// constants in consts.go are generated; this file is written by hand.
//
// The translated functions take a *libc.TLS as their first argument, C
// pointers are uintptr values into memory from libc.Xmalloc, and the C
// API types keep their header names (GhosttyString, GhosttyResult, ...)
// with struct fields prefixed by F.
package vt

import "unsafe"

// Fp returns the C function pointer for f, which must be a Go function
// whose signature is the translation of the C function pointer type: a
// *libc.TLS first, then the C parameters. The generated code calls through
// such pointers directly, so f must be a top-level function: the pointer
// is stored in C memory, which the garbage collector does not scan, and a
// closure could be collected while C still holds its pointer.
func Fp(f any) uintptr {
	type iface [2]uintptr
	return (*iface)(unsafe.Pointer(&f))[1]
}
